//go:build windows

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lxn/walk"
)

// IPTVPlayer ana uygulama yapısı
type IPTVPlayer struct {
	mw *walk.MainWindow

	vlc *VLCPlayer
	db  *DB

	channels         []Channel
	filteredChannels []Channel
	currentIndex     int
	isFullscreen     bool

	// Logo cache: URL -> PNG/JPEG baytları
	logoCache   map[string][]byte
	logoCacheMu sync.Mutex

	// EPG
	epgData map[string][]EPGProgram

	// Oto URL
	autoURLEnabled bool
	autoURL        string

	// Walk UI bileşenleri
	channelListBox *walk.ListBox
	searchEdit     *walk.LineEdit
	categoryCombo  *walk.ComboBox
	statusText     *walk.Label
	countText      *walk.Label
	logoView       *walk.ImageView
	videoComposite *walk.Composite // VLC buraya render eder

	// Tam ekran için panel referansları
	topPanel        *walk.Composite
	searchPanel     *walk.Composite
	leftPanel       *walk.Composite
	videoTitlePanel *walk.Composite // "OYNATICI" başlık çubuğu
	bottomPanel     *walk.Composite
	statusPanel     *walk.Composite

	// Dil desteği: çevrilebilir widget referansları
	lblTitle        *walk.Label   // "📺  IPTV Player"
	lblSearch       *walk.Label   // "Ara:" / "Search:"
	lblCategory     *walk.Label   // "Kategori:" / "Category:"
	lblChannels     *walk.Label   // "KANALLAR" / "CHANNELS"
	lblPlayer       *walk.Label   // "OYNATICI" / "PLAYER"
	btnFileOpen     *walk.PushButton
	btnURLOpen      *walk.PushButton
	btnPlay         *walk.PushButton
	btnStop         *walk.PushButton
	btnEPG          *walk.PushButton
	btnAddFav       *walk.PushButton
	btnFavorites    *walk.PushButton
	btnFullscreen   *walk.PushButton
	btnExit         *walk.PushButton

	// Menü action referansları
	actFileOpen    *walk.Action
	actURLOpen     *walk.Action
	actPlay        *walk.Action
	actStop        *walk.Action
	actEPG         *walk.Action
	actMenuExit    *walk.Action
	actAutoURL     *walk.Action
	actFullscreen  *walk.Action
	actLangTR      *walk.Action
	actLangEN      *walk.Action
	actAbout       *walk.Action

	// Menü başlığı action referansları (AssignActionTo ile alınır, SetText destekler)
	actMenuFile     *walk.Action // "&Dosya" / "&File"
	actMenuView     *walk.Action // "&Görünüm" / "&View"
	actMenuHelp     *walk.Action // "&Yardım" / "&Help"
	actMenuLanguage *walk.Action // "Dil Desteği" / "Language"
}

// EPGProgram tek EPG programı
type EPGProgram struct {
	Title string
	Start time.Time
	End   time.Time
	Desc  string
}

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

	// Dil tercihini yükle
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

// Run pencereyi oluşturur ve mesaj döngüsünü başlatır
func (p *IPTVPlayer) Run() {
	if err := p.createMainWindow(); err != nil {
		walk.MsgBox(nil, Lang.DlgError, "Pencere oluşturulamadı: "+err.Error(), walk.MsgBoxIconError)
		return
	}

	p.mw.Show()

	// Video alanını siyah yap ve VLC'ye bağla
	if p.videoComposite != nil {
		paintVideoBlack(p.videoComposite)
		p.vlc.SetHWND(uintptr(p.videoComposite.Handle()))
		// Çift tıklama ile tam ekran: VLC'nin child penceresi WM_PARENTNOTIFY gönderir
		SubclassVideoComposite(p.videoComposite, func() {
			p.mw.Synchronize(p.toggleFullscreen)
		})
	}

	// Otomatik URL varsa yükle
	if p.autoURLEnabled && p.autoURL != "" {
		go p.loadAutoURL()
	}

	p.mw.Run()
}

// --- M3U Yükleme ---

