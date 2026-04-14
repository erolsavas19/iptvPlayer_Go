package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// AppLogger tüm uygulama boyunca kullanılan logger
var AppLogger *log.Logger

// setupLog exe dizininde iptvplayer.log dosyasını açar/oluşturur.
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
	logFile := setupLog()
	if logFile != nil {
		defer logFile.Close()
	}

	// Yakalanmayan panicler için kurtarma
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("Kritik Hata (panic):\n%v\n\nStack:\n%s", r, debug.Stack())
			if AppLogger != nil {
				AppLogger.Println(msg)
			}
			showFatalError("Kritik Hata", msg)
		}
	}()

	// Çalışma dizinini exe'nin bulunduğu yere ayarla.
	// go run kullanıldığında exe geçici dizinde olur; o durumda chdir atlanır,
	// böylece favorites.db kaynak dizininde korunur.
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		if !strings.Contains(filepath.ToSlash(dir), "/tmp/") &&
			!strings.HasPrefix(dir, os.TempDir()) {
			os.Chdir(dir)
		}
	}

	runApp()
}
