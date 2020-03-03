package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"./winctl"

	"github.com/kbinani/screenshot"
	"github.com/yasutakatou/ishell"
	"github.com/yasutakatou/string2keyboard"
)

/*  - - - global variable, json interface structure  - - - */

var rs1Letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var Debug bool
var Cli bool
var AnimetionGif bool
var Token string
var targetHwnd uintptr
var cliHwnd uintptr
var stopCall = make(chan bool)

type (
	HANDLE uintptr
	HWND   HANDLE
)

type RECTdata struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type receiveData struct {
	Command string `json:"command"`
	Params  string `json:"params"`
	Token   string `json:"token"`
}

type configData struct {
	Target            string `json:"Target"`
	AutoCapture       bool   `json:"AutoCapture"`
	CapturePath       string `json:"CapturePath"`
	SeparateChar      string `json:"SeparateChar"`
	ReturnWindow      int    `json:"ReturnWindow"`
	AnimationDuration int    `json:"AnimationDuration"`
	AnimationDelay    int    `json:"AnimationDelay"`
	vcsDevice         string `json:"Target"`
}

var config = configData{}

type cliConfigData struct {
	LiveExitAsciiCode int    `json:"LiveExitAsciiCode"`
	Shebang           string `json:"Shebang"`
	ExportFormat      string `json:"ExportFormat"`
	Record            bool   `json:"Record"`
	LoopWait          int    `json:"LoopWait"`
}

var cliConfig = cliConfigData{}

type historyData struct {
	Command string `json:"Command"`
	Params  string `json:"Params"`
}

var History = []historyData{}

type responseData struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type responseListData struct {
	Status  string   `json:"status"`
	Message []string `json:"message"`
}

/*  - - - - - - - - - - - - - - - - - - - - - - - - - - -  */

func init() {
	rand.Seed(time.Now().UnixNano())

	config.Target = "Chrome"
	config.AutoCapture = false
	config.CapturePath = ""
	config.SeparateChar = ";"
	config.ReturnWindow = 1000
	config.AnimationDuration = 250
	config.AnimationDelay = 50
	config.vcsDevice = "/dev/vcs1"

	Debug = false
	Cli = false
	AnimetionGif = false

	cliConfig.LiveExitAsciiCode = 27
	cliConfig.Shebang = ""
	cliConfig.ExportFormat = "curl -H \"Content-type: application/json\" -X POST http://127.0.0.1:8080/ -d \"{\\\"token\\\":\\\"%1\\\",\\\"command\\\":\\\"#COMMAND#\\\",\\\"params\\\":\\\"#PARAMS#\\\"}\""
	cliConfig.Record = true
	cliConfig.LoopWait = 500
}

func main() {
	_import := flag.String("import", "", "[-import=import your operation history (must formated tsv)]")
	_cli := flag.Bool("cli", false, "[-cli=cli mode for recording operation (true is enable)]")
	_https := flag.Bool("https", false, "[-https=https mode (true is enable)]")
	_debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_token := flag.String("token", "", "[-token=authentication token (if this value is null, is set random)]")
	_port := flag.String("port", "8080", "[-port=port number]")
	_cert := flag.String("cert", "./localhost.pem", "[-cert=ssl_certificate file path (if you don't use https, haven't to use this option)]")
	_key := flag.String("key", "./localhost-key.pem", "[-key=ssl_certificate_key file path (if you don't use https, haven't to use this option)]")
	_vcs := flag.String("vcs", "/dev/vcs1", "[-vcs=set target vcs(linux only. use to teminal capture)]")

	flag.Parse()

	if len(string(*_import)) > 0 {
		ImportHistory(string(*_import))
	}

	if len(string(*_vcs)) > 0 && Exists(string(*_vcs)) {
		config.vcsDevice = string(*_vcs)
	}

	Cli = bool(*_cli)
	Debug = bool(*_debug)
	Token = string(*_token)

	if Token == "" {
		Token = RandStr(8)
	}

	if Debug == true {
		if runtime.GOOS == "linux" {
			fmt.Println(" - - - OS: Linux - - -")
		} else if runtime.GOOS == "windows" {
			fmt.Println(" - - - OS: Windows - - -")

			cliHwnd = winctl.GetWindow("GetForegroundWindow", Debug)
			targetHwnd = winctl.FocusWindow(targetHwnd, cliHwnd, config.Target, Debug)
		} else {
			fmt.Println("Error: not support this os.")
			os.Exit(-1)
		}
	}

	if Cli == false {
		fmt.Println("access token: ", Token)
		StartAPI(*_https, *_port, *_cert, *_key)
	}

	fmt.Println("Shebang: ", cliConfig.Shebang)
	fmt.Println("ExportFormat: ", cliConfig.ExportFormat)

	var shell = ishell.New()

	if runtime.GOOS == "windows" {
		shell.AddCmd(&ishell.Cmd{Name: "AnimetionGif",
			Help: "this option start to record target window into animated gif. when twice, stop to record",
			Func: CliHandler})

		shell.AddCmd(&ishell.Cmd{Name: "titles",
			Help: "process titles and handles are list up",
			Func: CliHandler})

		shell.AddCmd(&ishell.Cmd{Name: "liveRecord",
			Help: "this option start to live record",
			Func: CliHandler})
	}

	// History and Config section - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

	shell.AddCmd(&ishell.Cmd{Name: "capture",
		Help: "target window are captured and save",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "insertHistory",
		Help: "this option insert single your operation history",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "importHistory",
		Help: "this option import your operation history (must formated tsv)",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "runHistory",
		Help: "this option replay your historys",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "exportHistory",
		Help: "this option export your operation history to file(1st arg is file format: tsv(default) / shell. can omitted)",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "clearHistory",
		Help: "this option clear your operation history",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "deleteHistory",
		Help: "this option delete a part of your operation history",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "displayHistory",
		Help: "this option display your operation history",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "cliConfigGet",
		Help: "cli config option are display",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "cliConfigSet",
		Help: "usecase: (LoopWait,ExportFormat,shebang,record)=(int,strings,strings,boolean).",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "configGet",
		Help: "config option are display",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "configSet",
		Help: "usecase: (SeparateChar,Target,AutoCapture,CapturePath)=(char,strings,boolean,strings)",
		Func: CliHandler})

	// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

	shell.AddCmd(&ishell.Cmd{Name: "exec",
		Help: "execute command",
		Func: CliHandler})

	shell.AddCmd(&ishell.Cmd{Name: "default",
		Help: "default input is execute to emulate keyboard (in case of api server, same api is 'ops')",
		Func: CliHandler})

	// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

	shell.Run()
}

func RandStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = rs1Letters[rand.Intn(len(rs1Letters))]
	}
	return string(b)
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func OptionSetting(cli bool, params string) bool {
	if cli == true {
		ret := SetCliOptions(params)
		if ret == "" {
			fmt.Println(string(ConfigToByte(cli)))
			return true
		}
		fmt.Println(ret)
		return false
	}

	ret := SetOptions(params)
	if ret == "" {
		fmt.Println(string(ConfigToByte(cli)))
		return true
	}
	fmt.Println(ret)
	return false
}

func ChangeTarget(setHwnd uintptr) bool {
	if runtime.GOOS == "linux" {
		return false
	}

	breakCounter := 10

	for {
		if cliHwnd != winctl.GetWindow("GetForegroundWindow", Debug) {
			winctl.SetActiveWindow(winctl.HWND(cliHwnd), Debug)
			time.Sleep(time.Duration(100) * time.Millisecond)
		} else {
			break
		}
		breakCounter--
		if breakCounter == 0 {
			break
		}
	}

	breakCounter = 10

	for {
		if setHwnd != winctl.GetWindow("GetForegroundWindow", Debug) {
			winctl.SetActiveWindow(winctl.HWND(setHwnd), Debug)
			time.Sleep(time.Duration(100) * time.Millisecond)
		} else {
			break
		}
		breakCounter--
		if breakCounter == 0 {
			return false
		}
	}

	return true
}

func StringDo(doCmd string) bool {
	if len(doCmd) == 0 {
		return false
	}

	var foregroundWindow uintptr

	if runtime.GOOS == "windows" {
		foregroundWindow = winctl.GetWindow("GetForegroundWindow", Debug)
	}

	if strings.Index(doCmd, config.SeparateChar) != -1 {
		params := strings.Split(doCmd, config.SeparateChar)
		for r := 0; r < len(params); r++ {
			if Do(params[r]) == false {
				return false
			}
		}
	} else {
		if Do(doCmd) == false {
			return false
		}
	}

	if runtime.GOOS == "windows" {
		if config.AutoCapture == true {
			fmt.Println("capture: ", CaptureOnly(""))
		}

		time.Sleep(time.Duration(config.ReturnWindow) * time.Millisecond)
		ChangeTarget(foregroundWindow)
	}

	if Debug == true {
		data := &historyData{Command: "ops", Params: doCmd}
		outputJson, err := json.Marshal(data)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s", err))
		} else {
			fmt.Println(string(outputJson))
		}
	}

	if Cli == true && cliConfig.Record == true {
		History = append(History, historyData{Command: "ops", Params: doCmd})
	}
	return true
}

