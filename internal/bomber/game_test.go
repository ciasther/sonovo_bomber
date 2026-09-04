package bomber

import (
	"math"
	"testing"
	"time"
)

func TestBombFuseExplosionAndCrateStop(t *testing.T) {
	g := NewGame(75, 1)
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			if x == 0 || y == 0 || x == GridW-1 || y == GridH-1 {
				g.Grid[y][x] = Wall
			} else {
				g.Grid[y][x] = Floor
			}
		}
	}
	g.Player.X, g.Player.Y = 1, 1
	g.Player.ToX, g.Player.ToY = 1, 1
	g.Player.Range = 4
	g.Player.Invulnerable = 10
	g.Grid[1][3] = Crate
	g.Enemies = []*Enemy{{ID: 99, Mover: Mover{X: 11, Y: 9, ToX: 11, ToY: 9, MoveT: 1, MoveDuration: 1}, Alive: true, DecisionIn: 100}}

	if !g.PlaceBomb() {
		t.Fatal("bomba nie zostala ustawiona")
	}
	for i := 0; i < 37; i++ {
		g.Update(.05)
	}
	if g.Grid[1][3] != Floor {
		t.Fatal("skrzynka nie zostala zniszczona")
	}
	if g.Score != 10 {
		t.Fatalf("wynik za skrzynke = %d, oczekiwano 10", g.Score)
	}
	if g.Player.ActiveBombs != 0 {
		t.Fatalf("aktywne bomby = %d, oczekiwano 0", g.Player.ActiveBombs)
	}
	if len(g.Explosions) == 0 {
		t.Fatal("brak aktywnej eksplozji")
	}
	ex := g.Explosions[0]
	if !g.explosionContains(ex, 2, 1) || !g.explosionContains(ex, 3, 1) {
		t.Fatal("eksplozja nie objela oczekiwanych pol")
	}
	if g.explosionContains(ex, 4, 1) {
		t.Fatal("eksplozja przeszla przez skrzynke")
	}
}

func TestShieldAndTwoLives(t *testing.T) {
	g := NewGame(75, 2)
	g.Player.Shield = true
	g.hitPlayer()
	if g.Player.Lives != 2 || g.Player.Shield {
		t.Fatalf("tarcza: lives=%d shield=%v", g.Player.Lives, g.Player.Shield)
	}
	g.Player.Invulnerable = 0
	g.hitPlayer()
	if g.Player.Lives != 1 || g.Finished {
		t.Fatalf("pierwsze zycie: lives=%d finished=%v", g.Player.Lives, g.Finished)
	}
	g.Player.Invulnerable = 0
	g.hitPlayer()
	if g.Player.Lives != 0 || !g.Finished || g.Result.CompletedRun {
		t.Fatalf("drugie zycie: lives=%d finished=%v completed=%v", g.Player.Lives, g.Finished, g.Result.CompletedRun)
	}
}

func TestOneBombMultiplierCapsAtFour(t *testing.T) {
	g := NewGame(75, 3)
	g.Score = 0
	g.Enemies = nil
	ex := &Explosion{TTL: .5, Duration: .5}
	for i := 0; i < 5; i++ {
		p := Point{2 + i, 3}
		ex.Cells = append(ex.Cells, p)
		g.Enemies = append(g.Enemies, &Enemy{ID: i + 1, Mover: Mover{X: p.X, Y: p.Y, ToX: p.X, ToY: p.Y, MoveT: 1, MoveDuration: 1}, Alive: true})
	}
	g.Player.Invulnerable = 10
	g.applyExplosionHits(ex)
	if g.Score != 1400 {
		t.Fatalf("wynik mnoznika = %d, oczekiwano 1400", g.Score)
	}
	if g.CurrentCombo != 4 || g.BestCombo != 4 || ex.EnemyKills != 5 {
		t.Fatalf("combo=%d best=%d kills=%d", g.CurrentCombo, g.BestCombo, ex.EnemyKills)
	}
}

func TestCompletedRoundAddsBonus(t *testing.T) {
	g := NewGame(75, 4)
	g.Score = 125
	g.Elapsed = 74.98
	g.Player.Invulnerable = 10
	g.Enemies = []*Enemy{{ID: 1, Mover: Mover{X: 10, Y: 10, ToX: 10, ToY: 10, MoveT: 1, MoveDuration: 1}, Alive: true, DecisionIn: 100}}
	g.Update(.05)
	if !g.Finished || !g.Result.CompletedRun {
		t.Fatalf("finished=%v completed=%v", g.Finished, g.Result.CompletedRun)
	}
	if g.Result.Score != 625 || g.Result.SurvivalMS != 75000 {
		t.Fatalf("score=%d survival=%d", g.Result.Score, g.Result.SurvivalMS)
	}
}