func (p *IPTVPlayer) loadM3ULines(lines []string) {
	channels, cats := ParseM3U(lines)
	p.channels = channels
	p.filteredChannels = make([]Channel, len(channels))
	copy(p.filteredChannels, channels)
	p.currentIndex = -1

	allCats := append([]string{Lang.CategoryAll}, cats...)
	p.categoryCombo.SetModel(allCats)
	p.categoryCombo.SetCurrentIndex(0)

	p.channelListBox.SetModel(newChannelModel(p.filteredChannels))
	p.countText.SetText(fmt.Sprintf(Lang.StatusTotal, len(channels)))
}

// filterChannels arama + kategori filtresi uygular
func (p *IPTVPlayer) filterChannels() {
	searchText := strings.ToLower(p.searchEdit.Text())
	cat := p.categoryCombo.Text()

	result := make([]Channel, 0, len(p.channels))
	for _, ch := range p.channels {
		nameOK := searchText == "" || strings.Contains(strings.ToLower(ch.Name), searchText)
		catOK := cat == "" || cat == Lang.CategoryAll || ch.Group == cat
		if nameOK && catOK {
			result = append(result, ch)
		}
	}
	p.filteredChannels = result
	p.channelListBox.SetModel(newChannelModel(p.filteredChannels))
}

// onChannelSelected listeden kanal seçildiğinde çağrılır
func (p *IPTVPlayer) onChannelSelected() {
	idx := p.channelListBox.CurrentIndex()
	if idx < 0 || idx >= len(p.filteredChannels) {
		return
	}
	p.currentIndex = idx
	ch := p.filteredChannels[idx]
	p.setStatus(Lang.MsgSelected + ch.Name)

	// Logoyu arka planda yükle
	if ch.Logo != "" && p.logoView != nil {
		go p.loadLogo(ch.Logo)
	} else if p.logoView != nil {
		p.mw.Synchronize(func() { p.logoView.SetVisible(false) })
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
			p.mw.Synchronize(func() { p.logoView.SetVisible(false) })
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
		p.mw.Synchronize(func() { p.logoView.SetVisible(false) })
		return
	}

	bmp, err := walk.NewBitmapFromImage(img)
	if err != nil {
		p.mw.Synchronize(func() { p.logoView.SetVisible(false) })
		return
	}

	p.mw.Synchronize(func() {
		p.logoView.SetImage(bmp)
		p.logoView.SetVisible(true)
	})
}

// --- Oynatma ---

func (p *IPTVPlayer) play() {
	idx := p.channelListBox.CurrentIndex()
	if idx < 0 || idx >= len(p.filteredChannels) {
		walk.MsgBox(p.mw, Lang.DlgWarning, Lang.MsgSelectChannel, walk.MsgBoxIconWarning)
		return
	}
	p.currentIndex = idx
	ch := p.filteredChannels[idx]
	p.vlc.SetHWND(uintptr(p.videoComposite.Handle()))
	if err := p.vlc.PlayURL(ch.URL); err != nil {
		p.setStatus(Lang.MsgError + err.Error())
		return
	}
	p.setStatus(Lang.MsgPlaying + ch.Name)
}

func (p *IPTVPlayer) playURL(url, name string) {
	p.vlc.SetHWND(uintptr(p.videoComposite.Handle()))
	if err := p.vlc.PlayURL(url); err != nil {
		p.setStatus(Lang.MsgError + err.Error())
		return
	}
	p.setStatus(Lang.MsgPlaying + name)
}

func (p *IPTVPlayer) stop() {
	p.vlc.Stop()
	p.setStatus(Lang.MsgStopped)
}

// --- Dosya / URL ---

func (p *IPTVPlayer) openFile() {
	dlg := new(walk.FileDialog)
	dlg.Title = Lang.DlgOpenFile
	dlg.Filter = Lang.DlgOpenFileFilter
	ok, err := dlg.ShowOpen(p.mw)
	if err != nil || !ok {
		return
	}
	lines, err := readFileLines(dlg.FilePath)
	if err != nil {
		walk.MsgBox(p.mw, Lang.DlgError, Lang.DlgFileReadFail+err.Error(), walk.MsgBoxIconError)
		return
	}
	p.loadM3ULines(lines)
	p.setStatus(Lang.MsgLoaded + dlg.FilePath)
}

