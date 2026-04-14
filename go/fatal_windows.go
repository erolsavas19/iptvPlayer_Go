//go:build windows

package main

import "github.com/lxn/walk"

// showFatalError Windows'ta MsgBox ile kritik hata gösterir.
func showFatalError(title, msg string) {
	walk.MsgBox(nil, title, msg, walk.MsgBoxOK|walk.MsgBoxIconError)
}

// runApp Windows uygulamasını başlatır (walk mesaj döngüsü).
func runApp() {
	app, err := NewIPTVPlayer()
	if err != nil {
		if AppLogger != nil {
			AppLogger.Println("Başlatma hatası:", err)
		}
		walk.MsgBox(nil, "Başlatma Hatası", err.Error(), walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}
	defer app.Close()
	app.Run()
	if AppLogger != nil {
		AppLogger.Println("=== Uygulama kapandı ===")
	}
}