func TestRankingTieBreakers(t *testing.T) {
	base := time.Date(2026, 9, 1, 10, 0, 0, 0, time.Local)
	entries := []ScoreEntry{
		{ID: "later", EventID: "e", Score: 1000, SurvivalMS: 70000, CreatedAt: base.Add(time.Minute)},
		{ID: "shorter", EventID: "e", Score: 1000, SurvivalMS: 69000, CreatedAt: base.Add(-time.Minute)},
		{ID: "earlier", EventID: "e", Score: 1000, SurvivalMS: 70000, CreatedAt: base},
		{ID: "higher", EventID: "e", Score: 1100, SurvivalMS: 1000, CreatedAt: base.Add(2 * time.Minute)},
		{ID: "other", EventID: "x", Score: 9999, SurvivalMS: 75000, CreatedAt: base},
	}
	ranked := RankEntries(entries, "e", nil)
	want := []string{"higher", "earlier", "later", "shorter"}
	if len(ranked) != len(want) {
		t.Fatalf("ranking ma %d pozycji", len(ranked))
	}
	for i := range want {
		if ranked[i].ID != want[i] {
			t.Fatalf("pozycja %d = %s, oczekiwano %s", i, ranked[i].ID, want[i])
		}
	}
}

func TestParticleBurstIsCapped(t *testing.T) {
	g := NewGame(75, 9)
	g.Particles = nil
	g.spawnBurst(1, 1, RGB(255, 255, 255), 30)
	if len(g.Particles) != 12 {
		t.Fatalf("liczba czastek = %d, oczekiwano 12", len(g.Particles))
	}
}

func TestSummaryWaitsForPlayerAction(t *testing.T) {
	a := &App{Screen: ScreenSummary}
	a.Update(30)
	if a.Screen != ScreenSummary {
		t.Fatalf("ekran wyniku zmienil sie automatycznie na %d", a.Screen)
	}
}

func TestTutorialIsShownOnlyForFirstRun(t *testing.T) {
	a := &App{Brand: &Branding{Config: DefaultConfig()}}
	a.startRun()
	if !a.TutorialVisible {
		t.Fatal("brak podpowiedzi w pierwszej rundzie")
	}
	a.startRun()
	if a.TutorialVisible {
		t.Fatal("podpowiedz wrocila w kolejnej rundzie")
	}
}

func TestBombActionQueuesFeedbackSound(t *testing.T) {
	a := &App{Screen: ScreenPlay, ViewW: 1440, ViewH: 810, Game: NewGame(75, 10)}
	a.Tap(100, 100)
	if sounds := a.DrainSounds(); len(sounds) != 1 || sounds[0] != SoundBombPlaced {
		t.Fatalf("dzwiek poprawnej bomby = %v", sounds)
	}
	a.Tap(100, 100)
	if sounds := a.DrainSounds(); len(sounds) != 1 || sounds[0] != SoundBombBlocked {
		t.Fatalf("dzwiek zablokowanej bomby = %v", sounds)
	}
}

func TestSwipeQueuesNextMoveDuringAnimation(t *testing.T) {
	g := NewGame(75, 11)
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			if x == 0 || y == 0 || x == GridW-1 || y == GridH-1 {
				g.Grid[y][x] = Wall
			} else {
				g.Grid[y][x] = Floor
			}
		}
	}
	g.Enemies = nil

	if !g.Swipe(1, 0) || !g.Swipe(1, 0) {
		t.Fatal("ruch nie zostal przyjety podczas animacji")
	}
	g.Release()
	for i := 0; i < 3; i++ {
		g.Update(.05)
	}
	if g.Player.X != 2 || g.Player.ToX != 3 {
		t.Fatalf("pierwszy ruch nie uruchomil drugiego: x=%d to=%d", g.Player.X, g.Player.ToX)
	}
	for i := 0; i < 3; i++ {
		g.Update(.05)
	}
	if g.Player.X != 3 || g.Player.IsMoving() {
		t.Fatalf("drugi ruch nie zostal wykonany: x=%d moving=%v", g.Player.X, g.Player.IsMoving())
	}
}

func openGrid(g *Game) {
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			if x == 0 || y == 0 || x == GridW-1 || y == GridH-1 {
				g.Grid[y][x] = Wall
			} else {
				g.Grid[y][x] = Floor
			}
		}
	}
	g.Enemies = nil
}

func TestHeldSwipeKeepsWalkingUntilRelease(t *testing.T) {
	g := NewGame(75, 12)
	openGrid(g)
	g.Swipe(1, 0)
	for i := 0; i < 10; i++ {
		g.Update(.05)
	}
	if g.Player.X < 4 {
		t.Fatalf("przytrzymany gest nie prowadzi dalej: x=%d", g.Player.X)
	}
	g.Release()
	for i := 0; i < 10; i++ {
		g.Update(.05)
	}
	stopped := g.Player.X
	for i := 0; i < 10; i++ {
		g.Update(.05)
	}
	if g.Player.X != stopped || g.Player.IsMoving() {
		t.Fatalf("po zwolnieniu postac dalej idzie: x=%d", g.Player.X)
	}
}

