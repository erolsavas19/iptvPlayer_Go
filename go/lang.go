package main

// UIStrings uygulamanın tüm kullanıcı arayüzü metinlerini içerir.
type UIStrings struct {
	// ── Menü: Dosya ──────────────────────────────────────────────
	MenuFile        string // &Dosya / &File
	MenuFileOpen    string // Dosya Aç / Open File
	MenuURLOpen     string // URL Aç / Open URL
	MenuPlay        string // Çalıştır / Play
	MenuStop        string // Durdur / Stop
	MenuEPG         string // Kanal EPG'si / Channel EPG
	MenuExit        string // Çıkış / Exit

	// ── Menü: Görünüm ─────────────────────────────────────────────
	MenuView        string // &Görünüm / &View
	MenuAutoURL     string // Otomatik Açılacak URL / Auto-open URL
	MenuFullscreen  string // Tam Ekran / Full Screen
	MenuLanguage    string // Dil Desteği / Language
	MenuLangTR      string // Türkçe
	MenuLangEN      string // English

	// ── Menü: Yardım ─────────────────────────────────────────────
	MenuHelp        string // &Yardım / &Help
	MenuAbout       string // Program Hakkında / About

	// ── Üst bar ──────────────────────────────────────────────────
	BtnFileOpen     string // 📂  Dosya Aç / 📂  Open File
	BtnURLOpen      string // 🌐  URL Aç / 🌐  Open URL

	// ── Arama / Kategori ─────────────────────────────────────────
	LabelSearch     string // Ara: / Search:
	SearchCue       string // Kanal adı girin... / Enter channel name...
	LabelCategory   string // Kategori: / Category:
	CategoryAll     string // Tümü / All

	// ── Kanal listesi başlığı ─────────────────────────────────────
	LabelChannels   string // KANALLAR / CHANNELS

	// ── Video başlığı ─────────────────────────────────────────────
	LabelPlayer     string // OYNATICI / PLAYER

	// ── Alt butonlar ─────────────────────────────────────────────
	BtnPlay         string // ▶  Oynat / ▶  Play
	BtnStop         string // ⏹  Durdur / ⏹  Stop
	BtnEPG          string // 📋  EPG
	BtnAddFav       string // ★  Favori Ekle / ★  Add Favorite
	BtnFavorites    string // ★  Favoriler / ★  Favorites
	BtnFullscreen   string // ⛶  Tam Ekran / ⛶  Full Screen
	BtnExit         string // Çıkış / Exit

	// ── Durum çubuğu ─────────────────────────────────────────────
	StatusReady     string // Hazır / Ready
	StatusTotal     string // Toplam Kanal: %d / Total Channels: %d

	// ── Uygulama mesajları ───────────────────────────────────────
	MsgSelectChannel   string // Lütfen önce bir kanal seçin!
	MsgPlaying         string // Oynatılıyor: / Playing:
	MsgStopped         string // Durduruldu / Stopped
	MsgSelected        string // Seçili: / Selected:
	MsgLoaded          string // Yüklendi: / Loaded:
	MsgLoadingURL      string // URL yükleniyor... / Loading URL...
	MsgURLLoaded       string // URL'den yüklendi / Loaded from URL
	MsgAutoURLLoading  string // Otomatik URL yükleniyor... / Loading auto URL...
	MsgAutoURLLoaded   string // Otomatik URL yüklendi / Auto URL loaded
	MsgAutoURLFailed   string // Otomatik URL yüklenemedi: / Failed to load auto URL:
	MsgError           string // Hata: / Error:

	// ── MsgBox başlıkları / mesajları ────────────────────────────
	DlgWarning         string // Uyarı / Warning
	DlgError           string // Hata / Error
	DlgInfo            string // Bilgi / Information
	DlgSuccess         string // Başarılı / Success
	DlgAboutTitle      string // Program Hakkında / About
	DlgFavExists       string // '%s' zaten favorilerinizde! / '%s' is already in favorites!
	DlgFavAdded        string // '%s' favorilere eklendi! / '%s' added to favorites!
	DlgFavAddFail      string // Favori eklenemedi: / Failed to add favorite:
	DlgFileReadFail    string // Dosya okunamadı: / Could not read file:
	DlgOpenFile        string // M3U Dosyası Aç / Open M3U File
	DlgOpenFileFilter  string // M3U Dosyaları (*.m3u;*.m3u8)|*.m3u;*.m3u8|Tüm Dosyalar (*.*)|*.* / M3U Files...
	DlgOpenURL         string // URL Aç / Open URL
	DlgOpenURLPrompt   string // M3U URL adresini girin: / Enter M3U URL:
	DlgAutoURLTitle    string // Otomatik Açılacak URL / Auto-open URL
	DlgAutoURLPrompt   string // M3U URL adresini girin: / Enter M3U URL:

	// ── Favoriler penceresi ──────────────────────────────────────
	FavWindowTitle     string // Favoriler / Favorites
	FavColName         string // Kanal Adı / Channel Name
	FavColURL          string // URL
	FavBtnPlay         string // ▶  Oynat / ▶  Play
	FavBtnRemove       string // 🗑  Sil / 🗑  Remove
	FavBtnClose        string // Kapat / Close
	FavMsgSelect       string // Lütfen bir kanal seçin. / Please select a channel.
	FavMsgConfirmDel   string // '%s' favorilerden silinsin mi? / Remove '%s' from favorites?
	FavMsgConfirmTitle string // Favori Sil / Remove Favorite
	FavMsgEmpty        string // Favori listeniz boş. / Your favorites list is empty.

	// ── EPG penceresi ───────────────────────────────────────────
	EPGWindowTitle     string // EPG: %s
	EPGColTime         string // Saat / Time
	EPGColTitle        string // Program / Show
	EPGColDesc         string // Açıklama / Description
	EPGBtnClose        string // Kapat / Close
	EPGMsgNoData       string // Bu kanal için EPG verisi bulunamadı. / No EPG data for this channel.
	EPGMsgNoDataTitle  string // EPG Yok / No EPG

	// ── Input dialog ────────────────────────────────────────────
	DlgBtnOK     string // Tamam / OK
	DlgBtnCancel string // İptal / Cancel
}

