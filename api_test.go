package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"./winctl"
)

func TestListToByte(t *testing.T) {
	if runtime.GOOS == "linux" {
		return
	}

	cliHwnd = winctl.GetWindow("GetForegroundWindow", Debug)

	Execmd("winver")
	time.Sleep(time.Duration(1000) * time.Millisecond)
	targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)

	list := string(ListToByte(true))
	if strings.Index(list, "Windows") == -1 {
		t.Errorf("can't pick up running window")
	}
	StringDo(" ")
}

func TestJsonResponseToByte(t *testing.T) {
	var test responseData

	if err := json.Unmarshal(JsonResponseToByte("Test", "Test"), &test); err != nil {
		t.Errorf("can't convert correct value to json")
	}

	if test.Status != "Test" || test.Message != "Test" {
		t.Errorf("can't set correct value to json")
	}
}

func TestConfigToByte(t *testing.T) {
	var apiTest configData

	if err := json.Unmarshal(ConfigToByte(false), &apiTest); err != nil {
		t.Errorf("can't convert config in json to []byte")
	}
	if apiTest.SeparateChar != ";" || apiTest.AutoCapture != false {
		t.Errorf("can't set config in json to []byte")
	}

	var cliTest cliConfigData

	if err := json.Unmarshal(ConfigToByte(true), &cliTest); err != nil {
		t.Errorf("can't convert cliconfig in json to []byte")
	}
	if cliTest.LiveExitAsciiCode != 27 || cliTest.LoopWait != 500 {
		t.Errorf("can't set cliconfig in json to []byte")
	}
}

func TestUploadHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(UploadHandler))
	defer ts.Close()

	Token = "test"

	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	err := w.WriteField("name", "check.txt")
	if err != nil {
		t.Errorf("can't write 'name' field")
	}

	err = w.WriteField("token", "test")
	if err != nil {
		t.Errorf("can't write 'token' field")
	}

	f, err := os.Open("main.go")
	if err != nil {
		t.Errorf("should exist 'main.go'")
	}
	defer f.Close()
	fw, err := w.CreateFormFile("file", "main.go")
	if err != nil {
		t.Errorf("can't write 'file' field")
	}
	if _, err = io.Copy(fw, f); err != nil {
		t.Errorf("can't read 'main.go'")
	}
	w.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL, &buf)
	if err != nil {
		t.Errorf("can't create api request")
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	buff, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("can't call api request")
	}

	body, err := ioutil.ReadAll(buff.Body)
	if err != nil {
		t.Errorf("no reply from api")
	}

	var result responseData
	if err := json.Unmarshal(body, &result); err != nil {
		t.Errorf("api reply can't convert to json")
	}

	if result.Status == "Error" {
		t.Errorf("api reply is 'Error'")
	}

	if Filesize("main.go") != Filesize("check.txt") {
		t.Errorf("upload file isn't same size original file")
	}

	if runtime.GOOS == "windows" {
		Execmd("del check.txt")
	} else {
		Execmd("rm check.txt")
	}
	time.Sleep(time.Duration(1000) * time.Millisecond)
}

func Filesize(filename string) string {
	fileinfo, staterr := os.Stat(filename)

	if staterr != nil {
		fmt.Println(staterr)
		return "Error!"
	}

	return strconv.FormatInt(fileinfo.Size(), 10)
}

func requestAPI(endpoint, token, command, params string) responseData {
	data := receiveData{
		Token:   token,
		Command: command,
		Params:  params,
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return responseData{Status: "Error", Message: "Marshal error"}
	}

	client := &http.Client{}

	resp, err := client.Post(endpoint, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return responseData{Status: "Error", Message: "not send rest api " + endpoint}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return responseData{Status: "Error", Message: "not send rest api " + endpoint}
	}

	var result responseData
	if err := json.Unmarshal(body, &result); err != nil {
		return responseData{Status: "Error", Message: "Unmarshal error"}
	}

	return result
}

func TestApiHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(ApiHandler))
	defer ts.Close()

	var result responseData
	Token = "test"

	result = requestAPI(ts.URL, "fail", "tokenAuth", "params")
	if result.Status != "Error" {
		t.Errorf("error token is allow")
	}

	if runtime.GOOS == "linux" {
		requestAPI(ts.URL, "test", "exec", "touch test.txt")
		time.Sleep(time.Duration(1000) * time.Millisecond)
		if Exists("test.txt") == false {
			t.Errorf("os command not executed correctly")
		}

		result = requestAPI(ts.URL, "test", "capture", "test2")
		time.Sleep(time.Duration(1000) * time.Millisecond)
		if Exists("test2.txt") == false {
			t.Errorf("fail to capture tty or you wrong to seted vcs of value")
		}

		Execmd("rm test.txt test2.txt")
		time.Sleep(time.Duration(1000) * time.Millisecond)

		return
	}

	cliHwnd = winctl.GetWindow("GetForegroundWindow", Debug)

	requestAPI(ts.URL, "test", "exec", "winver")
	time.Sleep(time.Duration(1000) * time.Millisecond)

	targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, "Windows", Debug)
	if targetHwnd == 0 {
		t.Errorf("can't focus target window")
	}

	result = requestAPI(ts.URL, "test", "capture", "test")
	if Exists("test.png") == false {
		t.Errorf("can't create capture file")
	}

	Execmd("del test.png")
	time.Sleep(time.Duration(1000) * time.Millisecond)

	result = requestAPI(ts.URL, "test", "AnimetionGif", "test2")

	time.Sleep(time.Duration(1000) * time.Millisecond)

	result = requestAPI(ts.URL, "test", "AnimetionGif", "")

	time.Sleep(time.Duration(3000) * time.Millisecond)

	if Exists("test2.gif") == false {
		t.Errorf("can't create animate capture file")
	}

	Execmd("del test2.gif")

	result = requestAPI(ts.URL, "test", "ops", "\\n")
	if result.Status == "Error" {
		t.Errorf("target window not found")
	}
}

func TestDownloadHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(DownloadHandler))
	defer ts.Close()

	Token = "test"

	values := url.Values{}
	values.Add("name", "main.go")
	values.Add("token", "test")

	resp, err := http.PostForm(ts.URL, values)
	if err != nil {
		t.Errorf("can't create post form")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("no reply from api")
	}

	if Filesize("main.go") != fmt.Sprintf("%d", len(body)) {
		t.Errorf("download file isn't same size original file")
	}
}
