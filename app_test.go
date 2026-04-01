package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// RU: Функция `withTempDB`.
// EN: Function `withTempDB`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: withTempDB creates an isolated application instance backed by a temporary SQLite file for repeatable tests.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func withTempDB(t *testing.T) *App {
	t.Helper()
	originalResolver := resolveDatabasePath
	originalLegacyResolver := resolveLegacyDatabasePath
	tempDir := t.TempDir()
	resolveDatabasePath = func() (string, error) {
		return filepath.Join(tempDir, "test.sqlite"), nil
	}
	resolveLegacyDatabasePath = func() (string, error) {
		return filepath.Join(tempDir, "legacy.sqlite"), nil
	}
	t.Cleanup(func() {
		resolveDatabasePath = originalResolver
		resolveLegacyDatabasePath = originalLegacyResolver
	})

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})
	return app
}

// RU: Функция `loginAsAdmin`.
// EN: Function `loginAsAdmin`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: loginAsAdmin is a helper that authenticates using the seeded admin account.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func loginAsAdmin(t *testing.T, app *App) {
	t.Helper()
	state, err := app.Login(LoginRequest{Username: "admin", Password: "#@7pcehQCSpR"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !state.Authenticated || state.User == nil || state.User.Username != "admin" {
		t.Fatalf("unexpected session state: %+v", state)
	}
}

// RU: Функция `loginAsAdmin1`.
// EN: Function `loginAsAdmin1`.
//
// RU: Что делает: выполняет вход под тестовой admin-учёткой, которая нужна для проверки обычного admin-интерфейса и сценариев.
// EN: What it does: loginAsAdmin1 authenticates using the seeded test admin account for ordinary admin-role scenarios.
//
// RU: Ключевые моменты: использует видимую тестовую admin-учётку; не трогает защищённый username `admin`; упрощает регрессионные проверки.
// EN: Key points: uses the visible regular admin account; does not target the protected `admin` username; keeps regression checks concise.
func loginAsAdmin1(t *testing.T, app *App) {
	t.Helper()
	state, err := app.Login(LoginRequest{Username: "admin1", Password: "admin1"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !state.Authenticated || state.User == nil || state.User.Username != "admin1" {
		t.Fatalf("unexpected session state: %+v", state)
	}
}

// RU: Функция `loginAsUser`.
// EN: Function `loginAsUser`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: loginAsUser authenticates as an arbitrary test user and fails the test immediately on error.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func loginAsUser(t *testing.T, app *App, username string, password string) {
	t.Helper()
	state, err := app.Login(LoginRequest{Username: username, Password: password})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !state.Authenticated || state.User == nil || state.User.Username != username {
		t.Fatalf("unexpected session state: %+v", state)
	}
}

// RU: Функция `createUserService`.
// EN: Function `createUserService`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: createUserService logs in as a test user and creates one owned service for archive and visibility scenarios.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func createUserService(t *testing.T, app *App, username string, password string, name string, rate int, category string) {
	t.Helper()
	loginAsUser(t, app, username, password)
	_, err := app.UpsertService(UpsertServiceRequest{
		Name:     name,
		Unit:     "\u0447.",
		Rate:     rate,
		Category: category,
	})
	if err != nil {
		t.Fatalf("UpsertService() error = %v", err)
	}
}

// RU: Тест `TestSeededAdminLoginWorks`.
// EN: Test `TestSeededAdminLoginWorks`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestSeededAdminLoginWorks verifies that a brand-new database always contains a usable admin account.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestSeededAdminLoginWorks(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)
}

