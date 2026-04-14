//go:build linux

// vlc_linux.go — Linux için libvlc.so tabanlı gömülü video oynatıcı.
//
// VLC, her frame'i bir Go tamponuna render eder.
// Bu tampon Fyne canvas'ına kopyalanarak görüntü uygulama içinde gösterilir.
// Harici VLC penceresi açılmaz.
//
// Gereksinim: libvlc5 (vlc paketi yüklenince otomatik gelir)
//   sudo apt install vlc

package main

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"sync"
	"unsafe"

	"fyne.io/fyne/v2/canvas"
	"github.com/ebitengine/purego"
)

// ─── Frame boyutları ─────────────────────────────────────────────────────────

const (
	vlcW     = 1280
	vlcH     = 720
	vlcPitch = vlcW * 4 // RGBA: 4 bayt/piksel
)

// ─── libvlc.so fonksiyon tanımları ───────────────────────────────────────────

var vlcLibHandle uintptr

var (
	libvlcNew                 func(argc int32, argv uintptr) uintptr
	libvlcRelease             func(instance uintptr)
	libvlcMediaPlayerNew      func(instance uintptr) uintptr
	libvlcMediaPlayerRelease  func(player uintptr)
	libvlcMediaPlayerPlay     func(player uintptr) int32
	libvlcMediaPlayerStop     func(player uintptr)
	libvlcMediaNewLocation    func(instance, urlPtr uintptr) uintptr
	libvlcMediaRelease        func(media uintptr)
	libvlcMediaPlayerSetMedia func(player, media uintptr)
	libvlcVideoSetCallbacks   func(player, lockFn, unlockFn, displayFn, opaque uintptr)
	libvlcVideoSetFormat      func(player, chroma uintptr, w, h, pitch uint32)
)

// ─── Yapı ────────────────────────────────────────────────────────────────────

// VLCPlayer Linux için gömülü VLC player
type VLCPlayer struct {
	instance uintptr
	player   uintptr
	media    uintptr

	videoCanvas *canvas.Image

	frameMu  sync.Mutex
	writeBuf []byte        // VLC bu tampona yazar
	readyImg *image.NRGBA  // unlockCb'den sonra güvenle okunabilir

	refreshChan chan struct{}
	stopChan    chan struct{}

	// purego callback pointer'ları (GC koruması için alanda saklanır)
	lockCb    uintptr
	unlockCb  uintptr
	displayCb uintptr
}

// ─── libvlc.so yükleme ───────────────────────────────────────────────────────

