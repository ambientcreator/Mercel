package appcore_test

import (
	"os"
	"path/filepath"
	"strings"

	. "statistic/appcore"

	"testing"
)

// readFrontendScripts concatenates the modular frontend scripts (and the legacy
// monolithic app.js when it still exists) so content tests keep working after the
// frontend was split into separate blocks.
func readFrontendScripts(t *testing.T) string {
	t.Helper()
	assetsDir := filepath.Join("..", "frontend", "dist", "assets")
	ordered := []string{"app_state.js", "app_utils.js", "app_render.js", "app_archive.js", "app_bootstrap.js", "app.js"}
	var builder strings.Builder
	for _, name := range ordered {
		content, err := os.ReadFile(filepath.Join(assetsDir, name))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("read frontend script %s: %v", name, err)
		}
		builder.Write(content)
		builder.WriteString("\n")
	}
	return builder.String()
}

func readIndexHTML(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "frontend", "dist", "index.html"))
	if err != nil {
		t.Fatalf("ReadFile index.html error = %v", err)
	}
	return string(content)
}

func TestConfirmModalMarkupUsesReadableUTF8(t *testing.T) {
	content := readIndexHTML(t)
	for _, word := range []string{"Подтвердите", "Отмена", "Удалить"} {
		if !strings.Contains(content, word) {
			t.Fatalf("confirm modal is missing expected word %q", word)
		}
	}
	// 0xC3 0x90 / 0xC3 0x91 are the tell-tale bytes of UTF-8 text re-encoded as
	// Windows-1251 mojibake; they must never appear in clean markup.
	for _, broken := range []string{"\xc3\x90", "\xc3\x91"} {
		if strings.Contains(content, broken) {
			t.Fatalf("confirm modal still contains mojibake marker bytes %q", broken)
		}
	}
}

func TestDepartmentIsolationForUsersAndArchives(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	supportEmployee, err := app.CreateUser(UserWithPassword{Username: "support_emp", Password: "secret", Role: RoleSupportEmployee})
	if err != nil {
		t.Fatalf("CreateUser support employee error = %v", err)
	}
	supportSenior, err := app.CreateUser(UserWithPassword{Username: "support_senior", Password: "secret", Role: RoleSupportSeniorSpecialist})
	if err != nil {
		t.Fatalf("CreateUser support senior error = %v", err)
	}
	supportManager, err := app.CreateUser(UserWithPassword{Username: "support_manager", Password: "secret", Role: RoleSupportManager})
	if err != nil {
		t.Fatalf("CreateUser support manager error = %v", err)
	}
	techEmployee, err := app.CreateUser(UserWithPassword{Username: "tech_emp", Password: "secret", Role: RoleTechnician})
	if err != nil {
		t.Fatalf("CreateUser technician error = %v", err)
	}
	techSenior, err := app.CreateUser(UserWithPassword{Username: "tech_senior", Password: "secret", Role: RoleSeniorTech})
	if err != nil {
		t.Fatalf("CreateUser senior technician error = %v", err)
	}
	mrkManager, err := app.CreateUser(UserWithPassword{Username: "mrk_manager", Password: "secret", Role: RoleMRKManager})
	if err != nil {
		t.Fatalf("CreateUser mrk manager error = %v", err)
	}

	for _, tc := range []struct {
		username string
		password string
		service  string
		target   int
	}{
		{supportEmployee.Username, "secret", "support service", 1000},
		{supportSenior.Username, "secret", "support senior service", 1100},
		{supportManager.Username, "secret", "support manager service", 1200},
		{techEmployee.Username, "secret", "tech service", 1300},
		{techSenior.Username, "secret", "tech senior service", 1400},
		{mrkManager.Username, "secret", "mrk service", 1500},
	} {
		loginAsUser(t, app, tc.username, tc.password)
		createUserService(t, app, tc.username, tc.password, tc.service, 100, CategoryPrimary)
		result, err := app.CalculateAmount(CalculationRequest{TargetAmount: tc.target, Weights: map[string]int{}})
		if err != nil {
			t.Fatalf("CalculateAmount %s error = %v", tc.username, err)
		}
		if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: result.TargetAmount, Items: result.Items}); err != nil {
			t.Fatalf("SaveCalculation %s error = %v", tc.username, err)
		}
	}

	loginAsUser(t, app, supportManager.Username, "secret")
	supportUsers, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers support manager error = %v", err)
	}
	for _, item := range supportUsers {
		if RoleDepartmentForTest(item.Role) != RoleDepartmentForTest(RoleSupportManager) {
			t.Fatalf("support manager must not see other departments: %+v", supportUsers)
		}
	}
	supportArchives, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations support manager error = %v", err)
	}
	for _, item := range supportArchives {
		if item.CreatedBy == techEmployee.Username || item.CreatedBy == techSenior.Username || item.CreatedBy == mrkManager.Username {
			t.Fatalf("support manager must not see foreign archives: %+v", supportArchives)
		}
	}

	loginAsUser(t, app, mrkManager.Username, "secret")
	mrkUsers, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers mrk manager error = %v", err)
	}
	if len(mrkUsers) != 0 {
		t.Fatalf("mrk manager should not see support or tech users, got %+v", mrkUsers)
	}
	mrkArchives, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations mrk manager error = %v", err)
	}
	if len(mrkArchives) != 1 || mrkArchives[0].CreatedBy != mrkManager.Username {
		t.Fatalf("mrk manager should see only own archive, got %+v", mrkArchives)
	}
}

