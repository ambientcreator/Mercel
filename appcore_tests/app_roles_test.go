package appcore_test

import (
	. "statistic/appcore"
	"strings"
	"testing"
)

func TestSeniorSpecialistCanDeleteEmployeeOnly(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	employee, err := app.CreateUser(UserWithPassword{Username: "employee1", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser employee error = %v", err)
	}
	senior, err := app.CreateUser(UserWithPassword{Username: "senior1", Password: "secret", Role: RoleSeniorSpecialist})
	if err != nil {
		t.Fatalf("CreateUser senior error = %v", err)
	}
	manager, err := app.CreateUser(UserWithPassword{Username: "manager1", Password: "secret", Role: RoleManager})
	if err != nil {
		t.Fatalf("CreateUser manager error = %v", err)
	}

	_, err = app.Login(LoginRequest{Username: senior.Username, Password: "secret"})
	if err != nil {
		t.Fatalf("Login senior error = %v", err)
	}
	if err := app.DeleteUser(employee.ID); err != nil {
		t.Fatalf("DeleteUser employee error = %v", err)
	}
	if err := app.DeleteUser(manager.ID); err == nil {
		t.Fatalf("expected senior specialist to be blocked from deleting manager")
	}
}

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestManagerCanDeleteSeniorSpecialist`.
// EN: Test `TestManagerCanDeleteSeniorSpecialist`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РІвЂћвЂ“ Р В РЎвЂ Р РЋРІР‚С›Р В РЎвЂР В РЎвЂќР РЋР С“Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В Р’В±Р В Р’ВµР В Р’В· Р РЋР вЂљР РЋРЎвЂњР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР В РЎвЂќР В РЎвЂ.
// EN: What it does: TestManagerCanDeleteSeniorSpecialist validates the broader deletion rights granted to managers.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В РЎвЂР В Р’В·Р В РЎвЂўР В Р’В»Р В РЎвЂР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В РЎВ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РЎвЂ; Р В Р вЂ¦Р РЋРЎвЂњР В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РўвЂР В Р’В»Р РЋР РЏ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚С™Р РЋРІР‚в„– Р В РЎвЂўР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РІвЂћвЂ“; Р В РўвЂР В РЎвЂўР В РЎвЂќР РЋРЎвЂњР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestSeniorRolesCanResetEmployeePasswordsWithinOwnDepartment(t *testing.T) {
	cases := []struct {
		seniorRole   string
		employeeRole string
		seniorName   string
		employeeName string
	}{
		{RoleSupportSeniorSpecialist, RoleSupportEmployee, "support_senior_reset", "support_employee_reset"},
		{RoleSeniorTech, RoleTechnician, "tech_senior_reset", "tech_employee_reset"},
		{RoleSeniorMRK, RoleMRKEmployee, "mrk_senior_reset", "mrk_employee_reset"},
	}

	for _, tc := range cases {
		app := withTempDB(t)
		loginAsAdmin(t, app)

		senior, err := app.CreateUser(UserWithPassword{Username: tc.seniorName, Password: "secret", Role: tc.seniorRole})
		if err != nil {
			t.Fatalf("CreateUser senior error = %v", err)
		}
		employee, err := app.CreateUser(UserWithPassword{Username: tc.employeeName, Password: "secret", Role: tc.employeeRole})
		if err != nil {
			t.Fatalf("CreateUser employee error = %v", err)
		}

		loginAsUser(t, app, senior.Username, "secret")
		if _, err := app.ResetUserPassword(ResetUserPasswordRequest{UserID: employee.ID, Password: "updated123"}); err != nil {
			t.Fatalf("ResetUserPassword() error = %v", err)
		}

		app.Logout()
		if _, err := app.Login(LoginRequest{Username: employee.Username, Password: "updated123"}); err != nil {
			t.Fatalf("Login() with updated password error = %v", err)
		}
	}
}

func TestResetUserPasswordRejectsEmptyPassword(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	employee, err := app.CreateUser(UserWithPassword{Username: "empty_pwd_employee", Password: "secret", Role: RoleSupportEmployee})
	if err != nil {
		t.Fatalf("CreateUser employee error = %v", err)
	}

	adminUser, err := app.RequireAuthForTest()
	if err != nil {
		t.Fatalf("requireAuth() error = %v", err)
	}

	if _, err := app.ResetUserPassword(ResetUserPasswordRequest{UserID: employee.ID, Password: ""}); err == nil {
		t.Fatalf("expected empty password to be rejected")
	} else {
		message := err.Error()
		if strings.Contains(message, "?") {
			t.Fatalf("error contains placeholder characters: %q", message)
		}
		if strings.TrimSpace(message) == "" {
			t.Fatalf("unexpected empty error message")
		}
	}

	if adminUser == nil {
		t.Fatalf("expected authenticated admin session")
	}
}

func TestUpdateOwnLastActNumberPersistsInSessionAndDatabase(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin1(t, app)

	current, err := app.RequireAuthForTest()
	if err != nil {
		t.Fatalf("requireAuth() error = %v", err)
	}

	updated, err := app.UpdateUserLastActNumber(UpdateUserLastActNumberRequest{
		UserID:        current.ID,
		LastActNumber: 12,
	})
	if err != nil {
		t.Fatalf("UpdateUserLastActNumber() error = %v", err)
	}
	if updated.LastActNumber != 12 {
		t.Fatalf("updated.LastActNumber = %d, want 12", updated.LastActNumber)
	}

	session := app.GetSession()
	if session.User == nil || session.User.LastActNumber != 12 {
		t.Fatalf("session user last act number = %+v, want 12", session.User)
	}

	stored, err := app.GetUserByIDForTest(current.ID)
	if err != nil {
		t.Fatalf("getUserByID() error = %v", err)
	}
	if stored.LastActNumber != 12 {
		t.Fatalf("stored.LastActNumber = %d, want 12", stored.LastActNumber)
	}
}

func TestCreateUserRejectsUnsafeUsername(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	if _, err := app.CreateUser(UserWithPassword{Username: "' OR 1=1 --", Password: "secret123", Role: RoleSupportEmployee}); err == nil {
		t.Fatalf("expected unsafe username to be rejected")
	}

	users, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	for _, user := range users {
		if user.Username == "' OR 1=1 --" {
			t.Fatalf("unsafe username must not be persisted")
		}
	}
}

func TestLegacyPasswordHashMigratesOnLogin(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	created, err := app.CreateUser(UserWithPassword{Username: "legacy_user", Password: "secret123", Role: RoleSupportEmployee})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if _, err := app.DBForTest().Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, LegacyHashPasswordForTest("secret123"), created.ID); err != nil {
		t.Fatalf("force legacy hash error = %v", err)
	}

	app.Logout()
	if _, err := app.Login(LoginRequest{Username: "legacy_user", Password: "secret123"}); err != nil {
		t.Fatalf("Login() legacy user error = %v", err)
	}

	var stored string
	if err := app.DBForTest().QueryRow(`SELECT password_hash FROM users WHERE id = ?`, created.ID).Scan(&stored); err != nil {
		t.Fatalf("QueryRow() password_hash error = %v", err)
	}
	if !strings.HasPrefix(stored, "$2") {
		t.Fatalf("expected bcrypt hash after login migration, got %q", stored)
	}
}

func TestSeniorCannotResetPasswordOutsideOwnDepartment(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	supportSenior, err := app.CreateUser(UserWithPassword{Username: "support_senior_pwd", Password: "secret", Role: RoleSupportSeniorSpecialist})
	if err != nil {
		t.Fatalf("CreateUser support senior error = %v", err)
	}
	techEmployee, err := app.CreateUser(UserWithPassword{Username: "tech_employee_pwd", Password: "secret", Role: RoleTechnician})
	if err != nil {
		t.Fatalf("CreateUser technician error = %v", err)
	}

	loginAsUser(t, app, supportSenior.Username, "secret")
	if _, err := app.ResetUserPassword(ResetUserPasswordRequest{UserID: techEmployee.ID, Password: "updated123"}); err == nil {
		t.Fatalf("expected cross-department password reset to be blocked")
	}
}

func TestManagerCanDeleteSeniorSpecialist(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	manager, err := app.CreateUser(UserWithPassword{Username: "manager2", Password: "secret", Role: RoleManager})
	if err != nil {
		t.Fatalf("CreateUser manager error = %v", err)
	}
	senior, err := app.CreateUser(UserWithPassword{Username: "senior2", Password: "secret", Role: RoleSeniorSpecialist})
	if err != nil {
		t.Fatalf("CreateUser senior error = %v", err)
	}

	loginAsUser(t, app, manager.Username, "secret")
	if err := app.DeleteUser(senior.ID); err != nil {
		t.Fatalf("DeleteUser senior by manager error = %v", err)
	}
}

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestArchiveVisibilityByRole`.
// EN: Test `TestArchiveVisibilityByRole`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РІвЂћвЂ“ Р В РЎвЂ Р РЋРІР‚С›Р В РЎвЂР В РЎвЂќР РЋР С“Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В Р’В±Р В Р’ВµР В Р’В· Р РЋР вЂљР РЋРЎвЂњР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР В РЎвЂќР В РЎвЂ.
// EN: What it does: TestArchiveVisibilityByRole checks who may see which archived calculations across the role hierarchy.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В РЎвЂР В Р’В·Р В РЎвЂўР В Р’В»Р В РЎвЂР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В РЎВ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РЎвЂ; Р В Р вЂ¦Р РЋРЎвЂњР В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РўвЂР В Р’В»Р РЋР РЏ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚С™Р РЋРІР‚в„– Р В РЎвЂўР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РІвЂћвЂ“; Р В РўвЂР В РЎвЂўР В РЎвЂќР РЋРЎвЂњР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestArchiveVisibilityByRole(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	employee, err := app.CreateUser(UserWithPassword{Username: "employee2", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser employee error = %v", err)
	}
	senior, err := app.CreateUser(UserWithPassword{Username: "senior3", Password: "secret", Role: RoleSeniorSpecialist})
	if err != nil {
		t.Fatalf("CreateUser senior error = %v", err)
	}
	manager, err := app.CreateUser(UserWithPassword{Username: "manager3", Password: "secret", Role: RoleManager})
	if err != nil {
		t.Fatalf("CreateUser manager error = %v", err)
	}

	adminResult, err := app.CalculateAmount(CalculationRequest{TargetAmount: 50000, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount admin error = %v", err)
	}
	if _, err := app.SaveCalculation(SaveCalculationRequest{Title: "admin calc", TargetAmount: adminResult.TargetAmount, Items: adminResult.Items}); err != nil {
		t.Fatalf("SaveCalculation admin error = %v", err)
	}

	loginAsUser(t, app, employee.Username, "secret")
	createUserService(t, app, employee.Username, "secret", "employee service", 100, CategoryPrimary)
	employeeResult, err := app.CalculateAmount(CalculationRequest{TargetAmount: 50100, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount employee error = %v", err)
	}
	if _, err := app.SaveCalculation(SaveCalculationRequest{Title: "employee calc", TargetAmount: employeeResult.TargetAmount, Items: employeeResult.Items}); err != nil {
		t.Fatalf("SaveCalculation employee error = %v", err)
	}

	loginAsUser(t, app, senior.Username, "secret")
	createUserService(t, app, senior.Username, "secret", "senior service", 100, CategoryPrimary)
	seniorResult, err := app.CalculateAmount(CalculationRequest{TargetAmount: 50200, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount senior error = %v", err)
	}
	if _, err := app.SaveCalculation(SaveCalculationRequest{Title: "senior calc", TargetAmount: seniorResult.TargetAmount, Items: seniorResult.Items}); err != nil {
		t.Fatalf("SaveCalculation senior error = %v", err)
	}
	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations senior error = %v", err)
	}
	if len(calculations) != 2 {
		t.Fatalf("expected senior specialist to see own and employee archives, got %d", len(calculations))
	}
	seen := map[string]bool{}
	for _, item := range calculations {
		seen[item.CreatedBy] = true
	}
	if !seen[employee.Username] || !seen[senior.Username] {
		t.Fatalf("expected senior specialist to see own and employee archive, got %+v", calculations)
	}

	loginAsUser(t, app, manager.Username, "secret")
	managerCalculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations manager error = %v", err)
	}
	if len(managerCalculations) != 2 {
		t.Fatalf("expected manager to see employee and senior archives from the same department, got %d", len(managerCalculations))
	}
	seenManager := map[string]bool{}
	for _, item := range managerCalculations {
		seenManager[item.CreatedBy] = true
		if item.CreatedBy == "admin" {
			t.Fatalf("manager must not see admin archive: %+v", managerCalculations)
		}
	}
	if !seenManager[employee.Username] || !seenManager[senior.Username] {
		t.Fatalf("manager should see employee and senior archives, got %+v", managerCalculations)
	}

	loginAsAdmin(t, app)
	adminCalculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations admin error = %v", err)
	}
	if len(adminCalculations) != 3 {
		t.Fatalf("expected admin to see all archives, got %d", len(adminCalculations))
	}
}

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestManagerCannotAssignManagerRole`.
// EN: Test `TestManagerCannotAssignManagerRole`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РІвЂћвЂ“ Р В РЎвЂ Р РЋРІР‚С›Р В РЎвЂР В РЎвЂќР РЋР С“Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В Р’В±Р В Р’ВµР В Р’В· Р РЋР вЂљР РЋРЎвЂњР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР В РЎвЂќР В РЎвЂ.
// EN: What it does: TestManagerCannotAssignManagerRole prevents managers from creating peers by role reassignment.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В РЎвЂР В Р’В·Р В РЎвЂўР В Р’В»Р В РЎвЂР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В РЎВ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РЎвЂ; Р В Р вЂ¦Р РЋРЎвЂњР В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РўвЂР В Р’В»Р РЋР РЏ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚С™Р РЋРІР‚в„– Р В РЎвЂўР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РІвЂћвЂ“; Р В РўвЂР В РЎвЂўР В РЎвЂќР РЋРЎвЂњР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestManagerCannotAssignManagerRole(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	manager, err := app.CreateUser(UserWithPassword{Username: "manager4", Password: "secret", Role: RoleManager})
	if err != nil {
		t.Fatalf("CreateUser manager error = %v", err)
	}
	employee, err := app.CreateUser(UserWithPassword{Username: "employee4", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser employee error = %v", err)
	}

	loginAsUser(t, app, manager.Username, "secret")
	if _, err := app.UpdateUserRole(employee.ID, RoleManager); err == nil {
		t.Fatalf("expected manager to be blocked from assigning manager role")
	}
}

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestListUsersRespectsRoleVisibility`.
// EN: Test `TestListUsersRespectsRoleVisibility`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РІвЂћвЂ“ Р В РЎвЂ Р РЋРІР‚С›Р В РЎвЂР В РЎвЂќР РЋР С“Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В Р’В±Р В Р’ВµР В Р’В· Р РЋР вЂљР РЋРЎвЂњР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР В РЎвЂќР В РЎвЂ.
// EN: What it does: TestListUsersRespectsRoleVisibility verifies that each role sees only the users they are allowed to manage.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В РЎвЂР В Р’В·Р В РЎвЂўР В Р’В»Р В РЎвЂР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В РЎВ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РЎвЂ; Р В Р вЂ¦Р РЋРЎвЂњР В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РўвЂР В Р’В»Р РЋР РЏ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚С™Р РЋРІР‚в„– Р В РЎвЂўР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РІвЂћвЂ“; Р В РўвЂР В РЎвЂўР В РЎвЂќР РЋРЎвЂњР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestListUsersRespectsRoleVisibility(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	_, err := app.CreateUser(UserWithPassword{Username: "employee5", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser employee error = %v", err)
	}
	senior, err := app.CreateUser(UserWithPassword{Username: "senior5", Password: "secret", Role: RoleSeniorSpecialist})
	if err != nil {
		t.Fatalf("CreateUser senior error = %v", err)
	}
	manager, err := app.CreateUser(UserWithPassword{Username: "pasha", Password: "secret", Role: RoleManager})
	if err != nil {
		t.Fatalf("CreateUser manager error = %v", err)
	}

	loginAsUser(t, app, senior.Username, "secret")
	seniorUsers, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers senior error = %v", err)
	}
	for _, user := range seniorUsers {
		if user.Role != RoleEmployee {
			t.Fatalf("senior specialist must only see employees, got %+v", seniorUsers)
		}
	}

	loginAsUser(t, app, manager.Username, "secret")
	managerUsers, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers manager error = %v", err)
	}
	for _, user := range managerUsers {
		if user.Username == "pasha" {
			t.Fatalf("manager must not see own account, got %+v", managerUsers)
		}
		if user.Role == RoleManager || user.Role == RoleAdmin {
			t.Fatalf("manager must not see manager/admin accounts, got %+v", managerUsers)
		}
	}
}

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestListUsersSortsByRolePriority`.
// EN: Test `TestListUsersSortsByRolePriority`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РІвЂћвЂ“ Р В РЎвЂ Р РЋРІР‚С›Р В РЎвЂР В РЎвЂќР РЋР С“Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В Р’В±Р В Р’ВµР В Р’В· Р РЋР вЂљР РЋРЎвЂњР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР В РЎвЂќР В РЎвЂ.
// EN: What it does: TestListUsersSortsByRolePriority locks down the UI ordering of managers, seniors and employees.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В РЎвЂР В Р’В·Р В РЎвЂўР В Р’В»Р В РЎвЂР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В РЎВ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РЎвЂ; Р В Р вЂ¦Р РЋРЎвЂњР В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РўвЂР В Р’В»Р РЋР РЏ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚С™Р РЋРІР‚в„– Р В РЎвЂўР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РІвЂћвЂ“; Р В РўвЂР В РЎвЂўР В РЎвЂќР РЋРЎвЂњР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestListUsersSortsByRolePriority(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	if _, err := app.CreateUser(UserWithPassword{Username: "employee6", Password: "secret", Role: RoleEmployee}); err != nil {
		t.Fatalf("CreateUser employee error = %v", err)
	}
	if _, err := app.CreateUser(UserWithPassword{Username: "manager6", Password: "secret", Role: RoleManager}); err != nil {
		t.Fatalf("CreateUser manager error = %v", err)
	}
	if _, err := app.CreateUser(UserWithPassword{Username: "senior6", Password: "secret", Role: RoleSeniorSpecialist}); err != nil {
		t.Fatalf("CreateUser senior error = %v", err)
	}

	adminUsers, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers admin error = %v", err)
	}
	if len(adminUsers) < 3 {
		t.Fatalf("expected at least 3 users, got %+v", adminUsers)
	}

	if adminUsers[0].Role != RoleAdmin || adminUsers[1].Role != RoleManager || adminUsers[2].Role != RoleSeniorSpecialist || adminUsers[3].Role != RoleEmployee {
		t.Fatalf("expected admin -> manager -> senior -> employee order, got %+v", adminUsers[:4])
	}

	loginAsUser(t, app, "manager6", "secret")
	managerUsers, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers manager error = %v", err)
	}
	if len(managerUsers) < 2 {
		t.Fatalf("expected manager to see at least senior and employee, got %+v", managerUsers)
	}
	if managerUsers[0].Role != RoleSeniorSpecialist {
		t.Fatalf("expected senior specialists first for manager, got %+v", managerUsers)
	}
	for i := 1; i < len(managerUsers); i++ {
		if managerUsers[i].Role != RoleEmployee {
			t.Fatalf("expected employees after seniors for manager, got %+v", managerUsers)
		}
	}
}

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestConfirmModalMarkupUsesReadableUTF8`.
// EN: Test `TestConfirmModalMarkupUsesReadableUTF8`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РІвЂћвЂ“ Р В РЎвЂ Р РЋРІР‚С›Р В РЎвЂР В РЎвЂќР РЋР С“Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ Р В Р’В±Р В Р’ВµР В Р’В· Р РЋР вЂљР РЋРЎвЂњР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР В РЎвЂќР В РЎвЂ.
// EN: What it does: TestConfirmModalMarkupUsesReadableUTF8 protects the modal template from accidental mojibake regressions.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ  Р В РЎвЂР В Р’В·Р В РЎвЂўР В Р’В»Р В РЎвЂР РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В РЎВ Р РЋР С“Р РЋРІР‚В Р В Р’ВµР В Р вЂ¦Р В Р’В°Р РЋР вЂљР В РЎвЂР В РЎвЂ; Р В Р вЂ¦Р РЋРЎвЂњР В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В РўвЂР В Р’В»Р РЋР РЏ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚С™Р РЋРІР‚в„– Р В РЎвЂўР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РІвЂћвЂ“; Р В РўвЂР В РЎвЂўР В РЎвЂќР РЋРЎвЂњР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р В РЎвЂР РЋР вЂљР РЋРЎвЂњР В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР В Р’В¶Р В РЎвЂР В РўвЂР В Р’В°Р В Р’ВµР В РЎВР В РЎвЂўР В Р’Вµ Р В РЎвЂ”Р В РЎвЂўР В Р вЂ Р В Р’ВµР В РўвЂР В Р’ВµР В Р вЂ¦Р В РЎвЂР В Р’Вµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
