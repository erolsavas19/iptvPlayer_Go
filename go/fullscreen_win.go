//go:build windows

package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

// subclassedWndProc alt sınıflandırılmış WndProc'u GC'den korumak için saklar
var subclassedWndProc uintptr

// user32'de tanımlı olmayan GetMenu için syscall
var procGetMenu = windows.NewLazyDLL("user32.dll").NewProc("GetMenu")

func getMenuHandle(hwnd win.HWND) win.HMENU {
	ret, _, _ := procGetMenu.Call(uintptr(hwnd))
	return win.HMENU(ret)
}

// savedWinState tam ekrana girmeden önceki pencere durumunu saklar
type savedWinState struct {
	style int32
	rect  win.RECT
	menu  win.HMENU
}

var winState *savedWinState

// enterTrueFullscreen başlık çubuğu, kenarlık ve menüyü kaldırarak
// pencereyi monitörü tam kaplayan borderless modda gösterir.
func enterTrueFullscreen(hwnd win.HWND) {
	// Mevcut durumu kaydet
	style := win.GetWindowLong(hwnd, win.GWL_STYLE)
	var rect win.RECT
	win.GetWindowRect(hwnd, &rect)
	menu := getMenuHandle(hwnd)
	winState = &savedWinState{style: style, rect: rect, menu: menu}

	// Pencerenin bulunduğu monitörü bul ve boyutlarını al
	monitor := win.MonitorFromWindow(hwnd, win.MONITOR_DEFAULTTONEAREST)
	var mi win.MONITORINFO
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	win.GetMonitorInfo(monitor, &mi)

	// Başlık çubuğu ve yeniden boyutlandırma kenarlığını kaldır
	newStyle := style &^ int32(win.WS_CAPTION|win.WS_THICKFRAME)
	win.SetWindowLong(hwnd, win.GWL_STYLE, newStyle)

	// Menü çubuğunu gizle
	win.SetMenu(hwnd, 0)
	win.DrawMenuBar(hwnd)

	// Monitörü tam kaplayan boyuta getir (üstte en)
	win.SetWindowPos(hwnd, win.HWND_TOP,
		mi.RcMonitor.Left, mi.RcMonitor.Top,
		mi.RcMonitor.Right-mi.RcMonitor.Left,
		mi.RcMonitor.Bottom-mi.RcMonitor.Top,
		win.SWP_FRAMECHANGED|win.SWP_NOACTIVATE)
}

// exitTrueFullscreen pencere stilini, menüsünü ve boyutunu geri yükler.
func exitTrueFullscreen(hwnd win.HWND) {
	if winState == nil {
		return
	}
	state := winState
	winState = nil

	// Stili geri yükle (başlık + kenarlık geri gelsin)
	win.SetWindowLong(hwnd, win.GWL_STYLE, state.style)

	// Menüyü geri yükle
	win.SetMenu(hwnd, state.menu)
	win.DrawMenuBar(hwnd)

	// Önceki boyut ve konumu geri yükle
	win.SetWindowPos(hwnd, 0,
		state.rect.Left, state.rect.Top,
		state.rect.Right-state.rect.Left,
		state.rect.Bottom-state.rect.Top,
		win.SWP_FRAMECHANGED|win.SWP_NOACTIVATE|win.SWP_NOZORDER)
}

// SubclassVideoComposite: videoComposite'in WndProc'unu alt sınıflandırır.
// VLC'nin içindeki pencere tıklandığında WM_PARENTNOTIFY gönderir;
// bunu yakalayarak tam ekranı tetikleriz. (Tek tıklama: WM_LBUTTONDOWN)
func SubclassVideoComposite(composite *walk.Composite, onClick func()) {
	hwnd := win.HWND(composite.Handle())

	var oldProc uintptr

	newProc := syscall.NewCallback(func(h win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
		if msg == win.WM_PARENTNOTIFY {
			childMsg := uint32(wParam & 0xFFFF)
			if childMsg == win.WM_LBUTTONDOWN {
				onClick()
				return 0
			}
		}
		return win.CallWindowProc(oldProc, h, msg, wParam, lParam)
	})

	subclassedWndProc = newProc // GC'den koru
	oldProc = win.SetWindowLongPtr(hwnd, win.GWLP_WNDPROC, newProc)
}