func TestTechnicalDirectorSeesOnlyTechnicalAndTelecom(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	techDirector, err := app.CreateUser(UserWithPassword{Username: "tech_director_1", Password: "secret", Role: RoleTechnicalDirector})
	if err != nil {
		t.Fatalf("CreateUser technical director error = %v", err)
	}
	techUser, err := app.CreateUser(UserWithPassword{Username: "tech_visible_1", Password: "secret", Role: RoleTechnicalEmployee})
	if err != nil {
		t.Fatalf("CreateUser technical employee error = %v", err)
	}
	telecomUser, err := app.CreateUser(UserWithPassword{Username: "telecom_visible_1", Password: "secret", Role: RoleTelecomEmployeeVOLS})
	if err != nil {
		t.Fatalf("CreateUser telecom employee error = %v", err)
	}
	skudUser, err := app.CreateUser(UserWithPassword{Username: "skud_hidden_1", Password: "secret", Role: RoleSKUDInstaller})
	if err != nil {
		t.Fatalf("CreateUser skud employee error = %v", err)
	}

	createUserService(t, app, techUser.Username, "secret", "tech service", 100, CategoryPrimary)
	loginAsUser(t, app, techUser.Username, "secret")
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 100, Items: []CalculationItem{{Name: "tech service", Unit: "ч.", Rate: 100, Quantity: 1, LineTotal: 100, Category: CategoryPrimary, ServiceCode: "tech-service"}}}); err != nil {
		t.Fatalf("SaveCalculation technical employee error = %v", err)
	}

	createUserService(t, app, telecomUser.Username, "secret", "telecom service", 110, CategorySecondary)
	loginAsUser(t, app, telecomUser.Username, "secret")
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 110, Items: []CalculationItem{{Name: "telecom service", Unit: "ч.", Rate: 110, Quantity: 1, LineTotal: 110, Category: CategorySecondary, ServiceCode: "telecom-service"}}}); err != nil {
		t.Fatalf("SaveCalculation telecom employee error = %v", err)
	}

	createUserService(t, app, skudUser.Username, "secret", "skud service", 120, CategoryClosing)
	loginAsUser(t, app, skudUser.Username, "secret")
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 120, Items: []CalculationItem{{Name: "skud service", Unit: "ч.", Rate: 120, Quantity: 1, LineTotal: 120, Category: CategoryClosing, ServiceCode: "skud-service"}}}); err != nil {
		t.Fatalf("SaveCalculation skud employee error = %v", err)
	}

	loginAsUser(t, app, techDirector.Username, "secret")
	users, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers technical director error = %v", err)
	}
	seen := map[string]bool{}
	for _, user := range users {
		seen[user.Username] = true
		if dept := RoleDepartmentForTest(user.Role); dept != DepartmentTechnical && dept != DepartmentTelecom {
			t.Fatalf("technical director must not see foreign department user: %+v", users)
		}
	}
	if !seen[techUser.Username] || !seen[telecomUser.Username] {
		t.Fatalf("technical director must see technical and telecom users, got %+v", users)
	}
	if seen[skudUser.Username] {
		t.Fatalf("technical director must not see skud users, got %+v", users)
	}

	archives, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations technical director error = %v", err)
	}
	archiveSeen := map[string]bool{}
	for _, item := range archives {
		archiveSeen[item.CreatedBy] = true
		if dept := RoleDepartmentForTest(item.CreatedRole); dept != DepartmentTechnical && dept != DepartmentTelecom {
			t.Fatalf("technical director must not see foreign archive: %+v", archives)
		}
	}
	if !archiveSeen[techUser.Username] || !archiveSeen[telecomUser.Username] {
		t.Fatalf("technical director must see technical and telecom archives, got %+v", archives)
	}
	if archiveSeen[skudUser.Username] {
		t.Fatalf("technical director must not see skud archives, got %+v", archives)
	}
}

