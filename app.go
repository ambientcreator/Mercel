package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// RU: Набор констант приложения.
// EN: Application constant group.
//
// RU: Что делает: задаёт общие константы приложения, от которых зависят роли, категории услуг и базовые правила логики.
// EN: What it does: Application-wide constants define supported roles and service categories used by permissions and calculation logic.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
const (
	CategoryPrimary   = "primary"
	CategorySecondary = "secondary"
	CategoryClosing   = "closing"

	RoleAdmin            = "admin"
	RoleManager          = "manager"
	RoleSeniorSpecialist = "senior_specialist"
	RoleEmployee         = "employee"
)

// RU: Тип данных `Service`.
// EN: Data type `Service`.
//
// RU: Что делает: описывает структуру данных `Service`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: Service describes one billable position available to the current user: pricing, unit, ownership and allocation settings.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type Service struct {
	ID                int64    `json:"id"`
	Code              string   `json:"code"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Description       string   `json:"description"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent,omitempty"`
	CreatedBy         string   `json:"createdBy,omitempty"`
	CreatedAt         string   `json:"createdAt,omitempty"`
}

// RU: Тип данных `UpsertServiceRequest`.
// EN: Data type `UpsertServiceRequest`.
//
// RU: Что делает: описывает структуру данных `UpsertServiceRequest`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: UpsertServiceRequest is the payload used by the frontend when a user creates or edits a service definition.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type UpsertServiceRequest struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent"`
}

// RU: Тип данных `User`.
// EN: Data type `User`.
//
// RU: Что делает: описывает структуру данных `User`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: User stores public account information that can be safely returned to the frontend without password data.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

// RU: Тип данных `UserWithPassword`.
// EN: Data type `UserWithPassword`.
//
// RU: Что делает: описывает структуру данных `UserWithPassword`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: UserWithPassword extends a user payload with a plain-text password for creation workflows only.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type UserWithPassword struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// RU: Тип данных `LoginRequest`.
// EN: Data type `LoginRequest`.
//
// RU: Что делает: описывает структуру данных `LoginRequest`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: LoginRequest carries credentials from the login form to the backend session logic.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RU: Тип данных `SessionState`.
// EN: Data type `SessionState`.
//
// RU: Что делает: описывает структуру данных `SessionState`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: SessionState is the frontend-facing snapshot of the current authenticated user and their capabilities.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SessionState struct {
	Authenticated bool   `json:"authenticated"`
	User          *User  `json:"user,omitempty"`
	CanManage     bool   `json:"canManage"`
	CanAdmin      bool   `json:"canAdmin"`
	CanModerate   bool   `json:"canModerate"`
	Message       string `json:"message,omitempty"`
}

// RU: Тип данных `CalculationRequest`.
// EN: Data type `CalculationRequest`.
//
// RU: Что делает: описывает структуру данных `CalculationRequest`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: CalculationRequest contains the target amount and per-service weights used to build a calculation.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationRequest struct {
	TargetAmount int            `json:"targetAmount"`
	Weights      map[string]int `json:"weights"`
}

// RU: Тип данных `CalculationItem`.
// EN: Data type `CalculationItem`.
//
// RU: Что делает: описывает структуру данных `CalculationItem`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: CalculationItem is one line of the generated or archived calculation with both business and UI-oriented fields.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationItem struct {
	ServiceID         int64    `json:"serviceId"`
	ServiceCode       string   `json:"serviceCode"`
	Name              string   `json:"name"`
	Unit              string   `json:"unit"`
	Rate              int      `json:"rate"`
	Quantity          int      `json:"quantity"`
	LineTotal         int      `json:"lineTotal"`
	Description       string   `json:"description"`
	Weight            int      `json:"weight"`
	Category          string   `json:"category"`
	AllocationPercent *float64 `json:"allocationPercent,omitempty"`
}

// RU: Тип данных `CalculationResult`.
// EN: Data type `CalculationResult`.
//
// RU: Что делает: описывает структуру данных `CalculationResult`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: CalculationResult is the full response returned after a calculation attempt, including totals and metadata.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type CalculationResult struct {
	TargetAmount   int               `json:"targetAmount"`
	TotalAmount    int               `json:"totalAmount"`
	Items          []CalculationItem `json:"items"`
	FoundExact     bool              `json:"foundExact"`
	GeneratedAt    string            `json:"generatedAt"`
	ActiveServices int               `json:"activeServices"`
	Weights        map[string]int    `json:"weights"`
}

// RU: Тип данных `SaveCalculationRequest`.
// EN: Data type `SaveCalculationRequest`.
//
// RU: Что делает: описывает структуру данных `SaveCalculationRequest`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: SaveCalculationRequest is sent when the frontend persists the current calculation into the archive.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SaveCalculationRequest struct {
	Title        string            `json:"title"`
	TargetAmount int               `json:"targetAmount"`
	Items        []CalculationItem `json:"items"`
}

// RU: Тип данных `SavedCalculation`.
// EN: Data type `SavedCalculation`.
//
// RU: Что делает: описывает структуру данных `SavedCalculation`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: SavedCalculation represents one archived calculation entry as stored in SQLite and shown in history.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type SavedCalculation struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	TargetAmount int               `json:"targetAmount"`
	TotalAmount  int               `json:"totalAmount"`
	Items        []CalculationItem `json:"items"`
	CreatedAt    string            `json:"createdAt"`
	CreatedBy    string            `json:"createdBy"`
}

// RU: Тип данных `AppBootstrap`.
// EN: Data type `AppBootstrap`.
//
// RU: Что делает: описывает структуру данных `AppBootstrap`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: AppBootstrap aggregates all initial data the frontend needs after startup or refresh.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type AppBootstrap struct {
	Session             SessionState       `json:"session"`
	Services            []Service          `json:"services"`
	Users               []User             `json:"users"`
	SavedCalculations   []SavedCalculation `json:"savedCalculations"`
	DefaultGroupPercent map[string]float64 `json:"defaultGroupPercent"`
}

// RU: Тип данных `allocationState`.
// EN: Data type `allocationState`.
//
// RU: Что делает: описывает структуру данных `allocationState`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: allocationState is an internal DP cell used while searching for a good quantity distribution.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type allocationState struct {
	distance int
	score    int
	count    int
	prev     int
	idx      int
	ok       bool
}

// RU: Тип данных `groupAllocation`.
// EN: Data type `groupAllocation`.
//
// RU: Что делает: описывает структуру данных `groupAllocation`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: groupAllocation stores the best per-group exact allocation candidate found during structured solving.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type groupAllocation struct {
	amount     int
	score      int
	quantities []int
	ok         bool
}

// RU: Тип данных `App`.
// EN: Data type `App`.
//
// RU: Что делает: описывает структуру данных `App`, которая участвует в бизнес-логике, API или тестах.
// EN: What it does: App owns application state, the database handle and the current in-memory session.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
type App struct {
	ctx            context.Context
	db             *sql.DB
	mu             sync.RWMutex
	currentSession *User
}

// RU: Функция `NewApp`.
// EN: Function `NewApp`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: NewApp prepares the application object, resolves the working database file and initializes schema/data.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func NewApp() (*App, error) {
	dbPath, err := ensureDatabasePath()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	app := &App{db: db}
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

// RU: Переменная `resolveDatabasePath`.
// EN: Variable `resolveDatabasePath`.
//
// RU: Что делает: хранит ресурсы или глобальное состояние, которое нужно другим частям программы.
// EN: What it does: resolveDatabasePath returns the canonical SQLite path for the current Mercel installation.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
var resolveDatabasePath = func() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	appDir := filepath.Join(baseDir, "Mercel")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", fmt.Errorf("create app dir: %w", err)
	}
	return filepath.Join(appDir, "mercel.sqlite"), nil
}

