//go:build linux

package notificator

import (
	"os/exec"
)

type linuxNotificator struct {
	AppName string
}

func New(o Options) *Notificator {
	return &Notificator{
		notifier:    linuxNotificator{AppName: o.AppName},
		defaultIcon: o.DefaultIcon,
	}
}

func (l linuxNotificator) push(title string, text string, iconPath string, redirectUrl string) error {
	return exec.Command("notify-send", "-a", l.AppName, "-i", iconPath, title, text).Run()
}

// Causes the notification to stick around until clicked.
func (l linuxNotificator) pushCritical(title string, text string, iconPath string, redirectUrl string) error {
	return exec.Command("notify-send", "-a", l.AppName, "-i", iconPath, title, text, "-u", "critical").Run()
}
