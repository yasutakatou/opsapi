package winctl

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

func ListWindow(Debug bool) []string {
	return nil
}

func FocusWindow(targetHwnd,cliHwnd uintptr ,title string, Debug bool) uintptr {
	return 0
}

func GetWindow(funcName string, Debug bool) uintptr {
	return 0
}

func SetActiveWindow(hwnd HWND, Debug bool) {
}

func GetWindowRect(hwnd HWND, rect *RECTdata, Debug bool) (err error) {
	return
}