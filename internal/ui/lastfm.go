package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

type Lastfm struct {
	baseMenu
}

func NewLastfm(base baseMenu) *Lastfm {
	return &Lastfm{baseMenu: base}
}

func (m *Lastfm) GetMenuKey() string {
	return "last_fm"
}

func (m *Lastfm) MenuViews() []model.MenuItem {
	getControlTitle := func() string {
		if m.netease.lastfm.Tracker.Status() {
			return "关闭功能"
		}
		return "启用功能"
	}

	return []model.MenuItem{
		{Title: "管理授权"},
		{Title: "前往主页"},
		{Title: getControlTitle()},
		{Title: "清空队列", Subtitle: fmt.Sprintf("[共 %d 条]", m.netease.lastfm.Tracker.Count())},
	}
}

func (m *Lastfm) SubMenu(app *model.App, index int) model.Menu {
	switch index {
	case 0:
		return NewLastfmProfile(m.baseMenu)
	case 1:
		m.netease.lastfm.OpenUserHomePage()
	case 2:
		action := func() {
			m.netease.lastfm.Tracker.Toggle()
		}
		return NewConfirmMenu(m.baseMenu, []ConfirmItem{
			{title: model.MenuItem{Title: "确定"}, action: action, backLevel: 1},
		})
	case 3:
		action := func() {
			m.netease.lastfm.Tracker.Clear()

			notify.Notify(notify.NotifyContent{
				Title:   "清除 last.fm Scrobble 队列成功",
				Text:    "Last.fm Scrobble 队列已清除",
				GroupId: types.GroupID,
			})
		}
		return NewConfirmMenu(m.baseMenu, []ConfirmItem{
			{title: model.MenuItem{Title: "确定"}, action: action, backLevel: 1},
		})
	}
	return nil
}

func (m *Lastfm) FormatMenuItem(item *model.MenuItem) {
	item.Subtitle = "[未授权]"
	if !m.netease.lastfm.NeedAuth() {
		if username := m.netease.lastfm.UserName(); username != "" {
			item.Subtitle = fmt.Sprintf("[%s]", username)
		}
	}
}
