package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/lxn/walk"
)

// AppLogger tüm uygulama boyunca kullanılan logger
var AppLogger *log.Logger

// setupLog exe dizininde iptvplayer.log dosyasını açar/oluşturur.
// Dosyayı döner; çağıran defer ile kapatmalıdır.
func setupLog() *os.File {
	exePath, err := os.Executable()
	if err != nil {
		return nil
	}
	logPath := filepath.Join(filepath.Dir(exePath), "iptvplayer.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}
	AppLogger = log.New(f, "", log.LstdFlags)
	AppLogger.Println("=== Uygulama başlatıldı ===")
	return f
}

func main() {
	// Log dosyasını kur
	logFile := setupLog()
	if logFile != nil {
		defer logFile.Close()
	}

	// Yakalanmayan panicler için kurtarma; log dosyasına yazar ve kullanıcıya gösterir
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("Kritik Hata (panic):\n%v\n\nStack:\n%s", r, debug.Stack())
			if AppLogger != nil {
				AppLogger.Println(msg)
			}
			walk.MsgBox(nil, "Kritik Hata", msg, walk.MsgBoxOK|walk.MsgBoxIconError)
		}
	}()

	// Çalışma dizinini exe'nin bulunduğu yere ayarla
	if exePath, err := os.Executable(); err == nil {
		os.Chdir(filepath.Dir(exePath))
	}

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
