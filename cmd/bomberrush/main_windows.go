//go:build windows

package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"bomberrush/internal/bomber"
)

const (
	csVRedraw               = 0x0001
	csHRedraw               = 0x0002
	csOwnDC                 = 0x0020
	wsPopup                 = 0x80000000
	wsVisible               = 0x10000000
	wsOverlappedWindow      = 0x00CF0000
	wsExTopmost             = 0x00000008
	swShow                  = 5
	swRestore               = 9
	pmRemove                = 0x0001
	wmDestroy               = 0x0002
	wmSize                  = 0x0005
	wmActivate              = 0x0006
	wmKillFocus             = 0x0008
	wmCaptureChanged        = 0x0215
	wmPointerLeave          = 0x024A
	wmClose                 = 0x0010
	wmPaint                 = 0x000F
	wmEraseBkgnd            = 0x0014
	wmSetCursor             = 0x0020
	wmSysCommand            = 0x0112
	wmMouseMove             = 0x0200
	wmLButtonDown           = 0x0201
	wmLButtonUp             = 0x0202
	wmPointerUpdate         = 0x0245
	wmPointerDown           = 0x0246
	wmPointerUp             = 0x0247
	wmPointerCaptureChanged = 0x024C
	scScreenSave            = 0xF140
	scMonitorPower          = 0xF170
	smCXScreen              = 0
	smCYScreen              = 1
	dibRGBColors            = 0
	srcCopy                 = 0x00CC0020
	colorOnColor            = 3
	ptTouch                 = 0x00000002
	esSystemRequired        = 0x00000001
	esDisplayRequired       = 0x00000002
	esContinuous            = 0x80000000
	mbIconError             = 0x00000010
	hwndTopmost             = ^uintptr(0)
	swpShowWindow           = 0x0040
	monitorDefaultToPrimary = 0x00000002
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	winmm    = syscall.NewLazyDLL("winmm.dll")

	procRegisterClassExW              = user32.NewProc("RegisterClassExW")
	procCreateWindowExW               = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procDestroyWindow                 = user32.NewProc("DestroyWindow")
	procShowWindow                    = user32.NewProc("ShowWindow")
	procUpdateWindow                  = user32.NewProc("UpdateWindow")
	procPeekMessageW                  = user32.NewProc("PeekMessageW")
	procTranslateMessage              = user32.NewProc("TranslateMessage")
	procDispatchMessageW              = user32.NewProc("DispatchMessageW")
	procPostQuitMessage               = user32.NewProc("PostQuitMessage")
	procGetDC                         = user32.NewProc("GetDC")
	procReleaseDC                     = user32.NewProc("ReleaseDC")
	procGetClientRect                 = user32.NewProc("GetClientRect")
	procGetSystemMetrics              = user32.NewProc("GetSystemMetrics")
	procMonitorFromPoint              = user32.NewProc("MonitorFromPoint")
	procGetMonitorInfoW               = user32.NewProc("GetMonitorInfoW")
	procSetWindowPos                  = user32.NewProc("SetWindowPos")
	procSetForegroundWindow           = user32.NewProc("SetForegroundWindow")
	procSetCapture                    = user32.NewProc("SetCapture")
	procReleaseCapture                = user32.NewProc("ReleaseCapture")
	procScreenToClient                = user32.NewProc("ScreenToClient")
	procShowCursor                    = user32.NewProc("ShowCursor")
	procMessageBoxW                   = user32.NewProc("MessageBoxW")
	procMessageBeep                   = user32.NewProc("MessageBeep")
	procRegisterPointerInputTarget    = user32.NewProc("RegisterPointerInputTarget")
	procSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	procSetProcessDPIAware            = user32.NewProc("SetProcessDPIAware")
	procStretchDIBits                 = gdi32.NewProc("StretchDIBits")
	procSetStretchBltMode             = gdi32.NewProc("SetStretchBltMode")
	procGetModuleHandleW              = kernel32.NewProc("GetModuleHandleW")
	procLoadIconW                     = user32.NewProc("LoadIconW")
	procSetThreadExecutionState       = kernel32.NewProc("SetThreadExecutionState")
	procTimeBeginPeriod               = winmm.NewProc("timeBeginPeriod")
	procTimeEndPeriod                 = winmm.NewProc("timeEndPeriod")
)

