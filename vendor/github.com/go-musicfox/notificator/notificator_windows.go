//go:build windows

package notificator

import (
	"github.com/gen2brain/beeep"
)

type windowsNotificator struct{}

func New(o Options) *Notificator {
	return &Notificator{
		notifier:    windowsNotificator{},
		defaultIcon: o.DefaultIcon,
	}
}

func (w windowsNotificator) push(title string, text string, iconPath string, redirectUrl string) error {
	return beeep.Notify(title, text, iconPath)
}

// Causes the notification to stick around until clicked.
func (w windowsNotificator) pushCritical(title string, text string, iconPath string, redirectUrl string) error {
	return beeep.Notify(title, text, iconPath)
}
