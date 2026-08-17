package appcore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func NewApp() (*App, error) {
	dbPath, err := ensureDatabasePath()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	app := &App{db: db, dbPath: dbPath}
	if err := app.initDatabase(); err != nil {
		db.Close()
		return nil, err
	}
	if err := app.seedDefaultData(); err != nil {
		db.Close()
		return nil, err
	}
	return app, nil
}

// EN: Variable `resolveDatabasePath`.
//
// EN: What it does: resolveDatabasePath returns the canonical SQLite path for the current Mercel installation.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
var resolveDatabasePath = func() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	appDir := filepath.Join(baseDir, AppStorageDirName)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", fmt.Errorf("create app dir: %w", err)
	}
	return filepath.Join(appDir, AppStorageDBName), nil
}

// EN: Variable `resolveLegacyDatabasePath`.
//
// EN: What it does: resolveLegacyDatabasePath points to the pre-rename Statistic database so data can be recovered automatically.
//
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
var resolveLegacyDatabasePath = func() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(baseDir, "Statistic", "statistic.sqlite"), nil
}

var resolvePreviousMercelDatabasePath = func() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(baseDir, "Mercel", "mercel.sqlite"), nil
}

// EN: Function `ensureDatabasePath`.
//
// EN: What it does: ensureDatabasePath decides which database file should be used and performs legacy recovery when needed.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func ensureDatabasePath() (string, error) {
	dbPath, err := resolveDatabasePath()
	if err != nil {
		return "", err
	}
	legacyPath, err := resolveLegacyDatabasePath()
	if err != nil {
		return "", err
	}
	previousMercelPath, err := resolvePreviousMercelDatabasePath()
	if err != nil {
		return "", err
	}

	legacyCandidates := make([]string, 0, 2)
	for _, candidate := range []string{previousMercelPath, legacyPath} {
		if candidate == "" || candidate == dbPath {
			continue
		}
		duplicate := false
		for _, existing := range legacyCandidates {
			if existing == candidate {
				duplicate = true
				break
			}
		}
		if !duplicate {
			legacyCandidates = append(legacyCandidates, candidate)
		}
	}

	dbExists := false
	if _, err := os.Stat(dbPath); err == nil {
		dbExists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat database: %w", err)
	}

	existingLegacyCandidates := make([]string, 0, len(legacyCandidates))
	for _, candidate := range legacyCandidates {
		if _, err := os.Stat(candidate); err == nil {
			existingLegacyCandidates = append(existingLegacyCandidates, candidate)
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat legacy database: %w", err)
		}
	}

	if !dbExists {
		if len(existingLegacyCandidates) > 0 {
			if err := copyFile(existingLegacyCandidates[0], dbPath); err != nil {
				return "", fmt.Errorf("copy legacy database: %w", err)
			}
		}
		return dbPath, nil
	}

	for _, candidate := range existingLegacyCandidates {
		shouldRecover, err := shouldRecoverFromLegacy(dbPath, candidate)
		if err != nil {
			return "", fmt.Errorf("compare legacy database: %w", err)
		}
		if shouldRecover {
			if err := copyFile(candidate, dbPath); err != nil {
				return "", fmt.Errorf("restore legacy database: %w", err)
			}
			break
		}
	}
	return dbPath, nil
}

// EN: Function `shouldRecoverFromLegacy`.
//
// EN: What it does: shouldRecoverFromLegacy compares the new and legacy databases to detect when the new DB is only a fresh shell.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func shouldRecoverFromLegacy(targetPath string, legacyPath string) (bool, error) {
	targetUsers, targetCalcs, targetServices, err := databaseCounts(targetPath)
	if err != nil {
		return false, err
	}
	legacyUsers, legacyCalcs, legacyServices, err := databaseCounts(legacyPath)
	if err != nil {
		return false, err
	}

	targetLooksFresh := targetUsers <= 1 && targetCalcs == 0 && targetServices <= 8
	legacyHasMoreData := legacyUsers > targetUsers || legacyCalcs > targetCalcs || legacyServices > targetServices
	return targetLooksFresh && legacyHasMoreData, nil
}

