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

// RU: Р СџР ВµРЎР‚Р ВµР СР ВµР Р…Р Р…Р В°РЎРЏ `assets`.
// EN: Variable `assets`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: РЎвЂ¦РЎР‚Р В°Р Р…Р С‘РЎвЂљ РЎР‚Р ВµРЎРѓРЎС“РЎР‚РЎРѓРЎвЂ№ Р С‘Р В»Р С‘ Р С–Р В»Р С•Р В±Р В°Р В»РЎРЉР Р…Р С•Р Вµ РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ, Р С”Р С•РЎвЂљР С•РЎР‚Р С•Р Вµ Р Р…РЎС“Р В¶Р Р…Р С• Р Т‘РЎР‚РЎС“Р С–Р С‘Р С РЎвЂЎР В°РЎРѓРЎвЂљРЎРЏР С Р С—РЎР‚Р С•Р С–РЎР‚Р В°Р СР СРЎвЂ№.
// EN: What it does: assets bundles the prebuilt frontend so Wails can serve it from the executable without external files.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р С”Р В°Р С” Р С”Р С•Р Р…РЎвЂљРЎР‚Р В°Р С”РЎвЂљ Р С‘Р В»Р С‘ Р С•Р С—Р С•РЎР‚Р Р…Р В°РЎРЏ РЎвЂљР С•РЎвЂЎР С”Р В° Р Т‘Р В»РЎРЏ Р Т‘РЎР‚РЎС“Р С–Р С‘РЎвЂ¦ РЎвЂЎР В°РЎРѓРЎвЂљР ВµР в„– Р С—РЎР‚Р С•Р ВµР С”РЎвЂљР В°; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В·Р Т‘Р ВµРЎРѓРЎРЉ РЎвЂЎР В°РЎРѓРЎвЂљР С• РЎвЂљРЎР‚Р ВµР В±РЎС“РЎР‹РЎвЂљ Р С•РЎРѓРЎвЂљР С•РЎР‚Р С•Р В¶Р Р…Р С•РЎРѓРЎвЂљР С‘.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
//
//go:embed all:frontend/dist
var assets embed.FS

// RU: Р В¤РЎС“Р Р…Р С”РЎвЂ Р С‘РЎРЏ `resolveWebviewUserDataPath`.
// EN: Function `resolveWebviewUserDataPath`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р Р†РЎвЂ№РЎвЂЎР С‘РЎРѓР В»РЎРЏР ВµРЎвЂљ РЎРѓРЎвЂљР В°Р В±Р С‘Р В»РЎРЉР Р…РЎвЂ№Р в„– Р С—РЎС“РЎвЂљРЎРЉ Р Т‘Р В»РЎРЏ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦ WebView2, Р С”Р С•РЎвЂљР С•РЎР‚РЎвЂ№Р в„– Р Р…Р Вµ Р В·Р В°Р Р†Р С‘РЎРѓР С‘РЎвЂљ Р С•РЎвЂљ Р С‘Р СР ВµР Р…Р С‘ `.exe`.
// EN: What it does: resolves a stable WebView2 user data path that does not depend on the current executable name.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµРЎвЂљ Р С•Р Т‘Р С‘Р Р… Р С‘ РЎвЂљР С•РЎвЂљ Р В¶Р Вµ Р С”Р В°РЎвЂљР В°Р В»Р С•Р С– `MercelData`; Р С—РЎР‚Р ВµР Т‘Р С•РЎвЂљР Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ РЎРѓР С•Р В·Р Т‘Р В°Р Р…Р С‘Р Вµ Р Р…Р С•Р Р†РЎвЂ№РЎвЂ¦ Р С—Р В°Р С—Р С•Р С” Р С—РЎР‚Р С‘ Р С—Р ВµРЎР‚Р ВµР С‘Р СР ВµР Р…Р С•Р Р†Р В°Р Р…Р С‘Р С‘ Р С—РЎР‚Р С‘Р В»Р С•Р В¶Р ВµР Р…Р С‘РЎРЏ; РЎРѓР С•Р В·Р Т‘Р В°РЎвЂРЎвЂљ Р С”Р В°РЎвЂљР В°Р В»Р С•Р С– Р В·Р В°РЎР‚Р В°Р Р…Р ВµР Вµ.
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

// RU: Р В¤РЎС“Р Р…Р С”РЎвЂ Р С‘РЎРЏ `main`.
// EN: Function `main`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р Р†РЎРѓР С—Р С•Р СР С•Р С–Р В°РЎвЂљР ВµР В»РЎРЉР Р…Р С•Р Вµ Р С—РЎР‚Р ВµР С•Р В±РЎР‚Р В°Р В·Р С•Р Р†Р В°Р Р…Р С‘Р Вµ, Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚Р С”РЎС“ Р С‘Р В»Р С‘ Р С—Р С•Р Т‘Р С–Р С•РЎвЂљР С•Р Р†Р С”РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦.
// EN: What it does: main bootstraps the backend application, wires it into Wails, and starts the desktop window lifecycle.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р Т‘Р В»РЎРЏ РЎС“РЎРѓРЎвЂљР С•Р в„–РЎвЂЎР С‘Р Р†Р С•РЎРѓРЎвЂљР С‘ Р В»Р С•Р С–Р С‘Р С”Р С‘; Р СР С•Р В¶Р ВµРЎвЂљ Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљРЎРЉРЎРѓРЎРЏ РЎРѓРЎР‚Р В°Р В·РЎС“ Р Р† Р Р…Р ВµРЎРѓР С”Р С•Р В»РЎРЉР С”Р С‘РЎвЂ¦ Р СР ВµРЎРѓРЎвЂљР В°РЎвЂ¦; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂљР С•Р С‘РЎвЂљ Р Т‘Р ВµР В»Р В°РЎвЂљРЎРЉ Р С•РЎРѓР С•Р В·Р Р…Р В°Р Р…Р Р…Р С•.
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
		Width:            1366,
		Height:           860,
		MinWidth:         1120,
		MinHeight:        680,
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