func SendKey(doCmd string) bool {
	if runtime.GOOS == "windows" && ChangeTarget(targetHwnd) == false {
		return false
	}

	if Debug == true {
		fmt.Printf("KeyInput: ")
	}

	cCtrl := false
	cAlt := false

	if strings.Index(doCmd, "ctrl+") != -1 {
		cCtrl = true
		doCmd = strings.Replace(doCmd, "ctrl+", "", 1)
		if Debug == true {
			fmt.Printf("ctrl+")
		}
	}

	if strings.Index(doCmd, "alt+") != -1 {
		cAlt = true
		doCmd = strings.Replace(doCmd, "alt+", "", 1)
		if Debug == true {
			fmt.Printf("alt+")
		}
	}

	string2keyboard.KeyboardWrite(doCmd, cCtrl, cAlt)
	if Debug == true {
		fmt.Println(doCmd, cCtrl, cAlt)
	}
	return true
}

func Execmd(command string) []byte {
	var out []byte
	var err error

	if len(command) == 0 {
		return []byte("Error: no parameters.")
	}

	switch runtime.GOOS {
	case "linux":
		go func() {
			out, err := exec.Command(os.Getenv("SHELL"), "-c", command).Output()
			if err != nil {
				fmt.Println(err)
				return
			}

			if config.AutoCapture == true {
				t := time.Now()
				const layout = "2006-01-02-15-04-05"
				filename := config.CapturePath + t.Format(layout) + ".txt"

				file, err := os.Create(filename)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer file.Close()

				_, err = file.WriteString(string(out))
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("capture: ", filename)
			}
		}()
	case "windows":
		go func() {
			out, err = exec.Command("cmd", "/C", command).Output()
			if err != nil {
				fmt.Println(err)
			}
			if Debug == true {
				fmt.Printf("Stdout: %s\n", string(out))
			}
		}()
	}
	if Cli == true && cliConfig.Record == true {
		History = append(History, historyData{Command: "exec", Params: command})
	}
	return out
}

func CaptureOnly(filename string) string {
	ret := ""

	if Cli == true && cliConfig.Record == true {
		History = append(History, historyData{Command: "capture", Params: filename})
	}

	if len(filename) > 0 {
		ret = config.CapturePath + filename
	} else {
		t := time.Now()
		const layout = "2006-01-02-15-04-05"
		ret = config.CapturePath + t.Format(layout)
	}

	if runtime.GOOS == "linux" {
		ret = ret + ".txt"
		if Exists(config.vcsDevice) == true {
			Execmd("cat " + config.vcsDevice + " > " + ret)
		}
	} else {
		ret = ret + ".png"
		pngSave(GetScreenCapture(), ret)
	}
	return ret
}

func Do(doCmd string) bool {
	if strings.Index(doCmd, "wait") != -1 {
		waits := strings.Replace(doCmd, "wait", "", 1)
		cnt, err := strconv.Atoi(waits)
		if err == nil && cnt > 0 {
			if Debug == true {
				fmt.Println("wait: " + waits)
			}
			time.Sleep(time.Duration(cnt) * time.Second)
		}
	} else {
		if SendKey(doCmd) == false {
			return false
		}
	}

	return true
}

func GetScreenCapture() *image.RGBA {
	if runtime.GOOS == "linux" {
		return nil
	}

	if ChangeTarget(targetHwnd) == false {
		return nil
	}

	var rect winctl.RECTdata
	winctl.GetWindowRect(winctl.HWND(targetHwnd), &rect, Debug)

	if Debug == true {
		fmt.Printf("window rect: ")
		fmt.Println(rect)
	}

	img, err := screenshot.Capture(int(rect.Left), int(rect.Top), int(rect.Right-rect.Left), int(rect.Bottom-rect.Top))
	if err != nil {
		panic(err)
	}

	return img
}

func pngSave(img *image.RGBA, filePath string) {
	if runtime.GOOS == "linux" {
		return
	}

	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}

func CreateAnimationGif(filename string) string {
	if runtime.GOOS == "linux" {
		return ""
	}

	if ChangeTarget(targetHwnd) == false {
		return ""
	}

	if len(filename) == 0 {
		t := time.Now()
		const layout = "2006-01-02-15-04-05"
		filename = t.Format(layout) + ".gif"
	} else {
		filename = filename + ".gif"
	}

	outGif := &gif.GIF{}

	var rect winctl.RECTdata
	winctl.GetWindowRect(winctl.HWND(targetHwnd), &rect, Debug)

	go func() {
		for {
			select {
			case <-stopCall:
				f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
				defer f.Close()
				gif.EncodeAll(f, outGif)
				return
			default:
				img, err := screenshot.Capture(int(rect.Left), int(rect.Top), int(rect.Right-rect.Left), int(rect.Bottom-rect.Top))
				if err != nil {
					fmt.Println(err)
				}

				pm := image.NewPaletted(img.Bounds(), palette.Plan9)
				draw.FloydSteinberg.Draw(pm, img.Bounds(), img, image.ZP)

				outGif.Image = append(outGif.Image, pm)
				outGif.Delay = append(outGif.Delay, config.AnimationDelay)
			}
			time.Sleep(time.Duration(config.AnimationDuration) * time.Millisecond)
		}
	}()
	return filename
}
