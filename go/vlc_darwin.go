//go:build darwin

// vlc_darwin.go — macOS için VLC entegrasyonu.
// VLC kendi penceresinde açılır (NSView gömme ileride eklenebilir).

package main

import (
	"fmt"
	"os"
	"os/exec"

	"fyne.io/fyne/v2/canvas"
)

// VLCPlayer macOS için exec.Command tabanlı VLC player
type VLCPlayer struct {
	cmd *exec.Cmd
}

func findVLCBinary() (string, error) {
	for _, p := range []string{
		"/Applications/VLC.app/Contents/MacOS/VLC",
		"/usr/local/bin/vlc",
		"/opt/homebrew/bin/vlc",
	} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	path, err := exec.LookPath("vlc")
	if err != nil {
		return "", fmt.Errorf(
			"VLC bulunamadı.\nbrew install --cask vlc\nhttps://www.videolan.org/vlc/",
		)
	}
	return path, nil
}

func NewVLCPlayer() (*VLCPlayer, error) {
	if _, err := findVLCBinary(); err != nil {
		return nil, err
	}
	return &VLCPlayer{}, nil
}

// SetVideoCanvas macOS'ta no-op (henüz NSView gömme uygulanmadı)
func (v *VLCPlayer) SetVideoCanvas(_ *canvas.Image) {}

func (v *VLCPlayer) SetHWND(_ uintptr) {}

func (v *VLCPlayer) PlayURL(url string) error {
	v.Stop()
	bin, err := findVLCBinary()
	if err != nil {
		return err
	}
	v.cmd = exec.Command(bin, "--no-video-title-show", "--quiet", "--play-and-exit", url)
	if err := v.cmd.Start(); err != nil {
		v.cmd = nil
		return fmt.Errorf("VLC başlatılamadı: %w", err)
	}
	go func() {
		if v.cmd != nil {
			v.cmd.Wait()
		}
	}()
	return nil
}

func (v *VLCPlayer) Stop() {
	if v.cmd != nil && v.cmd.Process != nil {
		v.cmd.Process.Kill()
		v.cmd.Wait()
		v.cmd = nil
	}
}

func (v *VLCPlayer) SetFullscreen(_ bool) {}

func (v *VLCPlayer) Release() { v.Stop() }