// RU: Переменная `resolveLegacyDatabasePath`.
// EN: Variable `resolveLegacyDatabasePath`.
//
// RU: Что делает: хранит ресурсы или глобальное состояние, которое нужно другим частям программы.
// EN: What it does: resolveLegacyDatabasePath points to the pre-rename Statistic database so data can be recovered automatically.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
var resolveLegacyDatabasePath = func() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(baseDir, "Statistic", "statistic.sqlite"), nil
}

// RU: Функция `ensureDatabasePath`.
// EN: Function `ensureDatabasePath`.
//
// RU: Что делает: помогает работать с файлами, SQLite и миграциями.
// EN: What it does: ensureDatabasePath decides which database file should be used and performs legacy recovery when needed.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
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
	if legacyPath == dbPath {
		return dbPath, nil
	}

	dbExists := false
	if _, err := os.Stat(dbPath); err == nil {
		dbExists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat database: %w", err)
	}

	legacyExists := false
	if _, err := os.Stat(legacyPath); err == nil {
		legacyExists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat legacy database: %w", err)
	}

	if !dbExists {
		if legacyExists {
			if err := copyFile(legacyPath, dbPath); err != nil {
				return "", fmt.Errorf("copy legacy database: %w", err)
			}
		}
		return dbPath, nil
	}

	if legacyExists {
		shouldRecover, err := shouldRecoverFromLegacy(dbPath, legacyPath)
		if err != nil {
			return "", fmt.Errorf("compare legacy database: %w", err)
		}
		if shouldRecover {
			if err := copyFile(legacyPath, dbPath); err != nil {
				return "", fmt.Errorf("restore legacy database: %w", err)
			}
		}
	}
	return dbPath, nil
}

// RU: Функция `shouldRecoverFromLegacy`.
// EN: Function `shouldRecoverFromLegacy`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: shouldRecoverFromLegacy compares the new and legacy databases to detect when the new DB is only a fresh shell.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
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

// RU: Функция `databaseCounts`.
// EN: Function `databaseCounts`.
//
// RU: Что делает: помогает работать с файлами, SQLite и миграциями.
// EN: What it does: databaseCounts opens a database file read-only enough for diagnostics and reports key table sizes.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
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

// RU: Функция `countTableRows`.
// EN: Function `countTableRows`.
//
// RU: Что делает: помогает работать с файлами, SQLite и миграциями.
// EN: What it does: countTableRows is a small helper that counts rows in one table during migration and recovery checks.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
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
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// RU: Функция `copyFile`.
// EN: Function `copyFile`.
//
// RU: Что делает: помогает работать с файлами, SQLite и миграциями.
// EN: What it does: copyFile copies a database file byte-for-byte into a destination path, creating parent folders first.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
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

// RU: Метод `startup`.
// EN: Method `startup`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: startup stores the Wails startup context so backend methods can interact with the runtime if needed later.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// RU: Метод `Close`.
// EN: Method `Close`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: Close releases persistent resources such as the SQLite connection when the app shuts down.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// RU: Метод `initDatabase`.
// EN: Method `initDatabase`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: initDatabase creates the schema, runs lightweight migrations and seeds baseline application data.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) initDatabase() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
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
			created_by TEXT NOT NULL DEFAULT ''
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
	return nil
}

// RU: Метод `migrateDatabase`.
// EN: Method `migrateDatabase`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: migrateDatabase upgrades older SQLite files by adding missing columns required by newer builds.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) migrateDatabase() error {
	if err := ensureColumnExists(a.db, "calculations", "created_by", `ALTER TABLE calculations ADD COLUMN created_by TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("migrate calculations.created_by: %w", err)
	}
	if err := ensureColumnExists(a.db, "services", "created_by", `ALTER TABLE services ADD COLUMN created_by TEXT NOT NULL DEFAULT 'admin'`); err != nil {
		return fmt.Errorf("migrate services.created_by: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE services SET created_by = 'admin' WHERE created_by = '' OR created_by IS NULL`); err != nil {
		return fmt.Errorf("backfill services.created_by: %w", err)
	}
	return nil
}

// RU: Функция `ensureColumnExists`.
// EN: Function `ensureColumnExists`.
//
// RU: Что делает: помогает работать с файлами, SQLite и миграциями.
// EN: What it does: ensureColumnExists executes a defensive one-column migration only when the target column is absent.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func ensureColumnExists(db *sql.DB, table string, column string, alterSQL string) error {
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

// RU: Метод `seedDefaultData`.
// EN: Method `seedDefaultData`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: seedDefaultData inserts baseline records that every installation expects to exist after startup.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) seedDefaultData() error {
	if err := a.seedAdmin(); err != nil {
		return err
	}
	if err := a.seedTestAdmin(); err != nil {
		return err
	}
	return a.seedServices()
}

// RU: Метод `seedAdmin`.
// EN: Method `seedAdmin`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: seedAdmin guarantees the protected admin account exists with the expected credentials and role.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) seedAdmin() error {
	var (
		count int
		role  string
	)
	if err := a.db.QueryRow(`SELECT COUNT(*), COALESCE(MAX(role), '') FROM users WHERE username = 'admin'`).Scan(&count, &role); err != nil {
		return fmt.Errorf("check admin user: %w", err)
	}
	if count == 0 {
		_, err := a.db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin", hashPassword("#@7pcehQCSpR"), RoleAdmin, time.Now().Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("seed admin user: %w", err)
		}
		return nil
	}
	_, err := a.db.Exec(`UPDATE users SET password_hash = ?, role = ? WHERE username = ?`, hashPassword("#@7pcehQCSpR"), RoleAdmin, "admin")
	if err != nil {
		return fmt.Errorf("restore admin user: %w", err)
	}
	return nil
}

// RU: Метод `seedServices`.
// EN: Method `seedServices`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: seedServices populates default admin-owned services for first start and legacy empty databases.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
// RU: ????? `seedTestAdmin`.
// EN: Method `seedTestAdmin`.
//
// RU: ??? ??????: ???????????? ??????? ??????? ???????? admin-?????? ??? ???????? ????????? ?????????.
// EN: What it does: seedTestAdmin guarantees that a regular admin-role account for UI testing exists with known credentials.
//
// RU: ???????? ???????: ??? ?????? ?? ???????? ??????????; ?? ????? ? ??????? ? ????? ???????; ?????????? ???????? ?????? username `admin`.
// EN: Key points: this account is not protected; it remains visible in lists and deletable; only the username `admin` stays protected.
func (a *App) seedTestAdmin() error {
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = 'admin1'`).Scan(&count); err != nil {
		return fmt.Errorf("check test admin user: %w", err)
	}
	if count == 0 {
		_, err := a.db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, "admin1", hashPassword("admin1"), RoleAdmin, time.Now().Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("seed test admin user: %w", err)
		}
		return nil
	}
	_, err := a.db.Exec(`UPDATE users SET password_hash = ?, role = ? WHERE username = ?`, hashPassword("admin1"), RoleAdmin, "admin1")
	if err != nil {
		return fmt.Errorf("restore test admin user: %w", err)
	}
	return nil
}

func (a *App) seedServices() error {
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM services WHERE created_by = 'admin'`).Scan(&count); err != nil {
		return fmt.Errorf("count services: %w", err)
	}
	if count > 0 {
		return nil
	}

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

	for _, service := range defaults {
		if _, err := a.saveServiceForOwner(service, "admin"); err != nil {
			return err
		}
	}
	return nil
}

