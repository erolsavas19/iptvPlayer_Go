# IPTV Player — Go Edition

A lightweight Windows IPTV player written in Go, with native UI (lxn/walk) and VLC-powered video playback.

> Türkçe açıklama aşağıdadır.

---

## Özellikler

- M3U / M3U8 playlist desteği (yerel dosya veya URL)
- VLC tabanlı video oynatma (libvlc.dll, CGO gerektirmez)
- Kanal arama ve kategori filtreleme
- Favoriler (SQLite veritabanı, saf Go)
- EPG — Elektronik Program Rehberi görüntüleme
- Tam ekran modu (çift tıklama ile geçiş)
- Kanal logosu gösterimi
- Başlangıçta otomatik yüklenecek URL ayarı
- Türkçe / İngilizce kullanıcı arayüzü
- Windows DPI aware (yüksek çözünürlük desteği)

---

## Sistem Gereksinimleri

| Gereksinim | Sürüm | İndirme |
|---|---|---|
| İşletim Sistemi | Windows 10 / 11 (64-bit) | — |
| Go | 1.21 veya üzeri | https://golang.org/dl/ |
| VLC Media Player | 3.x veya üzeri (64-bit) | https://www.videolan.org/vlc/ |

> **Önemli:** Bu uygulama yalnızca Windows'ta çalışır. Kullanılan GUI kütüphanesi (`lxn/walk`) yalnızca Windows API'yi hedefler.

---

## Proje Yapısı

```
iptvPlayr_GO/
├── README.md
├── .gitignore
├── iptvPlayer_go.exe        ← Derlenmiş uygulama (git'e dahil değil)
└── go/                      ← Tüm kaynak kodlar
    ├── main.go              ← Giriş noktası; log, panic kurtarma
    ├── app.go               ← IPTVPlayer yapısı; M3U yükleme, oynatma, favoriler
    ├── mainwindow.go        ← Ana pencere ve tüm UI bileşenleri
    ├── video_win.go         ← VLC video render alanı (Windows subclass)
    ├── fullscreen_win.go    ← Tam ekran geçiş mantığı
    ├── vlc.go               ← VLC libvlc.dll entegrasyonu (saf Go, DLL yükleme)
    ├── db.go                ← SQLite: favoriler ve uygulama ayarları
    ├── m3u.go               ← M3U/M3U8 parser
    ├── dialogs.go           ← Favoriler, EPG, URL giriş pencereleri
    ├── lang.go              ← TR/EN dil dizileri ve çeviri altyapısı
    ├── manifest_win.go      ← Windows DPI aware manifest gömme
    ├── app.manifest         ← Windows uygulama manifestı (XML)
    ├── iptvPlayer.ico       ← Uygulama ikonu
    ├── rsrc.syso            ← Gömülü kaynaklar (ikon + manifest, önceden derlenmiş)
    ├── go.mod               ← Go modül tanımı
    ├── go.sum               ← Bağımlılık hash'leri (değiştirmeyin)
    ├── run.bat              ← Kaynak koddan çalıştırma scripti
    └── derle.bat            ← Release exe derleme scripti
```

---

## Kurulum ve Çalıştırma (Windows)

### Adım 1 — Go'yu Kurun

1. https://golang.org/dl/ adresinden Windows için `.msi` dosyasını indirin
2. Kurulumu tamamlayın (PATH otomatik eklenir)
3. Terminal açıp doğrulayın:
   ```
   go version
   ```

### Adım 2 — VLC'yi Kurun

1. https://www.videolan.org/vlc/ adresinden **64-bit** VLC'yi indirin
2. Varsayılan konuma kurun (`C:\Program Files\VideoLAN\VLC\`)

### Adım 3 — Projeyi İndirin

```bash
git clone https://github.com/KULLANICI/iptvPlayr_GO.git
cd iptvPlayr_GO
```

### Adım 4 — Bağımlılıkları İndirin

Python'daki `pip install -r requirements.txt` karşılığı:

```bash
cd go
go mod download
```

Bu komut `go.mod` dosyasındaki tüm kütüphaneleri otomatik olarak indirir. İnternetbağlantısı gerektirir, ilk çalıştırmada 1-2 dakika sürebilir.

### Adım 5 — Çalıştırın

**Yöntem A — Batch dosyası ile (en kolay):**

`go` klasöründeki `run.bat` dosyasına çift tıklayın.

**Yöntem B — Terminal ile:**

```bash
cd go
go run .
```

> `go run .` ile çalıştırıldığında arka planda küçük bir konsol penceresi açılır; geliştirme sırasında bu normaldir.

---

## Exe Derleme (Release Build)

### Hızlı Derleme

```bash
cd go
go build -ldflags="-H windowsgui -s -w" -o ../iptvPlayer_go.exe .
```

Bayrak açıklamaları:
- `-H windowsgui` — konsol penceresi açılmaz (saf GUI uygulaması)
- `-s -w` — hata ayıklama sembollerini çıkarır, exe dosyası küçülür

### İkon + Manifest Gömülü Tam Derleme

Önce `rsrc` aracını bir kere kurun:

```bash
go install github.com/akavel/rsrc@latest
```

Ardından `go/derle.bat` dosyasını çalıştırın. Bu script:
1. `rsrc.exe` ile ikonu ve manifestı `rsrc.syso` olarak gömer
2. `go build` ile release exe üretir

---

## Go Paket Yönetimi — Python ile Karşılaştırma

| Python | Go | Açıklama |
|--------|-----|----------|
| `pip install paket` | `go get github.com/paket/adı` | Bağımlılık ekle |
| `pip install -r requirements.txt` | `go mod download` | Mevcut bağımlılıkları indir |
| `python script.py` | `go run .` | Kaynak kodu çalıştır |
| `requirements.txt` | `go.mod` + `go.sum` | Bağımlılık listesi |
| PyInstaller ile exe | `go build -o program.exe .` | Derleme |

---

## Bağımlılıklar

| Paket | Amaç |
|-------|-------|
| `github.com/lxn/walk` | Windows native GUI framework |
| `github.com/lxn/win` | Windows API Go bağlamaları |
| `golang.org/x/sys` | Düşük seviye sistem çağrıları (DLL, registry) |
| `modernc.org/sqlite` | Saf Go SQLite (CGO gerektirmez) |

---

## VLC Neden Gerekli?

Uygulama, video oynatmak için sisteminizde kurulu VLC'nin `libvlc.dll` dosyasını çalışma zamanında yükler. CGO kullanılmaz; DLL doğrudan Windows API (`LoadLibrary`) ile çağrılır. Bu tasarım sayesinde:

- CGO toolchain kurulumu gerekmez (MinGW, MSYS2 vb.)
- Tek ve küçük bir `.exe` dosyası yeterlidir
- `libvlc.dll` projeye eklenmesi gerekmez

VLC kurulu değilse uygulama başlarken bir hata mesajı gösterir.

---

## Lisans

MIT License. Detaylar için [LICENSE](LICENSE) dosyasına bakın.
