
# OpsApi

 **OpsApi** is operation automation tool by *CLI* and *REST API*. not need environment set up, one binary solution! for Windows and Linux.

![demo](https://github.com/yasutakatou/opsapi/blob/pic/demo.gif)

## solution

 when you must boring web task(ex meny registration made of checkbox!), or input for monthly bussiness imfomation.<br> you should many input by human operation. and your team mate same. usual solution are

 - install RPA. and make flow.
 - python, ansible, such program language. make operation into code.
 - implemented OS Scripts.

and then, you distribute your way to teammate.

 in my opinion, there are must prepare enviroment, and flexible customize difficulty. moreover must login your OS.<br> 
and Above all, keyboard input can't emulate,and capture operation result easily (you need it, you implement by append code).<br>
*Is it easy way?*<br>
do you want operation share casually? or your absolute trust teammate controlled instead of you?
OpsApi solv it, easily.

## features

 - keyboard input emulation (same to hardware input)
 - automatic capture to picture and can create animetion gif
 - implemented by golang, one binary file.
 - support HTTPS
 - support multi os (Windows or Linux)

## install

```
git clone https://github.com/yasutakatou/opsapi
cd opsapi
go build .
```

or download binary from [release page](https://github.com/yasutakatou/opsapi/releases).<br>
save binary file, copy to entryed execute path directory.


## uninstall

delete that binary. del or rm command. (it's simple!)

## usecase

**important!**<br> 
 in advance, you must decide operations to what do.<br>
 (send text to your application. emulate your keystroke on your web browser. such more)<br>
if you don't know, cli mode help you. you can test operation  on cli. or liveRecord mode make your operation into code in auto.<br>
cli mode can use like linux shell. tab interpolation, upper key is history one before, etc. <br>

![cli](https://github.com/yasutakatou/opsapi/blob/pic/cli_mode.gif)

this tool run example in local. token is test. execute command on our OS for get file lists.<br>

```
$ opsapi -token=test 
```

Linux example:
```
curl -k -H "Content-type: application/json" -X POST http://127.0.0.1:8080/ -d '{"token":"test","command":"exec","params":"ls -la"}'
```

Windows example:
```
curl -k -H "Content-type: application/json" -X POST http://127.0.0.1:8080/ -d "{\"token\":\"test\",\"command\":\"exec\",\"params\":\"dir\"}"
```

if tool  run in another server(192.168.0.1) with https. token is customized and more changed port number.<br>

```
# opsapi -https -port=18080 -token=custompassword
```

Linux example:
```
curl -k -H "Content-type: application/json" -X POST https://192.168.0.1:18080/ -d '{"token":"custompassword","command":"exec","params":"ls -la"}'
```

Windows example:
```
curl -k -H "Content-type: application/json" -X POST https://192.168.0.1:18080/ -d "{\"token\":\"custompassword\",\"command\":\"exec\",\"params\":\"dir\"}"
```

## boot parameters

```
opsapi -h : display help 
```

|option name|default value|detail|
|:---|:---|:---|
|https|false|https mode (true is enable)|
|debug|false|debug mode (true is enable)|
|token|(random)|authentication token (if this value is null, is set random)|
|port|8080|port number|
|cert|./localhost.pem|ssl_certificate file path (if you don't use https, haven't to use this option)|
|key|./localhost-key.pem|ssl_certificate_key file path (if you don't use https, haven't to use this option)|
|import|(empty)|import your operation history (must formated tsv)|
|cli|false|cli mode for recording operation (true is enable)|
|vcs|/dev/vcs1|set target vcs(Linux only. use to teminal capture)|
	
*note: if you want to use https, you prepare cert file beforehand. (use mkcert and such more)*

## functions

functions is almost supported Windows and Linux. but, "liveRecord" "AnimetionGIF" and "titles" are only do Windows.<br>

### "configGet":
 view current config. argument isn't need.

### "configSet":
 config value change. you can set value one by one. format is left(option name)=right(value).
 each value detail are ["config" reference](https://github.com/yasutakatou/opsapi#config).

### "ops":
 keyboard emulation. argument is input strings by keyboard.<br>
 
*note: in what Windows, keyboard emulate on target window. in what Linux,  keyboard emulate on terminal console (/dev/ttyX such). therefore, must permted permisson to /dev/ttyX.*

example) [ops test] (keyboard emulaton to your terminal. input "test")<br>
*note: when running cli mode. and when your inputs aren't functions. your inputs are treat to "ops".*<br>

 special define.<br>

|strings|detail|
|:---|:---|
|ctrl+|ctrl key push at same time. example [ops ctrl+s] = ctrl&s|
|alt+|alt key push at same time. example [ops alt+a] = ctrl&a|
|\\\\n|enter, newline|
|\\\\\ |backslash|
|\\\\b|backspace|
|\\\\t|tab|
|\\\\"|double quote|

and can use **integer among LiveRawcodeChar value to ascii code emulate**.
example [\\\\\~38\~; \\\\\~38\~] <br>
in case of LiveRawcodeChar value is "~"(default).input are double "↑". <br>

*note: by what "SeparateChar" define, strings can input Continuously.*<br>
example [hoge; \\n ; wait10; fuga] -> input are "hoge" , newline , (10 second waits), "fuga"<br>
waitXX is embedded sleep function. program waits xx integer second. 

### "exec":
 execute os command. argument is os command.

### "capture":
 capture from target. argument is os file name. but, must exclude extension. <br>
if omit argument, filename is current year-month-days-hour-minits. example: now  2020/03/01 22:08 -> 2020-03-01-22-08<br>
*note: in case of Windows capturing is picture. (PNG format) in case of Linux capturing is text. (/dev/vcs device export)*

### "titles": (Windows Only)
 view running window title and handle id.<br>
example) "Google - Google Chrome : 4080c",<br>
window title: "Google - Google Chrome" handle id: "4080c"

### "AnimetionGif": (Windows Only)
 1st time: starting capture. 2nd time: capturing stop, and create file.<br>
argument is os file name. but, must exclude extension. <br>
*note: if you capture long time, your pc maybe memory depletion.*<br>
if omit argument, filename is current year-month-days-hour-minits. example: now  2020/03/01 22:08 -> 2020-03-01-22-08. 

### "clearHistory": (CLI only)
 history is cleared

### "deleteHistory": (CLI only)
 delete history one or more. argumet is single integer or use "-" range integer value.<br>
example)
```
[  1] Command:        ops Params: test
[  2] Command:        ops Params: \n
[  3] Command:        ops Params: input
```
if you want delete "\n", you input [deleteHistory 2]<br>
if you want delete "test" and "\n", you input [deleteHistory 1-2]

### "displayHistory": (CLI only)
 view current history for cli. argument isn't need. 

### "exportHistory": (CLI only)
 export current history. argument is two type. <br>
 1st: export format. 2nd: file name. (two arguments can omit.) <br>
export format are "tsv" and "shell". if you set "tsv", your history export to tab spread value format.<br>
if you set "shell", your history export to custom format. "ExportFormat" in cli option can set custom format.<br>
if omit file name argument, filename is current year-month-days-hour-minits. <br>
also export format is tsv by default. example: now  2020/03/01 22:08 -> 2020-03-01-22-08. 

example) <br>
in case of  export format is "tsv", and file name is "test.txt". you set [exportHistory tsv test.txt].

### "importHistory": (CLI only)
 import of exported history file before. argument is filename. also, file format must be "tsv".<br>
*note: when you run now history is all delete.*

### "insertHistory": (CLI only)
 insert one by one history into now historys. argument is three type. <br>
 1st: insert place 2nd: api name 3rd: api value. (three arguments can't omit every.)<br>

example)
```
[  1] Command:        ops Params: test
[  2] Command:        ops Params: \n
[  3] Command:        ops Params: input
```
if you want insert tab input before "\n", you input [insertHistory 2 \\t]

### "liveRecord": (Windows/CLI only)
 when by keyboard operating to target window, operation convert to and add history.<br>
*note: until what you press ascii code  "LiveExitAsciiCode" of cli config, program record your operation. (default is 27[Esc].)*<br>

### "runHistory": (CLI only)
 operation replay along now history.
note: when "runHistory" running, not add history.

## config

*note: if space string include, refer following.*

```
>>> configSet Shebang="## space string include example ##"
```

### "Target" (default: "Chrome". exists window title, or handle id.)
if you run "ops", keyboard emulation do to application of this value. you can set title of window, or id of handle.<br>
*note: value can get by "titles" function. also, value can set partially included. if not found value, return error.*<br>

example)
"Google - Google Chrome : 4080c",<br>
window title: "Google - Google Chrome" handle id: "4080c"<br>
if you want to set target this, you input [configSet Target=Chrome] or [configSet Target=4080c].