// RU: Функция `hashPassword`.
// EN: Function `hashPassword`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: hashPassword performs a simple deterministic password hash used by this local desktop application.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

// RU: Функция `stripSpaces`.
// EN: Function `stripSpaces`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: stripSpaces removes whitespace around and inside a string where credentials should not keep spaces.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func stripSpaces(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), "")
}

// RU: Функция `looksLikeMojibake`.
// EN: Function `looksLikeMojibake`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: looksLikeMojibake detects common UTF-8/Windows-1251 corruption markers in stored Russian text.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func looksLikeMojibake(value string) bool {
	patterns := []string{
		"Р ", "РЎ", "СЃ", "вЂ", "Р’В", "В Р", "Ћ", "™",
	}
	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}

// RU: Функция `sanitizeStoredText`.
// EN: Function `sanitizeStoredText`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: sanitizeStoredText normalizes archived titles and labels before returning them to the frontend.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func sanitizeStoredText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if looksLikeMojibake(value) {
		return ""
	}
	return value
}

// RU: Функция `defaultGroupPercent`.
// EN: Function `defaultGroupPercent`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: defaultGroupPercent returns the fallback allocation share for each service category.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func defaultGroupPercent() map[string]float64 {
	return map[string]float64{
		CategoryPrimary:   0.79,
		CategorySecondary: 0.20,
		CategoryClosing:   0.01,
	}
}

// RU: Функция `rolePower`.
// EN: Function `rolePower`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: rolePower maps roles to a comparable numeric hierarchy for authorization checks.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func rolePower(role string) int {
	switch role {
	case RoleAdmin:
		return 4
	case RoleManager:
		return 3
	case RoleSeniorSpecialist:
		return 2
	default:
		return 1
	}
}

// RU: Функция `userSortPriority`.
// EN: Function `userSortPriority`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: userSortPriority defines how users should be ordered in settings lists: managers, seniors, then employees.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func userSortPriority(role string) int {
	switch role {
	case RoleManager:
		return 1
	case RoleSeniorSpecialist:
		return 2
	case RoleEmployee:
		return 3
	default:
		return 9
	}
}

// RU: Функция `normalizeRole`.
// EN: Function `normalizeRole`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: normalizeRole accepts only known stored roles and falls back to employee for invalid data.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func normalizeRole(role string) string {
	switch role {
	case RoleAdmin, RoleManager, RoleSeniorSpecialist, RoleEmployee:
		return role
	default:
		return RoleEmployee
	}
}

// RU: Функция `normalizeAssignableRole`.
// EN: Function `normalizeAssignableRole`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: normalizeAssignableRole narrows role input from the UI to the set assignable to ordinary managed users.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func normalizeAssignableRole(role string) string {
	switch role {
	case RoleManager, RoleSeniorSpecialist, RoleEmployee:
		return role
	default:
		return RoleEmployee
	}
}

// RU: Функция `canCreateUsers`.
// EN: Function `canCreateUsers`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: canCreateUsers answers whether the current role may create subordinate accounts at all.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func canCreateUsers(role string) bool {
	return rolePower(role) >= rolePower(RoleSeniorSpecialist)
}

// RU: Функция `canCreateRole`.
// EN: Function `canCreateRole`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: canCreateRole checks whether one role is allowed to create another role.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func canCreateRole(actorRole string, targetRole string) bool {
	switch actorRole {
	case RoleAdmin:
		return targetRole == RoleEmployee || targetRole == RoleSeniorSpecialist || targetRole == RoleManager
	case RoleManager:
		return targetRole == RoleEmployee || targetRole == RoleSeniorSpecialist
	case RoleSeniorSpecialist:
		return targetRole == RoleEmployee
	default:
		return false
	}
}

// RU: Функция `normalizeCategory`.
// EN: Function `normalizeCategory`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: normalizeCategory sanitizes service category values coming from storage or the frontend.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func normalizeCategory(category string) string {
	switch category {
	case CategoryPrimary, CategorySecondary, CategoryClosing:
		return category
	default:
		return CategoryPrimary
	}
}

// RU: Функция `generateServiceCode`.
// EN: Function `generateServiceCode`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: generateServiceCode builds a stable service code from the service name for weight and mapping logic.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func generateServiceCode(name string) string {
	base := strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "(", "", ")", "", ",", "", ".", "")
	base = replacer.Replace(base)
	builder := strings.Builder{}
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	code := strings.Trim(builder.String(), "-")
	hashSource := sha256.Sum256([]byte(name))
	suffix := hex.EncodeToString(hashSource[:])[:8]
	if code == "" {
		return fmt.Sprintf("service-%s", suffix)
	}
	return fmt.Sprintf("%s-%s", code, suffix)
}

// RU: Функция `normalizeUnit`.
// EN: Function `normalizeUnit`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: normalizeUnit restricts service units to the two UI-supported values: hours and pieces.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func normalizeUnit(unit string) string {
	switch strings.TrimSpace(strings.ToLower(unit)) {
	case "\u0447", "\u0447.", "\u0447\u0430\u0441", "\u0447\u0430\u0441\u044b":
		return "\u0447."
	case "\u0448\u0442", "\u0448\u0442.", "\u0448\u0442\u0443\u043a\u0430", "\u0448\u0442\u0443\u043a\u0438":
		return "\u0448\u0442."
	default:
		return ""
	}
}

// RU: Функция `canViewAllArchives`.
// EN: Function `canViewAllArchives`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: canViewAllArchives indicates whether a role may browse archives beyond their own calculations.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func canViewAllArchives(role string) bool {
	return role == RoleAdmin
}

// RU: Метод `sessionStateLocked`.
// EN: Method `sessionStateLocked`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: sessionStateLocked builds a safe guest-like session snapshot used when no active session is available.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) sessionStateLocked(message string) SessionState {
	state := SessionState{Authenticated: a.currentSession != nil, Message: message}
	if a.currentSession != nil {
		copyUser := *a.currentSession
		state.User = &copyUser
		state.CanManage = rolePower(copyUser.Role) >= rolePower(RoleManager)
		state.CanAdmin = copyUser.Role == RoleAdmin
		state.CanModerate = rolePower(copyUser.Role) >= rolePower(RoleSeniorSpecialist)
	}
	return state
}

// RU: Метод `requireAuth`.
// EN: Method `requireAuth`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: requireAuth returns the current session user or an authorization error when nobody is logged in.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireAuth() (*User, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.currentSession == nil {
		return nil, errors.New("Сначала войдите в систему.")
	}
	copyUser := *a.currentSession
	return &copyUser, nil
}

// RU: Метод `requireManage`.
// EN: Method `requireManage`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: requireManage ensures the caller has management-level rights before continuing.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireManage() (*User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return nil, err
	}
	if rolePower(user.Role) < rolePower(RoleManager) {
		return nil, errors.New("Недостаточно прав для этого действия.")
	}
	return user, nil
}

// RU: Метод `requireAdmin`.
// EN: Method `requireAdmin`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: requireAdmin is the strictest guard and allows only the protected admin account.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireAdmin() (*User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return nil, err
	}
	if user.Role != RoleAdmin {
		return nil, errors.New("Доступно только администратору.")
	}
	return user, nil
}