type point struct {
	X int32
	Y int32
}

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type monitorInfo struct {
	Size    uint32
	Monitor rect
	Work    rect
	Flags   uint32
}

type msg struct {
	HWnd     uintptr
	Message  uint32
	_        uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       point
	LPrivate uint32
}

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}

type runner struct {
	hwnd          uintptr
	hdc           uintptr
	app           *bomber.App
	renderer      *bomber.Renderer
	clientW       int
	clientH       int
	logicalW      int
	logicalH      int
	running       bool
	gestures      *bomber.Gestures
	mouseDown     bool
	lastPointerAt time.Time
}

var activeRunner *runner

func lowWord(v uintptr) uint16  { return uint16(v & 0xffff) }
func highWord(v uintptr) uint16 { return uint16((v >> 16) & 0xffff) }
func signedWord(v uint16) int   { return int(int16(v)) }

func windowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	r := activeRunner
	switch message {
	case wmEraseBkgnd:
		return 1
	case wmSetCursor:
		return 1
	case wmSysCommand:
		command := wParam & 0xfff0
		if command == scScreenSave || command == scMonitorPower {
			return 0
		}
	case wmSize:
		if r != nil {
			r.clientW = int(lowWord(lParam))
			r.clientH = int(highWord(lParam))
		}
		return 0
	case wmPointerDown:
		if r != nil {
			x, y := r.pointerClientPosition(lParam)
			r.lastPointerAt = time.Now()
			r.gestures.Down(int(lowWord(wParam)), x, y)
		}
		return 0
	case wmPointerUpdate:
		if r != nil {
			x, y := r.pointerClientPosition(lParam)
			r.lastPointerAt = time.Now()
			r.gestures.Move(int(lowWord(wParam)), x, y)
		}
		return 0
	case wmPointerUp:
		if r != nil {
			x, y := r.pointerClientPosition(lParam)
			r.lastPointerAt = time.Now()
			r.gestures.Up(int(lowWord(wParam)), x, y)
		}
		return 0
	case wmPointerCaptureChanged, wmPointerLeave, wmKillFocus, wmCaptureChanged:
		if r != nil {
			r.mouseDown = false
			r.gestures.Cancel()
		}
		if message == wmKillFocus {
			return 0
		}
	case wmActivate:
		if r != nil && lowWord(wParam) == 0 {
			r.mouseDown = false
			r.gestures.Cancel()
		}
	case wmLButtonDown:
		if r != nil && !r.gestures.Active() && time.Since(r.lastPointerAt) > 350*time.Millisecond {
			r.mouseDown = true
			procSetCapture.Call(hwnd)
			r.gestures.Down(bomber.MousePointerID, signedWord(lowWord(lParam)), signedWord(highWord(lParam)))
		}
		return 0
	case wmMouseMove:
		if r != nil && r.mouseDown {
			r.gestures.Move(bomber.MousePointerID, signedWord(lowWord(lParam)), signedWord(highWord(lParam)))
		}
		return 0
	case wmLButtonUp:
		if r != nil && r.mouseDown {
			r.mouseDown = false
			procReleaseCapture.Call()
			r.gestures.Up(bomber.MousePointerID, signedWord(lowWord(lParam)), signedWord(highWord(lParam)))
		}
		return 0
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		if r != nil {
			r.running = false
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return ret
}

func (r *runner) pointerClientPosition(lParam uintptr) (int, int) {
	p := point{X: int32(signedWord(lowWord(lParam))), Y: int32(signedWord(highWord(lParam)))}
	procScreenToClient.Call(r.hwnd, uintptr(unsafe.Pointer(&p)))
	return int(p.X), int(p.Y)
}

func (r *runner) refreshViewport() {
	if r.clientW <= 0 || r.clientH <= 0 {
		var cr rect
		procGetClientRect.Call(r.hwnd, uintptr(unsafe.Pointer(&cr)))
		r.clientW = int(cr.Right - cr.Left)
		r.clientH = int(cr.Bottom - cr.Top)
	}
	lw, lh := bomber.LogicalSize(r.clientW, r.clientH)
	if lw != r.logicalW || lh != r.logicalH || r.renderer == nil {
		r.logicalW, r.logicalH = lw, lh
		r.renderer = bomber.NewRenderer(lw, lh)
		r.app.SetViewport(lw, lh)
	}
	r.gestures.SetClient(r.clientW, r.clientH)
}

func (r *runner) draw() {
	r.refreshViewport()
	surface := r.renderer.Render(r.app)
	if len(surface.Pix) == 0 || r.clientW <= 0 || r.clientH <= 0 {
		return
	}
	bmi := bitmapInfo{Header: bitmapInfoHeader{
		Size:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		Width:       int32(surface.W),
		Height:      -int32(surface.H),
		Planes:      1,
		BitCount:    32,
		Compression: 0,
	}}
	procSetStretchBltMode.Call(r.hdc, colorOnColor)
	procStretchDIBits.Call(
		r.hdc,
		0, 0, uintptr(r.clientW), uintptr(r.clientH),
		0, 0, uintptr(surface.W), uintptr(surface.H),
		uintptr(unsafe.Pointer(&surface.Pix[0])),
		uintptr(unsafe.Pointer(&bmi)),
		dibRGBColors,
		srcCopy,
	)
	runtime.KeepAlive(surface)
}

func main() {
	runtime.LockOSThread()
	windowed := flag.Bool("windowed", false, "uruchom w oknie diagnostycznym")
	width := flag.Int("width", 1600, "szerokosc okna diagnostycznego")
	height := flag.Int("height", 900, "wysokosc okna diagnostycznego")
	rootFlag := flag.String("root", "", "katalog aplikacji z assets i data")
	flag.Parse()

	root := *rootFlag
	if root == "" {
		exe, err := os.Executable()
		if err != nil {
			fatalError("Nie mozna ustalic katalogu aplikacji", err)
		}
		root = filepath.Dir(exe)
	}
	root, _ = filepath.Abs(root)

	if procSetProcessDpiAwarenessContext.Find() == nil {
		procSetProcessDpiAwarenessContext.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4
	} else {
		procSetProcessDPIAware.Call()
	}
	procSetThreadExecutionState.Call(esContinuous | esDisplayRequired | esSystemRequired)

	instance, _, _ := procGetModuleHandleW.Call(0)
	appIcon, _, _ := procLoadIconW.Call(instance, 1)
	className, _ := syscall.UTF16PtrFromString("NanoVoBomberRushWindow")
	title, _ := syscall.UTF16PtrFromString("BOMBER RUSH")
	wc := wndClassEx{
		Size:      uint32(unsafe.Sizeof(wndClassEx{})),
		Style:     csHRedraw | csVRedraw | csOwnDC,
		WndProc:   syscall.NewCallback(windowProc),
		Instance:  instance,
		Icon:      appIcon,
		ClassName: className,
		IconSm:    appIcon,
	}
	if atom, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); atom == 0 {
		fatalError("Nie mozna zarejestrowac okna", err)
	}

	style := uintptr(wsPopup | wsVisible)
	exStyle := uintptr(wsExTopmost)
	x, y, winW, winH := primaryMonitorBounds()
	if *windowed {
		style = wsOverlappedWindow | wsVisible
		exStyle = 0
		winW, winH = maxInt(640, *width), maxInt(640, *height)
		x, y = 80, 80
	}

	hwnd, _, createErr := procCreateWindowExW.Call(
		exStyle,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		style,
		uintptr(x), uintptr(y), uintptr(winW), uintptr(winH),
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		fatalError("Nie mozna utworzyc okna gry", createErr)
	}

	clientW, clientH := winW, winH
	var clientRect rect
	if ok, _, _ := procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&clientRect))); ok != 0 {
		if w := int(clientRect.Right - clientRect.Left); w > 0 {
			clientW = w
		}
		if h := int(clientRect.Bottom - clientRect.Top); h > 0 {
			clientH = h
		}
	}
	lw, lh := bomber.LogicalSize(clientW, clientH)
	app, err := bomber.NewApp(root, lw, lh)
	if err != nil {
		fatalError("Nie mozna wczytac assets/branding.json", err)
	}
	r := &runner{hwnd: hwnd, app: app, logicalW: lw, logicalH: lh, renderer: bomber.NewRenderer(lw, lh), clientW: clientW, clientH: clientH, running: true}
	r.gestures = bomber.NewGestures(app, clientW, clientH)
	activeRunner = r
	if procTimeBeginPeriod.Find() == nil {
		procTimeBeginPeriod.Call(1)
		defer procTimeEndPeriod.Call(1)
	}
	r.hdc, _, _ = procGetDC.Call(hwnd)
	if r.hdc == 0 {
		fatalError("Nie mozna utworzyc powierzchni obrazu", syscall.EINVAL)
	}
	defer procReleaseDC.Call(hwnd, r.hdc)
	defer procSetThreadExecutionState.Call(esContinuous)

	if procRegisterPointerInputTarget.Find() == nil {
		procRegisterPointerInputTarget.Call(hwnd, ptTouch)
	}
	procShowWindow.Call(hwnd, swShow)
	procSetForegroundWindow.Call(hwnd)
	procUpdateWindow.Call(hwnd)
	if !*windowed {
		procSetWindowPos.Call(hwnd, hwndTopmost, uintptr(x), uintptr(y), uintptr(winW), uintptr(winH), swpShowWindow)
		procShowWindow.Call(hwnd, swRestore)
		procSetForegroundWindow.Call(hwnd)
		procUpdateWindow.Call(hwnd)
		for {
			v, _, _ := procShowCursor.Call(0)
			if int32(v) < 0 {
				break
			}
		}
	}

	last := time.Now()
	frame := time.Second / 60
	for r.running {
		loopStarted := time.Now()
		var m msg
		for {
			has, _, _ := procPeekMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
			if has == 0 {
				break
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
		}
		now := time.Now()
		dt := now.Sub(last).Seconds()
		last = now
		r.gestures.Tick(dt)
		r.app.Update(dt)
		r.playSounds()
		r.draw()
		if remaining := frame - time.Since(loopStarted); remaining > 0 {
			time.Sleep(remaining)
		}
	}
}

