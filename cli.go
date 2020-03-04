package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yasutakatou/ishell"
)

func CliHandler(c *ishell.Context) {
	params := ""
	if len(c.Args) > 0 {
		params = c.Args[0]
		for i := 1; i < len(c.Args); i++ {
			if c.Cmd.Name == "insertHistory" || c.Cmd.Name == "exportHistory" || c.Cmd.Name == "exec" {
				params += " "
			} else {
				params += config.SeparateChar
			}
			params += c.Args[i]
		}
	}

	if Debug == true {
		fmt.Println("Command:" + c.Cmd.Name + " Params: " + params)
	}

	RunCliCmd(c.Cmd.Name, params)
}

func AnimetionSwitch(params string) string {
	if AnimetionGif == false {
		AnimetionGif = true
		return CreateAnimationGif(params)
	}
	stopCall <- true
	AnimetionGif = false
	return ""
}

func RunCliCmd(command, params string) bool {
	switch command {
	case "AnimetionGif":
		fmt.Println(AnimetionSwitch(params))
	case "insertHistory":
		return InsertHistory(params)
	case "runHistory":
		RunHistory()
	case "importHistory":
		return ImportHistory(params)
	case "exportHistory":
		ExportHistory(params)
	case "liveRecord":
		LiveRecord()
	case "clearHistory":
		History = nil
	case "deleteHistory":
		DeleteHistory(params)
	case "displayHistory":
		DisplayHistory()
	case "configGet":
		fmt.Println(string(ConfigToByte()))
	case "configSet":
		OptionSetting(params)
	case "exec":
		fmt.Println(Execmd(params))
	case "capture":
		fmt.Println(CaptureOnly(params))
	case "titles":
		fmt.Println(string(ListToByte(false)))
	default:
		return StringDo(params)
	}
	return true
}

func RunHistory() {
	recordFlag := config.Record
	config.Record = false

	for i := 0; i < len(History); i++ {
		fmt.Printf("Run! [%3d] Command: %10s Params: %s\n", i+1, History[i].Command, History[i].Params)
		switch History[i].Command {
		case "exec":
			fmt.Println(Execmd(History[i].Params))
		case "capture":
			fmt.Println(CaptureOnly(History[i].Params))
		case "configSet":
			SetOptions(History[i].Params)
		case "ops":
			StringDo(History[i].Params)
		}
		time.Sleep(time.Duration(config.LoopWait) * time.Millisecond)
	}
	config.Record = recordFlag
}

func InsertHistory(ranges string) bool {
	if strings.Index(ranges, " ") != -1 {
		params := strings.Split(ranges, " ")
		if len(params) == 3 {
			cnt, err := strconv.Atoi(params[0])
			if err == nil && cnt > 0 && len(History) >= cnt {
				History = Insert(History, cnt-1, params[1], params[2])
				fmt.Println(responseData{Status: "Success", Message: ""})
				return true
			}
		}
	}
	fmt.Println(responseData{Status: "Error", Message: "you set value out of range operation historys"})
	return false
}

func DeleteHistory(ranges string) {
	if strings.Index(ranges, "-") != -1 {
		params := strings.Split(ranges, "-")
		if len(params) == 2 {
			min, err := strconv.Atoi(params[0])
			max, err := strconv.Atoi(params[1])

			if err == nil && min > 0 && len(History) >= max && min < max {
				History = Unset(History, min-1, max)
				return
			}
		}
	} else {
		cnt, err := strconv.Atoi(ranges)
		if err == nil && cnt > 0 && len(History) >= cnt {
			cnt = cnt - 1
			History = Unset(History, cnt, cnt+1)
			return
		}
	}

	fmt.Println("Error: you set value out of range operation historys")
}

func Unset(s []historyData, min, max int) []historyData {
	return append(s[:min], s[max:]...)
}

func Insert(s []historyData, cnt int, command, params string) []historyData {
	s = append(s[:cnt+1], s[cnt:]...)
	s[cnt] = historyData{Command: command, Params: params}
	return s
}