// EN: Function `databaseCounts`.
//
// EN: What it does: databaseCounts opens a database file read-only enough for diagnostics and reports key table sizes.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func databaseCounts(path string) (users int, calculations int, services int, err error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return 0, 0, 0, err
	}
	defer db.Close()

	users, err = countTableRows(db, "users")
	if err != nil {
		return 0, 0, 0, err
	}
	calculations, err = countTableRows(db, "calculations")
	if err != nil {
		return 0, 0, 0, err
	}
	services, err = countTableRows(db, "services")
	if err != nil {
		return 0, 0, 0, err
	}
	return users, calculations, services, nil
}

// EN: Function `countTableRows`.
//
// EN: What it does: countTableRows is a small helper that counts rows in one table during migration and recovery checks.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func countTableRows(db *sql.DB, table string) (int, error) {
	var exists int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&exists); err != nil {
		return 0, err
	}
	if exists == 0 {
		return 0, nil
	}
	var count int
	if !isSafeSQLiteIdentifier(table) {
		return 0, fmt.Errorf("unsafe table identifier: %s", table)
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// EN: Function `copyFile`.
//
// EN: What it does: copyFile copies a database file byte-for-byte into a destination path, creating parent folders first.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func copyFile(src string, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	tempPath := dst + ".tmp"
	targetFile, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		targetFile.Close()
		_ = os.Remove(tempPath)
		return err
	}
	if err := targetFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, dst); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

// EN: Method `startup`.
//
// EN: What it does: startup stores the Wails startup context so backend methods can interact with the runtime if needed later.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// EN: Method `Close`.
//
// EN: What it does: Close releases persistent resources such as the SQLite connection when the app shuts down.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// EN: Method `initDatabase`.
//
// EN: What it does: initDatabase creates the schema, runs lightweight migrations and seeds baseline application data.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) initDatabase() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL DEFAULT '',
			last_act_number INTEGER NOT NULL DEFAULT 1,
			preferred_contract_code TEXT NOT NULL DEFAULT '1',
			contract_spbks_number TEXT NOT NULL DEFAULT '',
			contract_grizabl_number TEXT NOT NULL DEFAULT '',
			contract_signed_at TEXT NOT NULL DEFAULT '',
			act_template TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS services (
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
		`CREATE TABLE IF NOT EXISTS calculations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			target_amount INTEGER NOT NULL,
			total_amount INTEGER NOT NULL,
			items_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			created_by TEXT NOT NULL DEFAULT '',
			created_role TEXT NOT NULL DEFAULT ''
		);`,
	}
	for _, query := range queries {
		if _, err := a.db.Exec(query); err != nil {
			return fmt.Errorf("init database: %w", err)
		}
	}
	if err := a.migrateDatabase(); err != nil {
		return err
	}
	if err := a.ensureIndexes(); err != nil {
		return err
	}
	return nil
}

// EN: Method `ensureIndexes`.
//
// EN: What it does: ensureIndexes creates the lookup indexes the archive and service reads depend on.
//
// EN: Key points: it runs after migrateDatabase, not alongside the table DDL, because it indexes columns that older
// EN: databases only gain during migration — indexing created_role before the migration adds it fails outright.
// EN: CREATE INDEX IF NOT EXISTS is idempotent, so this is safe to repeat on every startup.
func (a *App) ensureIndexes() error {
	// The archive is read through archiveVisibilityFilter, which narrows rows by
	// created_by and created_role; services are always looked up by their owner.
	queries := []string{
		`CREATE INDEX IF NOT EXISTS idx_calculations_created_by_created_at
			ON calculations (created_by, created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_calculations_created_role
			ON calculations (created_role);`,
		`CREATE INDEX IF NOT EXISTS idx_services_created_by
			ON services (created_by);`,
	}
	for _, query := range queries {
		if _, err := a.db.Exec(query); err != nil {
			return fmt.Errorf("ensure indexes: %w", err)
		}
	}
	return nil
}