// RU: Тест `TestBootstrapSeedsServices`.
// EN: Test `TestBootstrapSeedsServices`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestBootstrapSeedsServices checks that default bootstrap data includes the expected seeded services.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestNewUserStartsWithoutServicesAndSeesOnlyOwnServices`.
// EN: Test `TestNewUserStartsWithoutServicesAndSeesOnlyOwnServices`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestNewUserStartsWithoutServicesAndSeesOnlyOwnServices ensures service ownership is isolated per account from the first login.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestAdminCanCreateService`.
// EN: Test `TestAdminCanCreateService`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestAdminCanCreateService covers the basic service creation workflow for the admin account.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestAdminCanCreateAndPromoteUser`.
// EN: Test `TestAdminCanCreateAndPromoteUser`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestAdminCanCreateAndPromoteUser verifies that admin can create a user and elevate their role afterwards.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestSeniorSpecialistCanCreateEmployee`.
// EN: Test `TestSeniorSpecialistCanCreateEmployee`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestSeniorSpecialistCanCreateEmployee checks that a senior specialist may only create an ordinary employee.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestManagerCanCreateEmployeeAndSeniorSpecialist`.
// EN: Test `TestManagerCanCreateEmployeeAndSeniorSpecialist`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestManagerCanCreateEmployeeAndSeniorSpecialist confirms manager-level account creation permissions.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestCreateUserAdminRoleFallsBackToEmployee`.
// EN: Test `TestCreateUserAdminRoleFallsBackToEmployee`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestCreateUserAdminRoleFallsBackToEmployee protects against assigning the admin role through the ordinary UI flow.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestAdminAccountIsRestoredAndHiddenFromUserList`.
// EN: Test `TestAdminAccountIsRestoredAndHiddenFromUserList`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestAdminAccountIsRestoredAndHiddenFromUserList ensures the protected admin account cannot disappear from the system.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestAuthenticatedUserCanCalculateAndSave`.
// EN: Test `TestAuthenticatedUserCanCalculateAndSave`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestAuthenticatedUserCanCalculateAndSave validates the happy path from calculation to archive persistence.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
// RU: Тест `TestSeededTestAdminIsVisibleAndDeletableByProtectedAdmin`.
// EN: Test `TestSeededTestAdminIsVisibleAndDeletableByProtectedAdmin`.
//
// RU: Что делает: проверяет, что тестовая учётка admin1 создаётся автоматически, видна как обычный пользователь и может быть удалена защищённым admin.
// EN: What it does: verifies that the admin1 test account is auto-seeded, remains visible as a regular user and can be deleted by the protected admin account.
//
// RU: Ключевые моменты: защищённый admin остаётся скрытым; admin1 имеет роль admin, но без специальной защиты; после удаления вход под admin1 должен перестать работать.
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

	saved, err := app.SaveCalculation(SaveCalculationRequest{Title: "Тестовый расчёт", TargetAmount: result.TargetAmount, Items: result.Items})
	if err != nil {
		t.Fatalf("SaveCalculation() error = %v", err)
	}
	if saved.ID == 0 || saved.CreatedBy != "admin" {
		t.Fatalf("unexpected saved calculation: %+v", saved)
	}
}

