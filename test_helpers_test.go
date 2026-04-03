package main

import (
	"path/filepath"
	"testing"
)

// RU: Р¤СѓРЅРєС†РёСЏ `withTempDB`.
// EN: Function `withTempDB`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: withTempDB creates an isolated application instance backed by a temporary SQLite file for repeatable tests.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
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

// RU: Р¤СѓРЅРєС†РёСЏ `loginAsAdmin`.
// EN: Function `loginAsAdmin`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: loginAsAdmin is a helper that authenticates using the seeded admin account.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
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

// RU: Р¤СѓРЅРєС†РёСЏ `loginAsAdmin1`.
// EN: Function `loginAsAdmin1`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІС…РѕРґ РїРѕРґ С‚РµСЃС‚РѕРІРѕР№ admin-СѓС‡С‘С‚РєРѕР№, РєРѕС‚РѕСЂР°СЏ РЅСѓР¶РЅР° РґР»СЏ РїСЂРѕРІРµСЂРєРё РѕР±С‹С‡РЅРѕРіРѕ admin-РёРЅС‚РµСЂС„РµР№СЃР° Рё СЃС†РµРЅР°СЂРёРµРІ.
// EN: What it does: loginAsAdmin1 authenticates using the seeded test admin account for ordinary admin-role scenarios.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РёСЃРїРѕР»СЊР·СѓРµС‚ РІРёРґРёРјСѓСЋ С‚РµСЃС‚РѕРІСѓСЋ admin-СѓС‡С‘С‚РєСѓ; РЅРµ С‚СЂРѕРіР°РµС‚ Р·Р°С‰РёС‰С‘РЅРЅС‹Р№ username `admin`; СѓРїСЂРѕС‰Р°РµС‚ СЂРµРіСЂРµСЃСЃРёРѕРЅРЅС‹Рµ РїСЂРѕРІРµСЂРєРё.
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

// RU: Р¤СѓРЅРєС†РёСЏ `loginAsUser`.
// EN: Function `loginAsUser`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: loginAsUser authenticates as an arbitrary test user and fails the test immediately on error.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
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

// RU: Р¤СѓРЅРєС†РёСЏ `createUserService`.
// EN: Function `createUserService`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: createUserService logs in as a test user and creates one owned service for archive and visibility scenarios.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
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

// RU: РўРµСЃС‚ `TestSeededAdminLoginWorks`.
// EN: Test `TestSeededAdminLoginWorks`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїСЂРѕРІРµСЂСЏРµС‚ РѕС‚РґРµР»СЊРЅС‹Р№ СЃС†РµРЅР°СЂРёР№ Рё С„РёРєСЃРёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ Р±РµР· СЂСѓС‡РЅРѕР№ РїСЂРѕРІРµСЂРєРё.
// EN: What it does: TestSeededAdminLoginWorks verifies that a brand-new database always contains a usable admin account.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЂР°Р±РѕС‚Р°РµС‚ РІ РёР·РѕР»РёСЂРѕРІР°РЅРЅРѕРј СЃС†РµРЅР°СЂРёРё; РЅСѓР¶РµРЅ РґР»СЏ Р·Р°С‰РёС‚С‹ РѕС‚ СЂРµРіСЂРµСЃСЃРёР№; РґРѕРєСѓРјРµРЅС‚РёСЂСѓРµС‚ РѕР¶РёРґР°РµРјРѕРµ РїРѕРІРµРґРµРЅРёРµ.
// EN: Key points: runs in an isolated scenario; protects against regressions; documents the expected behavior of the feature or rule.
