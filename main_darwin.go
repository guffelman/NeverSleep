//go:build darwin
// +build darwin

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/micmonay/keybd_event"
)

const (
	appName    = "NeverSleep"
	appID      = "com.garrettkeith." + appName
	timeFormat = "03:04 PM"
)

var (
	version = "2.5.0"
)

var (
	active   bool
	settings Settings
	logger   *log.Logger
)

type Settings struct {
	ActiveDays [7]bool
	StartTime  [7]string // An array to store start time for each day
	EndTime    [7]string // An array to store end time for each day
}

func main() {
	settingsDir, settingsFilePath := settingsPath()
	initializeLogger(settingsDir)
	appInstance := createAppInstance(appID)
	loadSettings(settingsFilePath)
	active = true
	go keepAwake()

	updateMenu(appInstance, settingsDir)
	//convertIcon()

	RunActivationPolicy(appInstance, func() {
		// Your code here
		logger.Println("Activation policy set")
	})

	appInstance.Run()

}

func initializeLogger(settingsDir string) {
	logFilePath := filepath.Join(settingsDir, "debug.log")
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	logger = log.New(logFile, "", log.LstdFlags)
}

func createAppInstance(appID string) fyne.App {
	myApp := app.NewWithID(appID)
	myApp.SetIcon(&fyne.StaticResource{
		StaticName:    "icon.ico",
		StaticContent: icon,
	})
	return myApp
}

func updateMenu(appInstance fyne.App, settingsDir string) {
	var activeMenu *fyne.MenuItem
	if active {
		activeMenu = fyne.NewMenuItem("Active", func() {
			active = false
			updateMenu(appInstance, settingsDir)
			logger.Println("Changed to Inactive")
		})
	} else {
		activeMenu = fyne.NewMenuItem("Inactive", func() {
			active = true
			go keepAwake()
			updateMenu(appInstance, settingsDir)
			logger.Println("Changed to Active")
		})
	}

	activeHours := fyne.NewMenuItem("Active Hours", func() {
		showActiveHours(appInstance, settingsDir)
	})

	about := fyne.NewMenuItem("About", func() {
		aboutWindow := appInstance.NewWindow("About")
		aboutWindow.SetContent(fyne.NewContainerWithLayout(
			layout.NewVBoxLayout(),
			container.NewVBox(
				widget.NewLabel("NeverSleep - v2.5"),
				widget.NewLabel("By Garrett Uffelman"),
				widget.NewLabel("Keeps your computer awake and Teams/Slack/Etc 'Active'."),
				container.NewHBox(layout.NewSpacer(), widget.NewButton("Close", func() {
					aboutWindow.Close()
				}), layout.NewSpacer()),
				widget.NewButton("Show Log", func() {
					showLog(settingsDir)
				}),
			),
		))
		aboutWindow.Resize(fyne.NewSize(400, 200))
		aboutWindow.Show()
	})

	quit := fyne.NewMenuItem("Quit", func() {
		appInstance.Quit()
	})

	if desk, ok := fyne.CurrentApp().(desktop.App); ok {
		desk.SetSystemTrayMenu(fyne.NewMenu("", activeMenu, activeHours, about, quit))
	}
}

func showActiveHours(myApp fyne.App, settingsDir string) {
	activeHoursWindow := myApp.NewWindow("Active Hours")
	activeHoursWindow.SetFixedSize(true) // Set the window to a fixed size
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	items := []fyne.CanvasObject{}

	for i, day := range days {
		dayIndex := i // Store the day index
		dayCheck := widget.NewCheck(day, nil)
		dayCheck.SetChecked(settings.ActiveDays[dayIndex])
		items = append(items, dayCheck)

		startTimeEntry := widget.NewEntry()
		startTimeEntry.SetText(settings.StartTime[dayIndex]) // Set the start time for this day
		items = append(items, startTimeEntry)

		endTimeEntry := widget.NewEntry()
		endTimeEntry.SetText(settings.EndTime[dayIndex]) // Set the end time for this day
		items = append(items, endTimeEntry)
	}

	saveButton := widget.NewButton("Save", func() {
		for i, item := range items[:21] { // Update the range to include all items
			switch i % 3 {
			case 0: // Checkboxes
				settings.ActiveDays[i/3] = item.(*widget.Check).Checked
			case 1: // Start times
				settings.StartTime[i/3] = item.(*widget.Entry).Text
			case 2: // End times
				settings.EndTime[i/3] = item.(*widget.Entry).Text
			}
		}

		logger.Printf("Saving settings:", settings)
		saveSettings(settingsDir) // Pass the settingsDir as an argument
		go keepAwake()
		activeHoursWindow.Close()
	})

	items = append(items, saveButton)
	vbox := container.NewVBox(items...)
	scrollContainer := container.NewScroll(vbox)
	scrollContainer.Resize(fyne.NewSize(200, 400))
	activeHoursWindow.SetContent(scrollContainer)
	activeHoursWindow.Resize(fyne.NewSize(400, 600)) // Set the initial size of the window
	activeHoursWindow.Show()
}

