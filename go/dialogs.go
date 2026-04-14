//go:build windows

package main

import (
	"fmt"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

// showInputDialog metin girişi için modal dialog gösterir
func showInputDialog(owner walk.Form, title, prompt, defaultVal string) string {
	var dlg *walk.Dialog
	var edit *walk.LineEdit
	result := ""

	Dialog{
		AssignTo: &dlg,
		Title:    title,
		MinSize:  Size{Width: 460, Height: 140},
		Layout:   VBox{Margins: Margins{Left: 12, Top: 10, Right: 12, Bottom: 10}},
		Children: []Widget{
			Label{Text: prompt},
			LineEdit{AssignTo: &edit, Text: defaultVal},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text:    Lang.DlgBtnOK,
						MinSize: Size{Width: 80},
						OnClicked: func() {
							result = edit.Text()
							dlg.Accept()
						},
					},
					PushButton{
						Text:      Lang.DlgBtnCancel,
						MinSize:   Size{Width: 80},
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Create(owner)

	dlg.Run()
	return result
}

// ShowFavoritesWindow favoriler penceresini açar (Dialog = modal)
func ShowFavoritesWindow(p *IPTVPlayer) {
	favorites, err := p.db.GetFavorites()
	if err != nil {
		walk.MsgBox(p.mw, Lang.DlgError, "Favoriler yüklenemedi!", walk.MsgBoxIconError)
		return
	}

	if len(favorites) == 0 {
		walk.MsgBox(p.mw, Lang.FavWindowTitle, Lang.FavMsgEmpty, walk.MsgBoxIconInformation)
		return
	}

	var dlg *walk.Dialog
	var listBox *walk.ListBox

	Dialog{
		AssignTo: &dlg,
		Title:    Lang.FavWindowTitle,
		MinSize:  Size{Width: 500, Height: 420},
		Layout:   VBox{Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}},
		Children: []Widget{
			Label{
				Text:      Lang.FavWindowTitle,
				Font:      Font{Bold: true, PointSize: 12},
				TextColor: walk.RGB(13, 115, 119),
			},
			ListBox{
				AssignTo: &listBox,
				Model:    newChannelModel(favorites),
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:    Lang.FavBtnPlay,
						MinSize: Size{Width: 90},
						OnClicked: func() {
							idx := listBox.CurrentIndex()
							if idx < 0 || idx >= len(favorites) {
								walk.MsgBox(dlg, Lang.DlgWarning, Lang.FavMsgSelect, walk.MsgBoxIconWarning)
								return
							}
							ch := favorites[idx]
							dlg.Accept()
							p.playURL(ch.URL, ch.Name)
						},
					},
					PushButton{
						Text:    Lang.FavBtnRemove,
						MinSize: Size{Width: 90},
						OnClicked: func() {
							idx := listBox.CurrentIndex()
							if idx < 0 || idx >= len(favorites) {
								walk.MsgBox(dlg, Lang.DlgWarning, Lang.FavMsgSelect, walk.MsgBoxIconWarning)
								return
							}
							ch := favorites[idx]
							if walk.MsgBox(dlg,
								Lang.FavMsgConfirmTitle,
								fmt.Sprintf(Lang.FavMsgConfirmDel, ch.Name),
								walk.MsgBoxYesNo|walk.MsgBoxIconQuestion,
							) == win.IDYES {
								if removeErr := p.db.RemoveFavorite(ch.Name, ch.URL); removeErr != nil {
									walk.MsgBox(dlg, Lang.DlgError, "Silinemedi: "+removeErr.Error(), walk.MsgBoxIconError)
									return
								}
								favorites, _ = p.db.GetFavorites()
								listBox.SetModel(newChannelModel(favorites))
							}
						},
					},
					HSpacer{},
					PushButton{
						Text:      Lang.FavBtnClose,
						MinSize:   Size{Width: 80},
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Create(p.mw)

	dlg.Run()
}

// ShowEPGWindow EPG rehber penceresini açar (Dialog = modal)
func ShowEPGWindow(p *IPTVPlayer, channelName string, programs []EPGProgram) {
	if len(programs) == 0 {
		walk.MsgBox(p.mw, Lang.EPGMsgNoDataTitle, Lang.EPGMsgNoData, walk.MsgBoxIconInformation)
		return
	}

	now := time.Now()
	epg := newEPGModel(now, programs)

	var dlg *walk.Dialog

	Dialog{
		AssignTo: &dlg,
		Title:    fmt.Sprintf(Lang.EPGWindowTitle, channelName),
		MinSize:  Size{Width: 800, Height: 520},
		Layout:   VBox{Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}},
		Children: []Widget{
			Label{
				Text:      channelName,
				Font:      Font{Bold: true, PointSize: 12},
				TextColor: walk.RGB(13, 115, 119),
			},
			Label{Text: now.Format("Monday, 02 January 2006")},
			TableView{
				AlternatingRowBG: true,
				Columns: []TableViewColumn{
					{Title: Lang.EPGColTime, Width: 120},
					{Title: Lang.EPGColTitle, Width: 200},
					{Title: Lang.EPGColDesc, Width: 350},
				},
				Model: epg,
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text:      Lang.EPGBtnClose,
						MinSize:   Size{Width: 80},
						OnClicked: func() { dlg.Cancel() },
					},
					HSpacer{},
				},
			},
		},
	}.Create(p.mw)

	dlg.Run()
}

// epgModel TableView için EPG modeli
type epgModel struct {
	walk.TableModelBase
	rows     []epgRow
	now      time.Time
	programs []EPGProgram
}

type epgRow struct {
	Time  string
	Title string
	Desc  string
}

func newEPGModel(now time.Time, programs []EPGProgram) *epgModel {
	m := &epgModel{now: now, programs: programs}
	for _, prog := range programs {
		m.rows = append(m.rows, epgRow{
			Time:  fmt.Sprintf("%s - %s", prog.Start.Format("15:04"), prog.End.Format("15:04")),
			Title: prog.Title,
			Desc:  prog.Desc,
		})
	}
	return m
}

func (m *epgModel) RowCount() int { return len(m.rows) }

func (m *epgModel) Value(row, col int) interface{} {
	if row >= len(m.rows) {
		return ""
	}
	r := m.rows[row]
	switch col {
	case 0:
		return r.Time
	case 1:
		return r.Title
	case 2:
		return r.Desc
	}
	return ""
}

func (m *epgModel) StyleCell(style *walk.CellStyle) {
	row := style.Row()
	if row < len(m.programs) {
		prog := m.programs[row]
		if !m.now.Before(prog.Start) && m.now.Before(prog.End) {
			style.BackgroundColor = walk.RGB(255, 230, 0)
			style.TextColor = walk.RGB(0, 0, 0)
		}
	}
}

// ShowAboutDialog hakkında dialog'unu gösterir
func ShowAboutDialog(owner *walk.MainWindow) {
	var dlg *walk.Dialog

	Dialog{
		AssignTo: &dlg,
		Title:    Lang.DlgAboutTitle,
		MinSize:  Size{Width: 420, Height: 240},
		Layout:   VBox{Margins: Margins{Left: 20, Top: 15, Right: 20, Bottom: 15}},
		Children: []Widget{
			Label{
				Text:      "IPTV Player  Ver. 1.0",
				Font:      Font{Bold: true, PointSize: 14},
				TextColor: walk.RGB(13, 115, 119),
			},
			VSpacer{Size: 4},
			Label{Text: "Bu program Bedavadır (Freeware), para ile satılmaz."},
			Label{Text: "Geliştirici:"},
			LinkLabel{Text: `<a href="https://www.freewaretr.com">https://www.freewaretr.com</a>`},
			LinkLabel{Text: `<a href="https://www.caprazbilgi.com">https://www.caprazbilgi.com</a>`},
			LinkLabel{Text: `<a href="https://www.youtube.com/caprazbilgi">https://www.youtube.com/caprazbilgi</a>`},
			VSpacer{},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text:      Lang.DlgBtnOK,
						MinSize:   Size{Width: 80},
						OnClicked: func() { dlg.Accept() },
					},
					HSpacer{},
				},
			},
		},
	}.Create(owner)

	dlg.Run()
}
