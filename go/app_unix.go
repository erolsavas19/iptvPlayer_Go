//go:build linux || darwin

// app_unix.go — Linux ve macOS için Fyne tabanlı IPTV Player.
//
// Windows sürümü (app.go + mainwindow.go) lxn/walk kullanır.
// Bu dosya aynı mantığı Fyne framework'ü ile yeniden uygular.
// İki sürüm arasında hiçbir paylaşılan UI tipi yoktur;
// build tag'ları derleme sırasında yalnızca birini seçer.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ─── Tipler ──────────────────────────────────────────────────────────────────

// EPGProgram tek EPG programını temsil eder
type EPGProgram struct {
	Title string
	Start time.Time
	End   time.Time
	Desc  string
}

// IPTVPlayer ana uygulama yapısı (Linux/macOS — Fyne)
type IPTVPlayer struct {
	fyneApp fyne.App
	window  fyne.Window

	vlc *VLCPlayer
	db  *DB

	channels         []Channel
	filteredChannels []Channel
	currentIndex int
	isFullscreen bool
	fsWindow     fyne.Window // tam ekran için ayrı pencere (ana pencere dokunulmaz)

	logoCache   map[string][]byte
	logoCacheMu sync.Mutex

	epgData map[string][]EPGProgram

	autoURLEnabled bool
	autoURL        string

	// Fyne UI bileşenleri
	channelList     *widget.List
	searchEntry     *widget.Entry
	categorySelect  *widget.Select
	statusLabel     *widget.Label
	countLabel      *widget.Label
	logoImage       *canvas.Image
	logoBox         *fyne.Container
	playerTitle *widget.Label // "OYNATICI (Kanal Adı)" başlık etiketi
	videoCanvas *canvas.Image // VLC frame'leri buraya render edilir (Linux)

	// Menü öğeleri (checkmark güncellemesi için)
	langMenuTR    *fyne.MenuItem
	langMenuEN    *fyne.MenuItem
	autoURLMenuItem *fyne.MenuItem

	// Kategori listesi (refresh için)
	categories []string
}

// ─── Başlatma ────────────────────────────────────────────────────────────────

// NewIPTVPlayer yeni bir IPTVPlayer oluşturur
func NewIPTVPlayer() (*IPTVPlayer, error) {
	db, err := NewDB("favorites.db")
	if err != nil {
		return nil, fmt.Errorf("veritabanı açılamadı: %w", err)
	}

	vlcPlayer, err := NewVLCPlayer()
	if err != nil {
		db.Close()
		return nil, err
	}

	p := &IPTVPlayer{
		vlc:          vlcPlayer,
		db:           db,
		currentIndex: -1,
		logoCache:    make(map[string][]byte),
		epgData:      make(map[string][]EPGProgram),
	}

	if url, ok := db.GetAutoURL(); ok {
		p.autoURLEnabled = true
		p.autoURL = url
	}

	if db.GetLanguage() == "en" {
		Lang = &LangEN
	} else {
		Lang = &LangTR
	}

	return p, nil
}

// Close kaynakları serbest bırakır
func (p *IPTVPlayer) Close() {
	if p.vlc != nil {
		p.vlc.Release()
	}
	if p.db != nil {
		p.db.Close()
	}
}

// Run Fyne uygulamasını başlatır
func (p *IPTVPlayer) Run() {
	p.fyneApp = fyneapp.New()
	p.fyneApp.Settings().SetTheme(theme.DarkTheme())

	title := fmt.Sprintf("IPTV Player (Ver. 1.0) — %s/%s", runtime.GOOS, runtime.GOARCH)
	p.window = p.fyneApp.NewWindow(title)
	p.window.Resize(fyne.NewSize(WindowWidth, WindowHeight))
	p.window.SetMaster()

	// Pencere ikonu: go run → go/ klasöründe, derlenmiş binary → go/ alt klasöründe
	for _, iconPath := range []string{"iptvPlayer.ico", "go/iptvPlayer.ico"} {
		if data, err := os.ReadFile(iconPath); err == nil {
			p.window.SetIcon(fyne.NewStaticResource("icon", data))
			break
		}
	}

	p.window.SetContent(p.buildMainContent())
	p.buildMenu()

	// Linux: VLC frame'lerini Fyne canvas'ına bağla
	p.vlc.SetVideoCanvas(p.videoCanvas)

	// ESC tuşu ile video tam ekrandan çık
	p.window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyEscape && p.isFullscreen {
			p.toggleFullscreen()
		}
	})

	p.window.SetOnClosed(func() {
		if p.vlc != nil {
			p.vlc.Stop()
		}
	})

	if p.autoURLEnabled && p.autoURL != "" {
		go p.loadAutoURL()
	}

	if AppLogger != nil {
		AppLogger.Printf("Fyne uygulaması başlatıldı (%s)", runtime.GOOS)
	}

	p.window.ShowAndRun()
}