func TestDepartmentRoleChangesStayInsideDepartment(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	supportManager, err := app.CreateUser(UserWithPassword{Username: "support_manager2", Password: "secret", Role: RoleSupportManager})
	if err != nil {
		t.Fatalf("CreateUser support manager error = %v", err)
	}
	techEmployee, err := app.CreateUser(UserWithPassword{Username: "tech_employee2", Password: "secret", Role: RoleTechnician})
	if err != nil {
		t.Fatalf("CreateUser technician error = %v", err)
	}

	loginAsUser(t, app, supportManager.Username, "secret")
	if _, err := app.UpdateUserRole(techEmployee.ID, RoleSeniorTech); err == nil {
		t.Fatalf("expected support manager to be blocked from changing tech department role")
	}
	if err := app.DeleteUser(techEmployee.ID); err == nil {
		t.Fatalf("expected support manager to be blocked from deleting tech department user")
	}
}

func TestFrontendRoleLabelsStayReadable(t *testing.T) {
	content := readFrontendScripts(t)

	requiredSnippets := []string{
		"global_director",
		"executive_director",
		"technical_director",
		"support_head",
		"support_sysadmin",
		"technical_senior",
		"telecom_construction_director",
		"telecom_senior_vols",
		"skud_head",
		"approval_employee",
		"marketing_courier",
		"commercial_senior_mrk",
		"finance_head",
		"legal_employee",
		"development_head",
		// readable Russian labels are stored as \uXXXX escapes; verify one decodes cleanly
		"\\u0413\\u0435\\u043d\\u0435\\u0440\\u0430\\u043b\\u044c\\u043d\\u044b\\u0439", // Генеральный
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("frontend scripts are missing readable role label %q", snippet)
		}
	}
	if strings.Contains(content, "????") {
		t.Fatalf("frontend scripts still contain placeholder question marks in role labels")
	}
	for _, broken := range []string{"\xc3\x90", "\xc3\x91"} {
		if strings.Contains(content, broken) {
			t.Fatalf("frontend scripts still contain mojibake marker bytes %q", broken)
		}
	}
}

func TestFrontendCoreUiFunctionsExist(t *testing.T) {
	content := readFrontendScripts(t)

	requiredSnippets := []string{
		"function resetServiceForm()",
		"function renderResult(result)",
		"function renderArchiveDetails(saved)",
		"function randomizeCalculation()",
		"const randomizeButton = document.getElementById(\"randomize-button\")",
		"function renderRoleOptions(roleChoices, selectedValue = \"\")",
		"function departmentSortPriority(department)",
		"<optgroup label=\"${departmentLabel(group.department)}\">${options}</optgroup>",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("frontend scripts are missing required snippet %q", snippet)
		}
	}

	if strings.Contains(content, "ReferenceError") {
		t.Fatalf("frontend scripts should not contain runtime error text leftovers")
	}
}

func TestFrontendAdminUsersCanSeeRoleControls(t *testing.T) {
	content := readFrontendScripts(t)

	required := "const canChangeRole = Boolean(appState.session?.canAdmin)"
	if !strings.Contains(content, required) {
		t.Fatalf("frontend scripts should allow admin users to see role controls")
	}
}