// RU: Метод `GetBootstrap`.
// EN: Method `GetBootstrap`.
//
// RU: Что делает: читает и возвращает данные для фронтенда или внутренней логики.
// EN: What it does: GetBootstrap returns the initial application payload used to hydrate the frontend state.
//
// RU: Ключевые моменты: используется в отрисовке интерфейса; должен возвращать только разрешённые данные; порядок и фильтрация важны для UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) GetBootstrap() (AppBootstrap, error) {
	services, err := a.GetServices()
	if err != nil {
		return AppBootstrap{}, err
	}
	calculations, err := a.ListCalculations()
	if err != nil {
		return AppBootstrap{}, err
	}
	users, err := a.ListUsers()
	if err != nil {
		return AppBootstrap{}, err
	}

	a.mu.RLock()
	session := a.sessionStateLocked("")
	a.mu.RUnlock()

	return AppBootstrap{
		Session:             session,
		Services:            services,
		Users:               users,
		SavedCalculations:   calculations,
		DefaultGroupPercent: defaultGroupPercent(),
	}, nil
}

// RU: Метод `Login`.
// EN: Method `Login`.
//
// RU: Что делает: управляет сессией и набором прав текущего пользователя.
// EN: What it does: Login validates credentials, opens a session and returns capability flags for the current user.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Login(req LoginRequest) (SessionState, error) {
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return SessionState{}, errors.New("Введите логин и пароль.")
	}

	var user User
	var passwordHash string
	err := a.db.QueryRow(`SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`, username).Scan(&user.ID, &user.Username, &passwordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionState{}, errors.New("Пользователь не найден.")
		}
		return SessionState{}, fmt.Errorf("login query: %w", err)
	}
	if hashPassword(password) != passwordHash {
		return SessionState{}, errors.New("Неверный пароль.")
	}

	a.mu.Lock()
	a.currentSession = &user
	state := a.sessionStateLocked("Вход выполнен успешно.")
	a.mu.Unlock()
	return state, nil
}

// RU: Метод `Logout`.
// EN: Method `Logout`.
//
// RU: Что делает: управляет сессией и набором прав текущего пользователя.
// EN: What it does: Logout clears the in-memory session and returns a locked guest state to the frontend.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Logout() SessionState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.currentSession = nil
	return a.sessionStateLocked("Вы вышли из системы.")
}

// RU: Метод `GetSession`.
// EN: Method `GetSession`.
//
// RU: Что делает: читает и возвращает данные для фронтенда или внутренней логики.
// EN: What it does: GetSession exposes the current session snapshot without reloading the full bootstrap payload.
//
// RU: Ключевые моменты: используется в отрисовке интерфейса; должен возвращать только разрешённые данные; порядок и фильтрация важны для UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) GetSession() SessionState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessionStateLocked("")
}

// RU: Метод `GetServices`.
// EN: Method `GetServices`.
//
// RU: Что делает: читает и возвращает данные для фронтенда или внутренней логики.
// EN: What it does: GetServices returns only the services belonging to the currently authenticated user.
//
// RU: Ключевые моменты: используется в отрисовке интерфейса; должен возвращать только разрешённые данные; порядок и фильтрация важны для UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) GetServices() ([]Service, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []Service{}, nil
	}
	rows, err := a.db.Query(`SELECT id, code, name, unit, rate, category, allocation_percent, created_by, created_at FROM services WHERE created_by = ? ORDER BY id ASC`, user.Username)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	services := make([]Service, 0)
	for rows.Next() {
		var item Service
		var allocation sql.NullFloat64
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Unit, &item.Rate, &item.Category, &allocation, &item.CreatedBy, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		item.Description = fmt.Sprintf("1 %s = %s", strings.TrimSuffix(item.Unit, "."), displayMoney(item.Rate))
		if allocation.Valid {
			value := allocation.Float64
			item.AllocationPercent = &value
		}
		services = append(services, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate services: %w", err)
	}
	return services, nil
}

// RU: Функция `formatMoney`.
// EN: Function `formatMoney`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: formatMoney produces a compact Russian-currency string for generated descriptions and archive titles.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func formatMoney(value int) string {
	return displayMoney(value)
}

// RU: Функция `displayMoney`.
// EN: Function `displayMoney`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: displayMoney is a UI-facing helper that mirrors money formatting where a separate name reads clearer.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func displayMoney(value int) string {
	return fmt.Sprintf("%d р", value)
}

