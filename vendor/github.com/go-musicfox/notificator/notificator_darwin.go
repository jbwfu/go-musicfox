//go:build darwin

package notificator

import (
	"fmt"
	"os/exec"
	"strings"
)

type osxNotificator struct {
	AppName string
	Sender  string
}

func New(o Options) *Notificator {
	return &Notificator{
		notifier:    osxNotificator{AppName: o.AppName, Sender: o.OSXSender},
		defaultIcon: o.DefaultIcon,
	}
}

func (o osxNotificator) push(title string, text string, iconPath string, redirectUrl string) error {
	// Checks if terminal-notifier exists, and is accessible.

	// if terminal-notifier exists, use it.
	// else, fall back to osascript. (Mavericks and later.)
	if CheckTermNotif() {
		if redirectUrl != "" {
			return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-contentImage", iconPath, "-open", redirectUrl).Run()
		}
		return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-contentImage", iconPath, "-sender", o.Sender).Run()
	} else if CheckMacOSVersion() {
		title = strings.ReplaceAll(title, `"`, `\"`)
		text = strings.ReplaceAll(text, `"`, `\"`)

		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.AppName, title)
		return exec.Command("osascript", "-e", notification).Run()
	}

	// finally falls back to growlnotify.

	return exec.Command("growlnotify", "-n", o.AppName, "--image", iconPath, "-m", title, "--url", redirectUrl).Run()
}

// Causes the notification to stick around until clicked.
func (o osxNotificator) pushCritical(title string, text string, iconPath string, redirectUrl string) error {
	// same function as above...
	if CheckTermNotif() {
		// timeout set to 30 seconds, to show the importance of the notification
		if redirectUrl != "" {
			return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-contentImage", iconPath, "-timeout", "30", "-open", redirectUrl).Run()
		}

		return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-contentImage", iconPath, "-timeout", "30", "-sender", o.Sender).Run()
	} else if CheckMacOSVersion() {
		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.AppName, title)
		return exec.Command("osascript", "-e", notification).Run()
	}

	return exec.Command("growlnotify", "-n", o.AppName, "--image", iconPath, "-m", title, "--url", redirectUrl).Run()
}
