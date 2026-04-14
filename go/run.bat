@echo off
cd /d %~dp0

echo [1/3] Go kurulumu kontrol ediliyor...
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo.
    echo  HATA: Go bulunamadi.
    echo  Lutfen Go kurun: https://golang.org/dl/
    echo  Kurulumdan sonra terminal/CMD'yi yeniden baslatin.
    pause
    exit /b 1
)
for /f "tokens=3" %%v in ('go version') do echo  Go surumu: %%v

echo.
echo [2/3] Bagimliliklar indiriliyor (ilk calistirmada 1-2 dakika surebilir)...
go mod download
if %errorlevel% neq 0 (
    echo.
    echo  HATA: Bagimliliklar indirilemedi.
    echo  Internet baglantinizi kontrol edin.
    pause
    exit /b 1
)
echo  Bagimliliklar hazir.

echo.
echo [3/3] IPTV Player baslatiliyor...
echo  (Gelistirme modunda arka planda konsol penceresi acik kalabilir)
echo.
go run .

if %errorlevel% neq 0 (
    echo.
    echo  HATA: Program baslatılamadi. Yukaridaki hata mesajini inceleyin.
    pause
)
