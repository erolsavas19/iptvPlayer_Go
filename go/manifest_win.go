//go:build windows

// manifest_win.go - Windows Common Controls v6 ve DPI farkındalığını
// çalışma zamanında etkinleştirir (rsrc.syso / harici araç gerektirmez).

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// manifestXML uygulamanın Windows manifest içeriği
const manifestXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <assemblyIdentity version="1.0.0.0" processorArchitecture="amd64" name="iptvplayer" type="win32"/>
  <dependency>
    <dependentAssembly>
      <assemblyIdentity type="win32" name="Microsoft.Windows.Common-Controls"
        version="6.0.0.0" processorArchitecture="amd64"
        publicKeyToken="6595b64144ccf1df" language="*"/>
    </dependentAssembly>
  </dependency>
  <application xmlns="urn:schemas-microsoft-com:asm.v3">
    <windowsSettings>
      <dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">PerMonitorV2, PerMonitor</dpiAwareness>
      <dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">True</dpiAware>
    </windowsSettings>
  </application>
</assembly>`

// ACTCTXW Win32 ACTCTXW yapısı
type ACTCTXW struct {
	cbSize                 uint32
	dwFlags                uint32
	lpSource               *uint16
	wProcessorArchitecture uint16
	wLangId                uint16
	lpAssemblyDirectory    *uint16
	lpResourceName         *uint16
	lpApplicationName      *uint16
	hModule                windows.Handle
}

var (
	kernel32         = windows.NewLazyDLL("kernel32.dll")
	user32dll2       = windows.NewLazyDLL("user32.dll")
	createActCtxW    = kernel32.NewProc("CreateActCtxW")
	activateActCtxFn = kernel32.NewProc("ActivateActCtx")
	setDPIAware      = user32dll2.NewProc("SetProcessDPIAware")
)

// activateManifest: manifest XML'ini geçici dosyaya yazar,
// CreateActCtx + ActivateActCtx ile Common Controls v6'yı etkinleştirir.
// Ayrıca DPI farkındalığını ayarlar.
func activateManifest() {
	// DPI farkındalığını ayarla (eski API, tüm Windows sürümlerinde çalışır)
	setDPIAware.Call()

	// Manifest'i geçici bir dosyaya yaz
	tmp := filepath.Join(os.TempDir(), "iptvplayer_manifest.xml")
	if err := os.WriteFile(tmp, []byte(manifestXML), 0600); err != nil {
		return
	}
	// Fonksiyon bitince sil
	defer os.Remove(tmp)

	srcPtr, err := syscall.UTF16PtrFromString(tmp)
	if err != nil {
		return
	}

	ctx := ACTCTXW{
		cbSize:  uint32(unsafe.Sizeof(ACTCTXW{})),
		dwFlags: 0,
		lpSource: srcPtr,
	}

	const INVALID_HANDLE_VALUE = ^uintptr(0)

	handle, _, _ := createActCtxW.Call(uintptr(unsafe.Pointer(&ctx)))
	if handle == 0 || handle == INVALID_HANDLE_VALUE {
		return
	}

	var cookie uintptr
	activateActCtxFn.Call(handle, uintptr(unsafe.Pointer(&cookie)))
	// Not: Aktivasyonu kapatmıyoruz — uygulama boyunca etkin kalmalı
}

func init() {
	// OS thread'ini kilitle: Go zamanlayıcısı goroutine'i farklı bir OS thread'ine
	// taşırsa ActivateActCtx'in etkinleştirdiği bağlam kaybolur.
	// runtime.LockOSThread() bunu önler; walk da kendi init'inde bunu yapıyor
	// ancak bu çağrıdan sonra — aradan önce thread değişebilir.
	runtime.LockOSThread()
	activateManifest()
}
