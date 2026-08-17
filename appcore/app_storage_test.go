package appcore_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	. "statistic/appcore"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrationAddsCreatedByColumnForLegacyDatabase(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "legacy.sqlite")
	usePathResolvers(t, dbPath, "", "")

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
	if err := app.DBForTest().QueryRow(`SELECT COUNT(*) FROM pragma_table_info('calculations') WHERE name = 'created_by'`).Scan(&count); err != nil {
		t.Fatalf("check created_by column: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected created_by column to be added, got %d", count)
	}
}

func TestNewAppStartsWithEmptyDatabaseAndSeedsDefaultServices(t *testing.T) {
	tempDir := t.TempDir()
	newPath := filepath.Join(tempDir, AppStorageDirName, AppStorageDBName)
	usePathResolvers(t, newPath, filepath.Join(tempDir, "missing-statistic.sqlite"), filepath.Join(tempDir, "missing-mercel.sqlite"))
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() on empty database error = %v", err)
	}
	defer app.Close()

	loginAsAdmin(t, app)
	services, err := app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}
	if len(services) != 8 {
		t.Fatalf("expected 8 seeded admin services, got %d", len(services))
	}
	for _, service := range services {
		if service.Unit != "ч." && service.Unit != "шт." {
			t.Fatalf("unexpected seeded unit %q for service %+v", service.Unit, service)
		}
		if service.Rate <= 0 {
			t.Fatalf("unexpected seeded rate %d for service %+v", service.Rate, service)
		}
	}

	state := app.Logout()
	if state.Authenticated {
		t.Fatalf("expected logged out state, got %+v", state)
	}

	var deprecatedUsers int
	if err := app.DBForTest().QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, "admin1").Scan(&deprecatedUsers); err != nil {
		t.Fatalf("check deprecated admin1 user: %v", err)
	}
	if deprecatedUsers != 0 {
		t.Fatalf("expected deprecated admin1 user to be absent, got %d", deprecatedUsers)
	}

	var deprecatedServices int
	if err := app.DBForTest().QueryRow(`SELECT COUNT(*) FROM services WHERE created_by = ?`, "admin1").Scan(&deprecatedServices); err != nil {
		t.Fatalf("check deprecated admin1 services: %v", err)
	}
	if deprecatedServices != 0 {
		t.Fatalf("expected deprecated admin1 services to be absent, got %d", deprecatedServices)
	}
}

