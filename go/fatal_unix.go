//go:build linux || darwin

package main

import (
	"fmt"
	"os"
)

// showFatalError Linux/macOS'ta stderr'e kritik hata yazar.
func showFatalError(title, msg string) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", title, msg)
}

// runApp Linux/macOS uygulamasını başlatır (Fyne mesaj döngüsü).
func runApp() {
	app, err := NewIPTVPlayer()
	if err != nil {
		if AppLogger != nil {
			AppLogger.Println("Başlatma hatası:", err)
		}
		showFatalError("Başlatma Hatası", err.Error())
		return
	}
	defer app.Close()
	app.Run()
	if AppLogger != nil {
		AppLogger.Println("=== Uygulama kapandı ===")
	}
}
