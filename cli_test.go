package main

import (
	"runtime"
	"testing"
	"time"

	"./winctl"
)

func TestSetShebang(t *testing.T) {
	if SetShebang("") != false {
		t.Errorf("can't detect incorrect shebang")
	}

	if SetShebang("#!/bin/bash") != true {
		t.Errorf("can't detect correct shebang")
	}
}

func TestSetExportFormat(t *testing.T) {
	if SetExportFormat("curl -H \"Content-type: application/json\" -X POST http://127.0.0.1:8080/ -d \"{\\\"token\\\":\\\"%1\\\",\\\"command\\\":\\\"##\\\",\\\"params\\\":\\\"##\\\"}\"") != false {
		t.Errorf("can't detect incorrect export format")
	}

	if SetExportFormat("curl -H \"Content-type: application/json\" -X POST http://127.0.0.1:8080/ -d \"{\\\"token\\\":\\\"%1\\\",\\\"command\\\":\\\"#COMMAND#\\\",\\\"params\\\":\\\"#PARAMS#\\\"}\"") != true {
		t.Errorf("can't detect correct export format")
	}
}

func TestRunHistory(t *testing.T) {
	//this code test t function RunHistory , InsertHistory, DeleteHistory, Unset, Insert,  DisplayHistory, ImportHistory and ExportHistory

	Cli = true

	if DisplayHistory() == true {
		t.Errorf("why? your history is not empty")
	}

	cliConfig.Record = true
	cliConfig.Shebang = ""
	History = append(History, historyData{Command: "cliConfigSet", Params: "Record=true"})

	if runtime.GOOS == "linux" {
		resp := InsertHistory("1 exec ls>test.txt")
		if resp.Status == "Error" {
			t.Errorf("can't insert your history")
		}
	} else {
		resp := InsertHistory("1 exec winver")
		if resp.Status == "Error" {
			t.Errorf("can't insert your history")
		}
	}

	time.Sleep(time.Duration(1000) * time.Millisecond)

	DeleteHistory("2")

	DisplayHistory()

	RunHistory()

	if runtime.GOOS == "linux" && Exists("test.txt") == false {
		t.Errorf("RunHistory command failed")
	}

	time.Sleep(time.Duration(1000) * time.Millisecond)

	if runtime.GOOS == "windows" {
		targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)
		if targetHwnd == 0 {
			t.Errorf("RunHistory command failed")
		}
		StringDo(" ")
	}

	ExportHistory("tsv history.txt")

	History = nil

	if DisplayHistory() == true {
		t.Errorf("why? your history is not empty")
	}

	if ImportHistory("history.txt") == false {
		t.Errorf("your history import or export failed")
	}

	DisplayHistory()

	RunHistory()

	time.Sleep(time.Duration(1000) * time.Millisecond)

	if runtime.GOOS == "windows" {
		StringDo(" ")
		targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)
		if targetHwnd == 0 {
			t.Errorf("can't focus target window")
		}
		StringDo(" ")

		Execmd("del history.txt")
	} else {
		Execmd("rm history.txt")
	}
	time.Sleep(time.Duration(1000) * time.Millisecond)
}

func TestRunCliCmd(t *testing.T) {
	if runtime.GOOS == "windows" {
		cliHwnd = winctl.GetWindow("GetForegroundWindow", Debug)
	}

	RunCliCmd("clearHistory", "")

	if DisplayHistory() == true {
		t.Errorf("why? your history is not empty")
	}

	if runtime.GOOS == "windows" {
		RunCliCmd("exec", "winver")
		time.Sleep(time.Duration(1000) * time.Millisecond)
		targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)

		if RunCliCmd("configSet", "TargetWindow=Windows") == false {
			t.Errorf("can't set correct value")
		}
	} else {
		RunCliCmd("exec", "touch test.txt")
		time.Sleep(time.Duration(1000) * time.Millisecond)
		if Exists("test.txt") == false {
			t.Errorf("os command not executed correctly")
		}
		RunCliCmd("exec", "rm test.txt")
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}

	RunCliCmd("capture", "test2")
	time.Sleep(time.Duration(2000) * time.Millisecond)

	if runtime.GOOS == "windows" && Exists("test2.png") == false {
		t.Errorf("can't create capture file")
	}

	if runtime.GOOS == "linux" && Exists("test2.txt") == false {
		t.Errorf("fail to capture tty or you wrong to seted vcs of value")
	}

	if runtime.GOOS == "windows" {
		Execmd("del test2.png")
	} else {
		Execmd("rm test2.txt")
	}
	time.Sleep(time.Duration(1000) * time.Millisecond)

	if runtime.GOOS == "linux" {
		return
	}

	RunCliCmd("AnimetionGif", "test2")

	time.Sleep(time.Duration(1000) * time.Millisecond)

	RunCliCmd("AnimetionGif", "")

	time.Sleep(time.Duration(3000) * time.Millisecond)

	if Exists("test2.gif") == false {
		t.Errorf("can't create animate capture file")
	}

	Execmd("del test2.gif")

	if RunCliCmd("", "\\n") == false {
		t.Errorf("target window not found")
	}
}

func TestLiveRecord(t *testing.T) {
	if runtime.GOOS == "linux" {
		return
	}

	cliHwnd = winctl.GetWindow("GetForegroundWindow", Debug)

	cliConfig.LiveExitAsciiCode = 32
	cliConfig.Record = true
	Cli = true

	RunCliCmd("exec", "winver")
	time.Sleep(time.Duration(1000) * time.Millisecond)
	targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)

	go func() {
		time.Sleep(time.Duration(2000) * time.Millisecond)
		StringDo("\\n")
		time.Sleep(time.Duration(1000) * time.Millisecond)
		targetHwnd = cliHwnd
		StringDo(" ")
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}()

	LiveRecord()

	if DisplayHistory() == false {
		t.Errorf("LiveRecord not work correctly")
	}
}
