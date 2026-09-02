package bomber

import "testing"

func newPlayGestures() (*Gestures, *Game) {
	g := NewGame(75, 21)
	openGrid(g)
	a := &App{Screen: ScreenPlay, ViewW: 1440, ViewH: 810, Game: g}
	return NewGestures(a, 3840, 2160), g
}

func TestSwipeFiresBeforeFingerLifts(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(7, 1000, 1000)
	gs.Move(7, 1010, 1000)
	if g.Player.IsMoving() {
		t.Fatal("ruch przed progiem")
	}
	gs.Move(7, 1000+gs.Threshold(), 1000)
	if !g.Player.IsMoving() || g.Player.ToX != 2 {
		t.Fatalf("brak ruchu w prawo: moving=%v to=%d", g.Player.IsMoving(), g.Player.ToX)
	}
	for i := 0; i < 10; i++ {
		g.Update(.05)
	}
	if g.Player.X < 4 {
		t.Fatalf("przytrzymany palec nie prowadzi dalej: x=%d", g.Player.X)
	}
	gs.Up(7, 1000+gs.Threshold(), 1000)
	if g.hold != (Point{}) || len(g.Bombs) != 0 {
		t.Fatalf("po oderwaniu: hold=%v bomby=%d", g.hold, len(g.Bombs))
	}
}

func TestDragChangesDirection(t *testing.T) {
	gs, g := newPlayGestures()
	thr := gs.Threshold()
	gs.Down(1, 500, 500)
	gs.Move(1, 500+thr, 500)
	gs.Move(1, 500+thr, 500+thr)
	if g.hold != (Point{0, 1}) {
		t.Fatalf("kierunek nie zmienil sie na dol: %v", g.hold)
	}
}

func TestTapWithoutMovePlacesBomb(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(3, 800, 800)
	gs.Move(3, 805, 803)
	gs.Up(3, 805, 803)
	if len(g.Bombs) != 1 || g.Bombs[0].X != 1 || g.Bombs[0].Y != 1 {
		t.Fatalf("bomba nie stoi pod graczem: %+v", g.Bombs)
	}
}

func TestSecondFingerPlacesBombWhileWalking(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 500+gs.Threshold(), 500)
	gs.Down(2, 3000, 1500)
	if len(g.Bombs) != 1 {
		t.Fatalf("drugi palec nie podlozyl bomby: %d", len(g.Bombs))
	}
	gs.Up(2, 3000, 1500)
	if len(g.Bombs) != 1 || g.hold != (Point{1, 0}) {
		t.Fatalf("oderwanie drugiego palca zmienilo stan: bomby=%d hold=%v", len(g.Bombs), g.hold)
	}
	gs.Up(1, 500+gs.Threshold(), 500)
	if g.hold != (Point{}) {
		t.Fatal("pierwszy palec nie zwolnil marszu")
	}
}

func TestSecondFingerOutsidePlayIsIgnored(t *testing.T) {
	a := &App{Screen: ScreenSummary, ViewW: 1440, ViewH: 810}
	gs := NewGestures(a, 1440, 810)
	gs.Down(1, 10, 10)
	gs.Down(2, 20, 20)
	gs.Up(2, 20, 20)
	if len(a.Pulses) != 0 {
		t.Fatal("drugi palec poza gra wywolal akcje")
	}
}

func TestCancelReleasesHold(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 500+gs.Threshold(), 500)
	gs.Cancel()
	if g.hold != (Point{}) || gs.Active() {
		t.Fatal("anulowanie nie zwolnilo gestu")
	}
}

func TestTapMapsClientToLogical(t *testing.T) {
	a := &App{Screen: ScreenAttract, ViewW: 1440, ViewH: 810}
	gs := NewGestures(a, 3840, 2160)
	gs.Down(1, 3839, 2159)
	gs.Up(1, 3839, 2159)
	if a.Screen != ScreenNick || len(a.Pulses) != 1 || a.Pulses[0].X > 1440 || a.Pulses[0].Y > 810 {
		t.Fatalf("mapowanie: screen=%d pulses=%+v", a.Screen, a.Pulses)
	}
}

func TestHoldExpiresWithoutPointerEvents(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 500+gs.Threshold(), 500)
	gs.Tick(HoldTimeout - .5)
	if g.hold == (Point{}) {
		t.Fatal("marsz wygasl za wczesnie")
	}
	gs.Tick(.6)
	if g.hold != (Point{}) || !gs.Active() {
		t.Fatalf("po timeout: hold=%v active=%v", g.hold, gs.Active())
	}
	gs.Up(1, 500+gs.Threshold(), 500)
	if len(g.Bombs) != 0 {
		t.Fatal("oderwanie po timeout podlozylo bombe")
	}
}
