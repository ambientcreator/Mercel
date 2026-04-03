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

// RU: РўРµСЃС‚ `TestDepartmentIsolationForUsersAndArchives`.
// EN: Test `TestDepartmentIsolationForUsersAndArchives`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЂСѓРєРѕРІРѕРґРёС‚РµР»Рё Рё СЃС‚Р°СЂС€РёРµ СЃРѕС‚СЂСѓРґРЅРёРєРё РІРёРґСЏС‚ С‚РѕР»СЊРєРѕ РїРѕР»СЊР·РѕРІР°С‚РµР»РµР№ Рё Р°СЂС…РёРІС‹ СЃРІРѕРµРіРѕ РѕС‚РґРµР»Р°.
// EN: What it does: verifies that managers and seniors only see users and archives from their own department.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РїРѕРєСЂС‹РІР°РµС‚ РЅРѕРІСѓСЋ РјРѕРґРµР»СЊ СЂРѕР»РµР№ РїРѕ РѕС‚РґРµР»Р°Рј РўРџ, С‚РµС…РѕС‚РґРµР»Р° Рё РњР Рљ; Р·Р°С‰РёС‰Р°РµС‚ РѕС‚ РјРµР¶РѕС‚РґРµР»СЊРЅС‹С… СЂРµРіСЂРµСЃСЃРёР№ РІРёРґРёРјРѕСЃС‚Рё.
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

// RU: РўРµСЃС‚ `TestDepartmentRoleChangesStayInsideDepartment`.
// EN: Test `TestDepartmentRoleChangesStayInsideDepartment`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЂСѓРєРѕРІРѕРґРёС‚РµР»СЊ РЅРµ РјРѕР¶РµС‚ РјРµРЅСЏС‚СЊ СЂРѕР»Рё РїРѕР»СЊР·РѕРІР°С‚РµР»РµР№ РёР· РґСЂСѓРіРѕРіРѕ РѕС‚РґРµР»Р°.
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
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 100, Items: []CalculationItem{{Name: "tech service", Unit: "С‡.", Rate: 100, Quantity: 1, LineTotal: 100, Category: CategoryPrimary, ServiceCode: "tech-service"}}}); err != nil {
		t.Fatalf("SaveCalculation technical employee error = %v", err)
	}

	createUserService(t, app, telecomUser.Username, "secret", "telecom service", 110, CategorySecondary)
	loginAsUser(t, app, telecomUser.Username, "secret")
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 110, Items: []CalculationItem{{Name: "telecom service", Unit: "С€С‚.", Rate: 110, Quantity: 1, LineTotal: 110, Category: CategorySecondary, ServiceCode: "telecom-service"}}}); err != nil {
		t.Fatalf("SaveCalculation telecom employee error = %v", err)
	}

	createUserService(t, app, skudUser.Username, "secret", "skud service", 120, CategoryClosing)
	loginAsUser(t, app, skudUser.Username, "secret")
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: 120, Items: []CalculationItem{{Name: "skud service", Unit: "С€С‚.", Rate: 120, Quantity: 1, LineTotal: 120, Category: CategoryClosing, ServiceCode: "skud-service"}}}); err != nil {
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

// RU: РўРµСЃС‚ `TestFrontendRoleLabelsStayReadable`.
// EN: Test `TestFrontendRoleLabelsStayReadable`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РїРѕРґРїРёСЃРё СЂРѕР»РµР№ Рё РІР°СЂРёР°РЅС‚С‹ РІС‹Р±РѕСЂР° СЂРѕР»РµР№ РІРѕ С„СЂРѕРЅС‚РµРЅРґРµ РѕСЃС‚Р°СЋС‚СЃСЏ С‡РёС‚Р°РµРјС‹РјРё.
// EN: What it does: verifies that frontend role labels and role choice captions remain readable text.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: Р·Р°С‰РёС‰Р°РµС‚ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№, РєРѕРіРґР° СЂСѓСЃСЃРєРёРµ СЃС‚СЂРѕРєРё РІРѕ С„СЂРѕРЅС‚РµРЅРґРµ РїСЂРµРІСЂР°С‰Р°СЋС‚СЃСЏ РІ `????`.
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
		[]byte(`Р“РµРЅРµСЂР°Р»СЊРЅС‹Р№ РґРёСЂРµРєС‚РѕСЂ`),
		[]byte(`РЎРёСЃС‚РµРјРЅС‹Р№ Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂ`),
		[]byte(`РЎС‚Р°СЂС€РёР№ РјРѕРЅС‚Р°Р¶РЅРёРє Р’РћР›РЎ`),
		[]byte(`Р СѓРєРѕРІРѕРґРёС‚РµР»СЊ РіСЂСѓРїРїС‹ СЂР°Р·СЂР°Р±РѕС‚РєРё`),
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

// RU: РўРµСЃС‚ `TestFrontendCoreUiFunctionsExist`.
// EN: Test `TestFrontendCoreUiFunctionsExist`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ РІРѕ С„СЂРѕРЅС‚РµРЅРґ-СЃРєСЂРёРїС‚Рµ РµСЃС‚СЊ РєР»СЋС‡РµРІС‹Рµ С„СѓРЅРєС†РёРё СЂР°СЃС‡С‘С‚Р°,
// RU: С„РѕСЂРјС‹ СѓСЃР»СѓРі Рё Р°СЂС…РёРІР°, Р° РїРѕРґРїРёСЃРё Р°СЂС…РёРІР° РѕСЃС‚Р°СЋС‚СЃСЏ С‡РёС‚Р°РµРјС‹РјРё.
// EN: What it does: verifies that the frontend script still contains the core
// EN: calculator, service-form, and archive functions, and that archive labels remain readable.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: Р»РѕРІРёС‚ СЂРµРіСЂРµСЃСЃРёРё РїРѕСЃР»Рµ СЂСѓС‡РЅС‹С… РїСЂР°РІРѕРє app.js; Р·Р°С‰РёС‰Р°РµС‚ РѕС‚
// RU: РїСЂРѕРїР°Р¶Рё С„СѓРЅРєС†РёР№ РІСЂРѕРґРµ renderResult/resetServiceForm Рё РѕС‚ РІРѕР·РІСЂР°С‚Р° Р±РёС‚С‹С… СЃС‚СЂРѕРє.
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
		[]byte("Р’С‹Р±РµСЂРёС‚Рµ СЂР°СЃС‡С‘С‚ РёР· Р°СЂС…РёРІР°"),
		[]byte("РђСЂС…РёРІ РїРѕРєР° РїСѓСЃС‚ РёР»Рё СЂР°СЃС‡С‘С‚ РµС‰С‘ РЅРµ РІС‹Р±СЂР°РЅ."),
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

	required := []byte("const canChangeRole = Boolean(appState.session.canAdmin)")
	if !bytes.Contains(content, required) {
		t.Fatalf("app.js should allow admin users to see role controls")
	}
}