func TestHitClearsHeldDirection(t *testing.T) {
	g := NewGame(75, 13)
	openGrid(g)
	g.Swipe(0, 1)
	g.hitPlayer()
	if g.hold != (Point{}) {
		t.Fatal("trafienie nie zatrzymalo marszu")
	}
}

func TestTwoBombsAtStart(t *testing.T) {
	g := NewGame(75, 14)
	openGrid(g)
	if !g.PlaceBomb() {
		t.Fatal("pierwsza bomba odrzucona")
	}
	g.Swipe(1, 0)
	g.Release()
	for i := 0; i < 4; i++ {
		g.Update(.05)
	}
	if !g.PlaceBomb() {
		t.Fatal("druga bomba odrzucona")
	}
	if g.PlaceBomb() {
		t.Fatal("trzecia bomba przyjeta ponad limit")
	}
}

func TestSimulationTempoIsSlowerThanRoundClock(t *testing.T) {
	g := NewGame(75, 5)
	if !g.PlaceBomb() {
		t.Fatal("bomba nie zostala ustawiona")
	}
	g.Update(.05)
	if math.Abs(g.Elapsed-.05) > 1e-9 {
		t.Fatalf("zegar rundy nie idzie w czasie rzeczywistym: %v", g.Elapsed)
	}
	if math.Abs(g.Bombs[0].Fuse-(BombFuse-.05*GameSpeed)) > 1e-9 {
		t.Fatalf("lont nie zwolnil o GameSpeed: %v", g.Bombs[0].Fuse)
	}
}

func TestBlastCellsMatchDetonationAndStopAtObstacles(t *testing.T) {
	g := NewGame(75, 3)
	openGrid(g)
	g.Grid[1][4] = Crate
	g.Grid[3][1] = Wall
	g.Player.X, g.Player.Y = 1, 1
	g.Player.Range = 4
	cells := g.BlastCells(1, 1, g.Player.Range)
	has := func(x, y int) bool {
		for _, p := range cells {
			if p.X == x && p.Y == y {
				return true
			}
		}
		return false
	}
	if !has(1, 1) || !has(4, 1) {
		t.Fatalf("podglad nie siega do skrzyni: %v", cells)
	}
	if has(5, 1) {
		t.Fatal("podglad przeszedl przez skrzynie")
	}
	if has(1, 3) || has(1, 4) {
		t.Fatalf("podglad przeszedl przez sciane: %v", cells)
	}
	if !g.PlaceBomb() {
		t.Fatal("bomba nie zostala postawiona")
	}
	g.detonate(g.Bombs[0])
	e := g.Explosions[len(g.Explosions)-1]
	if len(e.Cells) != len(cells) {
		t.Fatalf("podglad %d != detonacja %d", len(cells), len(e.Cells))
	}
	for i := range cells {
		if cells[i] != e.Cells[i] {
			t.Fatalf("pole %d: podglad %v != detonacja %v", i, cells[i], e.Cells[i])
		}
	}
}

func TestBombOriginFollowsPlayerPastHalfStep(t *testing.T) {
	g := NewGame(75, 3)
	openGrid(g)
	g.Player.X, g.Player.Y = 1, 1
	if x, y := g.BombOrigin(); x != 1 || y != 1 {
		t.Fatalf("stojac: %d,%d", x, y)
	}
	g.Swipe(1, 0)
	g.Update(.02)
	if x, _ := g.BombOrigin(); x != 1 {
		t.Fatalf("na poczatku kroku bomba powinna zostac pod graczem: %d", x)
	}
	for g.Player.IsMoving() && g.Player.MoveT/g.Player.MoveDuration <= .6 {
		g.Update(.01)
	}
	if x, _ := g.BombOrigin(); x != 2 {
		t.Fatalf("po polowie kroku bomba powinna isc na pole docelowe: %d", x)
	}
}

func TestTargetCellShowsHeldDirection(t *testing.T) {
	g := NewGame(75, 3)
	openGrid(g)
	g.Player.X, g.Player.Y = 1, 1
	if g.TargetCell() != (Point{1, 1}) {
		t.Fatalf("bez kierunku: %v", g.TargetCell())
	}
	g.Swipe(1, 0)
	if g.TargetCell() != (Point{2, 1}) {
		t.Fatalf("w ruchu: %v", g.TargetCell())
	}
	g.Release()
	for g.Player.IsMoving() {
		g.Update(.01)
	}
	if g.TargetCell() != (Point{2, 1}) {
		t.Fatalf("po zatrzymaniu: %v", g.TargetCell())
	}
}
