package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

type LastfmProfile struct {
	baseMenu
}

func NewLastfmProfile(base baseMenu) *LastfmProfile {
	return &LastfmProfile{
		baseMenu: base,
	}
}

func (m *LastfmProfile) GetMenuKey() string {
	return "lastfm_profile"
}

func (m *LastfmProfile) MenuViews() (menu []model.MenuItem) {
	if !lastfm.IsAvailable() {
		return []model.MenuItem{{Title: "设置 API account", Subtitle: "[待设置]"}}
	}

	getAuthTitle := func() string {
		if m.netease.lastfm.NeedAuth() {
			return "去授权"
		}
		return "取消授权"
	}
	return []model.MenuItem{{Title: "设置 API account", Subtitle: "[已设置]"}, {Title: getAuthTitle()}}
}

func (m *LastfmProfile) SubMenu(app *model.App, index int) model.Menu {
	switch index {
	case 0:
		page := NewLastfmCustomApiPage(m.netease)
		page.AfterAction = func() {
			app.MustMain().RefreshMenuList()
		}
		return NewMenuToPage(m.baseMenu, page)
	case 1:
		if m.netease.lastfm.NeedAuth() {
			page := NewLastfmAuthPage(m.netease)
			page.AfterAction = func() {
				app.MustMain().RefreshMenuList()
			}
			return NewMenuToPage(m.baseMenu, page)
		}

		action := func() {
			m.netease.lastfm.ClearUserInfo()

			notify.Notify(notify.NotifyContent{
				Title:   "清除授权成功",
				Text:    "Last.fm 授权已清除",
				GroupId: types.GroupID,
			})
		}

		return NewConfirmMenu(m.baseMenu, []ConfirmItem{
			{title: model.MenuItem{Title: "确定"}, action: action, backLevel: 2},
		})
	default:
		return nil
	}
}