// RU: Тест `TestStructuredCalculationKeepsServicesDistributed`.
// EN: Test `TestStructuredCalculationKeepsServicesDistributed`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestStructuredCalculationKeepsServicesDistributed checks that the structured solver avoids collapsing all value into one line.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestStructuredCalculationActivatesAllServicesWhenPossible`.
// EN: Test `TestStructuredCalculationActivatesAllServicesWhenPossible`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestStructuredCalculationActivatesAllServicesWhenPossible ensures all services stay active when the target can support them.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestMigrationAddsCreatedByColumnForLegacyDatabase`.
// EN: Test `TestMigrationAddsCreatedByColumnForLegacyDatabase`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestMigrationAddsCreatedByColumnForLegacyDatabase guards the migration that upgrades old calculation tables.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestMigrationAddsCreatedByColumnForLegacyDatabase(t *testing.T) {
	originalResolver := resolveDatabasePath
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "legacy.sqlite")
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
	_, err = db.Exec(`CREATE TABLE calculations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		target_amount INTEGER NOT NULL,
		total_amount INTEGER NOT NULL,
		items_json TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`)
	if err != nil {
		t.Fatalf("create legacy calculations table: %v", err)
	}
	_ = db.Close()

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	defer app.Close()

	var count int
	if err := app.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('calculations') WHERE name = 'created_by'`).Scan(&count); err != nil {
		t.Fatalf("check created_by column: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected created_by column to be added, got %d", count)
	}
}

// RU: Тест `TestNewDatabasePathCopiesLegacyData`.
// EN: Test `TestNewDatabasePathCopiesLegacyData`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestNewDatabasePathCopiesLegacyData verifies automatic data migration from the old Statistic path to Mercel.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestNewDatabasePathCopiesLegacyData(t *testing.T) {
	originalResolver := resolveDatabasePath
	originalLegacyResolver := resolveLegacyDatabasePath
	tempDir := t.TempDir()
	newPath := filepath.Join(tempDir, "Mercel", "mercel.sqlite")
	legacyPath := filepath.Join(tempDir, "Statistic", "statistic.sqlite")
	resolveDatabasePath = func() (string, error) {
		return newPath, nil
	}
	resolveLegacyDatabasePath = func() (string, error) {
		return legacyPath, nil
	}
	t.Cleanup(func() {
		resolveDatabasePath = originalResolver
		resolveLegacyDatabasePath = originalLegacyResolver
	})

	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	db, err := sql.Open("sqlite", legacyPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			unit TEXT NOT NULL,
			rate INTEGER NOT NULL,
			category TEXT NOT NULL,
			allocation_percent REAL,
			created_by TEXT NOT NULL DEFAULT 'admin',
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE calculations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			target_amount INTEGER NOT NULL,
			total_amount INTEGER NOT NULL,
			items_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			created_by TEXT NOT NULL DEFAULT ''
		);`,
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("create legacy schema: %v", err)
		}
	}
	if _, err := db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", hashPassword("#@7pcehQCSpR"), RoleAdmin, "2026-03-27T00:00:00Z"); err != nil {
		t.Fatalf("insert admin: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "dima", hashPassword("secret"), RoleEmployee, "2026-03-27T00:01:00Z"); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO calculations(title, target_amount, total_amount, items_json, created_at, created_by) VALUES(?, ?, ?, ?, ?, ?)`, "legacy calc", 50000, 50000, "[]", "2026-03-27T00:02:00Z", "dima"); err != nil {
		t.Fatalf("insert calculation: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("db.Close() error = %v", err)
	}

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	defer app.Close()

	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("expected new database to be created from legacy copy: %v", err)
	}
	loginAsAdmin(t, app)
	users, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected copied legacy user plus seeded admin1, got %+v", users)
	}
	usernames := map[string]bool{}
	for _, user := range users {
		usernames[user.Username] = true
	}
	if !usernames["dima"] || !usernames["admin1"] {
		t.Fatalf("expected copied user dima and seeded admin1, got %+v", users)
	}
	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	if len(calculations) != 1 || calculations[0].Title != "legacy calc" {
		t.Fatalf("expected copied archive, got %+v", calculations)
	}
}

