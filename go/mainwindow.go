//go:build windows

package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// teal rengini döndürür
func teal() walk.Color  { return walk.RGB(13, 115, 119) }
func white() walk.Color { return walk.RGB(255, 255, 255) }
func dark() walk.Color  { return walk.RGB(30, 30, 30) }
func gray() walk.Color  { return walk.RGB(136, 136, 136) }

// createMainWindow ana pencereyi ve tüm bileşenlerini oluşturur
func (p *IPTVPlayer) createMainWindow() error {
	var (
		searchEdit      *walk.LineEdit
		categoryCombo   *walk.ComboBox
		channelList     *walk.ListBox
		logoView        *walk.ImageView
		videoComp       *walk.Composite
		statusLbl       *walk.Label
		countLbl        *walk.Label
		topPanel        *walk.Composite
		searchPanel     *walk.Composite
		leftPanel       *walk.Composite
		videoTitlePanel *walk.Composite

		bottomPanel *walk.Composite
		statusPanel *walk.Composite

		// Çevrilebilir label'lar
		lblSearch   *walk.Label
		lblCategory *walk.Label
		lblChannels *walk.Label
		lblPlayer   *walk.Label

		// Çevrilebilir butonlar
		btnFileOpen   *walk.PushButton
		btnURLOpen    *walk.PushButton
		btnPlay       *walk.PushButton
		btnStop       *walk.PushButton
		btnEPG        *walk.PushButton
		btnAddFav     *walk.PushButton
		btnFavorites  *walk.PushButton
		btnFullscreen *walk.PushButton
		btnExit       *walk.PushButton

		// Menü action referansları
		actFileOpen   *walk.Action
		actURLOpen    *walk.Action
		actPlay       *walk.Action
		actStop       *walk.Action
		actEPG        *walk.Action
		actMenuExit   *walk.Action
		actAutoURL    *walk.Action
		actFullscreen *walk.Action
		actLangTR     *walk.Action
		actLangEN     *walk.Action
		actAbout      *walk.Action

		// Menü başlığı action referansları (AssignActionTo)
		actMenuFile     *walk.Action
		actMenuView     *walk.Action
		actMenuHelp     *walk.Action
		actMenuLanguage *walk.Action
	)

	err := MainWindow{
		AssignTo: &p.mw,
		Title:    "IPTV Player (Ver. 1.0)",
		MinSize:  Size{Width: WindowMinWidth, Height: WindowMinHeight},
		Size:     Size{Width: WindowWidth, Height: WindowHeight},
		Font:     Font{Family: "Segoe UI", PointSize: 9},

		MenuItems: []MenuItem{
			Menu{
				AssignActionTo: &actMenuFile,
				Text:           Lang.MenuFile,
				Items: []MenuItem{
					Action{AssignTo: &actFileOpen, Text: Lang.MenuFileOpen, OnTriggered: p.openFile, Shortcut: Shortcut{Modifiers: walk.ModControl, Key: walk.KeyO}},
					Action{AssignTo: &actURLOpen, Text: Lang.MenuURLOpen, OnTriggered: p.openURL, Shortcut: Shortcut{Modifiers: walk.ModControl, Key: walk.KeyU}},
					Separator{},
					Action{AssignTo: &actPlay, Text: Lang.MenuPlay, OnTriggered: p.play, Shortcut: Shortcut{Modifiers: walk.ModControl, Key: walk.KeyP}},
					Action{AssignTo: &actStop, Text: Lang.MenuStop, OnTriggered: p.stop, Shortcut: Shortcut{Modifiers: walk.ModControl, Key: walk.KeyS}},
					Separator{},
					Action{AssignTo: &actEPG, Text: Lang.MenuEPG, OnTriggered: p.showEPG, Shortcut: Shortcut{Modifiers: walk.ModControl, Key: walk.KeyE}},
					Separator{},
					Action{AssignTo: &actMenuExit, Text: Lang.MenuExit, OnTriggered: func() { p.mw.Close() }},
				},
			},
			Menu{
				AssignActionTo: &actMenuView,
				Text:           Lang.MenuView,
				Items: []MenuItem{
					Action{AssignTo: &actAutoURL, Text: Lang.MenuAutoURL, OnTriggered: p.toggleAutoURL},
					Separator{},
					Action{AssignTo: &actFullscreen, Text: Lang.MenuFullscreen, OnTriggered: p.toggleFullscreen, Shortcut: Shortcut{Key: walk.KeyF11}},
					Separator{},
					Menu{
						AssignActionTo: &actMenuLanguage,
						Text:           Lang.MenuLanguage,
						Items: []MenuItem{
							Action{AssignTo: &actLangTR, Text: Lang.MenuLangTR, OnTriggered: func() { p.setLanguage("tr") }},
							Action{AssignTo: &actLangEN, Text: Lang.MenuLangEN, OnTriggered: func() { p.setLanguage("en") }},
						},
					},
				},
			},
			Menu{
				AssignActionTo: &actMenuHelp,
				Text:           Lang.MenuHelp,
				Items: []MenuItem{
					Action{AssignTo: &actAbout, Text: Lang.MenuAbout, OnTriggered: func() { ShowAboutDialog(p.mw) }},
				},
			},
		},

		Layout: VBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{

			// ── ÜST BAR ──────────────────────────────────────────────────
			Composite{
				AssignTo:           &topPanel,
				AlwaysConsumeSpace: false,
				Layout:             HBox{Margins: Margins{Left: 10, Top: 6, Right: 10, Bottom: 6}, Spacing: 6},
				Children: []Widget{
					Label{
						Text:      "📺  IPTV Player",
						TextColor: teal(),
						Font:      Font{Family: "Segoe UI", Bold: true, PointSize: 13},
					},
					HSpacer{},
					PushButton{AssignTo: &btnFileOpen, Text: Lang.BtnFileOpen, OnClicked: p.openFile, MinSize: Size{Width: 100}},
					PushButton{AssignTo: &btnURLOpen, Text: Lang.BtnURLOpen, OnClicked: p.openURL, MinSize: Size{Width: 100}},
				},
			},

			// ── ARAMA + KATEGORİ ─────────────────────────────────────────
			Composite{
				AssignTo:           &searchPanel,
				AlwaysConsumeSpace: false,
				Layout:             HBox{Margins: Margins{Left: 10, Top: 2, Right: 10, Bottom: 4}, Spacing: 6},
				Children: []Widget{
					Label{AssignTo: &lblSearch, Text: Lang.LabelSearch, Font: Font{Family: "Segoe UI", PointSize: 9}},
					LineEdit{AssignTo: &searchEdit, CueBanner: Lang.SearchCue},
					Label{AssignTo: &lblCategory, Text: Lang.LabelCategory, Font: Font{Family: "Segoe UI", PointSize: 9}},
					ComboBox{
						AssignTo: &categoryCombo,
						Model:    []string{Lang.CategoryAll},
						MinSize:  Size{Width: 140},
					},
				},
			},

			// ── ORTA: SOL + SAĞ ──────────────────────────────────────────
			Composite{
				Layout: HBox{Spacing: 0, MarginsZero: true},
				Children: []Widget{

					// SOL PANEL
					Composite{
						AssignTo:           &leftPanel,
						AlwaysConsumeSpace: false,
						Layout:             VBox{MarginsZero: true, SpacingZero: true},
						MinSize:            Size{Width: 240},
						MaxSize:            Size{Width: 310},
						Children: []Widget{
							Composite{
								Background: SolidColorBrush{Color: teal()},
								Layout:     HBox{Margins: Margins{Left: 8, Top: 4, Right: 8, Bottom: 4}},
								Children: []Widget{
									Label{
										AssignTo:  &lblChannels,
										Text:      Lang.LabelChannels,
										TextColor: white(),
										Font:      Font{Family: "Segoe UI", Bold: true, PointSize: 9},
									},
									HSpacer{},
								},
							},
							ListBox{
								AssignTo:              &channelList,
								OnCurrentIndexChanged: p.onChannelSelected,
								OnItemActivated:       p.play,
							},
							ImageView{
								AssignTo:           &logoView,
								MinSize:            Size{Width: 50, Height: 80},
								MaxSize:            Size{Height: 100},
								Mode:               ImageViewModeZoom,
								Visible:            false,
								AlwaysConsumeSpace: false,
							},
						},
					},

					// SAĞ PANEL
					Composite{
						Layout:        VBox{MarginsZero: true, SpacingZero: true},
						StretchFactor: 1,
						Children: []Widget{
							Composite{
								AssignTo:   &videoTitlePanel,
								Background: SolidColorBrush{Color: teal()},
								Layout:     HBox{Margins: Margins{Left: 8, Top: 4, Right: 8, Bottom: 4}},
								Children: []Widget{
									Label{
										AssignTo:  &lblPlayer,
										Text:      Lang.LabelPlayer,
										TextColor: white(),
										Font:      Font{Family: "Segoe UI", Bold: true, PointSize: 9},
									},
									HSpacer{},
								},
							},
							Composite{
								AssignTo:      &videoComp,
								StretchFactor: 1,
								Layout:        VBox{MarginsZero: true, SpacingZero: true},
							},
						},
					},
				},
			},

			// ── ALT BUTONLAR ─────────────────────────────────────────────
			Composite{
				AssignTo:           &bottomPanel,
				AlwaysConsumeSpace: false,
				Layout:             HBox{Margins: Margins{Left: 10, Top: 4, Right: 10, Bottom: 4}, Spacing: 4},
				Children: []Widget{
					PushButton{AssignTo: &btnPlay, Text: Lang.BtnPlay, OnClicked: p.play, MinSize: Size{Width: 90}},
					PushButton{AssignTo: &btnStop, Text: Lang.BtnStop, OnClicked: p.stop, MinSize: Size{Width: 90}},
					PushButton{AssignTo: &btnEPG, Text: Lang.BtnEPG, OnClicked: p.showEPG, MinSize: Size{Width: 80}},
					PushButton{AssignTo: &btnAddFav, Text: Lang.BtnAddFav, OnClicked: p.addToFavorites, MinSize: Size{Width: 110}},
					PushButton{AssignTo: &btnFavorites, Text: Lang.BtnFavorites, OnClicked: p.showFavorites, MinSize: Size{Width: 100}},
					PushButton{AssignTo: &btnFullscreen, Text: Lang.BtnFullscreen, OnClicked: p.toggleFullscreen, MinSize: Size{Width: 105}},
					HSpacer{},
					PushButton{AssignTo: &btnExit, Text: Lang.BtnExit, OnClicked: func() { p.mw.Close() }, MinSize: Size{Width: 70}},
				},
			},

			// ── DURUM ÇUBUĞU ─────────────────────────────────────────────
			Composite{
				AssignTo:           &statusPanel,
				AlwaysConsumeSpace: false,
				Layout:             HBox{Margins: Margins{Left: 10, Top: 3, Right: 10, Bottom: 3}},
				Children: []Widget{
					Label{
						AssignTo:  &statusLbl,
						Text:      Lang.StatusReady,
						TextColor: gray(),
						Font:      Font{Family: "Segoe UI", PointSize: 9},
					},
					HSpacer{},
					Label{
						AssignTo:  &countLbl,
						Text:      "Toplam Kanal: 0",
						TextColor: teal(),
						Font:      Font{Family: "Segoe UI", Bold: true, PointSize: 9},
					},
				},
			},
		},
	}.Create()

	if err != nil {
		return err
	}

	// Widget referanslarını kaydet
	p.searchEdit = searchEdit
	p.categoryCombo = categoryCombo
	p.channelListBox = channelList
	p.logoView = logoView
	p.videoComposite = videoComp
	p.statusText = statusLbl
	p.countText = countLbl
	p.topPanel = topPanel
	p.searchPanel = searchPanel
	p.leftPanel = leftPanel
	p.videoTitlePanel = videoTitlePanel
	p.bottomPanel = bottomPanel
	p.statusPanel = statusPanel

	// Çevrilebilir widget referansları
	p.lblSearch = lblSearch
	p.lblCategory = lblCategory
	p.lblChannels = lblChannels
	p.lblPlayer = lblPlayer
	p.btnFileOpen = btnFileOpen
	p.btnURLOpen = btnURLOpen
	p.btnPlay = btnPlay
	p.btnStop = btnStop
	p.btnEPG = btnEPG
	p.btnAddFav = btnAddFav
	p.btnFavorites = btnFavorites
	p.btnFullscreen = btnFullscreen
	p.btnExit = btnExit

	// Menü action referansları
	p.actFileOpen = actFileOpen
	p.actURLOpen = actURLOpen
	p.actPlay = actPlay
	p.actStop = actStop
	p.actEPG = actEPG
	p.actMenuExit = actMenuExit
	p.actAutoURL = actAutoURL
	p.actFullscreen = actFullscreen
	p.actLangTR = actLangTR
	p.actLangEN = actLangEN
	p.actAbout = actAbout
	p.actMenuFile = actMenuFile
	p.actMenuView = actMenuView
	p.actMenuHelp = actMenuHelp
	p.actMenuLanguage = actMenuLanguage

	// Pencere ikonu
	for _, iconPath := range []string{"iptvPlayer.ico", "../iptvPlayer.ico"} {
		if icon, err := walk.NewIconFromFile(iconPath); err == nil {
			p.mw.SetIcon(icon)
			break
		}
	}

	// Video alanı boyutu değişince VLC render penceresini yeniden bağla
	videoComp.SizeChanged().Attach(func() {
		if p.vlc != nil {
			p.vlc.SetHWND(uintptr(p.videoComposite.Handle()))
		}
	})

	// Arama / kategori değişince filtrele
	searchEdit.TextChanged().Attach(func() { p.filterChannels() })
	categoryCombo.CurrentIndexChanged().Attach(func() { p.filterChannels() })

	// ESC → tam ekrandan çık
	p.mw.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyEscape {
			p.exitFullscreen()
		}
	})

	// Dil durumunu yansıt (checkmark)
	p.updateLangChecks()

	return nil
}

