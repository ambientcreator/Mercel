package appcore_test

import (
	"strings"
	"testing"

	. "statistic/appcore"
)

// seedArchiveForCopy creates an employee, gives them one service and stores one
// archived calculation, so the copy flow has something to import.
func seedArchiveForCopy(t *testing.T, app *App, username string, password string) SavedCalculation {
	t.Helper()
	if _, err := app.CreateUser(UserWithPassword{Username: username, Password: password, Role: RoleTechnicalEmployee}); err != nil {
		t.Fatalf("CreateUser(%s) error = %v", username, err)
	}
	createUserService(t, app, username, password, "Услуга сотрудника", 700, CategoryPrimary)

	result, err := app.CalculateAmount(CalculationRequest{TargetAmount: 2800})
	if err != nil {
		t.Fatalf("CalculateAmount() error = %v", err)
	}
	saved, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 2800, Items: result.Items})
	if err != nil {
		t.Fatalf("SaveCalculation() error = %v", err)
	}
	return saved
}

func TestGlobalDirectorCopiesArchiveServicesOfAnyEmployee(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	if _, err := app.CreateUser(UserWithPassword{Username: "gendir", Password: "Str0ng!Passw0rd", Role: RoleGlobalDirector}); err != nil {
		t.Fatalf("CreateUser(gendir) error = %v", err)
	}
	saved := seedArchiveForCopy(t, app, "worker_tech", "Str0ng!Passw0rd")

	loginAsUser(t, app, "gendir", "Str0ng!Passw0rd")
	session := app.GetSession()
	if !session.CanCopyArchiveServices {
		t.Fatal("expected the global director to be allowed to copy archive services")
	}

	copied, err := app.CopyArchiveServicesToAdmin(saved.ID)
	if err != nil {
		t.Fatalf("CopyArchiveServicesToAdmin() error = %v", err)
	}
	if copied.Created == 0 {
		t.Fatalf("expected copied services, got %+v", copied)
	}

	services, err := app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}
	if len(services) != 1 || services[0].Name != "Услуга сотрудника" {
		t.Fatalf("expected the archived service in the director's own list, got %+v", services)
	}
}

func TestGlobalDirectorCreatesOwnServices(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	if _, err := app.CreateUser(UserWithPassword{Username: "gendir_services", Password: "Str0ng!Passw0rd", Role: RoleGlobalDirector}); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	loginAsUser(t, app, "gendir_services", "Str0ng!Passw0rd")
	created, err := app.UpsertService(UpsertServiceRequest{
		Name:     "Услуга директора",
		Unit:     "ч.",
		Rate:     900,
		Category: CategoryPrimary,
	})
	if err != nil {
		t.Fatalf("UpsertService() error = %v", err)
	}
	if created.CreatedBy != "gendir_services" {
		t.Fatalf("expected the service to belong to the director, got %+v", created)
	}
}

func TestOnlyAdminAndDirectorsCopyArchiveServices(t *testing.T) {
	allowed := []string{RoleAdmin, RoleGlobalDirector, RoleExecutiveDirector, RoleTechnicalDirector, RoleCommercialDirector}
	for _, role := range allowed {
		if !RoleCanCopyArchiveServicesForTest(role) {
			t.Fatalf("expected %s to be allowed to copy archive services", role)
		}
	}
	denied := []string{RoleSupportHead, RoleSupportSenior, RoleSupportEmployee, RoleTechnicalHead, RoleTechnicalEmployee, RoleFinanceEmployee}
	for _, role := range denied {
		if RoleCanCopyArchiveServicesForTest(role) {
			t.Fatalf("expected %s to be denied copying archive services", role)
		}
	}
}

func TestTechnicalDirectorCannotCopyOutsideOwnDepartments(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	if _, err := app.CreateUser(UserWithPassword{Username: "techdir", Password: "Str0ng!Passw0rd", Role: RoleTechnicalDirector}); err != nil {
		t.Fatalf("CreateUser(techdir) error = %v", err)
	}
	if _, err := app.CreateUser(UserWithPassword{Username: "finance_worker", Password: "Str0ng!Passw0rd", Role: RoleFinanceEmployee}); err != nil {
		t.Fatalf("CreateUser(finance_worker) error = %v", err)
	}
	createUserService(t, app, "finance_worker", "Str0ng!Passw0rd", "Услуга финансов", 500, CategoryPrimary)
	result, err := app.CalculateAmount(CalculationRequest{TargetAmount: 1500})
	if err != nil {
		t.Fatalf("CalculateAmount() error = %v", err)
	}
	saved, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 1500, Items: result.Items})
	if err != nil {
		t.Fatalf("SaveCalculation() error = %v", err)
	}

	loginAsUser(t, app, "techdir", "Str0ng!Passw0rd")
	if _, err := app.CopyArchiveServicesToAdmin(saved.ID); err == nil {
		t.Fatal("expected the technical director to be denied an archive outside their departments")
	} else if !strings.Contains(err.Error(), "прав") {
		t.Fatalf("expected a permission error, got %v", err)
	}
}

func TestFrontendRoleChoicesFollowViewDepartments(t *testing.T) {
	content := readFrontendScripts(t)
	// Directors sit in the "global" department, so filtering candidates by the actor's
	// own department left them without a single assignable role.
	if strings.Contains(content, "roleDepartment(candidate) === actorDept") {
		t.Fatal("role choices still filter by the actor's own department, directors get an empty list")
	}
	if !strings.Contains(content, "const actorDepartments = roleViewDepartments(role);") {
		t.Fatal("role choices must be built from roleViewDepartments so directors cover every department they oversee")
	}
	if !strings.Contains(content, "return roleCanManageUsers(role) || roleLevel(role) === 2 || roleLevel(role) === 4;") {
		t.Fatal("roleCanCreateUsers must mirror the backend rule that includes directors")
	}
}

func TestFrontendCopyArchiveButtonUsesCapabilityFlag(t *testing.T) {
	content := readFrontendScripts(t)
	if !strings.Contains(content, "const canCopyArchiveServices = Boolean((appState.session?.canAdmin || appState.session?.canCopyArchiveServices) && saved);") {
		t.Fatal("the copy button must follow the canCopyArchiveServices capability, not the admin flag alone")
	}
	if !strings.Contains(content, "if (!(appState.session?.canAdmin || appState.session?.canCopyArchiveServices) || !calculationId) {") {
		t.Fatal("the copy handler must follow the canCopyArchiveServices capability")
	}
}

func TestFrontendExposesActTemplateChoice(t *testing.T) {
	markup := readIndexHTML(t)
	for _, id := range []string{"act-template-select", "save-act-template-button", "act-template-message"} {
		if !strings.Contains(markup, id) {
			t.Fatalf("settings markup is missing #%s", id)
		}
	}
	if !strings.Contains(markup, "Бланк акта") {
		t.Fatal("settings markup is missing the act blank panel title")
	}

	content := readFrontendScripts(t)
	if !strings.Contains(content, "api.UpdateUserActTemplate(") {
		t.Fatal("the settings screen must call UpdateUserActTemplate")
	}
	if !strings.Contains(content, "renderActTemplateChoice();") {
		t.Fatal("the blank selector must be rendered on every bootstrap refresh")
	}
	if !strings.Contains(content, "actTemplatePanel?.classList.toggle(\"hidden\", !canChooseActTemplate);") {
		t.Fatal("the blank selector must stay hidden for everybody but the admin")
	}
}