// RU: Метод `saveServiceForOwner`.
// EN: Method `saveServiceForOwner`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: saveServiceForOwner performs the actual insert/update of a service for a concrete user owner.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) saveServiceForOwner(req UpsertServiceRequest, owner string) (Service, error) {
	name := strings.TrimSpace(req.Name)
	unit := normalizeUnit(req.Unit)
	if name == "" || unit == "" || req.Rate <= 0 {
		return Service{}, errors.New("Заполните название, единицу и стоимость услуги.")
	}
	category := normalizeCategory(req.Category)
	var allocation interface{}
	if req.AllocationPercent != nil {
		if *req.AllocationPercent < 0 || *req.AllocationPercent > 100 {
			return Service{}, errors.New("Процент услуги должен быть в диапазоне от 0 до 100.")
		}
		allocation = *req.AllocationPercent
	}
	code := generateServiceCode(owner + "-" + name)
	createdAt := time.Now().Format(time.RFC3339)

	if req.ID == 0 {
		result, err := a.db.Exec(`INSERT INTO services(code, name, unit, rate, category, allocation_percent, created_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, code, name, unit, req.Rate, category, allocation, owner, createdAt)
		if err != nil {
			return Service{}, fmt.Errorf("create service: %w", err)
		}
		id, _ := result.LastInsertId()
		return a.getServiceByIDForOwner(id, owner)
	}

	_, err := a.db.Exec(`UPDATE services SET name = ?, unit = ?, rate = ?, category = ?, allocation_percent = ? WHERE id = ? AND created_by = ?`, name, unit, req.Rate, category, allocation, req.ID, owner)
	if err != nil {
		return Service{}, fmt.Errorf("update service: %w", err)
	}
	return a.getServiceByIDForOwner(req.ID, owner)
}

// RU: Метод `saveService`.
// EN: Method `saveService`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: saveService is a convenience wrapper that persists a service for the currently authenticated user.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) saveService(req UpsertServiceRequest) (Service, error) {
	user, err := a.requireAuth()
	if err != nil {
		return Service{}, err
	}
	return a.saveServiceForOwner(req, user.Username)
}

// RU: Метод `getServiceByIDForOwner`.
// EN: Method `getServiceByIDForOwner`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: getServiceByIDForOwner fetches one service while enforcing ownership boundaries.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) getServiceByIDForOwner(id int64, owner string) (Service, error) {
	var item Service
	var allocation sql.NullFloat64
	err := a.db.QueryRow(`SELECT id, code, name, unit, rate, category, allocation_percent, created_by, created_at FROM services WHERE id = ? AND created_by = ?`, id, owner).Scan(&item.ID, &item.Code, &item.Name, &item.Unit, &item.Rate, &item.Category, &allocation, &item.CreatedBy, &item.CreatedAt)
	if err != nil {
		return Service{}, err
	}
	item.Description = fmt.Sprintf("1 %s = %s", strings.TrimSuffix(item.Unit, "."), displayMoney(item.Rate))
	if allocation.Valid {
		value := allocation.Float64
		item.AllocationPercent = &value
	}
	return item, nil
}

// RU: Метод `UpsertService`.
// EN: Method `UpsertService`.
//
// RU: Что делает: выполняет изменение данных в приложении и проводит бизнес-операцию.
// EN: What it does: UpsertService is the public service create/update API used by the settings screen.
//
// RU: Ключевые моменты: меняет состояние SQLite или сессии; опирается на проверки ролей и владения; ошибки здесь заметны пользователю сразу.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) UpsertService(req UpsertServiceRequest) (Service, error) {
	if _, err := a.requireAuth(); err != nil {
		return Service{}, err
	}
	return a.saveService(req)
}

// RU: Метод `DeleteService`.
// EN: Method `DeleteService`.
//
// RU: Что делает: выполняет изменение данных в приложении и проводит бизнес-операцию.
// EN: What it does: DeleteService removes one owned service after authorization and ownership checks.
//
// RU: Ключевые моменты: меняет состояние SQLite или сессии; опирается на проверки ролей и владения; ошибки здесь заметны пользователю сразу.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) DeleteService(id int64) error {
	user, err := a.requireAuth()
	if err != nil {
		return err
	}
	if id <= 0 {
		return errors.New("Некорректный идентификатор услуги.")
	}
	_, err = a.db.Exec(`DELETE FROM services WHERE id = ? AND created_by = ?`, id, user.Username)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

// RU: Метод `CreateUser`.
// EN: Method `CreateUser`.
//
// RU: Что делает: выполняет изменение данных в приложении и проводит бизнес-операцию.
// EN: What it does: CreateUser creates a subordinate account according to the current actor permissions.
//
// RU: Ключевые моменты: меняет состояние SQLite или сессии; опирается на проверки ролей и владения; ошибки здесь заметны пользователю сразу.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) CreateUser(req UserWithPassword) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}
	if !canCreateUsers(current.Role) {
		return User{}, errors.New("Недостаточно прав для создания пользователей.")
	}
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return User{}, errors.New("Введите логин и пароль нового пользователя.")
	}
	role := normalizeAssignableRole(req.Role)
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Вы не можете создать пользователя с этой ролью.")
	}
	createdAt := time.Now().Format(time.RFC3339)
	result, err := a.db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, username, hashPassword(password), role, createdAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	id, _ := result.LastInsertId()
	return a.getUserByID(id)
}

// RU: Метод `getUserByID`.
// EN: Method `getUserByID`.
//
// RU: Что делает: выполняет один из ключевых шагов backend-логики внутри приложения.
// EN: What it does: getUserByID loads one user record by primary key for follow-up authorization logic.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) getUserByID(id int64) (User, error) {
	var user User
	err := a.db.QueryRow(`SELECT id, username, role, created_at FROM users WHERE id = ?`, id).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// RU: Метод `ListUsers`.
// EN: Method `ListUsers`.
//
// RU: Что делает: читает и возвращает данные для фронтенда или внутренней логики.
// EN: What it does: ListUsers returns the visible users for the current actor, already filtered and sorted for the UI.
//
// RU: Ключевые моменты: используется в отрисовке интерфейса; должен возвращать только разрешённые данные; порядок и фильтрация важны для UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) ListUsers() ([]User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []User{}, nil
	}
	if rolePower(user.Role) < rolePower(RoleSeniorSpecialist) {
		return []User{}, nil
	}
	rows, err := a.db.Query(`SELECT id, username, role, created_at FROM users WHERE username <> 'admin' ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.Username, &item.Role, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		if item.Username == user.Username {
			continue
		}
		switch user.Role {
		case RoleAdmin:
			users = append(users, item)
		case RoleManager:
			if item.Role == RoleEmployee || item.Role == RoleSeniorSpecialist {
				users = append(users, item)
			}
		case RoleSeniorSpecialist:
			if item.Role == RoleEmployee {
				users = append(users, item)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(users, func(i, j int) bool {
		leftPriority := userSortPriority(users[i].Role)
		rightPriority := userSortPriority(users[j].Role)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if users[i].CreatedAt != users[j].CreatedAt {
			return users[i].CreatedAt < users[j].CreatedAt
		}
		return users[i].Username < users[j].Username
	})
	return users, nil
}

// RU: Метод `UpdateUserRole`.
// EN: Method `UpdateUserRole`.
//
// RU: Что делает: выполняет изменение данных в приложении и проводит бизнес-операцию.
// EN: What it does: UpdateUserRole changes another user role while enforcing hierarchy restrictions.
//
// RU: Ключевые моменты: меняет состояние SQLite или сессии; опирается на проверки ролей и владения; ошибки здесь заметны пользователю сразу.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) UpdateUserRole(userID int64, role string) (User, error) {
	current, err := a.requireManage()
	if err != nil {
		return User{}, err
	}
	role = normalizeAssignableRole(role)
	user, err := a.getUserByID(userID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	if user.Username == "admin" || user.Username == current.Username {
		return User{}, errors.New("Эту учётную запись нельзя изменить.")
	}

	switch current.Role {
	case RoleAdmin:
	case RoleManager:
		if user.Role != RoleEmployee && user.Role != RoleSeniorSpecialist {
			return User{}, errors.New("Руководитель может менять роли только специалистам тех. поддержки и старшим специалистам ТП.")
		}
		if role != RoleEmployee && role != RoleSeniorSpecialist {
			return User{}, errors.New("Руководитель может назначать только роли специалиста тех. поддержки и старшего специалиста ТП.")
		}
	default:
		return User{}, errors.New("Недостаточно прав для смены роли.")
	}

	_, err = a.db.Exec(`UPDATE users SET role = ? WHERE id = ?`, role, userID)
	if err != nil {
		return User{}, fmt.Errorf("update user role: %w", err)
	}
	updated, err := a.getUserByID(userID)
	if err != nil {
		return User{}, err
	}
	return updated, nil
}

// RU: Метод `DeleteUser`.
// EN: Method `DeleteUser`.
//
// RU: Что делает: выполняет изменение данных в приложении и проводит бизнес-операцию.
// EN: What it does: DeleteUser removes a subordinate account if the current role is allowed to do so.
//
// RU: Ключевые моменты: меняет состояние SQLite или сессии; опирается на проверки ролей и владения; ошибки здесь заметны пользователю сразу.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) DeleteUser(userID int64) error {
	current, err := a.requireAuth()
	if err != nil {
		return err
	}
	if rolePower(current.Role) < rolePower(RoleSeniorSpecialist) {
		return errors.New("Недостаточно прав для удаления пользователя.")
	}
	user, err := a.getUserByID(userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user.Username == "admin" || user.Username == current.Username {
		return errors.New("Эту учётную запись удалять нельзя.")
	}
	if current.Role == RoleSeniorSpecialist && user.Role != RoleEmployee {
		return errors.New("Старший специалист ТП может удалять только специалистов тех. поддержки.")
	}
	if current.Role == RoleManager && user.Role != RoleEmployee && user.Role != RoleSeniorSpecialist {
		return errors.New("Руководитель может удалять только специалистов тех. поддержки и старших специалистов ТП.")
	}
	_, err = a.db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// RU: Метод `CalculateAmount`.
// EN: Method `CalculateAmount`.
//
// RU: Что делает: выполняет главный расчёт распределения суммы по услугам с учётом процентов и весов.
// EN: What it does: CalculateAmount is the main business entry point that builds a calculation for a requested total.
//
// RU: Ключевые моменты: это центральная бизнес-функция; она связывает проценты, веса и точное равенство сумм; её изменения нужно всегда проверять тестами.
// EN: Key points: is the central business operation; combines percentages, weights and exact totals; should always be rechecked with tests after changes.
func (a *App) CalculateAmount(req CalculationRequest) (CalculationResult, error) {
	if _, err := a.requireAuth(); err != nil {
		return CalculationResult{}, err
	}
	if req.TargetAmount <= 0 {
		return CalculationResult{}, errors.New("Введите сумму больше 0.")
	}
	if req.TargetAmount > 1_500_000 {
		return CalculationResult{}, fmt.Errorf("Сумма %d р превышает допустимый предел 1500000 р.", req.TargetAmount)
	}

	services, err := a.GetServices()
	if err != nil {
		return CalculationResult{}, err
	}
	if len(services) == 0 {
		return CalculationResult{}, errors.New("Сначала создайте хотя бы одну услугу.")
	}

	weights := normalizeWeights(req.Weights, services)
	quantities, ok := solveStructuredAllocation(req.TargetAmount, services, weights)
	if !ok {
		quantities, ok = solveExact(req.TargetAmount, services, weights)
	}
	if !ok {
		return CalculationResult{}, fmt.Errorf("Не удалось подобрать точный расчёт на сумму %s.", displayMoney(req.TargetAmount))
	}

	items := make([]CalculationItem, 0, len(services))
	total := 0
	active := 0
	for idx, service := range services {
		quantity := quantities[idx]
		lineTotal := service.Rate * quantity
		if quantity > 0 {
			active++
		}
		total += lineTotal
		items = append(items, CalculationItem{
			ServiceID:         service.ID,
			ServiceCode:       service.Code,
			Name:              service.Name,
			Unit:              service.Unit,
			Rate:              service.Rate,
			Quantity:          quantity,
			LineTotal:         lineTotal,
			Description:       service.Description,
			Weight:            weights[service.Code],
			Category:          service.Category,
			AllocationPercent: service.AllocationPercent,
		})
	}

	return CalculationResult{
		TargetAmount:   req.TargetAmount,
		TotalAmount:    total,
		Items:          items,
		FoundExact:     total == req.TargetAmount,
		GeneratedAt:    time.Now().Format(time.RFC3339),
		ActiveServices: active,
		Weights:        weights,
	}, nil
}

// RU: Функция `normalizeWeights`.
// EN: Function `normalizeWeights`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: normalizeWeights clamps and aligns frontend weights with the currently available services.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func normalizeWeights(input map[string]int, services []Service) map[string]int {
	result := make(map[string]int, len(services))
	for _, service := range services {
		result[service.Code] = 0
	}
	for code, weight := range input {
		if weight < 0 {
			weight = 0
		}
		result[code] = weight
	}
	return result
}

// RU: Функция `solveStructuredAllocation`.
// EN: Function `solveStructuredAllocation`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: solveStructuredAllocation tries the preferred group-based solver before falling back to a generic exact search.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func solveStructuredAllocation(target int, services []Service, weights map[string]int) ([]int, bool) {
	baseQuantities, reducedTarget, ok := reserveMinimumServices(target, services)
	if !ok {
		return nil, false
	}
	byCategory := map[string][]Service{
		CategoryPrimary:   {},
		CategorySecondary: {},
		CategoryClosing:   {},
	}
	indexMap := map[string][]int{
		CategoryPrimary:   {},
		CategorySecondary: {},
		CategoryClosing:   {},
	}
	for idx, service := range services {
		category := normalizeCategory(service.Category)
		byCategory[category] = append(byCategory[category], service)
		indexMap[category] = append(indexMap[category], idx)
	}

	groupTargets, ok := chooseGroupTargets(reducedTarget, byCategory)
	if !ok {
		return nil, false
	}
	targets := buildServiceTargets(byCategory, groupTargets, weights)
	result := append([]int(nil), baseQuantities...)
	for _, category := range []string{CategoryPrimary, CategorySecondary, CategoryClosing} {
		group := byCategory[category]
		if len(group) == 0 {
			continue
		}
		serviceTargets := make([]int, len(group))
		for i, service := range group {
			serviceTargets[i] = targets[service.Code]
		}
		allocation, ok := bestGroupAllocation(groupTargets[category], group, serviceTargets, weights)
		if !ok {
			return nil, false
		}
		mergeGroupQuantities(result, allocation.quantities, indexMap[category])
	}

	total := 0
	for idx, quantity := range result {
		total += quantity * services[idx].Rate
	}
	if total != target {
		return nil, false
	}
	return result, true
}

// RU: Функция `chooseGroupTargets`.
// EN: Function `chooseGroupTargets`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: chooseGroupTargets picks category subtotals close to configured percentages while staying exactly reachable.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func chooseGroupTargets(target int, byCategory map[string][]Service) (map[string]int, bool) {
	defaults := defaultGroupPercent()
	categoryTargets := allocateByRatios(target, []float64{defaults[CategoryPrimary], defaults[CategorySecondary], defaults[CategoryClosing]})
	desired := map[string]int{
		CategoryPrimary:   categoryTargets[0],
		CategorySecondary: categoryTargets[1],
		CategoryClosing:   categoryTargets[2],
	}
	reachable := map[string][]bool{}
	for _, category := range []string{CategoryPrimary, CategorySecondary, CategoryClosing} {
		reachable[category] = reachableAmounts(target, byCategory[category])
	}

	best := map[string]int{}
	bestScore := math.MaxInt
	windows := []int{500, 1500, 5000, 15000, target}
	for _, window := range windows {
		primaryCandidates := candidateAmounts(reachable[CategoryPrimary], desired[CategoryPrimary], window)
		secondaryCandidates := candidateAmounts(reachable[CategorySecondary], desired[CategorySecondary], window)
		for _, primary := range primaryCandidates {
			for _, secondary := range secondaryCandidates {
				closing := target - primary - secondary
				if closing < 0 || closing >= len(reachable[CategoryClosing]) || !reachable[CategoryClosing][closing] {
					continue
				}
				score := absInt(primary-desired[CategoryPrimary]) + absInt(secondary-desired[CategorySecondary]) + absInt(closing-desired[CategoryClosing])
				if score < bestScore {
					bestScore = score
					best[CategoryPrimary] = primary
					best[CategorySecondary] = secondary
					best[CategoryClosing] = closing
				}
			}
		}
		if bestScore != math.MaxInt {
			return best, true
		}
	}
	return nil, false
}

// RU: Функция `reserveMinimumServices`.
// EN: Function `reserveMinimumServices`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: reserveMinimumServices reserves one unit per service when the target allows all services to stay active.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func reserveMinimumServices(target int, services []Service) ([]int, int, bool) {
	result := make([]int, len(services))
	if len(services) == 0 {
		return result, target, true
	}
	minimumTotal := 0
	for _, service := range services {
		minimumTotal += service.Rate
	}
	if minimumTotal > target {
		return result, target, true
	}
	for idx := range services {
		result[idx] = 1
	}
	return result, target - minimumTotal, true
}

