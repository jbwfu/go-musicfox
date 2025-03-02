package notificator

import (
	"os/exec"
	"strconv"
	"strings"
)

type Options struct {
	DefaultIcon string
	AppName     string
	OSXSender   string
}

const (
	UrNormal   = "normal"
	UrCritical = "critical"
)

type notifier interface {
	push(title string, text string, iconPath string, redirectUrl string) error
	pushCritical(title string, text string, iconPath string, redirectUrl string) error
}

type Notificator struct {
	notifier    notifier
	defaultIcon string
}

func (n Notificator) Push(urgency string, title string, text string, iconPath string, redirectUrl string) error {
	if n.notifier == nil {
		return nil
	}
	icon := n.defaultIcon

	if iconPath != "" {
		icon = iconPath
	}

	if urgency == UrCritical {
		return n.notifier.pushCritical(title, text, icon, redirectUrl)
	}

	return n.notifier.push(title, text, icon, redirectUrl)
}

// Helper function for macOS

func CheckTermNotif() bool {
	// Checks if terminal-notifier exists, and is accessible.
	if err := exec.Command("which", "terminal-notifier"); err != nil {
		return false
	}
	// no error, so return true. (terminal-notifier exists)
	return true
}

func CheckMacOSVersion() bool {
	// Checks if the version of macOS is 10.9 or Higher (osascript support for notifications.)

	cmd := exec.Command("sw_vers", "-productVersion")
	check, _ := cmd.Output()

	version := strings.Split(strings.TrimSpace(string(check)), ".")

	// semantic versioning of macOS

	if len(version) < 2 {
		return false
	}

	major, _ := strconv.Atoi(version[0])
	minor, _ := strconv.Atoi(version[1])

	if major < 10 {
		return false
	} else if major == 10 && minor < 9 {
		return false
	} else {
		return true
	}
}
