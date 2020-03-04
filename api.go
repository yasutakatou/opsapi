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

	if checkLinuxFunction(p.Command) == true {
		w.Write(JsonResponseToByte("Error", "Error: ("+p.Command+") command can't use on linux"))
		return
	}

	switch p.Command {
	case "AnimetionGif":
		w.Write([]byte(AnimetionSwitch(p.Params)))
	case "configGet":
		w.Write(ConfigToByte())
	case "configSet":
		ret := SetOptions(p.Params)
		if ret != "" {
			w.Write(JsonResponseToByte("Error", ret))
			return
		}
		w.Write(ConfigToByte())
	case "ops":
		StringDo(p.Params)
		w.Write(JsonResponseToByte("Success", ""))
	case "exec":
		w.Write(JsonResponseToByte("Success", string(Execmd(p.Params))))
	case "capture":
		w.Write(JsonResponseToByte("Success", CaptureOnly(p.Params)))
	case "titles":
		w.Write(ListToByte(true))
	default:
		w.Write(JsonResponseToByte("Error", "command of you called isn't implement"))
	}
}

func checkLinuxFunction(params string) bool {
	if runtime.GOOS == "linux" {
		switch params {
		case "AnimationDelay":
			return true
		case "AnimationDuration":
			return true
		case "ReturnWindow":
			return true		
		case "Target":
			return true
		case "AnimetionGif":
			return true
		case "titles":
			return true
		}
	}
	return false
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

func ConfigToByte() []byte {
	var outputJson []byte
	var err error

	if runtime.GOOS == "windows" {
		if Cli == true {
			outputJson, err = json.Marshal(&configData{Target: config.Target, AutoCapture: config.AutoCapture, CapturePath: config.CapturePath, SeparateChar: config.SeparateChar, ReturnWindow: config.ReturnWindow, AnimationDuration: config.AnimationDuration, AnimationDelay: config.AnimationDelay, LiveExitAsciiCode: config.LiveExitAsciiCode, Shebang: config.Shebang, ExportFormat: config.ExportFormat, Record: config.Record, LoopWait: config.LoopWait})
		} else {
			outputJson, err = json.Marshal(&configData{Target: config.Target, AutoCapture: config.AutoCapture, CapturePath: config.CapturePath, SeparateChar: config.SeparateChar, ReturnWindow: config.ReturnWindow, AnimationDuration: config.AnimationDuration, AnimationDelay: config.AnimationDelay})
		}
	} else {
		if Cli == true {
			outputJson, err = json.Marshal(&configData{Target: config.Target, AutoCapture: config.AutoCapture, CapturePath: config.CapturePath, SeparateChar: config.SeparateChar, LiveExitAsciiCode: config.LiveExitAsciiCode, Shebang: config.Shebang, ExportFormat: config.ExportFormat, Record: config.Record, LoopWait: config.LoopWait})
		} else {
			outputJson, err = json.Marshal(&configData{Target: config.Target, AutoCapture: config.AutoCapture, CapturePath: config.CapturePath, SeparateChar: config.SeparateChar})
		}
	}

	if err != nil {
		return []byte(fmt.Sprintf("%s", err))
	}

	return []byte(outputJson)
}

func catUsecase() string {
	if runtime.GOOS == "windows" {
		if Cli == true {
			return "usecase: CLI(ExportFormat,Shebang,Record)=(strings,strings,boolean). API(ReturnWindow,SeparateChar,Target,AutoCapture,CapturePath,AnimationDuration,AnimationDelay)=(int,char,strings,boolean,strings,int,int)"
		}
		return "usecase: API(ReturnWindow,SeparateChar,Target,AutoCapture,CapturePath,AnimationDuration,AnimationDelay)=(int,char,strings,boolean,strings,int,int)"
	}

	if Cli == true {
		return "usecase: CLI(ExportFormat,Shebang,Record)=(strings,strings,boolean). API(SeparateChar,Target,AutoCapture,CapturePath)=(char,strings,boolean,strings)"

	}
	return "usecase: API(SeparateChar,Target,AutoCapture,CapturePath)=(char,strings,boolean,strings)"
}

func SetOptions(options string) string {
	params := strings.Split(options, "=")

	if len(params) < 2 || len(options) == 0 { return catUsecase() }

	if checkLinuxFunction(params[0]) == true {
		return "error"
	}

	switch params[0] {
	case "AnimationDelay":
		return setRange(&config.AnimationDelay,params[1],0,10000)
	case "AnimationDuration":
		return setRange(&config.AnimationDuration,params[1],0,10000)
	case "ReturnWindow":
		return setRange(&config.ReturnWindow,params[1],0,10000)
	case "SeparateChar":
		if len(params[1]) != 1 {
			return "SeparateChar set failure (usecase [SeparateChar=X {single char}])"
		}
		config.SeparateChar = params[1]
	case "Target":
		setWindow := winctl.FocusWindow(targetHwnd, cliHwnd, params[1], Debug)
		if len(params[1]) < 1 || setWindow == 0 {
			return "Target set failure. you seted title is not found. (usecase [Target=Chrome or XXXXXX])"
		}
		targetHwnd = setWindow
		config.Target = params[1]
	case "AutoCapture":
		return setTrueFalse(&config.AutoCapture, params[1])
	case "CapturePath":
		if len(params[1]) < 1 {
			return "CapturePath set failure (usecase [CapturePath=./pictures])"
		}
		config.CapturePath = params[1]
	default:
		return "error"
	}

	if config.Record == true {
		History = append(History, historyData{Command: "configSet", Params: options})
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
