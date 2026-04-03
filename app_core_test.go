package main

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestSeededAdminLoginWorks(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)
}

// RU: РўРµСЃС‚ `TestBootstrapSeedsServices`.
// EN: Test `TestBootstrapSeedsServices`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestBootstrapSeedsServices checks that default bootstrap data includes the expected seeded services.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestBootstrapSeedsServices(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	services, err := app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}
	if len(services) != 8 {
		t.Fatalf("expected 8 default services, got %d", len(services))
	}
}

// RU: РўРµСЃС‚ `TestNewUserStartsWithoutServicesAndSeesOnlyOwnServices`.
// EN: Test `TestNewUserStartsWithoutServicesAndSeesOnlyOwnServices`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestNewUserStartsWithoutServicesAndSeesOnlyOwnServices ensures service ownership is isolated per account from the first login.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestNewUserStartsWithoutServicesAndSeesOnlyOwnServices(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	_, err := app.CreateUser(UserWithPassword{Username: "dima", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	_, err = app.CreateUser(UserWithPassword{Username: "sanya", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	loginAsUser(t, app, "dima", "secret")
	services, err := app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}
	if len(services) != 0 {
		t.Fatalf("expected new user to start without services, got %d", len(services))
	}
	created, err := app.UpsertService(UpsertServiceRequest{
		Name:     "personal-service-dima",
		Unit:     "\u0447.",
		Rate:     100,
		Category: CategoryPrimary,
	})
	if err != nil {
		t.Fatalf("UpsertService() error = %v", err)
	}
	if created.CreatedBy != "dima" {
		t.Fatalf("expected service owner dima, got %+v", created)
	}

	loginAsUser(t, app, "sanya", "secret")
	services, err = app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}
	if len(services) != 0 {
		t.Fatalf("expected sanya to not see dima services, got %d", len(services))
	}

	loginAsAdmin(t, app)
	adminServices, err := app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}
	if len(adminServices) != 8 {
		t.Fatalf("expected admin to keep seeded services, got %d", len(adminServices))
	}
}

// RU: РўРµСЃС‚ `TestAdminCanCreateService`.
// EN: Test `TestAdminCanCreateService`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestAdminCanCreateService covers the basic service creation workflow for the admin account.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestAdminCanCreateService(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	percent := 12.5
	created, err := app.UpsertService(UpsertServiceRequest{
		Name:              "test-service",
		Unit:              "\u0448\u0442.",
		Rate:              999,
		Category:          CategorySecondary,
		AllocationPercent: &percent,
	})
	if err != nil {
		t.Fatalf("UpsertService() error = %v", err)
	}
	if created.ID == 0 || created.Name != "test-service" {
		t.Fatalf("unexpected created service: %+v", created)
	}
	if created.AllocationPercent == nil || *created.AllocationPercent != percent {
		t.Fatalf("expected allocation percent %.2f, got %+v", percent, created.AllocationPercent)
	}
}

// RU: РўРµСЃС‚ `TestAdminCanCreateAndPromoteUser`.
// EN: Test `TestAdminCanCreateAndPromoteUser`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestAdminCanCreateAndPromoteUser verifies that admin can create a user and elevate their role afterwards.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestAdminCanCreateAndPromoteUser(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	created, err := app.CreateUser(UserWithPassword{
		Username: "worker",
		Password: "secret",
		Role:     RoleEmployee,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if created.Username != "worker" {
		t.Fatalf("unexpected created user: %+v", created)
	}

	updated, err := app.UpdateUserRole(created.ID, RoleSeniorSpecialist)
	if err != nil {
		t.Fatalf("UpdateUserRole() error = %v", err)
	}
	if updated.Role != RoleSeniorSpecialist {
		t.Fatalf("expected role %q, got %q", RoleSeniorSpecialist, updated.Role)
	}
}

// RU: РўРµСЃС‚ `TestSeniorSpecialistCanCreateEmployee`.
// EN: Test `TestSeniorSpecialistCanCreateEmployee`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestSeniorSpecialistCanCreateEmployee checks that a senior specialist may only create an ordinary employee.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestSeniorSpecialistCanCreateEmployee(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	senior, err := app.CreateUser(UserWithPassword{
		Username: "senior-create",
		Password: "secret",
		Role:     RoleSeniorSpecialist,
	})
	if err != nil {
		t.Fatalf("CreateUser() senior error = %v", err)
	}

	loginAsUser(t, app, senior.Username, "secret")
	created, err := app.CreateUser(UserWithPassword{
		Username: "employee-from-senior",
		Password: "secret",
		Role:     RoleEmployee,
	})
	if err != nil {
		t.Fatalf("CreateUser() by senior specialist error = %v", err)
	}
	if created.Role != RoleEmployee {
		t.Fatalf("expected role %q, got %q", RoleEmployee, created.Role)
	}
}