// RU: Функция `buildServiceTargets`.
// EN: Function `buildServiceTargets`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: buildServiceTargets distributes each category budget down to concrete services before exact solving.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func buildServiceTargets(byCategory map[string][]Service, categoryTargetMap map[string]int, weights map[string]int) map[string]int {
	categoryOrder := []string{CategoryPrimary, CategorySecondary, CategoryClosing}

	totalServices := 0
	for _, group := range byCategory {
		totalServices += len(group)
	}
	result := make(map[string]int, totalServices)
	for _, category := range categoryOrder {
		group := byCategory[category]
		if len(group) == 0 {
			continue
		}
		ratios := buildCategoryRatios(category, group, weights)
		allocated := allocateByRatios(categoryTargetMap[category], ratios)
		for idx, service := range group {
			result[service.Code] = allocated[idx]
		}
	}
	return result
}

// RU: Функция `buildCategoryRatios`.
// EN: Function `buildCategoryRatios`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: buildCategoryRatios calculates initial per-service ratios inside one category.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func buildCategoryRatios(category string, group []Service, weights map[string]int) []float64 {
	groupPercent := defaultGroupPercent()[category] * 100.0
	ratios := make([]float64, len(group))
	explicitTotal := 0.0
	unassigned := 0
	for idx, service := range group {
		if service.AllocationPercent != nil {
			ratios[idx] = *service.AllocationPercent
			explicitTotal += *service.AllocationPercent
		} else {
			unassigned++
		}
	}

	remainingPercent := groupPercent - explicitTotal
	if remainingPercent < 0 {
		remainingPercent = 0
	}
	fallback := 0.0
	if unassigned > 0 {
		fallback = remainingPercent / float64(unassigned)
	}
	for idx, service := range group {
		if service.AllocationPercent == nil {
			ratios[idx] = fallback
		}
	}
	applyWeightAdjustments(ratios, group, weights)
	return ratios
}

