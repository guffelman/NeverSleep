//go:build darwin
// +build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

int
SetActivationPolicy(void) {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    return 0;
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
)

func setActivationPolicy() {
	fmt.Println("Setting ActivationPolicy")
	C.SetActivationPolicy()
}

func RunActivationPolicy(appInstance fyne.App, callback func()) {
	if runtime.GOOS == "darwin" {
		appInstance.Lifecycle().SetOnStarted(func() {
			go func() {
				time.Sleep(10 * time.Millisecond)
				setActivationPolicy()
				callback()
			}()
		})
	} else {
		callback()
	}
}