// RU: Тест `TestExistingFreshMercelDatabaseIsRecoveredFromLegacy`.
// EN: Test `TestExistingFreshMercelDatabaseIsRecoveredFromLegacy`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestExistingFreshMercelDatabaseIsRecoveredFromLegacy covers recovery when a nearly empty new DB already exists.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestExistingFreshMercelDatabaseIsRecoveredFromLegacy(t *testing.T) {
	originalResolver := resolveDatabasePath
	originalLegacyResolver := resolveLegacyDatabasePath
	tempDir := t.TempDir()
	newPath := filepath.Join(tempDir, "Mercel", "mercel.sqlite")
	legacyPath := filepath.Join(tempDir, "Statistic", "statistic.sqlite")
	resolveDatabasePath = func() (string, error) {
		return newPath, nil
	}
	resolveLegacyDatabasePath = func() (string, error) {
		return legacyPath, nil
	}
	t.Cleanup(func() {
		resolveDatabasePath = originalResolver
		resolveLegacyDatabasePath = originalLegacyResolver
	})

	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		t.Fatalf("MkdirAll new path error = %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll legacy path error = %v", err)
	}

	newDB, err := sql.Open("sqlite", newPath)
	if err != nil {
		t.Fatalf("sql.Open new db error = %v", err)
	}
	freshQueries := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, role TEXT NOT NULL, created_at TEXT NOT NULL);`,
		`CREATE TABLE services (id INTEGER PRIMARY KEY AUTOINCREMENT, code TEXT NOT NULL UNIQUE, name TEXT NOT NULL, unit TEXT NOT NULL, rate INTEGER NOT NULL, category TEXT NOT NULL, allocation_percent REAL, created_by TEXT NOT NULL DEFAULT 'admin', created_at TEXT NOT NULL);`,
		`CREATE TABLE calculations (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, target_amount INTEGER NOT NULL, total_amount INTEGER NOT NULL, items_json TEXT NOT NULL, created_at TEXT NOT NULL, created_by TEXT NOT NULL DEFAULT '');`,
	}
	for _, query := range freshQueries {
		if _, err := newDB.Exec(query); err != nil {
			t.Fatalf("create fresh new schema: %v", err)
		}
	}
	if _, err := newDB.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", hashPassword("#@7pcehQCSpR"), RoleAdmin, "2026-03-28T00:56:47+03:00"); err != nil {
		t.Fatalf("insert fresh admin: %v", err)
	}
	for idx := 0; idx < 8; idx++ {
		code := fmt.Sprintf("service-%d", idx)
		if _, err := newDB.Exec(`INSERT INTO services(code, name, unit, rate, category, allocation_percent, created_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, code, code, "ч.", 100, CategoryPrimary, nil, "admin", "2026-03-28T00:56:47+03:00"); err != nil {
			t.Fatalf("insert fresh service: %v", err)
		}
	}
	if err := newDB.Close(); err != nil {
		t.Fatalf("close fresh new db: %v", err)
	}

	legacyDB, err := sql.Open("sqlite", legacyPath)
	if err != nil {
		t.Fatalf("sql.Open legacy db error = %v", err)
	}
	for _, query := range freshQueries {
		if _, err := legacyDB.Exec(query); err != nil {
			t.Fatalf("create legacy schema: %v", err)
		}
	}
	if _, err := legacyDB.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", hashPassword("#@7pcehQCSpR"), RoleAdmin, "2026-03-27T20:55:29+03:00"); err != nil {
		t.Fatalf("insert legacy admin: %v", err)
	}
	if _, err := legacyDB.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "dima", hashPassword("secret"), RoleEmployee, "2026-03-27T21:02:07+03:00"); err != nil {
		t.Fatalf("insert legacy dima: %v", err)
	}
	if _, err := legacyDB.Exec(`INSERT INTO calculations(title, target_amount, total_amount, items_json, created_at, created_by) VALUES(?, ?, ?, ?, ?, ?)`, "legacy archive", 78560, 78560, "[]", "2026-03-28T00:35:33+03:00", "admin"); err != nil {
		t.Fatalf("insert legacy calc: %v", err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	defer app.Close()

	loginAsAdmin(t, app)
	users, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected recovered legacy user plus seeded admin1, got %+v", users)
	}
	usernames := map[string]bool{}
	for _, user := range users {
		usernames[user.Username] = true
	}
	if !usernames["dima"] || !usernames["admin1"] {
		t.Fatalf("expected recovered legacy dima and seeded admin1, got %+v", users)
	}
	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	if len(calculations) != 1 || calculations[0].Title != "legacy archive" {
		t.Fatalf("expected recovered archive from legacy db, got %+v", calculations)
	}
}

// RU: Тест `TestBuildServiceTargetsDistributesDefaultsEvenly`.
// EN: Test `TestBuildServiceTargetsDistributesDefaultsEvenly`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestBuildServiceTargetsDistributesDefaultsEvenly checks default intra-group percentage distribution.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestBuildServiceTargetsDistributesDefaultsEvenly(t *testing.T) {
	services := []Service{
		{Code: "s1", Category: CategoryPrimary},
		{Code: "s2", Category: CategoryPrimary},
		{Code: "s3", Category: CategoryPrimary},
		{Code: "s4", Category: CategorySecondary},
		{Code: "s5", Category: CategorySecondary},
		{Code: "s6", Category: CategorySecondary},
		{Code: "s7", Category: CategoryClosing},
		{Code: "s8", Category: CategoryClosing},
	}

	byCategory := map[string][]Service{
		CategoryPrimary:   services[:3],
		CategorySecondary: services[3:6],
		CategoryClosing:   services[6:],
	}
	targets := buildServiceTargets(byCategory, map[string]int{
		CategoryPrimary:   79000,
		CategorySecondary: 20000,
		CategoryClosing:   1000,
	}, map[string]int{})
	primary := targets["s1"] + targets["s2"] + targets["s3"]
	secondary := targets["s4"] + targets["s5"] + targets["s6"]
	closing := targets["s7"] + targets["s8"]
	if primary != 79000 || secondary != 20000 || closing != 1000 {
		t.Fatalf("unexpected category totals: primary=%d secondary=%d closing=%d", primary, secondary, closing)
	}

	if absInt(targets["s1"]-targets["s2"]) > 1 || absInt(targets["s2"]-targets["s3"]) > 1 {
		t.Fatalf("primary group should be evenly distributed, got %v", []int{targets["s1"], targets["s2"], targets["s3"]})
	}
	if absInt(targets["s4"]-targets["s5"]) > 1 || absInt(targets["s5"]-targets["s6"]) > 1 {
		t.Fatalf("secondary group should be evenly distributed, got %v", []int{targets["s4"], targets["s5"], targets["s6"]})
	}
	if absInt(targets["s7"]-targets["s8"]) > 1 {
		t.Fatalf("closing group should be evenly distributed, got %v", []int{targets["s7"], targets["s8"]})
	}
}

