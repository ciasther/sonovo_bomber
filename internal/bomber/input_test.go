package bomber

import "testing"

func newPlayGestures() (*Gestures, *Game) {
	g := NewGame(75, 21)
	openGrid(g)
	a := &App{Screen: ScreenPlay, ViewW: 1440, ViewH: 810, Game: g, StateAge: 1}
	return NewGestures(a, 1440, 810), g
}

func TestLightGestureStartsWalk(t *testing.T) {
	gs, g := newPlayGestures()
	if gs.DeadZone() > 12 {
		t.Fatalf("martwa strefa za duza dla lekkiego gestu: %v", gs.DeadZone())
	}
	gs.Down(1, 500, 500)
	gs.Move(1, 504, 500)
	if g.Player.IsMoving() {
		t.Fatal("ruch w martwej strefie")
	}
	gs.Move(1, 512, 500)
	if !g.Player.IsMoving() || g.Player.ToX != 2 {
		t.Fatalf("krotki gest nie ruszyl postaci: moving=%v to=%d", g.Player.IsMoving(), g.Player.ToX)
	}
	for i := 0; i < 10; i++ {
		g.Update(.05)
	}
	if g.Player.X < 3 {
		t.Fatalf("przytrzymany drazek nie prowadzi dalej: x=%d", g.Player.X)
	}
}

func TestStickChangesDirectionWithoutLifting(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 512, 500)
	if g.hold != (Point{1, 0}) {
		t.Fatalf("brak marszu w prawo: %v", g.hold)
	}
	gs.Move(1, 512, 560)
	if g.hold != (Point{0, 1}) {
		t.Fatalf("kierunek nie zmienil sie na dol: %v", g.hold)
	}
	if !gs.Active() {
		t.Fatal("palec zostal zwolniony")
	}
}

func TestStickHysteresisKeepsAxis(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 540, 500)
	gs.Move(1, 540, 544)
	if g.hold != (Point{1, 0}) {
		t.Fatalf("histereza nie utrzymala osi poziomej: %v", g.hold)
	}
	gs.Move(1, 540, 570)
	if g.hold != (Point{0, 1}) {
		t.Fatalf("wyrazne wychylenie w dol nie przelaczylo osi: %v", g.hold)
	}
}

func TestStickAnchorFollowsFinger(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 900, 500)
	if gs.App.Stick.AnchorX <= 500 {
		t.Fatalf("kotwica nie podazyla za palcem: %v", gs.App.Stick.AnchorX)
	}
	if d := 900 - gs.App.Stick.AnchorX; d > gs.App.Stick.Radius+.001 {
		t.Fatalf("wychylenie poza promien: %v > %v", d, gs.App.Stick.Radius)
	}
	gs.Move(1, 820, 500)
	if g.hold != (Point{-1, 0}) {
		t.Fatalf("powrot palca nie odwrocil kierunku: %v", g.hold)
	}
}

func TestReleaseInsideDeadZoneStopsWalk(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 560, 500)
	gs.Move(1, 502, 500)
	if g.hold != (Point{}) {
		t.Fatalf("powrot do martwej strefy nie zatrzymal marszu: %v", g.hold)
	}
}

func TestTapWithoutMovePlacesBomb(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(3, 800, 800)
	gs.Move(3, 803, 802)
	gs.Up(3, 803, 802)
	if len(g.Bombs) != 1 || g.Bombs[0].X != 1 || g.Bombs[0].Y != 1 {
		t.Fatalf("bomba nie stoi pod graczem: %+v", g.Bombs)
	}
	if gs.Active() || g.hold != (Point{}) {
		t.Fatal("gest nie zostal zamkniety")
	}
}

func TestWalkGestureDoesNotPlaceBombOnLift(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 560, 500)
	gs.Up(1, 560, 500)
	if len(g.Bombs) != 0 {
		t.Fatalf("marsz zakonczyl sie bomba: %d", len(g.Bombs))
	}
	if g.hold != (Point{}) || gs.App.Stick.Active {
		t.Fatal("drazek nie zniknal po oderwaniu")
	}
}