func TestFrontendExportControlsExist(t *testing.T) {
	indexContent := readIndexHTML(t)
	for _, snippet := range []string{
		`id="archive-date-input"`,
		`id="act-number-input"`,
		`id="employee-full-name-input"`,
		`id="contract-details-button"`,
		`id="contract-details-summary"`,
		`id="contract-modal-spbks-number-input"`,
		`id="contract-modal-grizabl-number-input"`,
		`id="contract-modal-code-spbks"`,
		`id="contract-modal-code-grizabl"`,
		`id="contract-modal-date-input"`,
		`id="contract-modal-fullname-input"`,
		`id="act-number-visible-input"`,
		`id="previous-act-number"`,
		`id="download-pdf-button"`,
	} {
		if !strings.Contains(indexContent, snippet) {
			t.Fatalf("index.html is missing export control %q", snippet)
		}
	}

	jsContent := readFrontendScripts(t)
	for _, snippet := range []string{
		"UpdateUserFullName",
		"ExportCurrentCalculationPDF",
		"openContractModal",
		"saveContractDetails",
		"downloadCurrentCalculationPDF",
		"persistActNumber",
		"renderActNumberControls",
		"renderResult(result)",
		"\\u0414\\u043b\\u044f \\u0432\\u044b\\u0433\\u0440\\u0443\\u0437\\u043a\\u0438 \\u0430\\u043a\\u0442\\u0430", // Для выгрузки акта
		"\\u0412\\u044b\\u0431\\u0440\\u0430\\u043d \\u0434\\u043e\\u0433\\u043e\\u0432\\u043e\\u0440",                // Выбран договор
	} {
		if !strings.Contains(jsContent, snippet) {
			t.Fatalf("frontend scripts are missing export-related snippet %q", snippet)
		}
	}
}

func TestFrontendRandomizationSkipsZeroPercentServices(t *testing.T) {
	content := readFrontendScripts(t)

	requiredSnippets := []string{
		"function buildRandomWeightPayload()",
		"next[service.code] = 0;",
		"if (!serviceTakesPartInCalculation(service)) {",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("frontend scripts are missing zero-percent randomization guard %q", snippet)
		}
	}
}

func TestArchiveEditFrontendHelpersExist(t *testing.T) {
	text := readFrontendScripts(t)
	for _, expected := range []string{
		"function escapeAttribute(value)",
		"data-archive-name",
		"archive-edit-input",
		"function copyArchiveServicesToAdmin(calculationId)",
		"function canDeleteArchiveCalculation(saved)",
		"async function deleteCalculation(calculationId)",
		"api.DeleteCalculation(id)",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("frontend archive edit is missing %q", expected)
		}
	}
}

func TestContractTemplatePickerBackend(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	templates, err := app.ListContractTemplates()
	if err != nil {
		t.Fatalf("ListContractTemplates error = %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected both act blanks, got %+v", templates)
	}
	for _, item := range templates {
		if item.Code == "" || item.Title == "" || item.CustomerName == "" || item.DirectorShort == "" {
			t.Fatalf("blank %+v is missing display fields", item)
		}
	}

	// Switching the blank must not require the contract number, date and full name
	// that UpdateUserContractDetails validates — that is the whole point of the picker.
	updated, err := app.SetPreferredContractTemplate("2")
	if err != nil {
		t.Fatalf("SetPreferredContractTemplate error = %v", err)
	}
	if updated.PreferredContractCode != "2" {
		t.Fatalf("expected blank 2, got %q", updated.PreferredContractCode)
	}

	session := app.GetSession()
	if session.User == nil || session.User.PreferredContractCode != "2" {
		t.Fatalf("session should carry the new blank, got %+v", session.User)
	}

	if _, err := app.SetPreferredContractTemplate("99"); err == nil {
		t.Fatalf("expected an unknown blank to be rejected")
	}

	current := app.GetSession()
	if current.User.PreferredContractCode != "2" {
		t.Fatalf("a rejected blank must not change the stored one, got %q", current.User.PreferredContractCode)
	}
}

func TestContractTemplatePickerRequiresSession(t *testing.T) {
	app := withTempDB(t)

	if _, err := app.ListContractTemplates(); err == nil {
		t.Fatalf("expected ListContractTemplates to require a session")
	}
	if _, err := app.SetPreferredContractTemplate("1"); err == nil {
		t.Fatalf("expected SetPreferredContractTemplate to require a session")
	}
}