// EN: Test `TestNewDatabasePathCopiesLegacyData`.
//
// EN: What it does: TestNewDatabasePathCopiesLegacyData verifies automatic data migration from the old Statistic path to Mercel.
//
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestNewDatabasePathCopiesLegacyData(t *testing.T) {
	tempDir := t.TempDir()
	newPath := filepath.Join(tempDir, "Mercel", "mercel.sqlite")
	legacyPath := filepath.Join(tempDir, "Statistic", "statistic.sqlite")
	usePathResolvers(t, newPath, legacyPath, "")

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
	if _, err := db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", HashPasswordForTest(legacyFixturePassword), RoleAdmin, "2026-03-27T00:00:00Z"); err != nil {
		t.Fatalf("insert admin: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "dima", HashPasswordForTest("secret"), RoleEmployee, "2026-03-27T00:01:00Z"); err != nil {
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
	if len(users) != 1 {
		t.Fatalf("expected copied legacy user to remain visible, got %+v", users)
	}
	usernames := map[string]bool{}
	for _, user := range users {
		usernames[user.Username] = true
	}
	if !usernames["dima"] {
		t.Fatalf("expected copied user dima, got %+v", users)
	}
	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	if len(calculations) != 1 || calculations[0].Title != "legacy calc" {
		t.Fatalf("expected copied archive, got %+v", calculations)
	}
}

// EN: Test `TestExistingFreshMercelDatabaseIsRecoveredFromLegacy`.
//
// EN: What it does: TestExistingFreshMercelDatabaseIsRecoveredFromLegacy covers recovery when a nearly empty new DB already exists.
//
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestExistingFreshMercelDatabaseIsRecoveredFromLegacy(t *testing.T) {
	tempDir := t.TempDir()
	newPath := filepath.Join(tempDir, "Mercel", "mercel.sqlite")
	legacyPath := filepath.Join(tempDir, "Statistic", "statistic.sqlite")
	usePathResolvers(t, newPath, legacyPath, "")

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
	if _, err := newDB.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", HashPasswordForTest(legacyFixturePassword), RoleAdmin, "2026-03-28T00:56:47+03:00"); err != nil {
		t.Fatalf("insert fresh admin: %v", err)
	}
	for idx := 0; idx < 8; idx++ {
		code := fmt.Sprintf("service-%d", idx)
		if _, err := newDB.Exec(`INSERT INTO services(code, name, unit, rate, category, allocation_percent, created_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, code, code, "Р РЋРІР‚РЋ.", 100, CategoryPrimary, nil, "admin", "2026-03-28T00:56:47+03:00"); err != nil {
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
	if _, err := legacyDB.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", HashPasswordForTest(legacyFixturePassword), RoleAdmin, "2026-03-27T20:55:29+03:00"); err != nil {
		t.Fatalf("insert legacy admin: %v", err)
	}
	if _, err := legacyDB.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "dima", HashPasswordForTest("secret"), RoleEmployee, "2026-03-27T21:02:07+03:00"); err != nil {
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
	if len(users) != 1 {
		t.Fatalf("expected recovered legacy user to remain visible, got %+v", users)
	}
	usernames := map[string]bool{}
	for _, user := range users {
		usernames[user.Username] = true
	}
	if !usernames["dima"] {
		t.Fatalf("expected recovered legacy dima, got %+v", users)
	}
	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	if len(calculations) != 1 || calculations[0].Title != "legacy archive" {
		t.Fatalf("expected recovered archive from legacy db, got %+v", calculations)
	}
}

func TestExistingFreshStableStorageRecoversFromPreviousMercelPath(t *testing.T) {
	tempDir := t.TempDir()
	newPath := filepath.Join(tempDir, AppStorageDirName, AppStorageDBName)
	previousMercelPath := filepath.Join(tempDir, "Mercel", "mercel.sqlite")
	usePathResolvers(t, newPath, filepath.Join(tempDir, "Statistic", "statistic.sqlite"), previousMercelPath)

	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		t.Fatalf("MkdirAll new path error = %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(previousMercelPath), 0o755); err != nil {
		t.Fatalf("MkdirAll previous Mercel path error = %v", err)
	}

	db, err := sql.Open("sqlite", previousMercelPath)
	if err != nil {
		t.Fatalf("sql.Open() previous Mercel db error = %v", err)
	}
	queries := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, role TEXT NOT NULL, created_at TEXT NOT NULL);`,
		`CREATE TABLE services (id INTEGER PRIMARY KEY AUTOINCREMENT, code TEXT NOT NULL UNIQUE, name TEXT NOT NULL, unit TEXT NOT NULL, rate INTEGER NOT NULL, category TEXT NOT NULL, allocation_percent REAL, created_by TEXT NOT NULL DEFAULT 'admin', created_at TEXT NOT NULL);`,
		`CREATE TABLE calculations (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, target_amount INTEGER NOT NULL, total_amount INTEGER NOT NULL, items_json TEXT NOT NULL, created_at TEXT NOT NULL, created_by TEXT NOT NULL DEFAULT '', created_role TEXT NOT NULL DEFAULT '');`,
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("create previous Mercel schema: %v", err)
		}
	}
	if _, err := db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", HashPasswordForTest(legacyFixturePassword), RoleAdmin, "2026-03-27T00:00:00Z"); err != nil {
		t.Fatalf("insert admin: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "tech_user", HashPasswordForTest("secret"), RoleTechnicalEmployee, "2026-03-27T00:01:00Z"); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO calculations(title, target_amount, total_amount, items_json, created_at, created_by, created_role) VALUES(?, ?, ?, ?, ?, ?, ?)`, "previous mercel calc", 40000, 40000, "[]", "2026-03-27T00:02:00Z", "tech_user", RoleTechnicalEmployee); err != nil {
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
		t.Fatalf("expected stable storage database to be created from previous Mercel path: %v", err)
	}
	loginAsAdmin(t, app)
	users, err := app.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	foundTechUser := false
	for _, user := range users {
		if user.Username == "tech_user" {
			foundTechUser = true
			break
		}
	}
	if !foundTechUser {
		t.Fatalf("expected tech_user to be recovered from previous Mercel path, got %+v", users)
	}
	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	if len(calculations) != 1 || calculations[0].Title != "previous mercel calc" {
		t.Fatalf("expected recovered archive from previous Mercel db, got %+v", calculations)
	}
}

