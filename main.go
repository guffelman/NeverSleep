package main

import (
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/micmonay/keybd_event"
    "github.com/tawesoft/golib/v2/dialog"
)

var active bool

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("NeverAway v1.0")
	systray.SetTooltip("NeverAway v1.0 - by @garrettuffelman")
	mActive := systray.AddMenuItem("Inactive", "Click to activate")
	go func() {
		for {
			<-mActive.ClickedCh
			if mActive.Checked() {
				mActive.Uncheck()
				active = false
				mActive.SetTitle("Inactive")
				go keepawake()
			} else {
				mActive.Check()
				active = true
				mActive.SetTitle("Active")
				go keepawake()
			}
		}
	}()

	mAbout := systray.AddMenuItem("About", "About NeverAway")
	go func() {
		<-mAbout.ClickedCh {
			dialog.Alert("NeverAway v1.0 - by @garrettuffelman")
		}

	}()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

}

func onExit() {
	systray.Quit()
}

func main() {
	// create a new taskbar item
	systray.Run(onReady, onExit)
	// on exit, close the taskbar item

}

// when function is called with a 1, it will keep the computer awake
// when function is called with a 0, it will allow the computer to sleep

func keepawake() {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}

	// For linux, it is very important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(3 * time.Second)
	}

	// press and release the f13 key
	kb.SetKeys(keybd_event.VK_F13)

	for {
		if active {
			kb.Press()
			time.Sleep(5000 * time.Millisecond)
			kb.Release()
			time.Sleep(5000 * time.Millisecond)

		} else {
			break

		}

	}

	// when active = false, do nothing

}