// RU: Тест `TestBuildServiceTargetsWeightAddsShareWithinGroup`.
// EN: Test `TestBuildServiceTargetsWeightAddsShareWithinGroup`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestBuildServiceTargetsWeightAddsShareWithinGroup proves that per-service weight shifts share within a group.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestBuildServiceTargetsWeightAddsShareWithinGroup(t *testing.T) {
	services := []Service{
		{Code: "s1", Category: CategoryPrimary},
		{Code: "s2", Category: CategoryPrimary},
		{Code: "s3", Category: CategoryPrimary},
	}

	targets := buildServiceTargets(map[string][]Service{
		CategoryPrimary: services,
	}, map[string]int{
		CategoryPrimary: 79000,
	}, map[string]int{
		"s1": 1,
	})

	if targets["s1"] <= targets["s2"] {
		t.Fatalf("expected weighted service to receive a larger share, got s1=%d s2=%d s3=%d", targets["s1"], targets["s2"], targets["s3"])
	}
	if absInt(targets["s2"]-targets["s3"]) > 1 {
		t.Fatalf("expected remaining services to stay balanced, got s2=%d s3=%d", targets["s2"], targets["s3"])
	}
	if targets["s1"]+targets["s2"]+targets["s3"] != 79000 {
		t.Fatalf("expected total to stay unchanged, got %d", targets["s1"]+targets["s2"]+targets["s3"])
	}
}

// RU: Тест `TestSeniorSpecialistCanDeleteEmployeeOnly`.
// EN: Test `TestSeniorSpecialistCanDeleteEmployeeOnly`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestSeniorSpecialistCanDeleteEmployeeOnly enforces deletion boundaries for the senior specialist role.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
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