// setPanelsVisible tüm çevre panellerini gösterir veya gizler.
func setPanelsVisible(p *IPTVPlayer, visible bool) {
	panels := []*walk.Composite{
		p.topPanel, p.searchPanel, p.leftPanel,
		p.videoTitlePanel, p.bottomPanel, p.statusPanel,
	}
	for _, panel := range panels {
		if panel != nil {
			panel.SetVisible(visible)
		}
	}
}

// safeLog AppLogger varsa mesajı loglar
func safeLog(format string, args ...interface{}) {
	if AppLogger != nil {
		if len(args) == 0 {
			AppLogger.Println(format)
		} else {
			AppLogger.Println(fmt.Sprintf(format, args...))
		}
	}
}

// toggleFullscreen tam ekrana girer/çıkar.
//
// Strateji:
//   - Giriş: başlık/kenarlık/menüyü kaldır → monitörü kapla → panelleri gizle
//   - Çıkış: panelleri göster → stili/menüyü geri yükle → önceki boyuta dön
//
// win.ShowWindow(hwnd, ...) Synchronize içinden çağrılırsa walk'ın WndProc'u
// reentrant tetiklenir ve program çöker. Bu yüzden tüm pencere işlemleri
// doğrudan goroutine'den (cross-thread) yapılır; yalnızca walk widget
// değişiklikleri Synchronize ile UI thread'de çalıştırılır.
func toggleFullscreen(p *IPTVPlayer) {
	hwnd := win.HWND(p.mw.Handle())
	safeLog("toggleFullscreen: isFullscreen=%v", p.isFullscreen)

	if !p.isFullscreen {
		p.isFullscreen = true

		go func() {
			defer func() {
				if r := recover(); r != nil {
					safeLog("toggleFullscreen (enter) goroutine panic: %v", r)
				}
			}()

			// Adım 1: Walk panel değişiklikleri UI thread'de
			time.Sleep(30 * time.Millisecond)
			p.mw.Synchronize(func() {
				safeLog("Paneller gizleniyor")
				setPanelsVisible(p, false)
			})

			// Adım 2: Pencere chrome'unu kaldır ve tam ekrana geç (cross-thread)
			time.Sleep(80 * time.Millisecond)
			safeLog("enterTrueFullscreen çağrılıyor (cross-thread)")
			enterTrueFullscreen(hwnd)

			// Adım 3: VLC'yi yeni boyuta bağla
			time.Sleep(150 * time.Millisecond)
			p.mw.Synchronize(func() {
				rebindVLCDelayed(p)
			})
		}()

	} else {
		p.isFullscreen = false

		go func() {
			defer func() {
				if r := recover(); r != nil {
					safeLog("toggleFullscreen (exit) goroutine panic: %v", r)
				}
			}()

			// Adım 1: Pencere stilini/menüsünü geri yükle (cross-thread)
			time.Sleep(30 * time.Millisecond)
			safeLog("exitTrueFullscreen çağrılıyor (cross-thread)")
			exitTrueFullscreen(hwnd)

			// Adım 2: Walk panellerini geri getir (UI thread)
			time.Sleep(100 * time.Millisecond)
			p.mw.Synchronize(func() {
				safeLog("Paneller gösteriliyor")
				setPanelsVisible(p, true)
				rebindVLCDelayed(p)
			})
		}()
	}
}

// rebindVLCDelayed VLC render penceresini 200ms sonra yeniden bağlar.
func rebindVLCDelayed(p *IPTVPlayer) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				safeLog("rebindVLCDelayed goroutine panic: %v", r)
			}
		}()
		time.Sleep(200 * time.Millisecond)
		p.mw.Synchronize(func() {
			if p.videoComposite != nil && p.vlc != nil {
				hwnd := uintptr(p.videoComposite.Handle())
				p.vlc.SetHWND(hwnd)
				safeLog("VLC HWND yeniden bağlandı")
			}
		})
	}()
}

// paintVideoBlack video alanının arka planını siyah yapar.
func paintVideoBlack(composite *walk.Composite) {
	brush, err := walk.NewSolidColorBrush(walk.RGB(0, 0, 0))
	if err == nil {
		composite.SetBackground(brush)
	}
}

// isWindowVisible bir HWND'nin görünür olup olmadığını kontrol eder
func isWindowVisible(hwnd uintptr) bool {
	return win.IsWindowVisible(win.HWND(hwnd))
}