func TestSecondFingerPlacesBombWhileWalking(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 560, 500)
	gs.Down(2, 1200, 700)
	if len(g.Bombs) != 1 {
		t.Fatalf("drugi palec nie podlozyl bomby: %d", len(g.Bombs))
	}
	gs.Up(2, 1200, 700)
	if len(g.Bombs) != 1 || g.hold != (Point{1, 0}) {
		t.Fatalf("oderwanie drugiego palca zmienilo stan: bomby=%d hold=%v", len(g.Bombs), g.hold)
	}
	gs.Up(1, 560, 500)
	if g.hold != (Point{}) {
		t.Fatal("pierwszy palec nie zwolnil marszu")
	}
}

func TestUITapFiresOnPressAndOnlyOnce(t *testing.T) {
	a := &App{Screen: ScreenAttract, ViewW: 1440, ViewH: 810, StateAge: 1}
	gs := NewGestures(a, 1440, 810)
	gs.Down(1, 700, 400)
	if a.Screen != ScreenNick {
		t.Fatal("dotkniecie nie zadzialalo od razu przy przylozeniu palca")
	}
	gs.Up(1, 700, 400)
	if len(a.Pulses) != 1 {
		t.Fatalf("liczba akcji: %d", len(a.Pulses))
	}
}

func TestUITapIgnoredRightAfterScreenChange(t *testing.T) {
	a := &App{Screen: ScreenAttract, ViewW: 1440, ViewH: 810, StateAge: .1}
	gs := NewGestures(a, 1440, 810)
	gs.Down(1, 700, 400)
	if a.Screen != ScreenAttract || len(a.Pulses) != 0 {
		t.Fatalf("gest przeszedl przez blokade: screen=%d pulses=%d", a.Screen, len(a.Pulses))
	}
	// Palec przylozony w blokadzie jest zablokowany na cale dotkniecie, takze po jej minieciu.
	for i := 0; i < 5; i++ {
		a.Update(.1)
	}
	gs.Up(1, 700, 400)
	if a.Screen != ScreenAttract || len(a.Pulses) != 0 {
		t.Fatalf("oderwanie po blokadzie wywolalo akcje: screen=%d pulses=%d", a.Screen, len(a.Pulses))
	}
	gs.Down(2, 700, 400)
	if a.Screen != ScreenNick {
		t.Fatal("kolejne dotkniecie po blokadzie nie dziala")
	}
}

func TestSecondFingerOutsidePlayIsIgnored(t *testing.T) {
	a := &App{Screen: ScreenSummary, ViewW: 1440, ViewH: 810, StateAge: 1}
	gs := NewGestures(a, 1440, 810)
	gs.Down(1, 10, 10)
	gs.Down(2, 20, 20)
	gs.Up(2, 20, 20)
	if len(a.Pulses) != 1 {
		t.Fatalf("drugi palec poza gra wywolal akcje: %d", len(a.Pulses))
	}
}

func TestCancelReleasesHold(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 560, 500)
	gs.Cancel()
	if g.hold != (Point{}) || gs.Active() || gs.App.Stick.Active {
		t.Fatal("anulowanie nie zwolnilo gestu")
	}
}

func TestTapMapsClientToLogical(t *testing.T) {
	a := &App{Screen: ScreenAttract, ViewW: 1440, ViewH: 810, StateAge: 1}
	gs := NewGestures(a, 3840, 2160)
	gs.Down(1, 3839, 2159)
	gs.Up(1, 3839, 2159)
	if a.Screen != ScreenNick || len(a.Pulses) != 1 || a.Pulses[0].X > 1440 || a.Pulses[0].Y > 810 {
		t.Fatalf("mapowanie: screen=%d pulses=%+v", a.Screen, a.Pulses)
	}
}

func TestHoldExpiresWithoutPointerEventsAndResumes(t *testing.T) {
	gs, g := newPlayGestures()
	gs.Down(1, 500, 500)
	gs.Move(1, 560, 500)
	gs.Tick(HoldTimeout - .5)
	if g.hold == (Point{}) {
		t.Fatal("marsz wygasl za wczesnie")
	}
	gs.Tick(.6)
	if g.hold != (Point{}) || !gs.Active() {
		t.Fatalf("po timeout: hold=%v active=%v", g.hold, gs.Active())
	}
	gs.Move(1, 561, 500)
	if g.hold != (Point{1, 0}) {
		t.Fatalf("ruch palca nie wznowil marszu: %v", g.hold)
	}
	gs.Up(1, 561, 500)
	if len(g.Bombs) != 0 {
		t.Fatal("oderwanie po marszu podlozylo bombe")
	}
}