func DisplayHistory() bool {
	if len(History) == 0 {
		return false
	}
	for i := 0; i < len(History); i++ {
		fmt.Printf("[%3d] Command: %10s Params: %s\n", i+1, History[i].Command, History[i].Params)
	}
	return true
}

func ImportHistory(params string) bool {
	if len(params) == 0 {
		return false
	}

	file, err := os.Open(params)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	History = nil
	s := bufio.NewScanner(file)
	for s.Scan() {
		strs := strings.Split(s.Text(), "\t")
		if len(strs) != 2 {
			fmt.Println("Error: your tsv file broken")
			History = nil
			return false
		}
		History = append(History, historyData{Command: strs[0], Params: strs[1]})
	}
	fmt.Println("importFile: ", params)
	return true
}

func ExportHistory(params string) bool {
	tsvFormatFlag := true

	if len(History) == 0 {
		return false
	}

	if strings.Index(params, "tsv") == 0 {
		tsvFormatFlag = true
		params = strings.Replace(params, "tsv", "", 1)
		params = strings.Replace(params, " ", "", 1)
	} else if strings.Index(params, "shell") == 0 {
		tsvFormatFlag = false
		params = strings.Replace(params, "shell", "", 1)
		params = strings.Replace(params, " ", "", 1)
	}

	filename := ""
	if len(params) == 0 {
		t := time.Now()
		const layout = "2006-01-02-15-04-05"
		filename += t.Format(layout) + ".txt"
	} else {
		filename = params
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	if tsvFormatFlag == false && len(config.Shebang) > 0 {
		_, err = file.WriteString(config.Shebang + "\n\n")
		if err != nil {
			fmt.Println(err)
			return false
		}
	}

	for i := 0; i < len(History); i++ {
		strs := ""
		if tsvFormatFlag == false {
			strs = strings.Replace(config.ExportFormat, "#COMMAND#", History[i].Command, 1)
			strs = strings.Replace(strs, "#PARAMS#", History[i].Params, 1)
		} else {
			strs = History[i].Command + "\t" + History[i].Params
		}
		if Debug == true {
			fmt.Printf("[%3d]: %s\n", i+1, strs)
		}
		_, err = file.WriteString(strs + "\n")
		if err != nil {
			fmt.Println(err)
			return false
		}
	}

	fmt.Println("exportFile: ", filename)
	return true
}

func SetExportFormat(params string) bool {
	if len(params) == 0 || strings.Index(params, "#COMMAND#") == -1 || strings.Index(params, "#PARAMS#") == -1 {
		return false
	}
	config.ExportFormat = params
	fmt.Println("ExportFormat: ", config.ExportFormat)
	return true
}

func SetShebang(params string) bool {
	if len(params) == 0 {
		return false
	}
	config.Shebang = params
	fmt.Println("Shebang: ", config.Shebang)
	return true
}

func setRange(setint *int, valString string, min,max int) string {
	cnt, err := strconv.Atoi(valString)
	if cnt > min && cnt < max && err == nil {
		*setint = cnt
		return ""
	}
	return fmt.Sprintf("value set failure (usecase [%d > value=XX > %d]).",max,min)
}

func setTrueFalse(truefalse *bool, strs string) string {
	if strs == "true" {
		*truefalse = true
		return ""
	}
	
	if strs == "false" {
		*truefalse = false
		return ""
	}
	return "value set failure (usecase [value=true/false])"
}

func SetCliOptions(options string) string {
	params := strings.Split(options, "=")

	if len(params) < 2 || len(options) == 0 {
		return "error"
	}

	switch params[0] {
	case "LoopWait":
		return setRange(&config.LoopWait,params[1],0,10000)
	case "LiveExitAsciiCode":
		return setRange(&config.LiveExitAsciiCode,params[1],0,127)
	case "ExportFormat":
		if SetExportFormat(params[1]) == false {
			return "you set value is empty, or invalid value {include #COMMAND# and #PARAMS# ?}"
		}
	case "Shebang":
		if SetShebang(params[1]) == false {
			return "you set value is empty, or invalid value"
		}
	case "Record":
		return setTrueFalse(&config.Record, params[1])
	default:
		return "error"
	}
	return ""
}