// LangTR Türkçe metinler
var LangTR = UIStrings{
	MenuFile:       "&Dosya",
	MenuFileOpen:   "Dosya Aç\tCtrl+O",
	MenuURLOpen:    "URL Aç\tCtrl+U",
	MenuPlay:       "Çalıştır\tCtrl+P",
	MenuStop:       "Durdur\tCtrl+S",
	MenuEPG:        "Kanal EPG'si\tCtrl+E",
	MenuExit:       "Çıkış\tAlt+F4",
	MenuView:       "&Görünüm",
	MenuAutoURL:    "Otomatik Açılacak URL",
	MenuFullscreen: "Tam Ekran\tF11",
	MenuLanguage:   "Dil Desteği",
	MenuLangTR:     "Türkçe",
	MenuLangEN:     "English",
	MenuHelp:       "&Yardım",
	MenuAbout:      "Program Hakkında",

	BtnFileOpen:   "📂  Dosya Aç",
	BtnURLOpen:    "🌐  URL Aç",
	LabelSearch:   "Ara:",
	SearchCue:     "Kanal adı girin...",
	LabelCategory: "Kategori:",
	CategoryAll:   "Tümü",
	LabelChannels: "KANALLAR",
	LabelPlayer:   "OYNATICI",
	BtnPlay:       "▶  Oynat",
	BtnStop:       "⏹  Durdur",
	BtnEPG:        "📋  EPG",
	BtnAddFav:     "★  Favori Ekle",
	BtnFavorites:  "★  Favoriler",
	BtnFullscreen: "⛶  Tam Ekran",
	BtnExit:       "Çıkış",

	StatusReady: "Hazır",
	StatusTotal: "Toplam Kanal: %d",

	MsgSelectChannel:  "Lütfen önce bir kanal seçin!",
	MsgPlaying:        "Oynatılıyor: ",
	MsgStopped:        "Durduruldu",
	MsgSelected:       "Seçili: ",
	MsgLoaded:         "Yüklendi: ",
	MsgLoadingURL:     "URL yükleniyor...",
	MsgURLLoaded:      "URL'den yüklendi",
	MsgAutoURLLoading: "Otomatik URL yükleniyor...",
	MsgAutoURLLoaded:  "Otomatik URL yüklendi",
	MsgAutoURLFailed:  "Otomatik URL yüklenemedi: ",
	MsgError:          "Hata: ",

	DlgWarning:        "Uyarı",
	DlgError:          "Hata",
	DlgInfo:           "Bilgi",
	DlgSuccess:        "Başarılı",
	DlgAboutTitle:     "Program Hakkında",
	DlgFavExists:      "'%s' zaten favorilerinizde!",
	DlgFavAdded:       "'%s' favorilere eklendi!",
	DlgFavAddFail:     "Favori eklenemedi: ",
	DlgFileReadFail:   "Dosya okunamadı: ",
	DlgOpenFile:       "M3U Dosyası Aç",
	DlgOpenFileFilter: "M3U Dosyaları (*.m3u;*.m3u8)|*.m3u;*.m3u8|Tüm Dosyalar (*.*)|*.*",
	DlgOpenURL:        "URL Aç",
	DlgOpenURLPrompt:  "M3U URL adresini girin:",
	DlgAutoURLTitle:   "Otomatik Açılacak URL",
	DlgAutoURLPrompt:  "M3U URL adresini girin:",

	FavWindowTitle:     "Favoriler",
	FavColName:         "Kanal Adı",
	FavColURL:          "URL",
	FavBtnPlay:         "▶  Oynat",
	FavBtnRemove:       "🗑  Sil",
	FavBtnClose:        "Kapat",
	FavMsgSelect:       "Lütfen bir kanal seçin.",
	FavMsgConfirmDel:   "'%s' favorilerden silinsin mi?",
	FavMsgConfirmTitle: "Favori Sil",
	FavMsgEmpty:        "Favori listeniz boş.",

	EPGWindowTitle:    "EPG: %s",
	EPGColTime:        "Saat",
	EPGColTitle:       "Program",
	EPGColDesc:        "Açıklama",
	EPGBtnClose:       "Kapat",
	EPGMsgNoData:      "Bu kanal için EPG verisi bulunamadı.",
	EPGMsgNoDataTitle: "EPG Yok",

	DlgBtnOK:     "Tamam",
	DlgBtnCancel: "İptal",
}

