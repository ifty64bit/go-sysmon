// SysMon — a lightweight, cross-platform system resource monitor.
//
// Architecture overview:
//
//	Go (backend)
//	  └─ SysInfoService  — polls CPU, memory, disk, and network every second
//	                        and emits a "stats" event to the frontend.
//
//	Svelte (frontend, embedded in the binary)
//	  └─ App.svelte      — root shell that subscribes to "stats" events
//	       ├─ CpuCard    — ring gauge + per-core bars
//	       ├─ MemoryCard — usage bar + swap bar
//	       ├─ NetworkCard — live rates + sparkline graphs
//	       └─ DiskCard   — per-partition usage bars
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// assets embeds the compiled Svelte app (frontend/dist) into the binary so
// SysMon ships as a single, portable executable with no external dependencies.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	svc := &SysInfoService{}

	app := application.New(application.Options{
		Name:        "SysMon",
		Description: "Lightweight system resource monitor",
		Services: []application.Service{
			// Register SysInfoService so its exported methods are callable
			// from the frontend as typed RPC functions.
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			// On macOS, quit the process when the last window is closed.
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "SysMon",
		Width:            1280,
		Height:           820,
		MinWidth:         960,
		MinHeight:        640,
		BackgroundColour: application.NewRGB(10, 10, 15),
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		URL: "/",
	})

	// Run the polling loop in the background. It will call svc.setup() once,
	// then emit a "stats" event every PollInterval until the process exits.
	go svc.StartPolling(app)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
