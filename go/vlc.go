//go:build windows

// VLC'yi CGO kullanmadan, libvlc.dll üzerinden çağıran saf Go wrapper.
// VLC kurulu olmalı (kayıt defterinden veya standart konumdan bulunur).

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var vlcLib *windows.LazyDLL

// findVLCDir Windows kayıt defterinden veya standart yollardan VLC kurulum klasörünü bulur
func findVLCDir() string {
	// Önce kayıt defterini dene (64-bit VLC)
	for _, regPath := range []string{
		`SOFTWARE\VideoLAN\VLC`,
		`SOFTWARE\WOW6432Node\VideoLAN\VLC`, // 32-bit VLC on 64-bit Windows
	} {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		dir, _, err := key.GetStringValue("InstallDir")
		key.Close()
		if err == nil && dir != "" {
			return dir
		}
	}

	// Kayıt defteri çalışmadıysa standart yolları dene
	for _, dir := range []string{
		`C:\Program Files\VideoLAN\VLC`,
		`C:\Program Files (x86)\VideoLAN\VLC`,
	} {
		if _, err := os.Stat(filepath.Join(dir, "libvlc.dll")); err == nil {
			return dir
		}
	}

	return ""
}

// loadVLCLib libvlc.dll'i yükler; önce PATH'e bakar, sonra kayıt defteri/standart konumları dener
func loadVLCLib() error {
	if vlcLib != nil {
		return nil // Zaten yüklendi
	}

	// Önce exe dizininde ara (DLL'leri yanına kopyalayan kullanıcılar için)
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	if _, err := os.Stat(filepath.Join(exeDir, "libvlc.dll")); err == nil {
		if err2 := tryLoadVLC(exeDir); err2 == nil {
			return nil
		}
	}

	// VLC kurulum klasörünü bul
	vlcDir := findVLCDir()
	if vlcDir == "" {
		return fmt.Errorf("VLC kurulum klasörü bulunamadı.\n\nLütfen VLC'yi kurun:\nhttps://www.videolan.org/vlc/")
	}

	return tryLoadVLC(vlcDir)
}

// tryLoadVLC belirtilen klasörden libvlc.dll'i yüklemeyi dener
func tryLoadVLC(dir string) error {
	dllPath := filepath.Join(dir, "libvlc.dll")
	if _, err := os.Stat(dllPath); err != nil {
		return fmt.Errorf("libvlc.dll bulunamadı: %s", dllPath)
	}

	// KRITIK: VLC'nin bağımlı DLL'lerini yükleyebilmesi için önce
	// VLC klasörünü DLL arama yoluna ekle
	dirPtr, err := syscall.UTF16PtrFromString(dir)
	if err == nil {
		kernel32 := windows.NewLazyDLL("kernel32.dll")
		setDllDir := kernel32.NewProc("SetDllDirectoryW")
		setDllDir.Call(uintptr(unsafe.Pointer(dirPtr)))
	}

	// VLC plugin klasörünü ayarla
	os.Setenv("VLC_PLUGIN_PATH", filepath.Join(dir, "plugins"))

	// Şimdi libvlc.dll'i yükle
	lib := windows.NewLazyDLL(dllPath)
	if err := lib.NewProc("libvlc_new").Find(); err != nil {
		return fmt.Errorf("libvlc.dll yüklenemedi (%s): %v", dir, err)
	}

	vlcLib = lib
	return nil
}

// vlcProc lazy proc döner
func vlcProc(name string) *windows.LazyProc {
	return vlcLib.NewProc(name)
}

// cStr Go string'ini null-terminated byte dizisine çevirir
func cStr(s string) []byte {
	return append([]byte(s), 0)
}

// VLCPlayer libvlc handle'larını tutar
type VLCPlayer struct {
	instance uintptr // libvlc_instance_t*
	player   uintptr // libvlc_media_player_t*
	media    uintptr // libvlc_media_t*
}

// NewVLCPlayer VLC instance ve player oluşturur
func NewVLCPlayer() (*VLCPlayer, error) {
	if err := loadVLCLib(); err != nil {
		return nil, err
	}

	// libvlc_new ile VLC instance oluştur
	args := []string{"--no-video-title-show", "--quiet", "--no-fullscreen"}
	argc := int32(len(args))
	rawStrs := make([][]byte, len(args))
	ptrs := make([]uintptr, len(args))
	for i, a := range args {
		rawStrs[i] = cStr(a)
		ptrs[i] = uintptr(unsafe.Pointer(&rawStrs[i][0]))
	}

	instance, _, _ := vlcProc("libvlc_new").Call(
		uintptr(argc),
		uintptr(unsafe.Pointer(&ptrs[0])),
	)
	if instance == 0 {
		return nil, fmt.Errorf("VLC başlatılamadı (libvlc_new başarısız)")
	}

	player, _, _ := vlcProc("libvlc_media_player_new").Call(instance)
	if player == 0 {
		vlcProc("libvlc_release").Call(instance)
		return nil, fmt.Errorf("VLC oynatıcı oluşturulamadı")
	}

	return &VLCPlayer{instance: instance, player: player}, nil
}

// SetHWND video çıkışını Win32 penceresine bağlar
func (v *VLCPlayer) SetHWND(hwnd uintptr) {
	vlcProc("libvlc_media_player_set_hwnd").Call(v.player, hwnd)
}

// PlayURL verilen URL'yi oynatır
func (v *VLCPlayer) PlayURL(url string) error {
	v.Stop()
	v.releaseMedia()

	urlBytes := cStr(url)
	media, _, _ := vlcProc("libvlc_media_new_location").Call(
		v.instance,
		uintptr(unsafe.Pointer(&urlBytes[0])),
	)
	if media == 0 {
		return fmt.Errorf("medya oluşturulamadı")
	}
	v.media = media

	vlcProc("libvlc_media_player_set_media").Call(v.player, media)
	ret, _, _ := vlcProc("libvlc_media_player_play").Call(v.player)
	if int32(ret) < 0 {
		return fmt.Errorf("oynatma başlatılamadı")
	}
	return nil
}

// Stop oynatmayı durdurur
func (v *VLCPlayer) Stop() {
	if v.player != 0 {
		vlcProc("libvlc_media_player_stop").Call(v.player)
	}
}

// SetFullscreen VLC'nin kendi tam ekranını açar/kapatır.
// Bazı VLC sürümlerinde bu fonksiyon bulunmayabilir; hata sessizce yutulur.
// (Tam ekran yönetimi Windows SW_MAXIMIZE ile yapıldığı için bu çağrıya
// gerek yoktur; yalnızca geriye dönük uyumluluk amacıyla bırakılmıştır.)
func (v *VLCPlayer) SetFullscreen(on bool) {
	proc := vlcLib.NewProc("libvlc_media_player_set_fullscreen")
	if err := proc.Find(); err != nil {
		// Bu VLC sürümünde fonksiyon yok — sorun değil, yoksay
		return
	}
	val := uintptr(0)
	if on {
		val = 1
	}
	proc.Call(v.player, val)
}

// Release tüm kaynakları serbest bırakır
func (v *VLCPlayer) Release() {
	v.Stop()
	v.releaseMedia()
	if v.player != 0 {
		vlcProc("libvlc_media_player_release").Call(v.player)
		v.player = 0
	}
	if v.instance != 0 {
		vlcProc("libvlc_release").Call(v.instance)
		v.instance = 0
	}
}

func (v *VLCPlayer) releaseMedia() {
	if v.media != 0 {
		vlcProc("libvlc_media_release").Call(v.media)
		v.media = 0
	}
}
