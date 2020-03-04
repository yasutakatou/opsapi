package main

import (
	"fmt"

	"./winctl"

	hook "github.com/robotn/gohook"
)

func LiveRecord() {
	fmt.Printf(" - - live recording! you want to end this mode, key press ascii code (%d) - - \n", config.LiveExitAsciiCode)

	altFlag := 0

	EvChan := hook.Start()
	defer hook.End()

	for ev := range EvChan {
		//KeyHold = 4,KeyUp   = 5
		if ev.Kind == 4 || ev.Kind == 5 {
			switch int(ev.Rawcode) {
			case 162, 164: //Ctrl,Alt
				if ev.Kind == 4 {
					altFlag = int(ev.Rawcode)
				} else {
					altFlag = 0
				}
			case config.LiveExitAsciiCode: //Default Escape
				return
			}
		}

		//KeyDown = 3
		if ev.Kind == 3 {
			strs := ""
			if altFlag == 0 {
				switch ev.Rawcode {
				case 8:
					strs = "\\b"
				case 9:
					strs = "\\t"
				case 13:
					strs = "\\n"
				default:
					strs = string(ev.Keychar)
				}
			} else {
				switch altFlag {
				case 162:
					strs = "ctrl+" + string(ev.Keychar)
				case 164:
					strs = "alt+" + string(ev.Keychar)
				}
			}

			if len(strs) > 0 && targetHwnd == winctl.GetWindow("GetForegroundWindow", Debug) && config.Record == true {
				History = append(History, historyData{Command: "ops", Params: strs})
				if Debug == true {
					fmt.Println("liveRecord: ", strs)
				}
			}
		}
	}
}