// RU: Тест `TestManagerCanDeleteSeniorSpecialist`.
// EN: Test `TestManagerCanDeleteSeniorSpecialist`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestManagerCanDeleteSeniorSpecialist validates the broader deletion rights granted to managers.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

	adminUser, err := app.requireAuth()
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

	if _, err := app.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, legacyHashPassword("secret123"), created.ID); err != nil {
		t.Fatalf("force legacy hash error = %v", err)
	}

	app.Logout()
	if _, err := app.Login(LoginRequest{Username: "legacy_user", Password: "secret123"}); err != nil {
		t.Fatalf("Login() legacy user error = %v", err)
	}

	var stored string
	if err := app.db.QueryRow(`SELECT password_hash FROM users WHERE id = ?`, created.ID).Scan(&stored); err != nil {
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

// RU: Тест `TestArchiveVisibilityByRole`.
// EN: Test `TestArchiveVisibilityByRole`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestArchiveVisibilityByRole checks who may see which archived calculations across the role hierarchy.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestManagerCannotAssignManagerRole`.
// EN: Test `TestManagerCannotAssignManagerRole`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestManagerCannotAssignManagerRole prevents managers from creating peers by role reassignment.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestListUsersRespectsRoleVisibility`.
// EN: Test `TestListUsersRespectsRoleVisibility`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestListUsersRespectsRoleVisibility verifies that each role sees only the users they are allowed to manage.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

// RU: Тест `TestListUsersSortsByRolePriority`.
// EN: Test `TestListUsersSortsByRolePriority`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestListUsersSortsByRolePriority locks down the UI ordering of managers, seniors and employees.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
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

	if adminUsers[0].Role != RoleManager || adminUsers[1].Role != RoleSeniorSpecialist || adminUsers[2].Role != RoleEmployee {
		t.Fatalf("expected manager -> senior -> employee order, got %+v", adminUsers[:3])
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

// RU: Тест `TestConfirmModalMarkupUsesReadableUTF8`.
// EN: Test `TestConfirmModalMarkupUsesReadableUTF8`.
//
// RU: Что делает: проверяет отдельный сценарий и фиксирует ожидаемое поведение без ручной проверки.
// EN: What it does: TestConfirmModalMarkupUsesReadableUTF8 protects the modal template from accidental mojibake regressions.
//
// RU: Ключевые моменты: работает в изолированном сценарии; нужен для защиты от регрессий; документирует ожидаемое поведение.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
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

// RU: Тест `TestDepartmentIsolationForUsersAndArchives`.
// EN: Test `TestDepartmentIsolationForUsersAndArchives`.
//
// RU: Что делает: проверяет, что руководители и старшие сотрудники видят только пользователей и архивы своего отдела.
// EN: What it does: verifies that managers and seniors only see users and archives from their own department.
//
// RU: Ключевые моменты: покрывает новую модель ролей по отделам ТП, техотдела и МРК; защищает от межотдельных регрессий видимости.
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

// RU: Тест `TestDepartmentRoleChangesStayInsideDepartment`.
// EN: Test `TestDepartmentRoleChangesStayInsideDepartment`.
//
// RU: Что делает: проверяет, что руководитель не может менять роли пользователей из другого отдела.
// EN: What it does: ensures a manager cannot reassign roles for users from another department.
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

// RU: Тест `TestFrontendRoleLabelsStayReadable`.
// EN: Test `TestFrontendRoleLabelsStayReadable`.
//
// RU: Что делает: проверяет, что подписи ролей и варианты выбора ролей во фронтенде остаются читаемыми.
// EN: What it does: verifies that frontend role labels and role choice captions remain readable text.
//
// RU: Ключевые моменты: защищает от регрессий, когда русские строки во фронтенде превращаются в `????`.
// EN: Key points: protects against regressions where frontend Russian strings degrade into `????`.
func TestFrontendRoleLabelsStayReadable(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("frontend", "dist", "assets", "app.js"))
	if err != nil {
		t.Fatalf("ReadFile app.js error = %v", err)
	}

	requiredSnippets := [][]byte{
		[]byte(`\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u0422\u0435\u0445. \u041f\u043e\u0434\u0434\u0435\u0440\u0436\u043a\u0438`),
		[]byte(`\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u0421\u043f\u0435\u0446\u0438\u0430\u043b\u0438\u0441\u0442 \u0422\u0435\u0445. \u041f\u043e\u0434\u0434\u0435\u0440\u0436\u043a\u0438`),
		[]byte(`\u0421\u043e\u0442\u0440\u0443\u0434\u043d\u0438\u043a \u0422\u0435\u0445. \u041f\u043e\u0434\u0434\u0435\u0440\u0436\u043a\u0438`),
		[]byte(`\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u0422\u0435\u0445\u043d\u0438\u0447\u0435\u0441\u043a\u043e\u0433\u043e \u043e\u0442\u0434\u0435\u043b\u0430`),
		[]byte(`\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u0422\u0435\u0445\u043d\u0438\u043a`),
		[]byte(`\u0422\u0435\u0445\u043d\u0438\u043a`),
		[]byte(`\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u041c\u0420\u041a`),
		[]byte(`\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u041c\u0420\u041a`),
		[]byte(`\u041c\u0420\u041a`),
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

// RU: Тест `TestFrontendCoreUiFunctionsExist`.
// EN: Test `TestFrontendCoreUiFunctionsExist`.
//
// RU: Что делает: проверяет, что во фронтенд-скрипте есть ключевые функции расчёта,
// RU: формы услуг и архива, а подписи архива остаются читаемыми.
// EN: What it does: verifies that the frontend script still contains the core
// EN: calculator, service-form, and archive functions, and that archive labels remain readable.
//
// RU: Ключевые моменты: ловит регрессии после ручных правок app.js; защищает от
// RU: пропажи функций вроде renderResult/resetServiceForm и от возврата битых строк.
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
		[]byte("Выберите расчёт из архива"),
		[]byte("Архив пока пуст или расчёт ещё не выбран."),
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