// LangEN English strings
var LangEN = UIStrings{
	MenuFile:       "&File",
	MenuFileOpen:   "Open File\tCtrl+O",
	MenuURLOpen:    "Open URL\tCtrl+U",
	MenuPlay:       "Play\tCtrl+P",
	MenuStop:       "Stop\tCtrl+S",
	MenuEPG:        "Channel EPG\tCtrl+E",
	MenuExit:       "Exit\tAlt+F4",
	MenuView:       "&View",
	MenuAutoURL:    "Auto-open URL",
	MenuFullscreen: "Full Screen\tF11",
	MenuLanguage:   "Language",
	MenuLangTR:     "Türkçe",
	MenuLangEN:     "English",
	MenuHelp:       "&Help",
	MenuAbout:      "About",

	BtnFileOpen:   "📂  Open File",
	BtnURLOpen:    "🌐  Open URL",
	LabelSearch:   "Search:",
	SearchCue:     "Enter channel name...",
	LabelCategory: "Category:",
	CategoryAll:   "All",
	LabelChannels: "CHANNELS",
	LabelPlayer:   "PLAYER",
	BtnPlay:       "▶  Play",
	BtnStop:       "⏹  Stop",
	BtnEPG:        "📋  EPG",
	BtnAddFav:     "★  Add Favorite",
	BtnFavorites:  "★  Favorites",
	BtnFullscreen: "⛶  Full Screen",
	BtnExit:       "Exit",

	StatusReady: "Ready",
	StatusTotal: "Total Channels: %d",

	MsgSelectChannel:  "Please select a channel first!",
	MsgPlaying:        "Playing: ",
	MsgStopped:        "Stopped",
	MsgSelected:       "Selected: ",
	MsgLoaded:         "Loaded: ",
	MsgLoadingURL:     "Loading URL...",
	MsgURLLoaded:      "Loaded from URL",
	MsgAutoURLLoading: "Loading auto URL...",
	MsgAutoURLLoaded:  "Auto URL loaded",
	MsgAutoURLFailed:  "Failed to load auto URL: ",
	MsgError:          "Error: ",

	DlgWarning:        "Warning",
	DlgError:          "Error",
	DlgInfo:           "Information",
	DlgSuccess:        "Success",
	DlgAboutTitle:     "About",
	DlgFavExists:      "'%s' is already in your favorites!",
	DlgFavAdded:       "'%s' added to favorites!",
	DlgFavAddFail:     "Failed to add favorite: ",
	DlgFileReadFail:   "Could not read file: ",
	DlgOpenFile:       "Open M3U File",
	DlgOpenFileFilter: "M3U Files (*.m3u;*.m3u8)|*.m3u;*.m3u8|All Files (*.*)|*.*",
	DlgOpenURL:        "Open URL",
	DlgOpenURLPrompt:  "Enter M3U URL:",
	DlgAutoURLTitle:   "Auto-open URL",
	DlgAutoURLPrompt:  "Enter M3U URL:",

	FavWindowTitle:     "Favorites",
	FavColName:         "Channel Name",
	FavColURL:          "URL",
	FavBtnPlay:         "▶  Play",
	FavBtnRemove:       "🗑  Remove",
	FavBtnClose:        "Close",
	FavMsgSelect:       "Please select a channel.",
	FavMsgConfirmDel:   "Remove '%s' from favorites?",
	FavMsgConfirmTitle: "Remove Favorite",
	FavMsgEmpty:        "Your favorites list is empty.",

	EPGWindowTitle:    "EPG: %s",
	EPGColTime:        "Time",
	EPGColTitle:       "Show",
	EPGColDesc:        "Description",
	EPGBtnClose:       "Close",
	EPGMsgNoData:      "No EPG data found for this channel.",
	EPGMsgNoDataTitle: "No EPG",

	DlgBtnOK:     "OK",
	DlgBtnCancel: "Cancel",
}

// Lang aktif dil; uygulama boyunca bu değişken kullanılır
var Lang = &LangTR
