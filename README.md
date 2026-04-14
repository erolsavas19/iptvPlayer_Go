# IPTV Player — Go Edition

A lightweight, native Windows IPTV player built with Go. Plays M3U/M3U8 streams using VLC under the hood, with a clean native UI powered by the Walk framework.

---

## Screenshots

![Screenshot 1](ScreenShots01.png)
![Screenshot 2](ScreenShots02.png)

---

## Features

- **M3U / M3U8 playlist support** — load from a local file or any URL
- **VLC-powered playback** — uses `libvlc.dll` at runtime; no CGO, no extra toolchain
- **Channel search** — instant filtering by channel name
- **Category filter** — group channels by their M3U group-title tag
- **Favorites** — save and manage favorite channels (stored in a local SQLite database)
- **EPG** — view Electronic Program Guide data per channel
- **Fullscreen mode** — toggle with double-click on the video area or F11
- **Channel logo display** — fetches and caches logos from M3U `tvg-logo` tags
- **Auto-open URL** — configure a playlist URL that loads automatically on startup
- **Bilingual UI** — switch between Turkish and English from the menu
- **High DPI aware** — sharp rendering on high-resolution displays

---

## Requirements

| Requirement | Version | Download |
|---|---|---|
| OS | Windows 10 / 11 (64-bit) | — |
| Go | 1.21 or later | https://golang.org/dl/ |
| VLC Media Player | 3.x or later (64-bit) | https://www.videolan.org/vlc/ |

> **Note:** This application is Windows-only. The GUI framework (`lxn/walk`) targets the Windows API exclusively.

---

## Getting Started

### 1. Install Go

Download and run the Windows installer from https://golang.org/dl/  
After installation, verify in a terminal:

```
go version
```

### 2. Install VLC

Download the **64-bit** installer from https://www.videolan.org/vlc/ and install to the default location.

### 3. Clone the repository

```bash
git clone https://github.com/erolsavas19/iptvPlayer_GO.git
cd iptvPlayer_GO
```

### 4. Download dependencies

```bash
cd go
go mod download
```

This fetches all libraries listed in `go.mod`. Requires an internet connection; allow a minute or two on the first run.

### 5. Run from source

**Option A — batch file (easiest):**  
Double-click `go\run.bat`

**Option B — terminal:**

```bash
cd go
go run .
```

> A console window will appear alongside the app when using `go run`. This is normal in development mode.

---

## Building a Release Executable

```bash
cd go
go build -ldflags="-H windowsgui -s -w" -o ../iptvPlayer_go.exe .
```

Flag reference:
- `-H windowsgui` — suppresses the console window (pure GUI app)
- `-s -w` — strips debug symbols, reduces exe size

Alternatively, double-click `go\derle.bat` for a full build that also embeds the application icon and manifest.

---

## Project Structure