// EN: Method `migrateDatabase`.
//
// EN: What it does: migrateDatabase upgrades older SQLite files by adding missing columns required by newer builds.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) migrateDatabase() error {
	if err := ensureColumnExists(a.db, "users", "full_name", `ALTER TABLE users ADD COLUMN full_name TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate users.full_name: %w", err)
	}
	if err := ensureColumnExists(a.db, "users", "last_act_number", `ALTER TABLE users ADD COLUMN last_act_number INTEGER NOT NULL DEFAULT 1`); err != nil {
		return fmt.Errorf("migrate users.last_act_number: %w", err)
	}
	if err := ensureColumnExists(a.db, "users", "preferred_contract_code", `ALTER TABLE users ADD COLUMN preferred_contract_code TEXT NOT NULL DEFAULT '1'`); err != nil {
		return fmt.Errorf("migrate users.preferred_contract_code: %w", err)
	}
	if err := ensureColumnExists(a.db, "users", "contract_spbks_number", `ALTER TABLE users ADD COLUMN contract_spbks_number TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate users.contract_spbks_number: %w", err)
	}
	if err := ensureColumnExists(a.db, "users", "contract_grizabl_number", `ALTER TABLE users ADD COLUMN contract_grizabl_number TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate users.contract_grizabl_number: %w", err)
	}
	if err := ensureColumnExists(a.db, "users", "contract_signed_at", `ALTER TABLE users ADD COLUMN contract_signed_at TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate users.contract_signed_at: %w", err)
	}
	if err := ensureColumnExists(a.db, "users", "act_template", `ALTER TABLE users ADD COLUMN act_template TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate users.act_template: %w", err)
	}
	if err := ensureColumnExists(a.db, "calculations", "created_by", `ALTER TABLE calculations ADD COLUMN created_by TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate calculations.created_by: %w", err)
	}
	if err := ensureColumnExists(a.db, "calculations", "created_role", `ALTER TABLE calculations ADD COLUMN created_role TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate calculations.created_role: %w", err)
	}
	if err := ensureColumnExists(a.db, "services", "created_by", `ALTER TABLE services ADD COLUMN created_by TEXT NOT NULL DEFAULT 'admin'`); err != nil {
		return fmt.Errorf("migrate services.created_by: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE services SET created_by = 'admin' WHERE created_by = '' OR created_by IS NULL`); err != nil {
		return fmt.Errorf("backfill services.created_by: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'manager'`, RoleSupportManager); err != nil {
		return fmt.Errorf("normalize users.manager role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'senior_specialist'`, RoleSupportSeniorSpecialist); err != nil {
		return fmt.Errorf("normalize users.senior_specialist role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'employee'`, RoleSupportEmployee); err != nil {
		return fmt.Errorf("normalize users.employee role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = COALESCE((SELECT role FROM users WHERE username = calculations.created_by), created_role, '') WHERE created_role = '' OR created_role IS NULL`); err != nil {
		return fmt.Errorf("backfill calculations.created_role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'manager'`, RoleSupportManager); err != nil {
		return fmt.Errorf("normalize calculations.manager role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'senior_specialist'`, RoleSupportSeniorSpecialist); err != nil {
		return fmt.Errorf("normalize calculations.senior_specialist role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'employee'`, RoleSupportEmployee); err != nil {
		return fmt.Errorf("normalize calculations.employee role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'support_manager'`, RoleSupportHead); err != nil {
		return fmt.Errorf("normalize users.support_manager role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'support_senior_specialist'`, RoleSupportSenior); err != nil {
		return fmt.Errorf("normalize users.support_senior_specialist role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'tech_manager'`, RoleTechnicalHead); err != nil {
		return fmt.Errorf("normalize users.tech_manager role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'senior_technician'`, RoleTechnicalSenior); err != nil {
		return fmt.Errorf("normalize users.senior_technician role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'technician'`, RoleTechnicalEmployee); err != nil {
		return fmt.Errorf("normalize users.technician role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'mrk_manager'`, RoleCommercialSubscriberHead); err != nil {
		return fmt.Errorf("normalize users.mrk_manager role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'senior_mrk'`, RoleCommercialSeniorMRK); err != nil {
		return fmt.Errorf("normalize users.senior_mrk role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET role = ? WHERE role = 'mrk_employee'`, RoleCommercialEmployeeMRK); err != nil {
		return fmt.Errorf("normalize users.mrk_employee role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET last_act_number = 1 WHERE last_act_number IS NULL OR last_act_number <= 0`); err != nil {
		return fmt.Errorf("backfill users.last_act_number: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE users SET preferred_contract_code = '1' WHERE preferred_contract_code IS NULL OR TRIM(preferred_contract_code) = '' OR preferred_contract_code NOT IN ('1', '2')`); err != nil {
		return fmt.Errorf("backfill users.preferred_contract_code: %w", err)
	}
	for legacy, current := range legacyActTemplateIDs {
		if _, err := a.db.Exec(`UPDATE users SET act_template = ? WHERE act_template = ?`, current, legacy); err != nil {
			return fmt.Errorf("normalize users.act_template %s: %w", legacy, err)
		}
	}
	if err := a.backfillActTemplates(); err != nil {
		return fmt.Errorf("backfill users.act_template: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'support_manager'`, RoleSupportHead); err != nil {
		return fmt.Errorf("normalize calculations.support_manager role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'support_senior_specialist'`, RoleSupportSenior); err != nil {
		return fmt.Errorf("normalize calculations.support_senior_specialist role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'tech_manager'`, RoleTechnicalHead); err != nil {
		return fmt.Errorf("normalize calculations.tech_manager role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'senior_technician'`, RoleTechnicalSenior); err != nil {
		return fmt.Errorf("normalize calculations.senior_technician role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'technician'`, RoleTechnicalEmployee); err != nil {
		return fmt.Errorf("normalize calculations.technician role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'mrk_manager'`, RoleCommercialSubscriberHead); err != nil {
		return fmt.Errorf("normalize calculations.mrk_manager role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'senior_mrk'`, RoleCommercialSeniorMRK); err != nil {
		return fmt.Errorf("normalize calculations.senior_mrk role: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET created_role = ? WHERE created_role = 'mrk_employee'`, RoleCommercialEmployeeMRK); err != nil {
		return fmt.Errorf("normalize calculations.mrk_employee role: %w", err)
	}
	return nil
}
func ensureColumnExists(db *sql.DB, table string, column string, alterSQL string) error {
	if !isSafeSQLiteIdentifier(table) {
		return fmt.Errorf("unsafe table identifier: %s", table)
	}
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultV sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultV, &primaryKey); err != nil {
			return err
		}
		if strings.EqualFold(name, column) {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = db.Exec(alterSQL)
	return err
}

// EN: Method `seedDefaultData`.
//
// EN: What it does: seedDefaultData inserts baseline records that every installation expects to exist after startup.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) seedDefaultData() error {
	if err := a.seedAdmin(); err != nil {
		return err
	}
	if err := a.removeDeprecatedAdmin1(); err != nil {
		return err
	}
	return a.seedServices()
}

// AdminPasswordEnvVar is the environment variable an operator can set to override
// the built-in administrator password before the app starts. When it is empty,
// the precomputed defaultAdminPasswordHash is used instead.
const AdminPasswordEnvVar = "MERCEL_ADMIN_PASSWORD"

// defaultAdminPasswordHash is the bcrypt hash of the built-in administrator
// password. The plaintext is deliberately NOT stored in source: even with full
// access to this repository the password cannot be recovered from this hash.
// To rotate it, set AdminPasswordEnvVar at runtime, or regenerate this hash from
// a new secret with bcrypt.GenerateFromPassword.
const defaultAdminPasswordHash = "$2a$10$ljd4EP7tOwTKc40TO3X2sOwHMQ1R3USKVLl9zfoonZXNl3kDzrIl."

// adminSeedPasswordHash returns the bcrypt hash that should be applied to the
// protected admin account. An operator-supplied secret takes precedence over the
// built-in default so each deployment can use its own credentials.
func adminSeedPasswordHash() string {
	if override := stripSpaces(os.Getenv(AdminPasswordEnvVar)); override != "" {
		return hashPassword(override)
	}
	return defaultAdminPasswordHash
}

// EN: Method `seedAdmin`.
//
// EN: What it does: seedAdmin guarantees the protected admin account exists with the expected credentials and role.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) seedAdmin() error {
	passwordHash := adminSeedPasswordHash()
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = 'admin'`).Scan(&count); err != nil {
		return fmt.Errorf("check admin user: %w", err)
	}
	if count == 0 {
		_, err := a.db.Exec(`INSERT INTO users(username, password_hash, full_name, last_act_number, act_template, role, created_at) VALUES(?, ?, ?, ?, ?, ?, ?)`, "admin", passwordHash, "", 1, a.nextActTemplate(), RoleAdmin, time.Now().Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("seed admin user: %w", err)
		}
		return nil
	}
	_, err := a.db.Exec(`UPDATE users SET password_hash = ?, role = ? WHERE username = ?`, passwordHash, RoleAdmin, "admin")
	if err != nil {
		return fmt.Errorf("restore admin user: %w", err)
	}
	return nil
}

// EN: Method `seedServices`.
//
// EN: What it does: seedServices populates default admin-owned services for first start and legacy empty databases.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) removeDeprecatedAdmin1() error {
	if _, err := a.db.Exec(`DELETE FROM services WHERE created_by = ?`, "admin1"); err != nil {
		return fmt.Errorf("delete admin1 services: %w", err)
	}
	if _, err := a.db.Exec(`DELETE FROM users WHERE username = ?`, "admin1"); err != nil {
		return fmt.Errorf("delete admin1 user: %w", err)
	}
	return nil
}

func (a *App) seedServices() error {
	defaults := []UpsertServiceRequest{
		{Name: "Удалённая техническая поддержка клиентов", Unit: "ч.", Rate: 350, Category: CategoryPrimary},
		{Name: "Удалённый мониторинг сети", Unit: "ч.", Rate: 200, Category: CategoryPrimary},
		{Name: "Диагностика и устранение внештатных проблем коммутационного оборудования", Unit: "ч.", Rate: 100, Category: CategoryPrimary},
		{Name: "Удалённая настройка сетевого оборудования (Eltex MES2124M)", Unit: "шт.", Rate: 505, Category: CategorySecondary},
		{Name: "Удалённая настройка сетевого оборудования (TP-Link TL-SG3428X)", Unit: "шт.", Rate: 1015, Category: CategorySecondary},
		{Name: "Настройка и обслуживание сетевого оборудования (Linksys SPS 224G4)", Unit: "шт.", Rate: 1010, Category: CategorySecondary},
		{Name: "Настройка VLAN по запросу", Unit: "шт.", Rate: 15, Category: CategoryClosing},
		{Name: "Изменение описания порта на оборудовании", Unit: "шт.", Rate: 12, Category: CategoryClosing},
	}

	for _, owner := range []string{"admin"} {
		if err := a.seedDefaultServicesForOwner(owner, defaults); err != nil {
			return err
		}
	}
	return nil
}
func (a *App) seedDefaultServicesForOwner(owner string, defaults []UpsertServiceRequest) error {
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM services WHERE created_by = ?`, owner).Scan(&count); err != nil {
		return fmt.Errorf("count services for %s: %w", owner, err)
	}
	if count > 0 {
		return nil
	}
	for _, service := range defaults {
		if _, err := a.saveServiceForOwner(service, owner); err != nil {
			return fmt.Errorf("seed default services for %s: %w", owner, err)
		}
	}
	return nil
}

// EN: Function `hashPassword`.
//
// EN: What it does: hashPassword performs a simple deterministic password hash used by this local desktop application.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