func (p *IPTVPlayer) openURL() {
	url := showInputDialog(p.mw, Lang.DlgOpenURL, Lang.DlgOpenURLPrompt, "")
	if url == "" {
		return
	}
	p.setStatus(Lang.MsgLoadingURL)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			p.mw.Synchronize(func() { p.setStatus(Lang.MsgError + err.Error()) })
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		lines := strings.Split(string(data), "\n")
		p.mw.Synchronize(func() {
			p.loadM3ULines(lines)
			p.setStatus(Lang.MsgURLLoaded)
		})
	}()
}

func (p *IPTVPlayer) loadAutoURL() {
	p.mw.Synchronize(func() { p.setStatus(Lang.MsgAutoURLLoading) })
	resp, err := http.Get(p.autoURL)
	if err != nil {
		p.mw.Synchronize(func() { p.setStatus(Lang.MsgAutoURLFailed + err.Error()) })
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	lines := strings.Split(string(data), "\n")
	p.mw.Synchronize(func() {
		p.loadM3ULines(lines)
		p.setStatus(Lang.MsgAutoURLLoaded)
	})
}

// --- Favoriler ---

func (p *IPTVPlayer) addToFavorites() {
	idx := p.channelListBox.CurrentIndex()
	if idx < 0 || idx >= len(p.filteredChannels) {
		walk.MsgBox(p.mw, Lang.DlgWarning, Lang.MsgSelectChannel, walk.MsgBoxIconWarning)
		return
	}
	p.currentIndex = idx
	ch := p.filteredChannels[idx]
	added, err := p.db.AddFavorite(ch.Name, ch.URL)
	if err != nil {
		walk.MsgBox(p.mw, Lang.DlgError, Lang.DlgFavAddFail+err.Error(), walk.MsgBoxIconError)
		return
	}
	if !added {
		walk.MsgBox(p.mw, Lang.DlgInfo, fmt.Sprintf(Lang.DlgFavExists, ch.Name), walk.MsgBoxIconInformation)
		return
	}
	walk.MsgBox(p.mw, Lang.DlgSuccess, fmt.Sprintf(Lang.DlgFavAdded, ch.Name), walk.MsgBoxIconInformation)
}

func (p *IPTVPlayer) showFavorites() {
	ShowFavoritesWindow(p)
}

// --- EPG ---

func (p *IPTVPlayer) showEPG() {
	idx := p.channelListBox.CurrentIndex()
	if idx < 0 || idx >= len(p.filteredChannels) {
		walk.MsgBox(p.mw, Lang.DlgWarning, Lang.MsgSelectChannel, walk.MsgBoxIconWarning)
		return
	}
	p.currentIndex = idx
	ch := p.filteredChannels[idx]
	ShowEPGWindow(p, ch.Name, p.epgData[ch.Name])
}

// --- Oto URL ---

func (p *IPTVPlayer) toggleAutoURL() {
	if !p.autoURLEnabled {
		url := showInputDialog(p.mw, Lang.DlgAutoURLTitle, Lang.DlgAutoURLPrompt, p.autoURL)
		if url == "" {
			return
		}
		p.autoURL = url
		p.autoURLEnabled = true
		p.db.SetAutoURL(url)
		go p.loadAutoURL()
	} else {
		p.autoURLEnabled = false
		p.autoURL = ""
		p.db.ClearAutoURL()
	}
}

// --- Yardımcılar ---

func (p *IPTVPlayer) setStatus(msg string) {
	if p.statusText != nil {
		p.statusText.SetText(msg)
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

// --- Walk ListBox Modeli ---

type channelModel struct {
	walk.ListModelBase
	items []Channel
}

func newChannelModel(channels []Channel) *channelModel {
	return &channelModel{items: channels}
}

func (m *channelModel) ItemCount() int        { return len(m.items) }
func (m *channelModel) Value(i int) interface{} { return m.items[i].Name }