### "AutoCapture" (default: "false". true or false)
this option, capture target window after "ops" every. detail wrote in "capture" function above.<br>
*note: filename is current year-month-days-hour-minits fixed. example: now  2020/03/01 22:08 -> 2020-03-01-22-08. *<br>
if you want to set true this, you input [configSet AutoCapture=true].

### "CapturePath" (default "" [empty] file path.)
 can set save path, when you use capturing function.

### "SeparateChar" (default ";". single character.)
when use "ops" function, use this value strings can input Continuously.<br>
example [hoge; \\n ; wait10; fuga] <br>
input are "hoge" , newline , (10 second waits), "fuga"

### "LiveRawcodeChar" (default "~". single character.)
when use "ops" function, can input ascii code among this value.<br>
about ascii code, please refer [this site](http://shanabrian.com/web/javascript/keycode.php).<br>
example [\\\\\~38\~; \\\\\~38\~] <br>
input are double "↑".

### "ReturnWindow" (default "100". 0 < value < 10000.)
 when "AutoCapture" is true, wait this value after capture. program waits xx integer milli second. 

### "AnimationDuration" (default "250". 0 < value < 10000.)
 capturing interval. The smaller, the larger the file size. 

### "AnimationDelay" (default "50". 0 < value < 10000.)
 gif animation play delay time. The smaller, the fastest the playing.

## config for cli

### "LiveExitAsciiCode (default "27")
"liveRecord" function wait this value ascii code for recording exit.<br>
about ascii code, please refer [this site](http://shanabrian.com/web/javascript/keycode.php).

### "Shebang"  (default "" [empty])
if you use Shebang, set this value.<br>
example) Shebang=#!/bin/bash

### "ExportFormat" (default: ```curl -H \"Content-type: application/json\" -X POST http://127.0.0.1:8080/ -d \"{\\\"token\\\":\\\"%1\\\",\\\"command\\\":\\\"#COMMAND#\\\",\\\"params\\\":\\\"#PARAMS#\\\"}\"```)

when you export history by shell formatted, convert along this value.<br>

example: **Command: -> #COMMAND#**, **Params: #PARAMS#** replace.
displayHistory
```
[  1] Command:        ops Params: test
```
to exportHistory
```
curl -H "Content-type: application/json" -X POST http://127.0.0.1:8080/ -d "{\"token\":\"%1\",\"command\":\"ops\",\"params\":\"test\"}"
```

### "Record" (default true. true or false)
 if this value is true, record history every inputs. if you want to set true this, you input [configSet Record=true].

### "LoopWait" (defalut "100". 0 < value < 10000.)
 "runHistory" function wait this value every do. program waits xx integer milli second. 

## consideration

*if more request, I consider to implementation*.<br>

 - animetion gif is split save, and finally merge to one file.
 
therefore, can more record at long time. I don't know how, append gif data after once save on golang.<br>
not impossible to use external command (FFmpeg such more), but I don't like.<br>
(you must prepare external command, not one binary environment.)

 - see to real time operating screen from remote  client (such use websocket).

if you do that, I suggest to use RDP, VNC.