// RU: Функция `applyWeightAdjustments`.
// EN: Function `applyWeightAdjustments`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: applyWeightAdjustments shifts ratio share toward weighted services while preserving total category mass.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func applyWeightAdjustments(ratios []float64, group []Service, weights map[string]int) {
	if len(group) < 2 {
		return
	}
	for idx, service := range group {
		weight := weights[service.Code]
		if weight < 0 {
			weight = 0
		}
		if weight > 10 {
			weight = 10
		}
		for step := 0; step < weight; step++ {
			transferShare(ratios, idx, 2.0)
		}
	}
}

// RU: Функция `transferShare`.
// EN: Function `transferShare`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: transferShare moves percentage share from peer services to one preferred target service.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func transferShare(ratios []float64, target int, share float64) {
	if share <= 0 {
		return
	}
	donors := make([]int, 0, len(ratios)-1)
	for idx, value := range ratios {
		if idx == target || value <= 0 {
			continue
		}
		donors = append(donors, idx)
	}
	remaining := share
	for remaining > 0.0001 && len(donors) > 0 {
		slice := remaining / float64(len(donors))
		nextDonors := make([]int, 0, len(donors))
		moved := 0.0
		for _, donor := range donors {
			take := math.Min(slice, ratios[donor])
			if take <= 0 {
				continue
			}
			ratios[donor] -= take
			moved += take
			if ratios[donor] > 0.0001 {
				nextDonors = append(nextDonors, donor)
			}
		}
		if moved <= 0 {
			break
		}
		ratios[target] += moved
		remaining -= moved
		donors = nextDonors
	}
}

// RU: Функция `reachableAmounts`.
// EN: Function `reachableAmounts`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: reachableAmounts marks which totals can be composed from the service rates of one category.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func reachableAmounts(limit int, services []Service) []bool {
	reachable := make([]bool, limit+1)
	reachable[0] = true
	for amount := 0; amount <= limit; amount++ {
		if !reachable[amount] {
			continue
		}
		for _, service := range services {
			next := amount + service.Rate
			if next <= limit {
				reachable[next] = true
			}
		}
	}
	return reachable
}

// RU: Функция `candidateAmounts`.
// EN: Function `candidateAmounts`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: candidateAmounts collects reachable totals nearest to a desired subtotal for later scoring.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func candidateAmounts(reachable []bool, desired int, window int) []int {
	minAmount := desired - window
	if minAmount < 0 {
		minAmount = 0
	}
	maxAmount := desired + window
	if maxAmount >= len(reachable) {
		maxAmount = len(reachable) - 1
	}
	candidates := make([]int, 0)
	for amount := minAmount; amount <= maxAmount; amount++ {
		if reachable[amount] {
			candidates = append(candidates, amount)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return absInt(candidates[i]-desired) < absInt(candidates[j]-desired)
	})
	return candidates
}

// RU: Функция `allocateByRatios`.
// EN: Function `allocateByRatios`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: allocateByRatios converts floating-point target shares into integer money targets that still sum exactly.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func allocateByRatios(total int, ratios []float64) []int {
	result := make([]int, len(ratios))
	if total <= 0 || len(ratios) == 0 {
		return result
	}

	sum := 0.0
	for _, ratio := range ratios {
		if ratio > 0 {
			sum += ratio
		}
	}
	if sum <= 0 {
		base := total / len(ratios)
		rest := total % len(ratios)
		for idx := range ratios {
			result[idx] = base
			if idx < rest {
				result[idx]++
			}
		}
		return result
	}

	type remainderPart struct {
		idx       int
		remainder float64
	}
	remainders := make([]remainderPart, 0, len(ratios))
	allocated := 0
	for idx, ratio := range ratios {
		normalized := 0.0
		if ratio > 0 {
			normalized = ratio / sum
		}
		exact := normalized * float64(total)
		base := int(exact)
		result[idx] = base
		allocated += base
		remainders = append(remainders, remainderPart{idx: idx, remainder: exact - float64(base)})
	}

	sort.SliceStable(remainders, func(i, j int) bool {
		return remainders[i].remainder > remainders[j].remainder
	})
	for idx := 0; idx < total-allocated; idx++ {
		result[remainders[idx%len(remainders)].idx]++
	}
	return result
}

// RU: Функция `sumTargets`.
// EN: Function `sumTargets`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: sumTargets totals a slice of integer allocations.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func sumTargets(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

// RU: Функция `bestGroupAllocation`.
// EN: Function `bestGroupAllocation`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: bestGroupAllocation finds the best exact quantities for one category under a target-share preference.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func bestGroupAllocation(target int, services []Service, targets []int, weights map[string]int) (groupAllocation, bool) {
	if len(services) == 0 {
		return groupAllocation{ok: target == 0, amount: 0, quantities: []int{}}, target == 0
	}
	type groupStage struct {
		score  int
		count  int
		prev   int
		qty    int
		exists bool
	}

	for _, margin := range []int{6, 14, 32, 80, 180, 400} {
		current := map[int]groupStage{
			0: {exists: true},
		}
		parents := make([]map[int]groupStage, len(services))

		for idx, service := range services {
			next := map[int]groupStage{}
			parents[idx] = map[int]groupStage{}
			targetQty := float64(targets[idx]) / float64(service.Rate)
			qtyMargin := margin + int(math.Ceil(math.Sqrt(targetQty+1)))
			minQty := 0
			maxQty := int(math.Ceil(targetQty)) + qtyMargin
			limitQty := target / service.Rate
			if maxQty > limitQty {
				maxQty = limitQty
			}
			for amount, state := range current {
				for qty := minQty; qty <= maxQty; qty++ {
					nextAmount := amount + qty*service.Rate
					if nextAmount > target {
						break
					}
					score := state.score + serviceAllocationScore(service, qty, targets[idx], weights[service.Code])
					count := state.count + qty
					existing, ok := next[nextAmount]
					if !ok || score > existing.score || (score == existing.score && count < existing.count) {
						stage := groupStage{score: score, count: count, prev: amount, qty: qty, exists: true}
						next[nextAmount] = stage
						parents[idx][nextAmount] = stage
					}
				}
			}
			current = next
		}

		finalState, ok := current[target]
		if !ok || !finalState.exists {
			continue
		}
		quantities := make([]int, len(services))
		currentAmount := target
		for idx := len(services) - 1; idx >= 0; idx-- {
			stage := parents[idx][currentAmount]
			quantities[idx] = stage.qty
			currentAmount = stage.prev
		}
		return groupAllocation{amount: target, score: finalState.score, quantities: quantities, ok: true}, true
	}
	return groupAllocation{}, false
}