// ─── Menü ────────────────────────────────────────────────────────────────────

// ms (menuStr) lxn/walk'a özgü & akseleratörünü ve \t kısayolunu kaldırır.
// "&Dosya\tCtrl+O" → "Dosya"
func ms(s string) string {
	s = strings.ReplaceAll(s, "&", "")
	if i := strings.Index(s, "\t"); i >= 0 {
		s = s[:i]
	}
	return s
}

// shortcut — Ctrl+key kısayolu oluşturur
func shortcut(key fyne.KeyName) *desktop.CustomShortcut {
	return &desktop.CustomShortcut{KeyName: key, Modifier: fyne.KeyModifierControl}
}

func (p *IPTVPlayer) buildMenu() {
	// ── Dosya menüsü ──────────────────────────────────────────────────
	miOpen := fyne.NewMenuItem(ms(Lang.MenuFileOpen), p.openFile)
	miOpen.Shortcut = shortcut(fyne.KeyO)

	miURL := fyne.NewMenuItem(ms(Lang.MenuURLOpen), p.openURL)
	miURL.Shortcut = shortcut(fyne.KeyU)

	miPlay := fyne.NewMenuItem(ms(Lang.MenuPlay), p.play)
	miPlay.Shortcut = shortcut(fyne.KeyP)

	miStop := fyne.NewMenuItem(ms(Lang.MenuStop), p.stop)
	miStop.Shortcut = shortcut(fyne.KeyS)

	miEPG := fyne.NewMenuItem(ms(Lang.MenuEPG), p.showEPG)
	miEPG.Shortcut = shortcut(fyne.KeyE)

	fileMenu := fyne.NewMenu(ms(Lang.MenuFile),
		miOpen, miURL,
		fyne.NewMenuItemSeparator(),
		miPlay, miStop,
		fyne.NewMenuItemSeparator(),
		miEPG,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(ms(Lang.MenuExit), func() { p.window.Close() }),
	)

	// ── Görünüm menüsü ────────────────────────────────────────────────
	miFS := fyne.NewMenuItem(ms(Lang.MenuFullscreen), p.toggleFullscreen)
	miFS.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF11}

	isTR := Lang == &LangTR
	p.langMenuTR = fyne.NewMenuItem(ms(Lang.MenuLangTR), func() { p.setLanguage("tr") })
	p.langMenuTR.Checked = isTR
	p.langMenuEN = fyne.NewMenuItem(ms(Lang.MenuLangEN), func() { p.setLanguage("en") })
	p.langMenuEN.Checked = !isTR

	langSubMenu := fyne.NewMenu(ms(Lang.MenuLanguage), p.langMenuTR, p.langMenuEN)
	langItem := fyne.NewMenuItem(ms(Lang.MenuLanguage), nil)
	langItem.ChildMenu = langSubMenu

	p.autoURLMenuItem = fyne.NewMenuItem(ms(Lang.MenuAutoURL), p.toggleAutoURL)
	p.autoURLMenuItem.Checked = p.autoURLEnabled && p.autoURL != ""

	viewMenu := fyne.NewMenu(ms(Lang.MenuView),
		p.autoURLMenuItem,
		fyne.NewMenuItemSeparator(),
		miFS,
		fyne.NewMenuItemSeparator(),
		langItem,
	)

	// ── Yardım menüsü ─────────────────────────────────────────────────
	helpMenu := fyne.NewMenu(ms(Lang.MenuHelp),
		fyne.NewMenuItem(ms(Lang.MenuAbout), p.showAbout),
	)

	p.window.SetMainMenu(fyne.NewMainMenu(fileMenu, viewMenu, helpMenu))
}