func keepAwake() {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		logger.Panic(err)
	}

	// For linux, it is very important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	// press and release the f13 key
	kb.SetKeys(keybd_event.VK_F13)

	const settingsTimeFormat = "03:04 PM"
	const currentTimeFormat = "03:04 PM"

	for {
		now := time.Now()
		currentDay := int(now.Weekday())
		currentDayIndex := (currentDay + 6) % 7
		currentTime := now.Format(currentTimeFormat)
		currentTimeParsed, err := time.Parse(currentTimeFormat, currentTime) // Parse current time

		if err != nil {
			logger.Panic(err)
		}

		startTime, err := time.Parse(settingsTimeFormat, settings.StartTime[currentDayIndex])
		if err != nil {
			logger.Panic(err)
		}

		endTime, err := time.Parse(settingsTimeFormat, settings.EndTime[currentDayIndex])
		if err != nil {
			logger.Panic(err)
		}

		activeDay := settings.ActiveDays[currentDayIndex]
		activeTime := currentTimeParsed.After(startTime) && currentTimeParsed.Before(endTime) // Compare using Before and After methods

		if active && activeDay && activeTime {
			logger.Println("Pressing F13")
			kb.Press()
			time.Sleep(10 * time.Millisecond)
			logger.Println("Releasing F13")
			kb.Release()
			time.Sleep(10 * time.Second)
		} else {
			reason := "App set to inactive"
			if active {
				if !activeDay {
					reason = "Outside of day range"
				} else if !activeTime {
					reason = "Outside of time range"
				}
			}
			logger.Printf("Sleeping for 5 minutes - Reason: %s\n", reason)
			time.Sleep(5 * time.Minute)
		}
	}
}

func settingsPath() (string, string) {
	usr, err := user.Current()
	if err != nil {
		logger.Panic(err)
	}
	dir := filepath.Join(usr.HomeDir, "NeverSleep")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		logger.Panic(err)
	}
	if runtime.GOOS == "windows" {
		dir = filepath.Join(usr.HomeDir, "Documents", "NeverSleep")
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Panic(err)
		}
	}
	return dir, filepath.Join(dir, "settings.json")
}

func saveSettings(settingsDir string) {
	filePath := filepath.Join(settingsDir, "settings.json")
	logger.Printf("Saving settings to: %s", filePath) // Use logger.Printf instead of fmt.Printf
	file, err := os.Create(filePath)
	if err != nil {
		logger.Fatalf("error creating settings file: %v", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(settings)
	if err != nil {
		logger.Fatalf("error encoding settings: %v", err)
	}
	logger.Println("Settings saved successfully.") // Use logger.Println instead of fmt.Println
}

func loadSettings(settingsFilePath string) {
	file, err := os.Open(settingsFilePath)
	if err != nil {
		// File does not exist, use default settings
		settings = Settings{
			ActiveDays: [7]bool{true, true, true, true, true, false, false}, // Starting with Monday
			StartTime:  [7]string{"08:00 AM", "08:00 AM", "08:00 AM", "08:00 AM", "08:00 AM", "08:00 AM", "08:00 AM"},
			EndTime:    [7]string{"05:00 PM", "05:00 PM", "05:00 PM", "05:00 PM", "05:00 PM", "05:00 PM", "05:00 PM"},
		}
		saveSettings(filepath.Dir(settingsFilePath))
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&settings)
	if err != nil {
		logger.Panic(err)
	}
}

func extractIcon() []byte {
	iconFile, err := os.Open("./img/icon.ico")
	if err != nil {
		logger.Panic(err)
	}
	defer iconFile.Close()
	iconData, err := ioutil.ReadAll(iconFile)
	if err != nil {
		logger.Panic(err)
	}
	return iconData
}

func showLog(settingsDir string) {
	logFilePath := filepath.Join(settingsDir, "debug.log")
	cmd := exec.Command("notepad.exe", logFilePath)
	err := cmd.Run()
	if err != nil {
		logger.Printf("Error opening log file: %v\n", err)
	}
}