func loadVLCLib() error {
	if vlcLibHandle != 0 {
		return nil
	}

	arch := runtime.GOARCH
	var libPaths []string
	switch arch {
	case "arm64":
		libPaths = []string{
			"libvlc.so.5",
			"/usr/lib/aarch64-linux-gnu/libvlc.so.5",
		}
	default: // amd64
		libPaths = []string{
			"libvlc.so.5",
			"libvlc.so",
			"/usr/lib/x86_64-linux-gnu/libvlc.so.5",
		}
	}

	// VLC plugin klasörünü ayarla
	for _, dir := range []string{
		"/usr/lib/x86_64-linux-gnu/vlc/plugins",
		"/usr/lib/aarch64-linux-gnu/vlc/plugins",
		"/usr/lib/vlc/plugins",
	} {
		if _, err := os.Stat(dir); err == nil {
			os.Setenv("VLC_PLUGIN_PATH", dir)
			break
		}
	}

	var lastErr error
	for _, p := range libPaths {
		lib, err := purego.Dlopen(p, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
		if err == nil && lib != 0 {
			vlcLibHandle = lib
			break
		}
		lastErr = err
	}

	if vlcLibHandle == 0 {
		return fmt.Errorf("libvlc.so yüklenemedi (%v)\nVLC kurulu mu? sudo apt install vlc", lastErr)
	}

	purego.RegisterLibFunc(&libvlcNew, vlcLibHandle, "libvlc_new")
	purego.RegisterLibFunc(&libvlcRelease, vlcLibHandle, "libvlc_release")
	purego.RegisterLibFunc(&libvlcMediaPlayerNew, vlcLibHandle, "libvlc_media_player_new")
	purego.RegisterLibFunc(&libvlcMediaPlayerRelease, vlcLibHandle, "libvlc_media_player_release")
	purego.RegisterLibFunc(&libvlcMediaPlayerPlay, vlcLibHandle, "libvlc_media_player_play")
	purego.RegisterLibFunc(&libvlcMediaPlayerStop, vlcLibHandle, "libvlc_media_player_stop")
	purego.RegisterLibFunc(&libvlcMediaNewLocation, vlcLibHandle, "libvlc_media_new_location")
	purego.RegisterLibFunc(&libvlcMediaRelease, vlcLibHandle, "libvlc_media_release")
	purego.RegisterLibFunc(&libvlcMediaPlayerSetMedia, vlcLibHandle, "libvlc_media_player_set_media")
	purego.RegisterLibFunc(&libvlcVideoSetCallbacks, vlcLibHandle, "libvlc_video_set_callbacks")
	purego.RegisterLibFunc(&libvlcVideoSetFormat, vlcLibHandle, "libvlc_video_set_format")

	return nil
}

// ─── Yapıcı ──────────────────────────────────────────────────────────────────

func NewVLCPlayer() (*VLCPlayer, error) {
	if err := loadVLCLib(); err != nil {
		return nil, err
	}

	args := []string{"--no-xlib", "--no-video-title-show", "--quiet"}
	argc := int32(len(args))

	cStrs := make([][]byte, len(args))
	ptrs := make([]uintptr, len(args))
	for i, a := range args {
		cStrs[i] = append([]byte(a), 0)
		ptrs[i] = uintptr(unsafe.Pointer(&cStrs[i][0]))
	}

	instance := libvlcNew(argc, uintptr(unsafe.Pointer(&ptrs[0])))
	if instance == 0 {
		return nil, fmt.Errorf("VLC instance başlatılamadı")
	}

	player := libvlcMediaPlayerNew(instance)
	if player == 0 {
		libvlcRelease(instance)
		return nil, fmt.Errorf("VLC player oluşturulamadı")
	}

	vp := &VLCPlayer{
		instance:    instance,
		player:      player,
		writeBuf:    make([]byte, vlcW*vlcH*4),
		readyImg:    image.NewNRGBA(image.Rect(0, 0, vlcW, vlcH)),
		refreshChan: make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
	}

	vp.setupCallbacks()

	// Frame'leri Fyne canvas'ına gönderen goroutine
	go func() {
		for {
			select {
			case <-vp.stopChan:
				return
			case <-vp.refreshChan:
				if vp.videoCanvas == nil {
					continue
				}
				// readyImg'den bağımsız bir kopya yap (çizim sırasında veri değişmesin)
				dst := image.NewNRGBA(image.Rect(0, 0, vlcW, vlcH))
				vp.frameMu.Lock()
				copy(dst.Pix, vp.readyImg.Pix)
				vp.frameMu.Unlock()
				vp.videoCanvas.Image = dst
				vp.videoCanvas.Refresh()
			}
		}
	}()

	return vp, nil
}

func (vp *VLCPlayer) setupCallbacks() {
	// Lock: VLC tampon adresini alır
	vp.lockCb = purego.NewCallback(func(opaque, planesPtr uintptr) uintptr {
		vp.frameMu.Lock()
		buf := uintptr(unsafe.Pointer(&vp.writeBuf[0]))
		*(*uintptr)(unsafe.Pointer(planesPtr)) = buf
		return buf
	})

	// Unlock: VLC tampona yazmayı bitirdi; readyImg'i güncelle
	vp.unlockCb = purego.NewCallback(func(opaque, picture, planes uintptr) {
		copy(vp.readyImg.Pix, vp.writeBuf)
		vp.frameMu.Unlock()
	})

	// Display: frame hazır, refresh kanalına sinyal gönder
	vp.displayCb = purego.NewCallback(func(opaque, picture uintptr) {
		select {
		case vp.refreshChan <- struct{}{}:
		default: // bir önceki frame henüz işlenmediyse atla
		}
	})

	libvlcVideoSetCallbacks(vp.player, vp.lockCb, vp.unlockCb, vp.displayCb, 0)

	// RGBA: 4 bayt/piksel, Go image.NRGBA ile doğrudan uyumlu
	chroma := []byte("RGBA\x00")
	libvlcVideoSetFormat(vp.player,
		uintptr(unsafe.Pointer(&chroma[0])),
		vlcW, vlcH, vlcPitch)
}

// ─── Arayüz metodları ────────────────────────────────────────────────────────

// SetVideoCanvas Fyne canvas'ını VLC render hedefi olarak ayarlar
func (vp *VLCPlayer) SetVideoCanvas(c *canvas.Image) {
	vp.videoCanvas = c
}

func (vp *VLCPlayer) SetHWND(_ uintptr) {}

func (vp *VLCPlayer) PlayURL(url string) error {
	vp.Stop()
	vp.releaseMedia()

	cURL := append([]byte(url), 0)
	media := libvlcMediaNewLocation(vp.instance, uintptr(unsafe.Pointer(&cURL[0])))
	if media == 0 {
		return fmt.Errorf("medya oluşturulamadı: %s", url)
	}
	vp.media = media

	libvlcMediaPlayerSetMedia(vp.player, media)
	if ret := libvlcMediaPlayerPlay(vp.player); ret < 0 {
		return fmt.Errorf("oynatma başlatılamadı")
	}
	return nil
}

func (vp *VLCPlayer) Stop() {
	if vp.player != 0 {
		libvlcMediaPlayerStop(vp.player)
	}
}

func (vp *VLCPlayer) SetFullscreen(_ bool) {}

func (vp *VLCPlayer) Release() {
	close(vp.stopChan)
	vp.Stop()
	vp.releaseMedia()
	if vp.player != 0 {
		libvlcMediaPlayerRelease(vp.player)
		vp.player = 0
	}
	if vp.instance != 0 {
		libvlcRelease(vp.instance)
		vp.instance = 0
	}
}

func (vp *VLCPlayer) releaseMedia() {
	if vp.media != 0 {
		libvlcMediaRelease(vp.media)
		vp.media = 0
	}
}
