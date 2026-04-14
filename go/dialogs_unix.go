//go:build linux || darwin

// dialogs_unix.go — Linux ve macOS için Fyne tabanlı dialog pencereleri.
// Windows sürümündeki dialogs.go (lxn/walk) ile aynı işlevselliği sağlar.

package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ─── Giriş Dialogu ───────────────────────────────────────────────────────────

// showInputDialog metin girişi için Fyne dialog gösterir.
// Sonuç callback fonksiyonu aracılığıyla iletilir (Fyne asenkron dialog modeli).
// OK tıklanırsa girilen metin, İptal'de boş string döner.
func showInputDialog(win fyne.Window, title, prompt, defaultVal string, callback func(string)) {
	entry := widget.NewEntry()
	entry.SetText(defaultVal)
	entry.SetPlaceHolder(prompt)

	content := container.NewVBox(
		widget.NewLabel(prompt),
		entry,
	)

	d := dialog.NewCustomConfirm(title, Lang.DlgBtnOK, Lang.DlgBtnCancel, content,
		func(ok bool) {
			if ok {
				callback(entry.Text)
			} else {
				callback("")
			}
		},
		win,
	)
	d.Resize(fyne.NewSize(460, 180))
	d.Show()
}

// ─── Favoriler Penceresi ──────────────────────────────────────────────────────

// ShowFavoritesWindow favoriler listesini modal dialog olarak gösterir
func ShowFavoritesWindow(p *IPTVPlayer) {
	favorites, err := p.db.GetFavorites()
	if err != nil {
		dialog.ShowError(fmt.Errorf("Favoriler yüklenemedi: %w", err), p.window)
		return
	}

	if len(favorites) == 0 {
		dialog.ShowInformation(Lang.FavWindowTitle, Lang.FavMsgEmpty, p.window)
		return
	}

	selectedIdx := -1

	list := widget.NewList(
		func() int { return len(favorites) },
		func() fyne.CanvasObject { return widget.NewLabel("kanal") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if int(id) < len(favorites) {
				obj.(*widget.Label).SetText(favorites[id].Name)
			}
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		selectedIdx = int(id)
	}

	var dlg dialog.Dialog

	btnPlay := widget.NewButton(Lang.FavBtnPlay, func() {
		if selectedIdx < 0 || selectedIdx >= len(favorites) {
			dialog.ShowInformation(Lang.DlgWarning, Lang.FavMsgSelect, p.window)
			return
		}
		ch := favorites[selectedIdx]
		dlg.Hide()
		p.playURL(ch.URL, ch.Name)
	})

	btnRemove := widget.NewButton(Lang.FavBtnRemove, func() {
		if selectedIdx < 0 || selectedIdx >= len(favorites) {
			dialog.ShowInformation(Lang.DlgWarning, Lang.FavMsgSelect, p.window)
			return
		}
		ch := favorites[selectedIdx]
		dialog.ShowConfirm(
			Lang.FavMsgConfirmTitle,
			fmt.Sprintf(Lang.FavMsgConfirmDel, ch.Name),
			func(ok bool) {
				if !ok {
					return
				}
				if err := p.db.RemoveFavorite(ch.Name, ch.URL); err != nil {
					dialog.ShowError(err, p.window)
					return
				}
				// Listeyi güncelle
				favorites, _ = p.db.GetFavorites()
				selectedIdx = -1
				list.Refresh()
			},
			p.window,
		)
	})

	btnClose := widget.NewButton(Lang.FavBtnClose, func() {
		dlg.Hide()
	})

	buttonBar := container.NewHBox(btnPlay, btnRemove, layout.NewSpacer(), btnClose)
	content := container.NewBorder(nil, buttonBar, nil, nil, list)

	dlg = dialog.NewCustomWithoutButtons(Lang.FavWindowTitle, content, p.window)
	dlg.Resize(fyne.NewSize(520, 420))
	dlg.Show()
}

// ─── EPG Penceresi ───────────────────────────────────────────────────────────

// ShowEPGWindow EPG rehberini tablo olarak gösterir
func ShowEPGWindow(p *IPTVPlayer, channelName string, programs []EPGProgram) {
	if len(programs) == 0 {
		dialog.ShowInformation(Lang.EPGMsgNoDataTitle, Lang.EPGMsgNoData, p.window)
		return
	}

	now := time.Now()

	// Satır verileri
	type epgRow struct {
		timeStr string
		title   string
		desc    string
		current bool
	}
	rows := make([]epgRow, len(programs))
	for i, prog := range programs {
		rows[i] = epgRow{
			timeStr: fmt.Sprintf("%s - %s", prog.Start.Format("15:04"), prog.End.Format("15:04")),
			title:   prog.Title,
			desc:    prog.Desc,
			current: !now.Before(prog.Start) && now.Before(prog.End),
		}
	}

	table := widget.NewTable(
		func() (int, int) { return len(rows) + 1, 3 }, // +1 başlık satırı
		func() fyne.CanvasObject {
			return widget.NewLabel("metin")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)
			if id.Row == 0 {
				// Başlık satırı
				switch id.Col {
				case 0:
					lbl.TextStyle = fyne.TextStyle{Bold: true}
					lbl.SetText(Lang.EPGColTime)
				case 1:
					lbl.TextStyle = fyne.TextStyle{Bold: true}
					lbl.SetText(Lang.EPGColTitle)
				case 2:
					lbl.TextStyle = fyne.TextStyle{Bold: true}
					lbl.SetText(Lang.EPGColDesc)
				}
				return
			}
			row := rows[id.Row-1]
			lbl.TextStyle = fyne.TextStyle{}
			switch id.Col {
			case 0:
				lbl.SetText(row.timeStr)
			case 1:
				if row.current {
					lbl.TextStyle = fyne.TextStyle{Bold: true}
				}
				lbl.SetText(row.title)
			case 2:
				lbl.SetText(row.desc)
			}
		},
	)

	table.SetColumnWidth(0, 130)
	table.SetColumnWidth(1, 210)
	table.SetColumnWidth(2, 360)

	header := widget.NewLabelWithStyle(
		fmt.Sprintf("%s — %s", channelName, now.Format("02 January 2006")),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	content := container.NewBorder(header, nil, nil, nil, table)

	dlg := dialog.NewCustom(
		fmt.Sprintf(Lang.EPGWindowTitle, channelName),
		Lang.EPGBtnClose,
		content,
		p.window,
	)
	dlg.Resize(fyne.NewSize(740, 500))
	dlg.Show()
}
