package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"./winctl"
)

func TestRandStr(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	nameTest := RandStr(8)
	for i := 0; i < 1000; i++ {
		if nameTest == RandStr(8) {
			t.Errorf("result is same user name, this function not expect to random")
		}
	}
}

func TestExecmd(t *testing.T) {
	//this code test t function Execmd and Exists
	Execmd("mkdir test")
	time.Sleep(time.Duration(500) * time.Millisecond)
	if Exists("test") == false {
		t.Errorf("os command not executed correctly")
	}
	Execmd("rmdir test")
	time.Sleep(time.Duration(500) * time.Millisecond)
	if Exists("test") == true {
		t.Errorf("os command not executed correctly")
	}
}

func TestStringDo(t *testing.T) {
	//this code test t function StringDo , SendKey and Do

	if runtime.GOOS == "linux" {
		/* must tty user loggin
		  watch --interval=1 'cat /dev/vcs1'
		*/
		StringDo("\\n")
		StringDo("touch /tmp/test")
		StringDo("\\n")
		time.Sleep(time.Duration(3000) * time.Millisecond)
		if Exists("/tmp/test") == false {
			t.Errorf("can't control tty or you loggined tty?")
		}
		Execmd("rm /tmp/test")
		return
	}

	cliHwnd = winctl.GetWindow("GetForegroundWindow", Debug)

	Execmd("winver")
	time.Sleep(time.Duration(1000) * time.Millisecond)

	targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)
	if targetHwnd == 0 {
		t.Errorf("can't focus target window")
	}

	if StringDo("\\n") == false {
		t.Errorf("can't send command to target window")
	}
}

func TestCaptureOnly(t *testing.T) {
	//this code test t function CaptureOnly , GetScreenCapture and pngSave

	if runtime.GOOS == "linux" {
		CaptureOnly("test")
		time.Sleep(time.Duration(1000) * time.Millisecond)
		if Exists("test.txt") == false {
			t.Errorf("fail to capture tty or you wrong to seted vcs of value")
		}
		Execmd("rm test.txt")
		return
	}

	cliHwnd = winctl.GetWindow("GetForegroundWindow", Debug)

	Execmd("winver")
	time.Sleep(time.Duration(1000) * time.Millisecond)

	targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)

	CaptureOnly("test")
	if Exists("test.png") == false {
		t.Errorf("can't create capture file")
	}

	StringDo("\\n")

	Execmd("del test.png")
	time.Sleep(time.Duration(1000) * time.Millisecond)
}

func TestChangeTarget(t *testing.T) {
	if runtime.GOOS == "linux" { return }

	cliHwnd = winctl.GetWindow("GetForegroundWindow", Debug)

	Execmd("winver")
	time.Sleep(time.Duration(1000) * time.Millisecond)

	targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)

	if targetHwnd == 0 {
		t.Errorf("can't focus target window")
	}

	if ChangeTarget(targetHwnd) == false {
		t.Errorf("can't change target window")
	}

	StringDo("\\n")
}

func TestOptionSetting(t *testing.T) {
	//this code test t function SetOptions and SetCliOptions

	if runtime.GOOS == "windows" {
		Execmd("winver")
		time.Sleep(time.Duration(1000) * time.Millisecond)

		targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)
	}

	fmt.Println(" - - - -  api success case - - - -  ")
	if runtime.GOOS == "windows" && OptionSetting(false, "ReturnWindow=100") == false {
		t.Errorf("ReturnWindow: can't set correct value")
	}
	if OptionSetting(false, "SeparateChar=:") == false {
		t.Errorf("SeparateChar: can't set correct value")
	}
	if runtime.GOOS == "windows" && OptionSetting(false, "Target=Windows") == false {
		t.Errorf("Target: can't set correct value")
	}
	if OptionSetting(false, "AutoCapture=false") == false {
		t.Errorf("AutoCapture: can't set correct value")
	}
	if OptionSetting(false, "CapturePath=..\\") == false {
		t.Errorf("CapturePath: can't set correct value")
	}
	if runtime.GOOS == "windows" && OptionSetting(false, "AnimationDuration=1000") == false {
		t.Errorf("AnimationDuration: can't set correct value")
	}
	if runtime.GOOS == "windows" && OptionSetting(false, "AnimationDelay=500") == false {
		t.Errorf("AnimationDelay: can't set correct value")
	}
	fmt.Println("")

	fmt.Println(" - - - -  api fail case - - - -  ")
	if OptionSetting(false, "ReturnWindow=10001") == true {
		t.Errorf("ReturnWindow: can't detect incorrect value")
	}
	if OptionSetting(false, "SeparateChar=[]") == true {
		t.Errorf("SeparateChar: can't detect incorrect value")
	}
	if runtime.GOOS == "windows" && OptionSetting(false, "Target=XXXXXXX") == true {
		t.Errorf("Target: can't detect incorrect value")
	}
	if OptionSetting(false, "AutoCapture=fail") == true {
		t.Errorf("AutoCapture: can't detect incorrect value")
	}
	if OptionSetting(false, "AnimationDuration=0") == true {
		t.Errorf("AnimationDuration: can't detect incorrect value")
	}
	if OptionSetting(false, "AnimationDelay=0") == true {
		t.Errorf("AnimationDelay: can't detect incorrect value")
	}
	fmt.Println("")

	fmt.Println(" - - - -  cli success case - - - -  ")
	if OptionSetting(true, "LoopWait=100") == false {
		t.Errorf("LoopWait: can't set correct value")
	}
	if OptionSetting(true, "LiveExitAsciiCode=32") == false {
		t.Errorf("LiveExitAsciiCode: can't set correct value")
	}
	if OptionSetting(true, "ExportFormat=curl -H \"Content-type: application/json\" -X POST http://127.0.0.1:8080/ -d \"{\\\"token\\\":\\\"%1\\\",\\\"command\\\":\\\"#COMMAND#\\\",\\\"params\\\":\\\"#PARAMS#\\\"}\"") == false {
		t.Errorf("ExportFormat: can't set correct value")
	}
	if OptionSetting(true, "Shebang=#!/bin/bash") == false {
		t.Errorf("Shebang: can't set correct value")
	}
	if OptionSetting(true, "Record=false") == false {
		t.Errorf("Record: can't set correct value")
	}
	fmt.Println("")

	fmt.Println(" - - - -  cli fail case - - - -  ")
	if OptionSetting(true, "LoopWait=10001") == true {
		t.Errorf("LoopWait: can't detect incorrect value")
	}
	if OptionSetting(true, "LiveExitAsciiCode=200") == true {
		t.Errorf("LiveExitAsciiCode: can't detect incorrect value")
	}
	if OptionSetting(true, "ExportFormat=curl -H \"Content-type: application/json\" -X POST http://127.0.0.1:8080/ -d \"{\\\"token\\\":\\\"%1\\\",\\\"command\\\":\\\"#CMD#\\\",\\\"params\\\":\\\"#PRM#\\\"}\"") == true {
		t.Errorf("ExportFormat: can't detect incorrect value")
	}
	if OptionSetting(true, "Shebang=") == true {
		t.Errorf("Shebang: can't detect incorrect value")
	}
	if OptionSetting(true, "Record=fail") == true {
		t.Errorf("Record: can't detect incorrect value")
	}
	fmt.Println("")

	if runtime.GOOS == "windows" {
		StringDo("\\n")
	}
}
