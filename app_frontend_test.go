package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestConfirmModalMarkupUsesReadableUTF8(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("frontend", "dist", "index.html"))
	if err != nil {
		t.Fatalf("ReadFile index.html error = %v", err)
	}

	expectedWords := [][]byte{
		{0xD0, 0x9F, 0xD0, 0xBE, 0xD0, 0xB4, 0xD1, 0x82, 0xD0, 0xB2, 0xD0, 0xB5, 0xD1, 0x80, 0xD0, 0xB6, 0xD0, 0xB4, 0xD0, 0xB5, 0xD0, 0xBD, 0xD0, 0xB8, 0xD0, 0xB5},
		{0xD0, 0x9F, 0xD0, 0xBE, 0xD0, 0xB4, 0xD1, 0x82, 0xD0, 0xB2, 0xD0, 0xB5, 0xD1, 0x80, 0xD0, 0xB4, 0xD0, 0xB8, 0xD1, 0x82, 0xD0, 0xB5, 0x20, 0xD1, 0x83, 0xD0, 0xB4, 0xD0, 0xB0, 0xD0, 0xBB, 0xD0, 0xB5, 0xD0, 0xBD, 0xD0, 0xB8, 0xD0, 0xB5},
		{0xD0, 0x9E, 0xD1, 0x82, 0xD0, 0xBC, 0xD0, 0xB5, 0xD0, 0xBD, 0xD0, 0xB0},
		{0xD0, 0xA3, 0xD0, 0xB4, 0xD0, 0xB0, 0xD0, 0xBB, 0xD0, 0xB8, 0xD1, 0x82, 0xD1, 0x8C},
	}
	for _, expected := range expectedWords {
		if !bytes.Contains(content, expected) {
			t.Fatalf("confirm modal is missing expected utf-8 bytes %v", expected)
		}
	}
	for _, broken := range [][]byte{{0xC3, 0x90}, {0xC3, 0x91}} {
		if bytes.Contains(content, broken) {
			t.Fatalf("confirm modal still contains mojibake marker bytes %v", broken)
		}
	}
}

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestDepartmentIsolationForUsersAndArchives`.
// EN: Test `TestDepartmentIsolationForUsersAndArchives`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р В РЎвЂ Р В РЎвЂ Р РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В Р’Вµ Р РЋР С“Р В РЎвЂўР РЋРІР‚С™Р РЋР вЂљР РЋРЎвЂњР В РўвЂР В Р вЂ¦Р В РЎвЂР В РЎвЂќР В РЎвЂ Р В Р вЂ Р В РЎвЂР В РўвЂР РЋР РЏР РЋРІР‚С™ Р РЋРІР‚С™Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В РЎвЂќР В РЎвЂў Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р В РЎвЂўР В Р вЂ Р В Р’В°Р РЋРІР‚С™Р В Р’ВµР В Р’В»Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ Р В Р’В°Р РЋР вЂљР РЋРІР‚В¦Р В РЎвЂР В Р вЂ Р РЋРІР‚в„– Р РЋР С“Р В Р вЂ Р В РЎвЂўР В Р’ВµР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°.
// EN: What it does: verifies that managers and seniors only see users and archives from their own department.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В РЎвЂ”Р В РЎвЂўР В РЎвЂќР РЋР вЂљР РЋРІР‚в„–Р В Р вЂ Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р РЋРЎвЂњР РЋР вЂ№ Р В РЎВР В РЎвЂўР В РўвЂР В Р’ВµР В Р’В»Р РЋР Р‰ Р РЋР вЂљР В РЎвЂўР В Р’В»Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ”Р В РЎвЂў Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В РЎВ Р В РЎС›Р В РЎСџ, Р РЋРІР‚С™Р В Р’ВµР РЋРІР‚В¦Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В° Р В РЎвЂ Р В РЎС™Р В Р’В Р В РЎв„ў; Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚В°Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™ Р В РЎВР В Р’ВµР В Р’В¶Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РІвЂћвЂ“ Р В Р вЂ Р В РЎвЂР В РўвЂР В РЎвЂР В РЎВР В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В РЎвЂ.
// EN: Key points: covers the new department-based role model across support, tech and MRK; prevents cross-department visibility regressions.
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
		if roleDepartment(item.Role) != roleDepartment(RoleSupportManager) {
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

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestDepartmentRoleChangesStayInsideDepartment`.
// EN: Test `TestDepartmentRoleChangesStayInsideDepartment`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р РЋР вЂљР РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р В Р вЂ¦Р В Р’Вµ Р В РЎВР В РЎвЂўР В Р’В¶Р В Р’ВµР РЋРІР‚С™ Р В РЎВР В Р’ВµР В Р вЂ¦Р РЋР РЏР РЋРІР‚С™Р РЋР Р‰ Р РЋР вЂљР В РЎвЂўР В Р’В»Р В РЎвЂ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р В РЎвЂўР В Р вЂ Р В Р’В°Р РЋРІР‚С™Р В Р’ВµР В Р’В»Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂР В Р’В· Р В РўвЂР РЋР вЂљР РЋРЎвЂњР В РЎвЂ“Р В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂўР РЋРІР‚С™Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°.
// EN: What it does: ensures a manager cannot reassign roles for users from another department.
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
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 100, Items: []CalculationItem{{Name: "tech service", Unit: "Р РЋРІР‚РЋ.", Rate: 100, Quantity: 1, LineTotal: 100, Category: CategoryPrimary, ServiceCode: "tech-service"}}}); err != nil {
		t.Fatalf("SaveCalculation technical employee error = %v", err)
	}

	createUserService(t, app, telecomUser.Username, "secret", "telecom service", 110, CategorySecondary)
	loginAsUser(t, app, telecomUser.Username, "secret")
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 110, Items: []CalculationItem{{Name: "telecom service", Unit: "Р РЋРІвЂљВ¬Р РЋРІР‚С™.", Rate: 110, Quantity: 1, LineTotal: 110, Category: CategorySecondary, ServiceCode: "telecom-service"}}}); err != nil {
		t.Fatalf("SaveCalculation telecom employee error = %v", err)
	}

	createUserService(t, app, skudUser.Username, "secret", "skud service", 120, CategoryClosing)
	loginAsUser(t, app, skudUser.Username, "secret")
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 120, Items: []CalculationItem{{Name: "skud service", Unit: "Р РЋРІвЂљВ¬Р РЋРІР‚С™.", Rate: 120, Quantity: 1, LineTotal: 120, Category: CategoryClosing, ServiceCode: "skud-service"}}}); err != nil {
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
		if dept := roleDepartment(user.Role); dept != DepartmentTechnical && dept != DepartmentTelecom {
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
		if dept := roleDepartment(item.CreatedRole); dept != DepartmentTechnical && dept != DepartmentTelecom {
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

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestFrontendRoleLabelsStayReadable`.
// EN: Test `TestFrontendRoleLabelsStayReadable`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р В РЎвЂ”Р В РЎвЂўР В РўвЂР В РЎвЂ”Р В РЎвЂР РЋР С“Р В РЎвЂ Р РЋР вЂљР В РЎвЂўР В Р’В»Р В Р’ВµР В РІвЂћвЂ“ Р В РЎвЂ Р В Р вЂ Р В Р’В°Р РЋР вЂљР В РЎвЂР В Р’В°Р В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„– Р В Р вЂ Р РЋРІР‚в„–Р В Р’В±Р В РЎвЂўР РЋР вЂљР В Р’В° Р РЋР вЂљР В РЎвЂўР В Р’В»Р В Р’ВµР В РІвЂћвЂ“ Р В Р вЂ Р В РЎвЂў Р РЋРІР‚С›Р РЋР вЂљР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’ВµР В Р вЂ¦Р В РўвЂР В Р’Вµ Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋР вЂ№Р РЋРІР‚С™Р РЋР С“Р РЋР РЏ Р РЋРІР‚РЋР В РЎвЂР РЋРІР‚С™Р В Р’В°Р В Р’ВµР В РЎВР РЋРІР‚в„–Р В РЎВР В РЎвЂ.
// EN: What it does: verifies that frontend role labels and role choice captions remain readable text.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚В°Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РІвЂћвЂ“, Р В РЎвЂќР В РЎвЂўР В РЎвЂ“Р В РўвЂР В Р’В° Р РЋР вЂљР РЋРЎвЂњР РЋР С“Р РЋР С“Р В РЎвЂќР В РЎвЂР В Р’Вµ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР В РЎвЂўР В РЎвЂќР В РЎвЂ Р В Р вЂ Р В РЎвЂў Р РЋРІР‚С›Р РЋР вЂљР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’ВµР В Р вЂ¦Р В РўвЂР В Р’Вµ Р В РЎвЂ”Р РЋР вЂљР В Р’ВµР В Р вЂ Р РЋР вЂљР В Р’В°Р РЋРІР‚В°Р В Р’В°Р РЋР вЂ№Р РЋРІР‚С™Р РЋР С“Р РЋР РЏ Р В Р вЂ  `????`.
// EN: Key points: protects against regressions where frontend Russian strings degrade into `????`.
func TestFrontendRoleLabelsStayReadable(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("frontend", "dist", "assets", "app.js"))
	if err != nil {
		t.Fatalf("ReadFile app.js error = %v", err)
	}

	requiredSnippets := [][]byte{
		[]byte(`global_director`),
		[]byte(`executive_director`),
		[]byte(`technical_director`),
		[]byte(`support_head`),
		[]byte(`support_sysadmin`),
		[]byte(`technical_senior`),
		[]byte(`telecom_construction_director`),
		[]byte(`telecom_senior_vols`),
		[]byte(`skud_head`),
		[]byte(`approval_employee`),
		[]byte(`marketing_courier`),
		[]byte(`commercial_senior_mrk`),
		[]byte(`finance_head`),
		[]byte(`legal_employee`),
		[]byte(`development_head`),
		[]byte(`Р В РІР‚СљР В Р’ВµР В Р вЂ¦Р В Р’ВµР РЋР вЂљР В Р’В°Р В Р’В»Р РЋР Р‰Р В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РўвЂР В РЎвЂР РЋР вЂљР В Р’ВµР В РЎвЂќР РЋРІР‚С™Р В РЎвЂўР РЋР вЂљ`),
		[]byte(`Р В Р Р‹Р В РЎвЂР РЋР С“Р РЋРІР‚С™Р В Р’ВµР В РЎВР В Р вЂ¦Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В Р’В°Р В РўвЂР В РЎВР В РЎвЂР В Р вЂ¦Р В РЎвЂР РЋР С“Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋР вЂљ`),
		[]byte(`Р В Р Р‹Р РЋРІР‚С™Р В Р’В°Р РЋР вЂљР РЋРІвЂљВ¬Р В РЎвЂР В РІвЂћвЂ“ Р В РЎВР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’В°Р В Р’В¶Р В Р вЂ¦Р В РЎвЂР В РЎвЂќ Р В РІР‚в„ўР В РЎвЂєР В РІР‚С”Р В Р Р‹`),
		[]byte(`Р В Р’В Р РЋРЎвЂњР В РЎвЂќР В РЎвЂўР В Р вЂ Р В РЎвЂўР В РўвЂР В РЎвЂР РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР Р‰ Р В РЎвЂ“Р РЋР вЂљР РЋРЎвЂњР В РЎвЂ”Р В РЎвЂ”Р РЋРІР‚в„– Р РЋР вЂљР В Р’В°Р В Р’В·Р РЋР вЂљР В Р’В°Р В Р’В±Р В РЎвЂўР РЋРІР‚С™Р В РЎвЂќР В РЎвЂ`),
	}
	for _, snippet := range requiredSnippets {
		if !bytes.Contains(content, snippet) {
			t.Fatalf("app.js is missing readable role label %q", string(snippet))
		}
	}
	if bytes.Contains(content, []byte("????")) {
		t.Fatalf("app.js still contains placeholder question marks in role labels")
	}
}

// RU: Р В РЎС›Р В Р’ВµР РЋР С“Р РЋРІР‚С™ `TestFrontendCoreUiFunctionsExist`.
// EN: Test `TestFrontendCoreUiFunctionsExist`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В Р вЂ Р В Р’ВµР РЋР вЂљР РЋР РЏР В Р’ВµР РЋРІР‚С™, Р РЋРІР‚РЋР РЋРІР‚С™Р В РЎвЂў Р В Р вЂ Р В РЎвЂў Р РЋРІР‚С›Р РЋР вЂљР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’ВµР В Р вЂ¦Р В РўвЂ-Р РЋР С“Р В РЎвЂќР РЋР вЂљР В РЎвЂР В РЎвЂ”Р РЋРІР‚С™Р В Р’Вµ Р В Р’ВµР РЋР С“Р РЋРІР‚С™Р РЋР Р‰ Р В РЎвЂќР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р РЋРІР‚С›Р РЋРЎвЂњР В Р вЂ¦Р В РЎвЂќР РЋРІР‚В Р В РЎвЂР В РЎвЂ Р РЋР вЂљР В Р’В°Р РЋР С“Р РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™Р В Р’В°,
// RU: Р РЋРІР‚С›Р В РЎвЂўР РЋР вЂљР В РЎВР РЋРІР‚в„– Р РЋРЎвЂњР РЋР С“Р В Р’В»Р РЋРЎвЂњР В РЎвЂ“ Р В РЎвЂ Р В Р’В°Р РЋР вЂљР РЋРІР‚В¦Р В РЎвЂР В Р вЂ Р В Р’В°, Р В Р’В° Р В РЎвЂ”Р В РЎвЂўР В РўвЂР В РЎвЂ”Р В РЎвЂР РЋР С“Р В РЎвЂ Р В Р’В°Р РЋР вЂљР РЋРІР‚В¦Р В РЎвЂР В Р вЂ Р В Р’В° Р В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋР вЂ№Р РЋРІР‚С™Р РЋР С“Р РЋР РЏ Р РЋРІР‚РЋР В РЎвЂР РЋРІР‚С™Р В Р’В°Р В Р’ВµР В РЎВР РЋРІР‚в„–Р В РЎВР В РЎвЂ.
// EN: What it does: verifies that the frontend script still contains the core
// EN: calculator, service-form, and archive functions, and that archive labels remain readable.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В Р’В»Р В РЎвЂўР В Р вЂ Р В РЎвЂР РЋРІР‚С™ Р РЋР вЂљР В Р’ВµР В РЎвЂ“Р РЋР вЂљР В Р’ВµР РЋР С“Р РЋР С“Р В РЎвЂР В РЎвЂ Р В РЎвЂ”Р В РЎвЂўР РЋР С“Р В Р’В»Р В Р’Вµ Р РЋР вЂљР РЋРЎвЂњР РЋРІР‚РЋР В Р вЂ¦Р РЋРІР‚в„–Р РЋРІР‚В¦ Р В РЎвЂ”Р РЋР вЂљР В Р’В°Р В Р вЂ Р В РЎвЂўР В РЎвЂќ app.js; Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚В°Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂўР РЋРІР‚С™
// RU: Р В РЎвЂ”Р РЋР вЂљР В РЎвЂўР В РЎвЂ”Р В Р’В°Р В Р’В¶Р В РЎвЂ Р РЋРІР‚С›Р РЋРЎвЂњР В Р вЂ¦Р В РЎвЂќР РЋРІР‚В Р В РЎвЂР В РІвЂћвЂ“ Р В Р вЂ Р РЋР вЂљР В РЎвЂўР В РўвЂР В Р’Вµ renderResult/resetServiceForm Р В РЎвЂ Р В РЎвЂўР РЋРІР‚С™ Р В Р вЂ Р В РЎвЂўР В Р’В·Р В Р вЂ Р РЋР вЂљР В Р’В°Р РЋРІР‚С™Р В Р’В° Р В Р’В±Р В РЎвЂР РЋРІР‚С™Р РЋРІР‚в„–Р РЋРІР‚В¦ Р РЋР С“Р РЋРІР‚С™Р РЋР вЂљР В РЎвЂўР В РЎвЂќ.
// EN: Key points: catches regressions after manual edits to app.js; protects against
// EN: missing functions such as renderResult/resetServiceForm and against broken strings.
func TestFrontendCoreUiFunctionsExist(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("frontend", "dist", "assets", "app.js"))
	if err != nil {
		t.Fatalf("ReadFile app.js error = %v", err)
	}

	requiredSnippets := [][]byte{
		[]byte("function resetServiceForm()"),
		[]byte("function renderResult(result)"),
		[]byte("function renderArchiveDetails(saved)"),
		[]byte("function randomizeCalculation()"),
		[]byte("const randomizeButton = document.getElementById(\"randomize-button\")"),
		[]byte("function renderRoleOptions(roleChoices, selectedValue = \"\")"),
		[]byte("function departmentSortPriority(department)"),
		[]byte("<optgroup label=\"${departmentLabel(group.department)}\">${options}</optgroup>"),
		[]byte("Р В РІР‚в„ўР РЋРІР‚в„–Р В Р’В±Р В Р’ВµР РЋР вЂљР В РЎвЂР РЋРІР‚С™Р В Р’Вµ Р РЋР вЂљР В Р’В°Р РЋР С“Р РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™ Р В РЎвЂР В Р’В· Р В Р’В°Р РЋР вЂљР РЋРІР‚В¦Р В РЎвЂР В Р вЂ Р В Р’В°"),
		[]byte("Р В РЎвЂ™Р РЋР вЂљР РЋРІР‚В¦Р В РЎвЂР В Р вЂ  Р В РЎвЂ”Р В РЎвЂўР В РЎвЂќР В Р’В° Р В РЎвЂ”Р РЋРЎвЂњР РЋР С“Р РЋРІР‚С™ Р В РЎвЂР В Р’В»Р В РЎвЂ Р РЋР вЂљР В Р’В°Р РЋР С“Р РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™ Р В Р’ВµР РЋРІР‚В°Р РЋРІР‚В Р В Р вЂ¦Р В Р’Вµ Р В Р вЂ Р РЋРІР‚в„–Р В Р’В±Р РЋР вЂљР В Р’В°Р В Р вЂ¦."),
	}
	for _, snippet := range requiredSnippets {
		if !bytes.Contains(content, snippet) {
			t.Fatalf("app.js is missing required readable snippet %q", string(snippet))
		}
	}

	if bytes.Contains(content, []byte("ReferenceError")) {
		t.Fatalf("app.js should not contain runtime error text leftovers")
	}
}

// RU: ???? `TestFrontendAdminUsersCanSeeRoleControls`.
// EN: Test `TestFrontendAdminUsersCanSeeRoleControls`.
//
// RU: ??? ??????: ?????????, ??? ???????? ??????? canAdmin ??????????? ???????? ??? ?????? ?????????? ??????.
// EN: What it does: verifies that the frontend treats canAdmin as sufficient for rendering role management controls.
//
// RU: ???????? ???????: ???????? ???????? test-admin `admin1`, ? ???????? ???? admin, ?? UI ?? ?????? ???????? ?????? ?? canManage.
// EN: Key points: protects the admin1 scenario where the user has the admin role and the UI must not depend on canManage alone.
func TestFrontendAdminUsersCanSeeRoleControls(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("frontend", "dist", "assets", "app.js"))
	if err != nil {
		t.Fatalf("ReadFile app.js error = %v", err)
	}

	required := []byte("const canChangeRole = Boolean(appState.session?.canAdmin)")
	if !bytes.Contains(content, required) {
		t.Fatalf("app.js should allow admin users to see role controls")
	}
}

func TestFrontendExportControlsExist(t *testing.T) {
	indexContent, err := os.ReadFile(filepath.Join("frontend", "dist", "index.html"))
	if err != nil {
		t.Fatalf("ReadFile index.html error = %v", err)
	}
	for _, snippet := range [][]byte{
		[]byte(`id="archive-date-input"`),
		[]byte(`id="act-number-input"`),
		[]byte(`id="employee-full-name-input"`),
		[]byte(`id="contract-details-button"`),
		[]byte(`id="contract-details-summary"`),
		[]byte(`id="contract-modal-spbks-number-input"`),
		[]byte(`id="contract-modal-grizabl-number-input"`),
		[]byte(`id="contract-modal-code-spbks"`),
		[]byte(`id="contract-modal-code-grizabl"`),
		[]byte(`id="contract-modal-date-input"`),
		[]byte(`id="contract-modal-fullname-input"`),
		[]byte(`id="act-number-visible-input"`),
		[]byte(`id="previous-act-number"`),
		[]byte(`id="download-pdf-button"`),
	} {
		if !bytes.Contains(indexContent, snippet) {
			t.Fatalf("index.html is missing export control %q", string(snippet))
		}
	}

	jsContent, err := os.ReadFile(filepath.Join("frontend", "dist", "assets", "app.js"))
	if err != nil {
		t.Fatalf("ReadFile app.js error = %v", err)
	}
	for _, snippet := range [][]byte{
		[]byte("UpdateUserFullName"),
		[]byte("ExportCurrentCalculationPDF"),
		[]byte("openContractModal"),
		[]byte("saveContractDetails"),
		[]byte("downloadCurrentCalculationPDF"),
		[]byte("persistActNumber"),
		[]byte("renderActNumberControls"),
		[]byte("renderResult(result)"),
		[]byte("\u0414\u043b\u044f \u0432\u044b\u0433\u0440\u0443\u0437\u043a\u0438 \u0430\u043a\u0442\u0430 \u0443\u043a\u0430\u0436\u0438\u0442\u0435 \u043d\u043e\u043c\u0435\u0440 \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430"),
		[]byte("\u0412\u044b\u0431\u0440\u0430\u043d \u0434\u043e\u0433\u043e\u0432\u043e\u0440:"),
	} {
		if !bytes.Contains(jsContent, snippet) {
			t.Fatalf("app.js is missing export-related snippet %q", string(snippet))
		}
	}
}