// RU: Функция `serviceAllocationScore`.
// EN: Function `serviceAllocationScore`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: serviceAllocationScore ranks candidate quantities by closeness to target share and weight preference.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func serviceAllocationScore(service Service, qty int, targetShare int, weight int) int {
	allocated := qty * service.Rate
	score := -absInt(allocated-targetShare) * 100
	if targetShare > 0 && qty == 0 {
		score -= 50_000
	}
	score += weight * qty * 25
	if qty > 0 {
		score += 500
	}
	return score
}

// RU: Функция `absInt`.
// EN: Function `absInt`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: absInt is a tiny helper used by scoring and distance calculations.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// RU: Функция `solveExact`.
// EN: Function `solveExact`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: solveExact is the fallback exact solver used when the structured strategy cannot satisfy the target.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func solveExact(target int, services []Service, weights map[string]int) ([]int, bool) {
	baseQuantities, reducedTarget, _ := reserveMinimumServices(target, services)
	dp := make([]allocationState, reducedTarget+1)
	dp[0] = allocationState{ok: true, prev: -1, idx: -1}
	for amount := 0; amount <= reducedTarget; amount++ {
		if !dp[amount].ok {
			continue
		}
		for idx, service := range services {
			next := amount + service.Rate
			if next > reducedTarget {
				continue
			}
			candidate := allocationState{score: dp[amount].score + serviceAllocationScore(service, 1, service.Rate, weights[service.Code]), count: dp[amount].count + 1, prev: amount, idx: idx, ok: true}
			if !dp[next].ok || candidate.score > dp[next].score || (candidate.score == dp[next].score && candidate.count < dp[next].count) {
				dp[next] = candidate
			}
		}
	}
	if !dp[reducedTarget].ok {
		return nil, false
	}
	quantities := append([]int(nil), baseQuantities...)
	for current := reducedTarget; current > 0; {
		step := dp[current]
		quantities[step.idx]++
		current = step.prev
	}
	return quantities, true
}

// RU: Функция `mergeGroupQuantities`.
// EN: Function `mergeGroupQuantities`.
//
// RU: Что делает: участвует во внутренней математике расчёта и точном распределении суммы.
// EN: What it does: mergeGroupQuantities writes group-local quantities back into the full result slice.
//
// RU: Ключевые моменты: является частью расчётного пайплайна; чувствителен к граничным случаям; требует тестовой проверки после правок.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func mergeGroupQuantities(result []int, quantities []int, indexMap []int) {
	for idx, quantity := range quantities {
		result[indexMap[idx]] += quantity
	}
}

// RU: Метод `SaveCalculation`.
// EN: Method `SaveCalculation`.
//
// RU: Что делает: выполняет изменение данных в приложении и проводит бизнес-операцию.
// EN: What it does: SaveCalculation stores the current calculation in SQLite together with its author and line items.
//
// RU: Ключевые моменты: меняет состояние SQLite или сессии; опирается на проверки ролей и владения; ошибки здесь заметны пользователю сразу.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) SaveCalculation(req SaveCalculationRequest) (SavedCalculation, error) {
	user, err := a.requireAuth()
	if err != nil {
		return SavedCalculation{}, err
	}
	if req.TargetAmount <= 0 || len(req.Items) == 0 {
		return SavedCalculation{}, errors.New("Нельзя сохранить пустой расчёт.")
	}
	title := sanitizeStoredText(req.Title)
	if title == "" {
		title = fmt.Sprintf("Расчёт на %s", displayMoney(req.TargetAmount))
	}
	total := 0
	for _, item := range req.Items {
		total += item.LineTotal
	}
	payload, err := json.Marshal(req.Items)
	if err != nil {
		return SavedCalculation{}, fmt.Errorf("marshal calculation items: %w", err)
	}
	createdAt := time.Now().Format(time.RFC3339)
	result, err := a.db.Exec(`INSERT INTO calculations(title, target_amount, total_amount, items_json, created_at, created_by) VALUES(?, ?, ?, ?, ?, ?)`, title, req.TargetAmount, total, string(payload), createdAt, user.Username)
	if err != nil {
		return SavedCalculation{}, fmt.Errorf("save calculation: %w", err)
	}
	id, _ := result.LastInsertId()
	return SavedCalculation{ID: id, Title: title, TargetAmount: req.TargetAmount, TotalAmount: total, Items: req.Items, CreatedAt: createdAt, CreatedBy: user.Username}, nil
}

// RU: Метод `DeleteCalculation`.
// EN: Method `DeleteCalculation`.
//
// RU: Что делает: выполняет изменение данных в приложении и проводит бизнес-операцию.
// EN: What it does: DeleteCalculation removes one archived calculation if the current role may access that entry.
//
// RU: Ключевые моменты: меняет состояние SQLite или сессии; опирается на проверки ролей и владения; ошибки здесь заметны пользователю сразу.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) DeleteCalculation(id int64) error {
	user, err := a.requireAuth()
	if err != nil {
		return err
	}
	if id <= 0 {
		return errors.New("Некорректный идентификатор расчёта.")
	}
	if rolePower(user.Role) >= rolePower(RoleManager) {
		_, err = a.db.Exec(`DELETE FROM calculations WHERE id = ?`, id)
		return err
	}
	_, err = a.db.Exec(`DELETE FROM calculations WHERE id = ? AND created_by = ?`, id, user.Username)
	return err
}

// RU: Метод `ListCalculations`.
// EN: Method `ListCalculations`.
//
// RU: Что делает: читает и возвращает данные для фронтенда или внутренней логики.
// EN: What it does: ListCalculations returns the archive visible to the current role, already filtered by ownership rules.
//
// RU: Ключевые моменты: используется в отрисовке интерфейса; должен возвращать только разрешённые данные; порядок и фильтрация важны для UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) ListCalculations() ([]SavedCalculation, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []SavedCalculation{}, nil
	}
	query := `SELECT c.id, c.title, c.target_amount, c.total_amount, c.items_json, c.created_at, c.created_by FROM calculations c`
	args := []interface{}{}
	switch {
	case canViewAllArchives(user.Role):
	case user.Role == RoleManager:
		query += ` LEFT JOIN users u ON u.username = c.created_by WHERE u.role IN (?, ?)`
		args = append(args, RoleEmployee, RoleSeniorSpecialist)
	case user.Role == RoleSeniorSpecialist:
		query += ` LEFT JOIN users u ON u.username = c.created_by WHERE u.role = ?`
		args = append(args, RoleEmployee)
	default:
		query += ` WHERE c.created_by = ?`
		args = append(args, user.Username)
	}
	query += ` ORDER BY c.id DESC`
	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list calculations: %w", err)
	}
	defer rows.Close()
	result := make([]SavedCalculation, 0)
	for rows.Next() {
		var item SavedCalculation
		var payload string
		if err := rows.Scan(&item.ID, &item.Title, &item.TargetAmount, &item.TotalAmount, &payload, &item.CreatedAt, &item.CreatedBy); err != nil {
			return nil, fmt.Errorf("scan calculation: %w", err)
		}
		item.Title = sanitizeStoredText(item.Title)
		if item.Title == "" {
			item.Title = fmt.Sprintf("Расчёт на %s", displayMoney(item.TargetAmount))
		}
		if err := json.Unmarshal([]byte(payload), &item.Items); err != nil {
			return nil, fmt.Errorf("parse calculation items: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
