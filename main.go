package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
)

// RU: Переменная `assets`.
// EN: Variable `assets`.
//
// RU: Что делает: хранит ресурсы или глобальное состояние, которое нужно другим частям программы.
// EN: What it does: assets bundles the prebuilt frontend so Wails can serve it from the executable without external files.
//
// RU: Ключевые моменты: важен как контракт или опорная точка для других частей проекта; изменения здесь часто требуют осторожности.
// EN: Key points: serves as a shared contract or reference point; is reused across multiple areas of the project; changes here should be made carefully.
//
//go:embed all:frontend/dist
var assets embed.FS

// RU: Функция `resolveWebviewUserDataPath`.
// EN: Function `resolveWebviewUserDataPath`.
//
// RU: Что делает: вычисляет стабильный путь для данных WebView2, который не зависит от имени `.exe`.
// EN: What it does: resolves a stable WebView2 user data path that does not depend on the current executable name.
//
// RU: Ключевые моменты: использует один и тот же каталог `MercelData`; предотвращает создание новых папок при переименовании приложения; создаёт каталог заранее.
// EN: Key points: uses the fixed `MercelData` directory; prevents extra folders from appearing after renaming the executable; creates the directory ahead of time.
func resolveWebviewUserDataPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	webviewDir := filepath.Join(configDir, AppStorageDirName, "webview2")
	if err := os.MkdirAll(webviewDir, 0o755); err != nil {
		return "", err
	}

	return webviewDir, nil
}

// RU: Функция `main`.
// EN: Function `main`.
//
// RU: Что делает: выполняет вспомогательное преобразование, проверку или подготовку данных.
// EN: What it does: main bootstraps the backend application, wires it into Wails, and starts the desktop window lifecycle.
//
// RU: Ключевые моменты: важен для устойчивости логики; может использоваться сразу в нескольких местах; изменения стоит делать осознанно.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer app.Close()

	webviewUserDataPath, err := resolveWebviewUserDataPath()
	if err != nil {
		log.Fatalf("failed to resolve webview user data path: %v", err)
	}

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
			app.startup(ctx)
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}
