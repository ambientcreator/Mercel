package appcore_test

import (
	"path/filepath"
	"testing"

	. "statistic/appcore"
)

func usePathResolvers(t *testing.T, dbPath string, legacyPath string, previousPath string) {
	t.Helper()
	restore := OverridePathResolversForTest(PathResolvers{
		DatabasePath: func() (string, error) {
			return dbPath, nil
		},
		LegacyDatabasePath: func() (string, error) {
			return legacyPath, nil
		},
	})
	if previousPath != "" {
		restore = OverridePathResolversForTest(PathResolvers{
			DatabasePath: func() (string, error) {
				return dbPath, nil
			},
			LegacyDatabasePath: func() (string, error) {
				return legacyPath, nil
			},
			PreviousMercelDBPath: func() (string, error) {
				return previousPath, nil
			},
		})
	}
	t.Cleanup(restore)
}

func withTempDB(t *testing.T) *App {
	t.Helper()
	tempDir := t.TempDir()
	usePathResolvers(t, filepath.Join(tempDir, "test.sqlite"), filepath.Join(tempDir, "legacy.sqlite"), "")

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})
	return app
}

func loginAsAdmin(t *testing.T, app *App) {
	t.Helper()
	state, err := app.Login(LoginRequest{Username: "admin", Password: "siuW*R%wWkQS"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !state.Authenticated || state.User == nil || state.User.Username != "admin" {
		t.Fatalf("unexpected session state: %+v", state)
	}
}

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

func createUserService(t *testing.T, app *App, username string, password string, name string, rate int, category string) {
	t.Helper()
	loginAsUser(t, app, username, password)
	_, err := app.UpsertService(UpsertServiceRequest{
		Name:     name,
		Unit:     "ч.",
		Rate:     rate,
		Category: category,
	})
	if err != nil {
		t.Fatalf("UpsertService() error = %v", err)
	}
}
