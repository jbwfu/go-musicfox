//go:build windows

package notify

import (
	"github.com/gen2brain/beeep"
	"github.com/go-musicfox/notificator"
)

type Notificator struct {
	*notificator.Notificator
}

func NewNotificator(o notificator.Options) *Notificator {
	return &Notificator{
		Notificator: notificator.New(o),
	}
}

func (n Notificator) Push(urgency, title, text, iconPath, redirectUrl, groupId string) error {
	return beeep.Notify(title, text, iconPath)
}