// applyLang tüm widget metinlerini aktif dile göre günceller.
// Dil değiştiğinde çağrılır.
func (p *IPTVPlayer) applyLang() {
	// Tüm menü action metinleri (başlıklar dahil)
	actions := map[*walk.Action]string{
		p.actMenuFile:     Lang.MenuFile,
		p.actMenuView:     Lang.MenuView,
		p.actMenuHelp:     Lang.MenuHelp,
		p.actMenuLanguage: Lang.MenuLanguage,
		p.actFileOpen:     Lang.MenuFileOpen,
		p.actURLOpen:      Lang.MenuURLOpen,
		p.actPlay:         Lang.MenuPlay,
		p.actStop:         Lang.MenuStop,
		p.actEPG:          Lang.MenuEPG,
		p.actMenuExit:     Lang.MenuExit,
		p.actAutoURL:      Lang.MenuAutoURL,
		p.actFullscreen:   Lang.MenuFullscreen,
		p.actLangTR:       Lang.MenuLangTR,
		p.actLangEN:       Lang.MenuLangEN,
		p.actAbout:        Lang.MenuAbout,
	}
	for act, text := range actions {
		if act != nil {
			act.SetText(text)
		}
	}

	// Label metinleri
	if p.lblSearch != nil {
		p.lblSearch.SetText(Lang.LabelSearch)
	}
	if p.lblCategory != nil {
		p.lblCategory.SetText(Lang.LabelCategory)
	}
	if p.lblChannels != nil {
		p.lblChannels.SetText(Lang.LabelChannels)
	}
	if p.lblPlayer != nil {
		p.lblPlayer.SetText(Lang.LabelPlayer)
	}

	// Search placeholder
	if p.searchEdit != nil {
		p.searchEdit.SetCueBanner(Lang.SearchCue)
	}

	// "Tümü / All" kategorisi güncelle
	if p.categoryCombo != nil {
		idx := p.categoryCombo.CurrentIndex()
		// İlk item her zaman "Tümü/All"
		if p.categoryCombo.Model() != nil {
			// Mevcut kategori listesini al ve ilk elemanı güncelle
			if model, ok := p.categoryCombo.Model().([]string); ok && len(model) > 0 {
				model[0] = Lang.CategoryAll
				p.categoryCombo.SetModel(model)
				p.categoryCombo.SetCurrentIndex(idx)
			}
		}
	}

	// Buton metinleri
	buttons := map[*walk.PushButton]string{
		p.btnFileOpen:   Lang.BtnFileOpen,
		p.btnURLOpen:    Lang.BtnURLOpen,
		p.btnPlay:       Lang.BtnPlay,
		p.btnStop:       Lang.BtnStop,
		p.btnEPG:        Lang.BtnEPG,
		p.btnAddFav:     Lang.BtnAddFav,
		p.btnFavorites:  Lang.BtnFavorites,
		p.btnFullscreen: Lang.BtnFullscreen,
		p.btnExit:       Lang.BtnExit,
	}
	for btn, text := range buttons {
		if btn != nil {
			btn.SetText(text)
		}
	}

	// Durum çubuğu
	if p.statusText != nil {
		p.statusText.SetText(Lang.StatusReady)
	}

	// Dil seçim checkmark'ını güncelle
	p.updateLangChecks()
}

// updateLangChecks aktif dile göre menü checkmark'larını günceller.
func (p *IPTVPlayer) updateLangChecks() {
	isTR := Lang == &LangTR
	if p.actLangTR != nil {
		p.actLangTR.SetChecked(isTR)
	}
	if p.actLangEN != nil {
		p.actLangEN.SetChecked(!isTR)
	}
}

// setLanguage dili değiştirir, DB'ye kaydeder ve UI'ı günceller.
func (p *IPTVPlayer) setLanguage(code string) {
	if code == "en" {
		Lang = &LangEN
	} else {
		Lang = &LangTR
	}
	p.db.SetLanguage(code)
	p.applyLang()
}

// toggleFullscreen tam ekrana girer/çıkar (F11, buton, video tıklama)
func (p *IPTVPlayer) toggleFullscreen() {
	p.vlc.SetFullscreen(false) // VLC'nin kendi tam ekranını devre dışı bırak
	toggleFullscreen(p)
}

// exitFullscreen ESC ile tam ekrandan çıkar
func (p *IPTVPlayer) exitFullscreen() {
	if p.isFullscreen {
		p.toggleFullscreen()
	}
}