```
iptvPlayer_GO/
├── README.md
├── LICENSE
├── .gitignore
├── iptvPlayer_go.exe        ← compiled executable
└── go/                      ← all source files
    ├── main.go              ← entry point; logging, panic recovery
    ├── app.go               ← core struct; M3U loading, playback, favourites
    ├── mainwindow.go        ← main window layout and all UI widgets
    ├── video_win.go         ← VLC video render area (Windows subclassing)
    ├── fullscreen_win.go    ← fullscreen toggle logic
    ← VLC libvlc.dll integration (pure Go, no CGO)
    ├── db.go                ← SQLite: favourites and app settings
    ├── m3u.go               ← M3U/M3U8 parser
    ├── dialogs.go           ← favourites, EPG, URL input windows
    ├── lang.go              ← TR/EN string tables and language switching
    ├── manifest_win.go      ← Windows DPI-aware manifest embedding
    ├── app.manifest         ← Windows application manifest (XML)
    ├── iptvPlayer.ico       ← application icon
    ├── rsrc.syso            ← pre-compiled resources (icon + manifest)
    ├── go.mod               ← module definition
    ├── go.sum               ← dependency checksums
    ├── run.bat              ← run-from-source helper
    └── derle.bat            ← release build script
```

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/lxn/walk` | Native Windows GUI framework |
| `github.com/lxn/win` | Low-level Windows API bindings |
| `golang.org/x/sys` | System calls — DLL loading, registry access |
| `modernc.org/sqlite` | Pure-Go SQLite driver (no CGO required) |

---

## Why is VLC required?

The application loads VLC's `libvlc.dll` at runtime using the Windows `LoadLibrary` API — no CGO and no C compiler needed. This keeps the build simple (a single `go build` command) while still providing full VLC media engine capabilities. VLC is not bundled; it must be installed separately. If VLC is not found, the application shows an error message on startup.

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---
---

# IPTV Player — Go Sürümü

Go ile yazılmış, hafif ve yerel Windows IPTV oynatıcısı. M3U/M3U8 akışlarını VLC altyapısıyla oynatır; kullanıcı arayüzü Walk framework'ü ile oluşturulmuş tam yerel bir Windows uygulamasıdır.

---

## Ekran Görüntüleri

![Ekran Görüntüsü 1](ScreenShots01.png)
![Ekran Görüntüsü 2](ScreenShots02.png)

---

## Özellikler

- **M3U / M3U8 playlist desteği** — yerel dosya veya URL'den yükleme
- **VLC tabanlı oynatma** — `libvlc.dll` çalışma zamanında yüklenir; CGO veya ekstra araç gerekmez
- **Kanal arama** — kanal adına göre anlık filtreleme
- **Kategori filtresi** — M3U group-title etiketine göre gruplama
- **Favoriler** — favori kanalları kaydetme ve yönetme (yerel SQLite veritabanı)
- **EPG** — Elektronik Program Rehberi verilerini kanal bazında görüntüleme
- **Tam ekran modu** — video alanına çift tıklama veya F11 ile geçiş
- **Kanal logosu görüntüleme** — M3U `tvg-logo` etiketindeki logolar indirilir ve önbelleğe alınır
- **Otomatik URL açma** — başlangıçta otomatik yüklenecek playlist URL'si ayarı
- **İki dilli arayüz** — menüden Türkçe ve İngilizce arasında geçiş
- **Yüksek DPI desteği** — yüksek çözünürlüklü ekranlarda keskin görüntü

---

## Sistem Gereksinimleri

| Gereksinim | Sürüm | İndirme |
|---|---|---|
| İşletim Sistemi | Windows 10 / 11 (64-bit) | — |
| Go | 1.21 veya üzeri | https://golang.org/dl/ |
| VLC Media Player | 3.x veya üzeri (64-bit) | https://www.videolan.org/vlc/ |

> **Not:** Bu uygulama yalnızca Windows'ta çalışır. Kullanılan GUI kütüphanesi (`lxn/walk`) yalnızca Windows API'sini hedefler.

---

## Kurulum ve Çalıştırma

### 1. Go Kurulumu

https://golang.org/dl/ adresinden Windows için `.msi` dosyasını indirin ve kurun.  
Kurulumdan sonra terminalde doğrulayın:

```
go version
```

### 2. VLC Kurulumu

https://www.videolan.org/vlc/ adresinden **64-bit** VLC'yi indirin ve varsayılan konuma kurun.

### 3. Depoyu İndirin

```bash
git clone https://github.com/erolsavas19/iptvPlayer_GO.git
cd iptvPlayer_GO
```

### 4. Bağımlılıkları İndirin

```bash
cd go
go mod download
```

Bu komut `go.mod` dosyasındaki tüm kütüphaneleri otomatik olarak indirir. İnternet bağlantısı gerektirir; ilk çalıştırmada 1-2 dakika sürebilir.

### 5. Çalıştırın

**Yöntem A — Batch dosyası ile (en kolay):**  
`go\run.bat` dosyasına çift tıklayın.

**Yöntem B — Terminal ile:**

```bash
cd go
go run .
```

> `go run` ile çalıştırıldığında arka planda küçük bir konsol penceresi açılır; geliştirme sırasında bu normaldir.

---

## Exe Derleme (Release Build)

```bash
cd go
go build -ldflags="-H windowsgui -s -w" -o ../iptvPlayer_go.exe .
```

Bayrak açıklamaları:
- `-H windowsgui` — konsol penceresi açılmaz (saf GUI uygulaması)
- `-s -w` — hata ayıklama sembolleri çıkarılır, exe boyutu küçülür

Alternatif olarak `go\derle.bat` dosyasını çalıştırabilirsiniz; bu script uygulama ikonunu ve manifestını da gömerek tam bir release build üretir.

---

## Proje Yapısı

```
iptvPlayer_GO/
├── README.md
├── LICENSE
├── .gitignore
├── iptvPlayer_go.exe        ← derlenmiş uygulama
└── go/                      ← tüm kaynak kodlar
    ├── main.go              ← giriş noktası; log kurulumu, panic kurtarma
    ├── app.go               ← temel yapı; M3U yükleme, oynatma, favoriler
    ├── mainwindow.go        ← ana pencere düzeni ve tüm UI bileşenleri
    ├── video_win.go         ← VLC video render alanı (Windows subclassing)
    ├── fullscreen_win.go    ← tam ekran geçiş mantığı
    ├── vlc.go               ← VLC libvlc.dll entegrasyonu (saf Go, CGO yok)
    ├── db.go                ← SQLite: favoriler ve uygulama ayarları
    ├── m3u.go               ← M3U/M3U8 parser
    ├── dialogs.go           ← favoriler, EPG, URL giriş pencereleri
    ├── lang.go              ← TR/EN dil dizileri ve dil değiştirme
    ├── manifest_win.go      ← Windows DPI aware manifest gömme
    ├── app.manifest         ← Windows uygulama manifestı (XML)
    ├── iptvPlayer.ico       ← uygulama ikonu
    ├── rsrc.syso            ← önceden derlenmiş kaynaklar (ikon + manifest)
    ├── go.mod               ← modül tanımı
    ├── go.sum               ← bağımlılık hash'leri
    ├── run.bat              ← kaynak koddan çalıştırma scripti
    └── derle.bat            ← release derleme scripti
```

---

## Bağımlılıklar

| Paket | Amaç |
|---|---|
| `github.com/lxn/walk` | Windows yerel GUI framework'ü |
| `github.com/lxn/win` | Düşük seviye Windows API bağlamaları |
| `golang.org/x/sys` | Sistem çağrıları — DLL yükleme, kayıt defteri erişimi |
| `modernc.org/sqlite` | Saf Go SQLite sürücüsü (CGO gerektirmez) |

---

## VLC Neden Gerekli?

Uygulama, video oynatmak için sisteminizde kurulu VLC'nin `libvlc.dll` dosyasını Windows `LoadLibrary` API'si ile çalışma zamanında yükler. CGO veya C derleyicisi gerekmez; tek bir `go build` komutu yeterlidir. VLC uygulamaya dahil edilmez, ayrıca kurulması gerekir. VLC bulunamazsa uygulama başlarken bir hata mesajı gösterir.

---

## Lisans

MIT Lisansı — detaylar için [LICENSE](LICENSE) dosyasına bakın.