// RU: РўРµСЃС‚ `TestManagerCanCreateEmployeeAndSeniorSpecialist`.
// EN: Test `TestManagerCanCreateEmployeeAndSeniorSpecialist`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestManagerCanCreateEmployeeAndSeniorSpecialist confirms manager-level account creation permissions.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestManagerCanCreateEmployeeAndSeniorSpecialist(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	manager, err := app.CreateUser(UserWithPassword{
		Username: "manager-create",
		Password: "secret",
		Role:     RoleManager,
	})
	if err != nil {
		t.Fatalf("CreateUser() manager error = %v", err)
	}

	loginAsUser(t, app, manager.Username, "secret")
	employee, err := app.CreateUser(UserWithPassword{
		Username: "employee-from-manager",
		Password: "secret",
		Role:     RoleEmployee,
	})
	if err != nil {
		t.Fatalf("CreateUser() employee by manager error = %v", err)
	}
	if employee.Role != RoleEmployee {
		t.Fatalf("expected employee role, got %q", employee.Role)
	}

	senior, err := app.CreateUser(UserWithPassword{
		Username: "senior-from-manager",
		Password: "secret",
		Role:     RoleSeniorSpecialist,
	})
	if err != nil {
		t.Fatalf("CreateUser() senior by manager error = %v", err)
	}
	if senior.Role != RoleSeniorSpecialist {
		t.Fatalf("expected senior role, got %q", senior.Role)
	}
}

// RU: РўРµСЃС‚ `TestCreateUserAdminRoleFallsBackToEmployee`.
// EN: Test `TestCreateUserAdminRoleFallsBackToEmployee`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestCreateUserAdminRoleFallsBackToEmployee protects against assigning the admin role through the ordinary UI flow.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestCreateUserAdminRoleFallsBackToEmployee(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	created, err := app.CreateUser(UserWithPassword{
		Username: "candidate",
		Password: "secret",
		Role:     RoleAdmin,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if created.Role != RoleEmployee {
		t.Fatalf("expected fallback role %q, got %q", RoleEmployee, created.Role)
	}
}

