package bomber

const MousePointerID = -1

type gesturePointer struct {
	id               int
	anchorX, anchorY int
	swiped           bool
}

// Przesunięcie działa od razu po progu, bez czekania na oderwanie palca; drugi palec podkłada bombę.
const HoldTimeout = 3.0

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

func (t *Gestures) Threshold() int {
	return max(14, min(t.ClientW, t.ClientH)/90)
}

func (t *Gestures) Active() bool { return t.primary != nil }

func (t *Gestures) toLogical(x, y int) (int, int) {
	w, h := t.ClientW, t.ClientH
	if w <= 0 || h <= 0 {
		return 0, 0
	}
	x = clamp(x, 0, w-1)
	y = clamp(y, 0, h-1)
	return x * t.App.ViewW / w, y * t.App.ViewH / h
}

// Bez zdarzenia UP (utrata okna, sterownik) marsz wygasa sam, żeby postać nie szła w nieskończoność.
func (t *Gestures) Tick(dt float64) {
	if t.primary == nil {
		return
	}
	t.idle += dt
	if t.idle >= HoldTimeout {
		t.App.Release()
	}
}

func (t *Gestures) Down(id, x, y int) {
	t.idle = 0
	if t.primary == nil {
		t.primary = &gesturePointer{id: id, anchorX: x, anchorY: y}
		return
	}
	if t.primary.id == id {
		return
	}
	t.extras[id] = true
	if t.App.Screen == ScreenPlay {
		t.App.Tap(t.toLogical(x, y))
	}
}

func (t *Gestures) Move(id, x, y int) {
	p := t.primary
	if p == nil || p.id != id {
		return
	}
	t.idle = 0
	dx, dy := x-p.anchorX, y-p.anchorY
	thr := t.Threshold()
	if dx*dx+dy*dy < thr*thr {
		return
	}
	p.swiped = true
	p.anchorX, p.anchorY = x, y
	t.App.Swipe(dx, dy)
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
	t.Move(id, x, y)
	if !p.swiped {
		t.App.Tap(t.toLogical(x, y))
	}
	t.App.Release()
	t.primary = nil
}

func (t *Gestures) Cancel() {
	t.primary = nil
	t.extras = map[int]bool{}
	t.App.Release()
}