// EN: Test `TestBuildServiceTargetsDistributesDefaultsEvenly`.
//
// EN: What it does: TestBuildServiceTargetsDistributesDefaultsEvenly checks default intra-group percentage distribution.
//
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
	targets := BuildServiceTargetsForTest(byCategory, map[string]int{
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

	if AbsIntForTest(targets["s1"]-targets["s2"]) > 1 || AbsIntForTest(targets["s2"]-targets["s3"]) > 1 {
		t.Fatalf("primary group should be evenly distributed, got %v", []int{targets["s1"], targets["s2"], targets["s3"]})
	}
	if AbsIntForTest(targets["s4"]-targets["s5"]) > 1 || AbsIntForTest(targets["s5"]-targets["s6"]) > 1 {
		t.Fatalf("secondary group should be evenly distributed, got %v", []int{targets["s4"], targets["s5"], targets["s6"]})
	}
	if AbsIntForTest(targets["s7"]-targets["s8"]) > 1 {
		t.Fatalf("closing group should be evenly distributed, got %v", []int{targets["s7"], targets["s8"]})
	}
}

// EN: Test `TestBuildServiceTargetsWeightAddsShareWithinGroup`.
//
// EN: What it does: TestBuildServiceTargetsWeightAddsShareWithinGroup proves that per-service weight shifts share within a group.
//
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
func TestBuildServiceTargetsWeightAddsShareWithinGroup(t *testing.T) {
	services := []Service{
		{Code: "s1", Category: CategoryPrimary},
		{Code: "s2", Category: CategoryPrimary},
		{Code: "s3", Category: CategoryPrimary},
	}

	targets := BuildServiceTargetsForTest(map[string][]Service{
		CategoryPrimary: services,
	}, map[string]int{
		CategoryPrimary: 79000,
	}, map[string]int{
		"s1": 1,
	})

	if targets["s1"] <= targets["s2"] {
		t.Fatalf("expected weighted service to receive a larger share, got s1=%d s2=%d s3=%d", targets["s1"], targets["s2"], targets["s3"])
	}
	if AbsIntForTest(targets["s2"]-targets["s3"]) > 1 {
		t.Fatalf("expected remaining services to stay balanced, got s2=%d s3=%d", targets["s2"], targets["s3"])
	}
	if targets["s1"]+targets["s2"]+targets["s3"] != 79000 {
		t.Fatalf("expected total to stay unchanged, got %d", targets["s1"]+targets["s2"]+targets["s3"])
	}
}

func TestBuildServiceTargetsNegativeWeightRemovesShareWithinGroup(t *testing.T) {
	services := []Service{
		{Code: "s1", Category: CategoryPrimary},
		{Code: "s2", Category: CategoryPrimary},
		{Code: "s3", Category: CategoryPrimary},
	}

	targets := BuildServiceTargetsForTest(map[string][]Service{
		CategoryPrimary: services,
	}, map[string]int{
		CategoryPrimary: 79000,
	}, map[string]int{
		"s1": -1,
	})

	if targets["s1"] >= targets["s2"] {
		t.Fatalf("expected negatively weighted service to receive a smaller share, got s1=%d s2=%d s3=%d", targets["s1"], targets["s2"], targets["s3"])
	}
	if AbsIntForTest(targets["s2"]-targets["s3"]) > 1 {
		t.Fatalf("expected remaining services to stay balanced, got s2=%d s3=%d", targets["s2"], targets["s3"])
	}
	if targets["s1"]+targets["s2"]+targets["s3"] != 79000 {
		t.Fatalf("expected total to stay unchanged, got %d", targets["s1"]+targets["s2"]+targets["s3"])
	}
}

// EN: Test `TestSeniorSpecialistCanDeleteEmployeeOnly`.
//
// EN: What it does: TestSeniorSpecialistCanDeleteEmployeeOnly enforces deletion boundaries for the senior specialist role.
//
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
