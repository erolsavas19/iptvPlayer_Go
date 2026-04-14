@echo off
cd /d %~dp0

echo [1/2] Manifest ve simge gömülüyor...
"%USERPROFILE%\go\bin\rsrc.exe" -manifest app.manifest -ico iptvPlayer.ico -o rsrc.syso
if %errorlevel% neq 0 (
    echo UYARI: rsrc.exe bulunamadi, mevcut rsrc.syso kullanilacak.
)

echo [2/2] Derleniyor...
"C:\Program Files\Go\bin\go.exe" build -ldflags="-H windowsgui -s -w" -o ..\iptvPlayer_go.exe .

if %errorlevel% == 0 (
    echo.
    echo Derleme basarili!
    echo Dosya: iptvPlayer_go.exe
) else (
    echo.
    echo HATA: Derleme basarisiz!
)
pause
