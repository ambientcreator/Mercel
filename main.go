package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
	"statistic/appcore"
)

// RU: РџРµСЂРµРјРµРЅРЅР°СЏ `assets`.
// EN: Variable `assets`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: С…СЂР°РЅРёС‚ СЂРµСЃСѓСЂСЃС‹ РёР»Рё РіР»РѕР±Р°Р»СЊРЅРѕРµ СЃРѕСЃС‚РѕСЏРЅРёРµ, РєРѕС‚РѕСЂРѕРµ РЅСѓР¶РЅРѕ РґСЂСѓРіРёРј С‡Р°СЃС‚СЏРј РїСЂРѕРіСЂР°РјРјС‹.
// EN: What it does: assets bundles the prebuilt frontend so Wails can serve it from the executable without external files.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РєР°Рє РєРѕРЅС‚СЂР°РєС‚ РёР»Рё РѕРїРѕСЂРЅР°СЏ С‚РѕС‡РєР° РґР»СЏ РґСЂСѓРіРёС… С‡Р°СЃС‚РµР№ РїСЂРѕРµРєС‚Р°; РёР·РјРµРЅРµРЅРёСЏ Р·РґРµСЃСЊ С‡Р°СЃС‚Рѕ С‚СЂРµР±СѓСЋС‚ РѕСЃС‚РѕСЂРѕР¶РЅРѕСЃС‚Рё.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
//
//go:embed all:frontend/dist
var assets embed.FS

// RU: Р¤СѓРЅРєС†РёСЏ `resolveWebviewUserDataPath`.
// EN: Function `resolveWebviewUserDataPath`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹С‡РёСЃР»СЏРµС‚ СЃС‚Р°Р±РёР»СЊРЅС‹Р№ РїСѓС‚СЊ РґР»СЏ РґР°РЅРЅС‹С… WebView2, РєРѕС‚РѕСЂС‹Р№ РЅРµ Р·Р°РІРёСЃРёС‚ РѕС‚ РёРјРµРЅРё `.exe`.
// EN: What it does: resolves a stable WebView2 user data path that does not depend on the current executable name.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РёСЃРїРѕР»СЊР·СѓРµС‚ РѕРґРёРЅ Рё С‚РѕС‚ Р¶Рµ РєР°С‚Р°Р»РѕРі `MercelData`; РїСЂРµРґРѕС‚РІСЂР°С‰Р°РµС‚ СЃРѕР·РґР°РЅРёРµ РЅРѕРІС‹С… РїР°РїРѕРє РїСЂРё РїРµСЂРµРёРјРµРЅРѕРІР°РЅРёРё РїСЂРёР»РѕР¶РµРЅРёСЏ; СЃРѕР·РґР°С‘С‚ РєР°С‚Р°Р»РѕРі Р·Р°СЂР°РЅРµРµ.
// EN: Key points: uses the fixed `MercelData` directory; prevents extra folders from appearing after renaming the executable; creates the directory ahead of time.
func resolveWebviewUserDataPath() (string, error) {
	// WebView2 is more stable when its user-data folder lives in LocalAppData,
	// while our SQLite database remains in Roaming.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		configDir, fallbackErr := os.UserConfigDir()
		if fallbackErr != nil {
			return "", fmt.Errorf("resolve local app data: %w; fallback config dir: %v", err, fallbackErr)
		}
		cacheDir = configDir
	}

	webviewDir := filepath.Join(cacheDir, appcore.AppStorageDirName, "webview2")
	if err := os.MkdirAll(webviewDir, 0o755); err != nil {
		return "", err
	}

	return webviewDir, nil
}

func resolveStartupLogPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "mercel-startup.log"
	}
	logDir := filepath.Join(configDir, appcore.AppStorageDirName)
	if mkErr := os.MkdirAll(logDir, 0o755); mkErr != nil {
		return "mercel-startup.log"
	}
	return filepath.Join(logDir, "startup.log")
}

func appendStartupLog(message string) {
	line := fmt.Sprintf("[%s] %s%s", time.Now().Format(time.RFC3339), message, "\n")
	file, err := os.OpenFile(resolveStartupLogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(line)
}

// RU: Р¤СѓРЅРєС†РёСЏ `main`.
// EN: Function `main`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: main bootstraps the backend application, wires it into Wails, and starts the desktop window lifecycle.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func main() {
	appendStartupLog(fmt.Sprintf("starting app on %s", runtime.GOOS))

	app, err := appcore.NewApp()
	if err != nil {
		appendStartupLog(fmt.Sprintf("NewApp failed: %v", err))
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer app.Close()

	webviewUserDataPath, err := resolveWebviewUserDataPath()
	if err != nil {
		appendStartupLog(fmt.Sprintf("resolveWebviewUserDataPath failed: %v", err))
		log.Fatalf("failed to resolve webview user data path: %v", err)
	}
	appendStartupLog(fmt.Sprintf("webview path: %s", webviewUserDataPath))

	err = wails.Run(&options.App{
		Title:            "Mercel",
		Width:            1440,
		Height:           960,
		MinWidth:         1180,
		MinHeight:        760,
		Frameless:        false,
		DisableResize:    false,
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Windows: &windowsoptions.Options{
			WebviewUserDataPath: webviewUserDataPath,
		},
		OnStartup: func(ctx context.Context) {
			app.Startup(ctx)
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		appendStartupLog(fmt.Sprintf("wails.Run failed: %v", err))
		log.Fatalf("failed to run app: %v", err)
	}
	appendStartupLog("app stopped normally")
}