// RU: РўРµСЃС‚ `TestAdminAccountIsRestoredAndHiddenFromUserList`.
// EN: Test `TestAdminAccountIsRestoredAndHiddenFromUserList`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestAdminAccountIsRestoredAndHiddenFromUserList ensures the protected admin account cannot disappear from the system.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestAdminAccountIsRestoredAndHiddenFromUserList(t *testing.T) {
	originalResolver := resolveDatabasePath
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "admin-reset.sqlite")
	resolveDatabasePath = func() (string, error) {
		return dbPath, nil
	}
	t.Cleanup(func() {
		resolveDatabasePath = originalResolver
	})

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`)
	if err != nil {
		t.Fatalf("create users table: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE services (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		unit TEXT NOT NULL,
		rate INTEGER NOT NULL,
		category TEXT NOT NULL,
		allocation_percent REAL,
		created_at TEXT NOT NULL
	);`)
	if err != nil {
		t.Fatalf("create services table: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE calculations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		target_amount INTEGER NOT NULL,
		total_amount INTEGER NOT NULL,
		items_json TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`)
	if err != nil {
		t.Fatalf("create calculations table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", hashPassword("broken"), RoleEmployee, "2026-03-27T00:00:00Z")
	if err != nil {
		t.Fatalf("insert broken admin: %v", err)
	}
	_ = db.Close()

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	defer app.Close()

	state, err := app.Login(LoginRequest{Username: "admin", Password: "#@7pcehQCSpR"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if state.User == nil || state.User.Role != RoleAdmin {
		t.Fatalf("expected admin role to be restored, got %+v", state.User)
	}

	_, err = app.CreateUser(UserWithPassword{Username: "worker", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	users, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	visibleAdmin1 := false
	for _, user := range users {
		if user.Username == "admin" {
			t.Fatalf("admin should not be present in list users")
		}
		if user.Username == "admin1" {
			visibleAdmin1 = true
		}
	}
	if !visibleAdmin1 {
		t.Fatalf("expected admin1 to be present in list users after seeding")
	}
}

// RU: РўРµСЃС‚ `TestAuthenticatedUserCanCalculateAndSave`.
// EN: Test `TestAuthenticatedUserCanCalculateAndSave`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestAuthenticatedUserCanCalculateAndSave validates the happy path from calculation to archive persistence.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
// RU: РўРµСЃС‚ `TestSeededTestAdminIsVisibleAndDeletableByProtectedAdmin`.
// EN: Test `TestSeededTestAdminIsVisibleAndDeletableByProtectedAdmin`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ С‚РµСЃС‚РѕРІР°СЏ СѓС‡С‘С‚РєР° admin1 СЃРѕР·РґР°С‘С‚СЃСЏ Р°РІС‚РѕРјР°С‚РёС‡РµСЃРєРё, РІРёРґРЅР° РєР°Рє РѕР±С‹С‡РЅС‹Р№ РїРѕР»СЊР·РѕРІР°С‚РµР»СЊ Рё РјРѕР¶РµС‚ Р±С‹С‚СЊ СѓРґР°Р»РµРЅР° Р·Р°С‰РёС‰С‘РЅРЅС‹Рј admin.
// EN: What it does: verifies that the admin1 test account is auto-seeded, remains visible as a regular user and can be deleted by the protected admin account.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: Р·Р°С‰РёС‰С‘РЅРЅС‹Р№ admin РѕСЃС‚Р°С‘С‚СЃСЏ СЃРєСЂС‹С‚С‹Рј; admin1 РёРјРµРµС‚ СЂРѕР»СЊ admin, РЅРѕ Р±РµР· СЃРїРµС†РёР°Р»СЊРЅРѕР№ Р·Р°С‰РёС‚С‹; РїРѕСЃР»Рµ СѓРґР°Р»РµРЅРёСЏ РІС…РѕРґ РїРѕРґ admin1 РґРѕР»Р¶РµРЅ РїРµСЂРµСЃС‚Р°С‚СЊ СЂР°Р±РѕС‚Р°С‚СЊ.
// EN: Key points: the protected admin stays hidden; admin1 has the admin role but no special protection; login as admin1 must fail after deletion.
func TestSeededTestAdminIsVisibleAndDeletableByProtectedAdmin(t *testing.T) {
	app := withTempDB(t)

	loginAsAdmin1(t, app)
	app.Logout()

	loginAsAdmin(t, app)
	users, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	var admin1 *User
	for i := range users {
		if users[i].Username == "admin" {
			t.Fatalf("protected admin should not be present in list users")
		}
		if users[i].Username == "admin1" {
			admin1 = &users[i]
		}
	}
	if admin1 == nil {
		t.Fatalf("expected admin1 to be visible in list users, got %+v", users)
	}
	if admin1.Role != RoleAdmin {
		t.Fatalf("expected admin1 to keep admin role, got %+v", admin1)
	}

	if err := app.DeleteUser(admin1.ID); err != nil {
		t.Fatalf("DeleteUser(admin1) error = %v", err)
	}

	app.Logout()
	if _, err := app.Login(LoginRequest{Username: "admin1", Password: "admin1"}); err == nil {
		t.Fatalf("expected admin1 login to fail after deletion")
	}
}

func TestAuthenticatedUserCanCalculateAndSave(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	result, err := app.CalculateAmount(CalculationRequest{TargetAmount: 50000, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount() error = %v", err)
	}
	if result.TotalAmount != 50000 {
		t.Fatalf("expected total 50000, got %d", result.TotalAmount)
	}

	saved, err := app.SaveCalculation(SaveCalculationRequest{Title: "РўРµСЃС‚РѕРІС‹Р№ СЂР°СЃС‡С‘С‚", TargetAmount: result.TargetAmount, Items: result.Items})
	if err != nil {
		t.Fatalf("SaveCalculation() error = %v", err)
	}
	if saved.ID == 0 || saved.CreatedBy != "admin" {
		t.Fatalf("unexpected saved calculation: %+v", saved)
	}
}

// RU: РўРµСЃС‚ `TestStructuredCalculationKeepsServicesDistributed`.
// EN: Test `TestStructuredCalculationKeepsServicesDistributed`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestStructuredCalculationKeepsServicesDistributed checks that the structured solver avoids collapsing all value into one line.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestSaveCalculationUsesDailyTitleAndOverwritesSameDay(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	resultOne, err := app.CalculateAmount(CalculationRequest{TargetAmount: 50000, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount first error = %v", err)
	}
	savedOne, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: resultOne.TargetAmount, Items: resultOne.Items})
	if err != nil {
		t.Fatalf("SaveCalculation first error = %v", err)
	}

	resultTwo, err := app.CalculateAmount(CalculationRequest{TargetAmount: 50100, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount second error = %v", err)
	}
	savedTwo, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: resultTwo.TargetAmount, Items: resultTwo.Items})
	if err != nil {
		t.Fatalf("SaveCalculation second error = %v", err)
	}

	expectedTitle := archiveDateTitle(time.Now())
	if savedOne.Title != expectedTitle || savedTwo.Title != expectedTitle {
		t.Fatalf("expected daily title %q, got first=%q second=%q", expectedTitle, savedOne.Title, savedTwo.Title)
	}
	if savedOne.ID != savedTwo.ID {
		t.Fatalf("expected same calculation row to be overwritten, got ids %d and %d", savedOne.ID, savedTwo.ID)
	}

	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations error = %v", err)
	}
	if len(calculations) != 1 {
		t.Fatalf("expected one archived calculation for the day, got %+v", calculations)
	}
	if calculations[0].TotalAmount != resultTwo.TotalAmount || calculations[0].TargetAmount != resultTwo.TargetAmount {
		t.Fatalf("expected archive to contain the overwritten latest calculation, got %+v", calculations[0])
	}
}

func TestArchiveSurvivesAuthorDeletion(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	employee, err := app.CreateUser(UserWithPassword{Username: "employee_archive", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser employee error = %v", err)
	}

	loginAsUser(t, app, employee.Username, "secret")
	createUserService(t, app, employee.Username, "secret", "employee archived service", 100, CategoryPrimary)
	result, err := app.CalculateAmount(CalculationRequest{TargetAmount: 50100, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount employee error = %v", err)
	}
	if _, err := app.SaveCalculation(SaveCalculationRequest{TargetAmount: result.TargetAmount, Items: result.Items}); err != nil {
		t.Fatalf("SaveCalculation employee error = %v", err)
	}

	loginAsAdmin(t, app)
	if err := app.DeleteUser(employee.ID); err != nil {
		t.Fatalf("DeleteUser employee error = %v", err)
	}

	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations admin error = %v", err)
	}
	found := false
	for _, item := range calculations {
		if item.CreatedBy == employee.Username {
			found = true
			if item.CreatedRole != RoleEmployee {
				t.Fatalf("expected archived role to stay employee, got %+v", item)
			}
		}
	}
	if !found {
		t.Fatalf("expected archived calculation to remain visible after author deletion, got %+v", calculations)
	}
}

func TestAdminCanCopyServicesFromArchivedCalculation(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	if _, err := app.UpsertService(UpsertServiceRequest{Name: "\u0421\u0442\u0430\u0440\u0430\u044f \u0430\u0434\u043c\u0438\u043d\u0441\u043a\u0430\u044f \u0443\u0441\u043b\u0443\u0433\u0430", Unit: "\u0447\u0430\u0441\u044b", Rate: 90, Category: CategoryPrimary}); err != nil {
		t.Fatalf("seed admin service error = %v", err)
	}
	worker, err := app.CreateUser(UserWithPassword{Username: "archive_source", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser worker error = %v", err)
	}

	loginAsUser(t, app, worker.Username, "secret")
	saved, err := app.SaveCalculation(SaveCalculationRequest{
		TargetAmount: 470,
		Items: []CalculationItem{
			{Name: "Imported New", Unit: "\u0447.", Rate: 120, Quantity: 1, LineTotal: 120, Category: CategoryPrimary, ServiceCode: "imported-new"},
			{Name: "Imported Update", Unit: "\u0448\u0442.", Rate: 350, Quantity: 1, LineTotal: 350, Category: CategorySecondary, ServiceCode: "imported-update"},
		},
	})
	if err != nil {
		t.Fatalf("SaveCalculation worker error = %v", err)
	}

	loginAsAdmin(t, app)
	copied, err := app.CopyArchiveServicesToAdmin(saved.ID)
	if err != nil {
		t.Fatalf("CopyArchiveServicesToAdmin() error = %v", err)
	}
	if copied.Created != 2 || copied.Updated != 0 {
		t.Fatalf("expected full replacement with 2 created services, got %+v", copied)
	}

	services, err := app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}
	foundNew := false
	foundUpdated := false
	for _, service := range services {
		switch service.Name {
		case "Imported New":
			foundNew = true
			if service.Rate != 120 || service.Unit != "\u0447." || service.Category != CategoryPrimary {
				t.Fatalf("unexpected imported new service: %+v", service)
			}
		case "Imported Update":
			foundUpdated = true
			if service.Rate != 350 || service.Unit != "\u0448\u0442." || service.Category != CategorySecondary {
				t.Fatalf("expected admin service to be updated from archive, got %+v", service)
			}
		}
	}
	if !foundNew || !foundUpdated {
		t.Fatalf("expected copied services to be visible in admin list, got %+v", services)
	}
	for _, service := range services {
		if service.Name == "\u0421\u0442\u0430\u0440\u0430\u044f \u0430\u0434\u043c\u0438\u043d\u0441\u043a\u0430\u044f \u0443\u0441\u043b\u0443\u0433\u0430" {
			t.Fatalf("expected previous admin services to be removed before copy, got %+v", services)
		}
	}
}

func TestNonAdminCannotCopyServicesFromArchivedCalculation(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	worker, err := app.CreateUser(UserWithPassword{Username: "archive_source_2", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser worker error = %v", err)
	}
	manager, err := app.CreateUser(UserWithPassword{Username: "support_manager_copy", Password: "secret", Role: RoleManager})
	if err != nil {
		t.Fatalf("CreateUser manager error = %v", err)
	}

	loginAsUser(t, app, worker.Username, "secret")
	saved, err := app.SaveCalculation(SaveCalculationRequest{
		TargetAmount: 120,
		Items:        []CalculationItem{{Name: "Copy Guard", Unit: "\u0447.", Rate: 120, Quantity: 1, LineTotal: 120, Category: CategoryPrimary, ServiceCode: "copy-guard"}},
	})
	if err != nil {
		t.Fatalf("SaveCalculation worker error = %v", err)
	}

	loginAsUser(t, app, manager.Username, "secret")
	if _, err := app.CopyArchiveServicesToAdmin(saved.ID); err == nil {
		t.Fatalf("expected non-admin copy from archive to be rejected")
	}
}

func TestStructuredCalculationKeepsServicesDistributed(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	result, err := app.CalculateAmount(CalculationRequest{TargetAmount: 88640, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount() error = %v", err)
	}
	if result.TotalAmount != 88640 {
		t.Fatalf("expected exact total 88640, got %d", result.TotalAmount)
	}

	primaryTotal := 0
	secondaryTotal := 0
	closingTotal := 0
	activePrimary := 0
	activeSecondary := 0
	activeClosing := 0
	for _, item := range result.Items {
		switch item.Category {
		case CategoryPrimary:
			primaryTotal += item.LineTotal
			if item.Quantity > 0 {
				activePrimary++
			}
		case CategorySecondary:
			secondaryTotal += item.LineTotal
			if item.Quantity > 0 {
				activeSecondary++
			}
		case CategoryClosing:
			closingTotal += item.LineTotal
			if item.Quantity > 0 {
				activeClosing++
			}
		}
	}

	if activePrimary < 2 || activeSecondary < 2 || activeClosing < 1 {
		t.Fatalf("expected calculation to stay distributed, got active primary=%d secondary=%d closing=%d", activePrimary, activeSecondary, activeClosing)
	}
	if primaryTotal < 50000 || secondaryTotal < 10000 {
		t.Fatalf("expected primary and secondary groups to contribute materially, got primary=%d secondary=%d closing=%d", primaryTotal, secondaryTotal, closingTotal)
	}
}

// RU: РўРµСЃС‚ `TestStructuredCalculationActivatesAllServicesWhenPossible`.
// EN: Test `TestStructuredCalculationActivatesAllServicesWhenPossible`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestStructuredCalculationActivatesAllServicesWhenPossible ensures all services stay active when the target can support them.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestStructuredCalculationActivatesAllServicesWhenPossible(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	result, err := app.CalculateAmount(CalculationRequest{TargetAmount: 79640, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount() error = %v", err)
	}
	if result.TotalAmount != 79640 {
		t.Fatalf("expected exact total 79640, got %d", result.TotalAmount)
	}
	if result.ActiveServices != len(result.Items) {
		t.Fatalf("expected all services to be active, got %d of %d", result.ActiveServices, len(result.Items))
	}
}

// RU: РўРµСЃС‚ `TestMigrationAddsCreatedByColumnForLegacyDatabase`.
// EN: Test `TestMigrationAddsCreatedByColumnForLegacyDatabase`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestMigrationAddsCreatedByColumnForLegacyDatabase guards the migration that upgrades old calculation tables.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
