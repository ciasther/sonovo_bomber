package bomber

import "math"

const MousePointerID = -1

// Marsz wygasa sam, gdy sterownik przestanie słać zdarzenia; palec trzymany bez ruchu ma zapas czasu.
const HoldTimeout = 8.0

// Ekran po zmianie stanu przez chwilę nie przyjmuje dotyku, żeby jeden gest nie przeskoczył dwóch ekranów.
const screenTapLockout = .28

type gesturePointer struct {
	id     int
	moved  bool
	tapped bool
}

type Gestures struct {
	App              *App
	ClientW, ClientH int
	primary          *gesturePointer
	extras           map[int]bool
	idle             float64
}

func NewGestures(app *App, clientW, clientH int) *Gestures {
	return &Gestures{App: app, ClientW: clientW, ClientH: clientH, extras: map[int]bool{}}
}

func (t *Gestures) SetClient(w, h int) {
	if w > 0 && h > 0 {
		t.ClientW, t.ClientH = w, h
	}
}

func (t *Gestures) Active() bool { return t.primary != nil }

// Martwa strefa drążka w jednostkach logicznych; celowo mała, by lekki gest wystarczał.
func (t *Gestures) DeadZone() float64 {
	return math.Max(8, float64(min(t.App.ViewW, t.App.ViewH))/110)
}

func (t *Gestures) stickRadius() float64 {
	return math.Max(3*t.DeadZone(), float64(min(t.App.ViewW, t.App.ViewH))/16)
}

func (t *Gestures) toLogical(x, y int) (int, int) {
	w, h := t.ClientW, t.ClientH
	if w <= 0 || h <= 0 {
		return 0, 0
	}
	x = clamp(x, 0, w-1)
	y = clamp(y, 0, h-1)
	return x * t.App.ViewW / w, y * t.App.ViewH / h
}

func (t *Gestures) playMode() bool {
	return t.App.Screen == ScreenPlay && t.App.Game != nil
}

func (t *Gestures) Tick(dt float64) {
	if t.primary == nil {
		return
	}
	t.idle += dt
	if t.idle >= HoldTimeout {
		// Zatrzymaj marsz, ale zostaw drążek — ruch palca od razu go wznowi.
		t.App.Stick.DirX, t.App.Stick.DirY = 0, 0
		if t.App.Game != nil {
			t.App.Game.Release()
		}
	}
}

func (t *Gestures) Down(id, x, y int) {
	t.idle = 0
	lx, ly := t.toLogical(x, y)
	if t.primary == nil {
		p := &gesturePointer{id: id}
		t.primary = p
		if t.playMode() {
			t.App.Stick = Stick{Active: true, AnchorX: float64(lx), AnchorY: float64(ly), X: float64(lx), Y: float64(ly), Radius: t.stickRadius()}
			return
		}
		if t.App.StateAge >= screenTapLockout {
			p.tapped = true
			t.App.Tap(lx, ly)
		}
		return
	}
	if t.primary.id == id {
		return
	}
	t.extras[id] = true
	if t.playMode() {
		t.App.Tap(lx, ly)
	}
}

func (t *Gestures) Move(id, x, y int) {
	p := t.primary
	if p == nil || p.id != id {
		return
	}
	t.idle = 0
	if !t.App.Stick.Active {
		return
	}
	lx, ly := t.toLogical(x, y)
	s := &t.App.Stick
	s.X, s.Y = float64(lx), float64(ly)
	dx, dy := s.X-s.AnchorX, s.Y-s.AnchorY
	if r := math.Hypot(dx, dy); r > s.Radius {
		// Kotwica podąża za palcem, więc drążek nigdy nie ucieka poza zasięg.
		s.AnchorX += dx * (1 - s.Radius/r)
		s.AnchorY += dy * (1 - s.Radius/r)
		dx, dy = s.X-s.AnchorX, s.Y-s.AnchorY
	}
	nx, ny := t.resolveDir(dx, dy)
	if nx != 0 || ny != 0 {
		p.moved = true
	}
	if nx == s.DirX && ny == s.DirY {
		return
	}
	s.DirX, s.DirY = nx, ny
	if nx == 0 && ny == 0 {
		t.App.Release()
		return
	}
	t.App.Swipe(nx, ny)
}

// Histereza 1.25x: bieżąca oś wygrywa, dopóki druga wyraźnie jej nie przewyższy.
func (t *Gestures) resolveDir(dx, dy float64) (int, int) {
	dead := t.DeadZone()
	if dx*dx+dy*dy < dead*dead {
		return 0, 0
	}
	ax, ay := math.Abs(dx), math.Abs(dy)
	horizontal := ax >= ay
	switch {
	case t.App.Stick.DirX != 0:
		horizontal = ay <= ax*1.25
	case t.App.Stick.DirY != 0:
		horizontal = ax > ay*1.25
	}
	if horizontal {
		return signf(dx), 0
	}
	return 0, signf(dy)
}

func signf(v float64) int {
	if v < 0 {
		return -1
	}
	return 1
}

func (t *Gestures) Up(id, x, y int) {
	if t.extras[id] {
		delete(t.extras, id)
		return
	}
	p := t.primary
	if p == nil || p.id != id {
		return
	}
	if t.App.Stick.Active {
		t.Move(id, x, y)
		if !p.moved && !p.tapped {
			lx, ly := t.toLogical(x, y)
			t.App.Tap(lx, ly)
		}
	}
	t.App.Release()
	t.primary = nil
}

func (t *Gestures) Cancel() {
	t.primary = nil
	t.extras = map[int]bool{}
	t.App.Release()
}
