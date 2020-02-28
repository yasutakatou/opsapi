package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"./winctl"
)

func StartAPI(https bool, port, cert, key string) {
	fmt.Println("port number: ", port)

	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/download", DownloadHandler)
	http.HandleFunc("/", ApiHandler)

	if https == true {
		err := http.ListenAndServeTLS(":"+port, cert, key, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	} else {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")

	d := json.NewDecoder(r.Body)
	p := &receiveData{}
	err := d.Decode(p)
	if err != nil {
		w.Write(JsonResponseToByte("Error", "internal server error"))
		return
	}

	if Debug == true {
		fmt.Println("Command: " + p.Command + " Parameters: " + p.Params + " Token: " + p.Token)
	}

	if p.Token != Token {
		w.Write(JsonResponseToByte("Error", "token is invalid"))
		return
	}

	switch p.Command {
	case "AnimetionGif":
		if runtime.GOOS == "linux" {
			w.Write(JsonResponseToByte("Error", "Error: ("+p.Command+") command can't use on linux"))
			return
		}
		if AnimetionGif == false {
			w.Write([]byte(CreateAnimationGif(p.Params)))
			AnimetionGif = true
		}
		stopCall <- true
		AnimetionGif = false
	case "configGet":
		w.Write(ConfigToByte(false))
	case "configSet":
		ret := SetOptions(p.Params)
		if ret != "" {
			w.Write(JsonResponseToByte("Error", ret))
			return
		}
		if Cli == true && cliConfig.Record == true {
			History = append(History, historyData{Command: p.Command, Params: p.Params})
		}
		w.Write(ConfigToByte(false))
	case "ops":
		StringDo(p.Params)
		w.Write(JsonResponseToByte("Success", ""))
	case "exec":
		if Debug == true {
			w.Write(JsonResponseToByte("Success", string(Execmd(p.Params))))
		} else {
			Execmd(p.Params)
		}
	case "capture":
		w.Write(JsonResponseToByte("Success", CaptureOnly(p.Params)))
	case "titles":
		if runtime.GOOS == "linux" {
			w.Write(JsonResponseToByte("Error", "Error: ("+p.Command+") command can't use on linux"))
			return
		}
		w.Write(ListToByte(true))
	default:
		w.Write(JsonResponseToByte("Error", "command of you called isn't implement"))
	}
}

func ListToByte(statusFlag bool) []byte {
	if runtime.GOOS == "linux" {
		return nil
	}

	lists := winctl.ListWindow(Debug)

	if statusFlag == false {
		outputJson := ""

		for i := 0; i < len(lists); i++ {
			outputJson = outputJson + "\"" + lists[i] + "\", "
		}
		return []byte(outputJson)
	}

	data := &responseListData{Status: "Success", Message: lists}
	outputJson, err := json.Marshal(data)
	if err != nil {
		return []byte(fmt.Sprintf("%s", err))
	}
	return outputJson
}

func JsonResponseToByte(status, message string) []byte {
	data := &responseData{Status: status, Message: message}
	outputJson, err := json.Marshal(data)
	if err != nil {
		return []byte(fmt.Sprintf("%s", err))
	}
	return []byte(outputJson)
}

func ConfigToByte(cli bool) []byte {
	var outputJson []byte
	var err error

	if cli == true {
		outputJson, err = json.Marshal(&cliConfigData{LiveExitAsciiCode: cliConfig.LiveExitAsciiCode, Shebang: cliConfig.Shebang, ExportFormat: cliConfig.ExportFormat, Record: cliConfig.Record, LoopWait: cliConfig.LoopWait})
	} else {
		if runtime.GOOS == "windows" {
			outputJson, err = json.Marshal(&configData{Target: config.Target, AutoCapture: config.AutoCapture, CapturePath: config.CapturePath, SeparateChar: config.SeparateChar, ReturnWindow: config.ReturnWindow, AnimationDuration: config.AnimationDuration, AnimationDelay: config.AnimationDelay})
		} else {
			outputJson, err = json.Marshal(&configData{Target: config.Target, AutoCapture: config.AutoCapture, CapturePath: config.CapturePath, SeparateChar: config.SeparateChar})
		}
	}

	if err != nil {
		return []byte(fmt.Sprintf("%s", err))
	}
	return []byte(outputJson)
}

func SetOptions(options string) string {
	params := strings.Split(options, "=")

	if len(params) < 2 || len(options) == 0 {
		if runtime.GOOS == "windows" {
			return "usecase: (ReturnWindow,SeparateChar,Target,AutoCapture,CapturePath,AnimationDuration,AnimationDelay)=(int,char,strings,boolean,strings,int,int)"
		}
		return "usecase: (SeparateChar,Target,AutoCapture,CapturePath)=(char,strings,boolean,strings)"
	}

	switch params[0] {
	case "AnimationDelay":
		if runtime.GOOS == "linux" {
			return "usecase: (SeparateChar,Target,AutoCapture,CapturePath)=(char,strings,boolean,strings)"
		}
		cnt, err := strconv.Atoi(params[1])
		if cnt > 0 && cnt < 10000 && err == nil {
			config.AnimationDelay = cnt
			return ""
		}
		return "AnimationDelay set failure (usecase [10000 > AnimationDelay=XX {Milliseconds}> 0])."
	case "AnimationDuration":
		if runtime.GOOS == "linux" {
			return "usecase: (SeparateChar,Target,AutoCapture,CapturePath)=(char,strings,boolean,strings)"
		}
		cnt, err := strconv.Atoi(params[1])
		if cnt > 0 && cnt < 10000 && err == nil {
			config.AnimationDuration = cnt
			return ""
		}
		return "AnimationDuration set failure (usecase [10000 > AnimationDuration=XX {Milliseconds}> 0])."
	case "ReturnWindow":
		if runtime.GOOS == "linux" {
			return "usecase: (SeparateChar,Target,AutoCapture,CapturePath)=(char,strings,boolean,strings)"
		}
		cnt, err := strconv.Atoi(params[1])
		if cnt > 0 && cnt < 10000 && err == nil {
			config.ReturnWindow = cnt
			return ""
		}
		return "ReturnWindow set failure (usecase [ReturnWindow=XX {1-10000 Milliseconds}])"
	case "SeparateChar":
		if len(params[1]) == 1 {
			config.SeparateChar = params[1]
			return ""
		}
		return "SeparateChar set failure (usecase [SeparateChar=X {single char}])"
	case "Target":
		if runtime.GOOS == "linux" {
			return ""
		}
		setWindow := winctl.FocusWindow(targetHwnd, cliHwnd, params[1], Debug)
		if len(params[1]) < 1 || setWindow == 0 {
			return "Target set failure. you seted title is not found. (usecase [Target=Chrome or XXXXXX])"
		}
		targetHwnd = setWindow
		config.Target = params[1]
	case "AutoCapture":
		if params[1] == "true" {
			config.AutoCapture = true
			return ""
		} else if params[1] == "false" {
			config.AutoCapture = false
			return ""
		} else {
			return "AutoCapture set failure (usecase [AutoCapture=true/false])"
		}
	case "CapturePath":
		if len(params[1]) < 1 {
			return "CapturePath set failure (usecase [CapturePath=./pictures])"
		}
		config.CapturePath = params[1]
		return ""
	default:
		if runtime.GOOS == "windows" {
			return "usecase: (ReturnWindow,SeparateChar,Target,AutoCapture,CapturePath,AnimationDuration,AnimationDelay)=(int,char,strings,boolean,strings,int,int)"
		}
		return "usecase: (SeparateChar,Target,AutoCapture,CapturePath)=(char,strings,boolean,strings)"
	}
	return ""
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")

	userToken := r.FormValue("token")
	name := r.FormValue("name")
	file, _, err := r.FormFile("file")
	defer file.Close()

	if len(name) == 0 {
		w.Write(JsonResponseToByte("Error", "filename parameter(name) are empty"))
		return
	}

	if userToken != Token || len(userToken) == 0 {
		w.Write(JsonResponseToByte("Error", "token is invalid"))
		return
	}

	if err != nil {
		w.Write(JsonResponseToByte("Error", "file upload error"))
		return
	}

	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		w.Write(JsonResponseToByte("Error", "file create error"))
	} else {
		io.Copy(f, file)
	}
	w.Write(JsonResponseToByte("Success", "file upload success"))
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")

	userToken := r.FormValue("token")
	file := r.FormValue("name")

	if len(file) == 0 {
		w.Write(JsonResponseToByte("Error", "filename parameter(name) are empty"))
		return
	}

	if userToken != Token || len(userToken) == 0 {
		w.Write(JsonResponseToByte("Error", "token is invalid"))
		return
	}

	downloadBytes, err := ioutil.ReadFile(file)
	if err != nil {
		w.Write(JsonResponseToByte("Error", "file download error"))
		return
	}

	mime := http.DetectContentType(downloadBytes)

	fileSize := len(string(downloadBytes))

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", "attachment; filename="+file+"")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Length", strconv.Itoa(fileSize))
	w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

	http.ServeContent(w, r, file, time.Now(), bytes.NewReader(downloadBytes))
}