// ─── Çift Tıklama ile Tam Ekran ───────────────────────────────────────────────

// tapDetector video alanı üzerine saydam katman olarak eklenir.
// Çift tıklamayı yakalar → toggleFullscreen çağırır.
type tapDetector struct {
	widget.BaseWidget
	onDoubleTap func()
}

func newTapDetector(onDoubleTap func()) *tapDetector {
	t := &tapDetector{onDoubleTap: onDoubleTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tapDetector) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(bg)
}

func (t *tapDetector) DoubleTapped(_ *fyne.PointEvent) {
	if t.onDoubleTap != nil {
		t.onDoubleTap()
	}
}

// ─── Ana Pencere Düzeni ───────────────────────────────────────────────────────

func (p *IPTVPlayer) buildMainContent() fyne.CanvasObject {
	// ── Arama entry (genişleyen) ───────────────────────────────────────
	p.searchEntry = widget.NewEntry()
	p.searchEntry.SetPlaceHolder(Lang.SearchCue)
	p.searchEntry.OnChanged = func(_ string) { p.filterChannels() }

	// ── Kategori seçici (sabit genişlik, sağ tarafta) ──────────────────
	p.categorySelect = widget.NewSelect([]string{Lang.CategoryAll}, func(_ string) {
		p.filterChannels()
	})
	p.categorySelect.SetSelected(Lang.CategoryAll)

	// Arama çubuğu: [Ara:] [=======entry genişler=======] [Kategori:] [select]
	// NewBorder: sol=sabit etiket, sağ=sabit {etiket+select}, orta=genişleyen entry
	searchBar := container.NewBorder(
		nil, nil,
		widget.NewLabel(Lang.LabelSearch),
		container.NewHBox(widget.NewLabel(Lang.LabelCategory), p.categorySelect),
		p.searchEntry,
	)

	// ── Kanal listesi ──────────────────────────────────────────────────
	p.channelList = widget.NewList(
		func() int { return len(p.filteredChannels) },
		func() fyne.CanvasObject {
			return widget.NewLabel("kanal")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if int(id) < len(p.filteredChannels) {
				obj.(*widget.Label).SetText(p.filteredChannels[id].Name)
			}
		},
	)
	p.channelList.OnSelected = func(id widget.ListItemID) {
		p.onChannelSelected(int(id))
	}

	// ── Logo görüntüleyici ─────────────────────────────────────────────
	p.logoImage = &canvas.Image{}
	p.logoImage.FillMode = canvas.ImageFillContain
	p.logoImage.SetMinSize(fyne.NewSize(180, 80))
	p.logoBox = container.NewVBox(p.logoImage)
	p.logoBox.Hide()

	channelsTitle := widget.NewLabelWithStyle(
		Lang.LabelChannels, fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
	)
	leftPanel := container.NewBorder(channelsTitle, p.logoBox, nil, nil, p.channelList)

	// ── Video alanı (sağ panel) ────────────────────────────────────────
	// Siyah arka plan
	bg := canvas.NewRectangle(color.Black)

	// VLC render hedefi — başlangıçta siyah boş frame
	p.videoCanvas = canvas.NewImageFromImage(
		image.NewNRGBA(image.Rect(0, 0, 1280, 720)),
	)
	p.videoCanvas.FillMode = canvas.ImageFillContain

	// Çift tıklama ile tam ekrana geçiş (her platformda)
	tapLayer := newTapDetector(p.toggleFullscreen)

	var videoArea fyne.CanvasObject
	if runtime.GOOS == "darwin" {
		// macOS: VLC kendi penceresinde açılır; sadece siyah ekran + tap katmanı
		videoArea = container.NewStack(bg, tapLayer)
	} else {
		// Linux: VLC frame'leri canvas'a render edilir + tap katmanı üstte
		videoArea = container.NewStack(bg, p.videoCanvas, tapLayer)
	}

	// Başlık etiketi: "OYNATICI" → oynatınca "OYNATICI (Kanal Adı)"
	p.playerTitle = widget.NewLabelWithStyle(
		Lang.LabelPlayer, fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
	)
	rightPanel := container.NewBorder(p.playerTitle, nil, nil, nil, videoArea)

	// ── Bölünmüş görünüm ──────────────────────────────────────────────
	split := container.NewHSplit(leftPanel, rightPanel)
	split.SetOffset(0.30)

	// ── Üst çubuk: başlık solda, butonlar sağda ───────────────────────
	// NewBorder: sol=başlık, sağ={butonlar}, orta=boşluk
	topBar := container.NewBorder(
		nil, nil,
		widget.NewLabelWithStyle("📺  IPTV Player", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(
			widget.NewButton(Lang.BtnFileOpen, p.openFile),
			widget.NewButton(Lang.BtnURLOpen, p.openURL),
		),
		nil,
	)

	// ── Alt butonlar ──────────────────────────────────────────────────
	bottomBar := container.NewHBox(
		widget.NewButton(Lang.BtnPlay, p.play),
		widget.NewButton(Lang.BtnStop, p.stop),
		widget.NewButton(Lang.BtnEPG, p.showEPG),
		widget.NewButton(Lang.BtnAddFav, p.addToFavorites),
		widget.NewButton(Lang.BtnFavorites, p.showFavorites),
		widget.NewButton(Lang.BtnFullscreen, p.toggleFullscreen),
		layout.NewSpacer(),
		widget.NewButton(Lang.BtnExit, func() { p.window.Close() }),
	)

	// ── Durum çubuğu ──────────────────────────────────────────────────
	p.statusLabel = widget.NewLabel(Lang.StatusReady)
	p.countLabel = widget.NewLabelWithStyle(
		fmt.Sprintf(Lang.StatusTotal, 0),
		fyne.TextAlignTrailing,
		fyne.TextStyle{Bold: true},
	)
	statusBar := container.NewBorder(nil, nil, p.statusLabel, p.countLabel, nil)

	// ── Üst / alt bölümler ────────────────────────────────────────────
	topSection := container.NewVBox(
		topBar,
		widget.NewSeparator(),
		searchBar,
		widget.NewSeparator(),
	)
	bottomSection := container.NewVBox(
		widget.NewSeparator(),
		bottomBar,
		widget.NewSeparator(),
		statusBar,
	)

	return container.NewBorder(topSection, bottomSection, nil, nil, split)
}

// ─── Dil ─────────────────────────────────────────────────────────────────────

func (p *IPTVPlayer) setLanguage(code string) {
	if code == "en" {
		Lang = &LangEN
	} else {
		Lang = &LangTR
	}
	p.db.SetLanguage(code)
	p.setStatus(Lang.StatusReady)
	// Tam ekrandaysa önce kapat
	if p.isFullscreen {
		p.isFullscreen = false
		if p.fsWindow != nil {
			p.fsWindow.Close()
			p.fsWindow = nil
		}
		p.vlc.SetVideoCanvas(p.videoCanvas)
	}
	p.window.SetContent(p.buildMainContent())
	p.buildMenu()
	// Kanal listesini geri yükle
	if len(p.channels) > 0 {
		p.refreshChannelList()
	}
}

// ─── M3U Yükleme ─────────────────────────────────────────────────────────────

func (p *IPTVPlayer) loadM3ULines(lines []string) {
	channels, cats := ParseM3U(lines)
	p.channels = channels
	p.categories = cats
	p.filteredChannels = make([]Channel, len(channels))
	copy(p.filteredChannels, channels)
	p.currentIndex = -1

	allCats := append([]string{Lang.CategoryAll}, cats...)

	p.categorySelect.Options = allCats
	p.categorySelect.SetSelected(Lang.CategoryAll)
	p.categorySelect.Refresh()

	p.channelList.Refresh()
	p.countLabel.SetText(fmt.Sprintf(Lang.StatusTotal, len(channels)))
}

func (p *IPTVPlayer) refreshChannelList() {
	p.filteredChannels = make([]Channel, len(p.channels))
	copy(p.filteredChannels, p.channels)
	allCats := append([]string{Lang.CategoryAll}, p.categories...)
	p.categorySelect.Options = allCats
	p.categorySelect.SetSelected(Lang.CategoryAll)
	p.categorySelect.Refresh()
	p.channelList.Refresh()
	p.countLabel.SetText(fmt.Sprintf(Lang.StatusTotal, len(p.channels)))
}

// filterChannels arama + kategori filtresi uygular
func (p *IPTVPlayer) filterChannels() {
	// Widget'lar henüz oluşturulmamış olabilir (buildMainContent sırasında)
	if p.channelList == nil || p.searchEntry == nil || p.categorySelect == nil {
		return
	}
	searchText := strings.ToLower(p.searchEntry.Text)
	cat := p.categorySelect.Selected

	result := make([]Channel, 0, len(p.channels))
	for _, ch := range p.channels {
		nameOK := searchText == "" || strings.Contains(strings.ToLower(ch.Name), searchText)
		catOK := cat == "" || cat == Lang.CategoryAll || ch.Group == cat
		if nameOK && catOK {
			result = append(result, ch)
		}
	}
	p.filteredChannels = result
	p.channelList.Refresh()
}

// onChannelSelected listeden kanal seçildiğinde çağrılır
func (p *IPTVPlayer) onChannelSelected(idx int) {
	if idx < 0 || idx >= len(p.filteredChannels) {
		return
	}
	p.currentIndex = idx
	ch := p.filteredChannels[idx]
	p.setStatus(Lang.MsgSelected + ch.Name)

	if ch.Logo != "" {
		go p.loadLogo(ch.Logo)
	} else {
		p.logoBox.Hide()
		p.logoBox.Refresh()
	}
}

// loadLogo kanal logosunu URL'den indirir ve gösterir
func (p *IPTVPlayer) loadLogo(logoURL string) {
	p.logoCacheMu.Lock()
	data, cached := p.logoCache[logoURL]
	p.logoCacheMu.Unlock()

	if !cached {
		resp, err := http.Get(logoURL)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		data, _ = io.ReadAll(resp.Body)
		p.logoCacheMu.Lock()
		p.logoCache[logoURL] = data
		p.logoCacheMu.Unlock()
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}

	fyneImg := canvas.NewImageFromImage(img)
	fyneImg.FillMode = canvas.ImageFillContain
	fyneImg.SetMinSize(fyne.NewSize(180, 80))

	p.logoImage = fyneImg
	p.logoBox.Objects = []fyne.CanvasObject{fyneImg}
	p.logoBox.Show()
	p.logoBox.Refresh()
}

// ─── Oynatma ─────────────────────────────────────────────────────────────────

func (p *IPTVPlayer) play() {
	if p.currentIndex < 0 || p.currentIndex >= len(p.filteredChannels) {
		dialog.ShowInformation(Lang.DlgWarning, Lang.MsgSelectChannel, p.window)
		return
	}
	ch := p.filteredChannels[p.currentIndex]
	if err := p.vlc.PlayURL(ch.URL); err != nil {
		p.setStatus(Lang.MsgError + err.Error())
		return
	}
	p.setStatus(Lang.MsgPlaying + ch.Name)
	p.setNowPlaying("▶  " + ch.Name)
}

func (p *IPTVPlayer) playURL(url, name string) {
	if err := p.vlc.PlayURL(url); err != nil {
		p.setStatus(Lang.MsgError + err.Error())
		return
	}
	p.setStatus(Lang.MsgPlaying + name)
	p.setNowPlaying("▶  " + name)
}

func (p *IPTVPlayer) stop() {
	p.vlc.Stop()
	p.setStatus(Lang.MsgStopped)
	p.setNowPlaying("—")
}

// setNowPlaying başlık etiketini günceller: "▶  Kanal" → "OYNATICI (Kanal)", "—" → "OYNATICI"
func (p *IPTVPlayer) setNowPlaying(text string) {
	if p.playerTitle == nil {
		return
	}
	if strings.HasPrefix(text, "▶") {
		name := strings.TrimSpace(strings.TrimPrefix(text, "▶"))
		p.playerTitle.SetText(Lang.LabelPlayer + " (" + name + ")")
	} else {
		p.playerTitle.SetText(Lang.LabelPlayer)
	}
}

// ─── Video Tam Ekran ──────────────────────────────────────────────────────────
//
// Ana pencereye hiç dokunulmaz — pozisyon/boyut sorunu bu sayede tamamen ortadan
// kalkar. Tam ekran için ayrı bir pencere (fsWindow) açılır; VLC render hedefi
// fsCanvas'a yönlendirilir. Çıkışta fsWindow kapatılır, VLC ana canvas'a döner.

func (p *IPTVPlayer) toggleFullscreen() {
	p.isFullscreen = !p.isFullscreen

	if p.isFullscreen {
		// Tam ekran penceresi için yeni bir canvas oluştur
		fsCanvas := canvas.NewImageFromImage(
			image.NewNRGBA(image.Rect(0, 0, vlcW, vlcH)),
		)
		fsCanvas.FillMode = canvas.ImageFillContain

		// VLC'yi bu canvas'a yönlendir
		p.vlc.SetVideoCanvas(fsCanvas)

		// Yeni tam ekran penceresi
		p.fsWindow = p.fyneApp.NewWindow("")
		p.fsWindow.SetPadded(false)
		p.fsWindow.SetContent(container.NewStack(
			canvas.NewRectangle(color.Black),
			fsCanvas,
			newTapDetector(p.toggleFullscreen),
		))

		// ESC tuşu
		p.fsWindow.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
			if ev.Name == fyne.KeyEscape {
				p.toggleFullscreen()
			}
		})

		// Pencere kapatılırsa da düzgünce çık
		p.fsWindow.SetOnClosed(func() {
			if p.isFullscreen {
				p.isFullscreen = false
				p.vlc.SetVideoCanvas(p.videoCanvas)
				p.fsWindow = nil
			}
		})

		p.fsWindow.SetFullScreen(true)
		p.fsWindow.Show()
	} else {
		// VLC'yi ana canvas'a geri döndür, tam ekran penceresini kapat
		p.vlc.SetVideoCanvas(p.videoCanvas)
		if p.fsWindow != nil {
			p.fsWindow.SetOnClosed(nil) // çift tetiklenmeyi önle
			p.fsWindow.Close()
			p.fsWindow = nil
		}
		p.window.RequestFocus()
	}
}