func (r *runner) playSounds() {
	for _, cue := range r.app.DrainSounds() {
		kind := uintptr(0x00000000)
		switch cue {
		case bomber.SoundBombBlocked, bomber.SoundDamage:
			kind = 0x00000010
		case bomber.SoundPickup, bomber.SoundRoundEnd:
			kind = 0x00000040
		case bomber.SoundTimeWarning:
			kind = 0x00000030
		}
		procMessageBeep.Call(kind)
	}
}

func procGetSystemMetricsCall(index uintptr) uintptr {
	v, _, _ := procGetSystemMetrics.Call(index)
	return v
}

func primaryMonitorBounds() (int, int, int, int) {
	p := point{}
	h, _, _ := procMonitorFromPoint.Call(uintptr(unsafe.Pointer(&p)), monitorDefaultToPrimary)
	if h != 0 {
		info := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
		if ok, _, _ := procGetMonitorInfoW.Call(h, uintptr(unsafe.Pointer(&info))); ok != 0 {
			w := int(info.Monitor.Right - info.Monitor.Left)
			h := int(info.Monitor.Bottom - info.Monitor.Top)
			if w > 0 && h > 0 {
				return int(info.Monitor.Left), int(info.Monitor.Top), w, h
			}
		}
	}
	return 0, 0, int(procGetSystemMetricsCall(smCXScreen)), int(procGetSystemMetricsCall(smCYScreen))
}

func fatalError(message string, err error) {
	detail := message
	if err != nil {
		detail += ": " + err.Error()
	}
	if exe, e := os.Executable(); e == nil {
		_ = os.WriteFile(filepath.Join(filepath.Dir(exe), "bomber-rush-error.log"), []byte(time.Now().Format(time.RFC3339)+" "+detail+"\r\n"), 0o644)
	}
	text, _ := syscall.UTF16PtrFromString(detail)
	title, _ := syscall.UTF16PtrFromString("BOMBER RUSH - blad")
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), mbIconError)
	os.Exit(1)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