// ─── Dosya / URL ─────────────────────────────────────────────────────────────

func (p *IPTVPlayer) openFile() {
	fd := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
		if err != nil || f == nil {
			return
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			dialog.ShowError(fmt.Errorf(Lang.DlgFileReadFail+err.Error()), p.window)
			return
		}
		lines := strings.Split(string(data), "\n")
		p.loadM3ULines(lines)
		p.setStatus(Lang.MsgLoaded + f.URI().Path())
	}, p.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".m3u", ".m3u8"}))
	fd.Show()
}

func (p *IPTVPlayer) openURL() {
	showInputDialog(p.window, Lang.DlgOpenURL, Lang.DlgOpenURLPrompt, "", func(url string) {
		if url == "" {
			return
		}
		p.setStatus(Lang.MsgLoadingURL)
		go func() {
			resp, err := http.Get(url)
			if err != nil {
				p.setStatusAsync(Lang.MsgError + err.Error())
				return
			}
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			lines := strings.Split(string(data), "\n")
			p.window.Canvas().Refresh(p.window.Canvas().Content())
			p.loadM3ULines(lines)
			p.setStatus(Lang.MsgURLLoaded)
		}()
	})
}

func (p *IPTVPlayer) loadAutoURL() {
	p.setStatusAsync(Lang.MsgAutoURLLoading)
	resp, err := http.Get(p.autoURL)
	if err != nil {
		p.setStatusAsync(Lang.MsgAutoURLFailed + err.Error())
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	lines := strings.Split(string(data), "\n")
	p.loadM3ULines(lines)
	p.setStatusAsync(Lang.MsgAutoURLLoaded)
}

// ─── Favoriler ───────────────────────────────────────────────────────────────

func (p *IPTVPlayer) addToFavorites() {
	if p.currentIndex < 0 || p.currentIndex >= len(p.filteredChannels) {
		dialog.ShowInformation(Lang.DlgWarning, Lang.MsgSelectChannel, p.window)
		return
	}
	ch := p.filteredChannels[p.currentIndex]
	added, err := p.db.AddFavorite(ch.Name, ch.URL)
	if err != nil {
		dialog.ShowError(fmt.Errorf(Lang.DlgFavAddFail+err.Error()), p.window)
		return
	}
	if !added {
		dialog.ShowInformation(Lang.DlgInfo, fmt.Sprintf(Lang.DlgFavExists, ch.Name), p.window)
		return
	}
	dialog.ShowInformation(Lang.DlgSuccess, fmt.Sprintf(Lang.DlgFavAdded, ch.Name), p.window)
}

func (p *IPTVPlayer) showFavorites() {
	ShowFavoritesWindow(p)
}

// ─── EPG ─────────────────────────────────────────────────────────────────────

func (p *IPTVPlayer) showEPG() {
	if p.currentIndex < 0 || p.currentIndex >= len(p.filteredChannels) {
		dialog.ShowInformation(Lang.DlgWarning, Lang.MsgSelectChannel, p.window)
		return
	}
	ch := p.filteredChannels[p.currentIndex]
	ShowEPGWindow(p, ch.Name, p.epgData[ch.Name])
}

// ─── Oto URL ─────────────────────────────────────────────────────────────────

func (p *IPTVPlayer) toggleAutoURL() {
	if !p.autoURLEnabled {
		showInputDialog(p.window, Lang.DlgAutoURLTitle, Lang.DlgAutoURLPrompt, p.autoURL, func(url string) {
			if url == "" {
				return
			}
			p.autoURL = url
			p.autoURLEnabled = true
			p.db.SetAutoURL(url)
			p.buildMenu() // checkmark güncelle
			go p.loadAutoURL()
		})
	} else {
		dialog.ShowConfirm(
			Lang.MenuAutoURL,
			"Otomatik URL devre dışı bırakılsın mı?\n(Auto URL disable edilsin mi?)",
			func(ok bool) {
				if ok {
					p.autoURLEnabled = false
					p.autoURL = ""
					p.db.ClearAutoURL()
					p.setStatus(Lang.StatusReady)
					p.buildMenu() // checkmark güncelle
				}
			},
			p.window,
		)
	}
}

// ─── Hakkında ────────────────────────────────────────────────────────────────

func (p *IPTVPlayer) showAbout() {
	content := widget.NewRichTextFromMarkdown(
		"# IPTV Player  Ver. 1.0\n\n" +
			"Bu program Bedavadır (Freeware), para ile satılmaz.\n\n" +
			"**Platform:** " + runtime.GOOS + "/" + runtime.GOARCH + "\n\n" +
			"**Geliştirici:**\n" +
			"- https://www.freewaretr.com\n" +
			"- https://www.caprazbilgi.com\n" +
			"- https://www.youtube.com/caprazbilgi",
	)
	d := dialog.NewCustom(Lang.MenuAbout, Lang.DlgBtnOK, content, p.window)
	d.Resize(fyne.NewSize(420, 260))
	d.Show()
}

// ─── Yardımcılar ─────────────────────────────────────────────────────────────

func (p *IPTVPlayer) setStatus(msg string) {
	if p.statusLabel != nil {
		p.statusLabel.SetText(msg)
	}
}

// setStatusAsync goroutine'den güvenle çağrılabilir
func (p *IPTVPlayer) setStatusAsync(msg string) {
	if p.statusLabel != nil {
		p.statusLabel.SetText(msg)
	}
}

// readFileLines yerel dosyayı satır satır okur
func readFileLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}
