package bomber

import (
	"math"
	"math/rand"
	"time"
)

const (
	GridW    = 15
	GridH    = 13
	BombFuse = 1.6
)

type Cell uint8

const (
	Floor Cell = iota
	Wall
	Crate
)

type Point struct{ X, Y int }

type PickupType uint8

const (
	PickupRange PickupType = iota
	PickupBomb
	PickupShield
	PickupPartner
)

type Mover struct {
	X, Y         int
	FromX, FromY int
	ToX, ToY     int
	MoveT        float64
	MoveDuration float64
}

func (m *Mover) IsMoving() bool { return m.MoveT < m.MoveDuration }
func (m *Mover) Visual() (float64, float64) {
	if m.MoveDuration <= 0 || m.MoveT >= m.MoveDuration {
		return float64(m.X), float64(m.Y)
	}
	t := m.MoveT / m.MoveDuration
	t = 1 - math.Pow(1-t, 3)
	return float64(m.FromX) + float64(m.ToX-m.FromX)*t, float64(m.FromY) + float64(m.ToY-m.FromY)*t
}
func (m *Mover) Step(dt float64) {
	if !m.IsMoving() {
		return
	}
	m.MoveT += dt
	if m.MoveT >= m.MoveDuration {
		m.X, m.Y = m.ToX, m.ToY
		m.MoveT = m.MoveDuration
	}
}
func (m *Mover) StartMove(x, y int, duration float64) {
	m.FromX, m.FromY = m.X, m.Y
	m.ToX, m.ToY = x, y
	m.MoveT = 0
	m.MoveDuration = duration
}

type Player struct {
	Mover
	Lives         int
	Range         int
	MaxBombs      int
	ActiveBombs   int
	Shield        bool
	Invulnerable  float64
	RespawnFlash  float64
	LastDirection Point
}

type Enemy struct {
	ID int
	Mover
	DecisionIn  float64
	Alive       bool
	PulseOffset float64
}

type Bomb struct {
	ID       int
	X, Y     int
	Fuse     float64
	Range    int
	Exploded bool
}

type Explosion struct {
	ID         int
	BombID     int
	Cells      []Point
	TTL        float64
	Duration   float64
	EnemyKills int
}

type Pickup struct {
	X, Y int
	Type PickupType
	Age  float64
}

type Particle struct {
	X, Y      float64
	VX, VY    float64
	Life, Max float64
	Size      float64
	Color     Color
	Ring      bool
}

type FloatingText struct {
	X, Y  float64
	Text  string
	Color Color
	Life  float64
	Max   float64
}

type GameResult struct {
	Score        int
	SurvivalMS   int64
	BestCombo    int
	Waves        int
	CompletedRun bool
}

type Game struct {
	Grid         [GridH][GridW]Cell
	Player       Player
	Enemies      []*Enemy
	Bombs        []*Bomb
	Explosions   []*Explosion
	Pickups      []*Pickup
	Particles    []Particle
	Floating     []FloatingText
	Score        int
	BestCombo    int
	CurrentCombo int
	Wave         int
	Elapsed      float64
	RoundSeconds float64
	Finished     bool
	Result       GameResult
	Last20       bool
	MidWave      bool
	FinalWave    bool
	WavePending  float64
	Shake        float64
	Flash        float64
	Rng          *rand.Rand
	queuedMove   Point
	hold         Point
	nextID       int
}

func NewGame(roundSeconds int, seed int64) *Game {
	if roundSeconds <= 0 {
		roundSeconds = 75
	}
	g := &Game{
		RoundSeconds: float64(roundSeconds),
		Wave:         1,
		Rng:          rand.New(rand.NewSource(seed)),
	}
	g.Player = Player{Mover: Mover{X: 1, Y: 1, ToX: 1, ToY: 1, MoveDuration: .12, MoveT: .12}, Lives: 2, Range: 2, MaxBombs: 2}
	g.generateGrid()
	g.spawnEnemies(1)
	return g
}

func (g *Game) next() int { g.nextID++; return g.nextID }

func (g *Game) generateGrid() {
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			switch {
			case x == 0 || y == 0 || x == GridW-1 || y == GridH-1:
				g.Grid[y][x] = Wall
			case x%2 == 0 && y%2 == 0:
				g.Grid[y][x] = Wall
			default:
				g.Grid[y][x] = Floor
			}
		}
	}
	safe := map[Point]bool{{1, 1}: true, {2, 1}: true, {1, 2}: true, {GridW - 2, GridH - 2}: true, {GridW - 3, GridH - 2}: true, {GridW - 2, GridH - 3}: true}
	for y := 1; y < GridH-1; y++ {
		for x := 1; x < GridW-1; x++ {
			p := Point{x, y}
			if g.Grid[y][x] == Floor && !safe[p] && g.Rng.Float64() < .58 {
				g.Grid[y][x] = Crate
			}
		}
	}
}

func (g *Game) Remaining() float64 { return maxFloat(0, g.RoundSeconds-g.Elapsed) }

func (g *Game) Swipe(dx, dy int) bool {
	if g.Finished {
		return false
	}
	if abs(dx)+abs(dy) != 1 {
		return false
	}
	direction := Point{dx, dy}
	g.hold = direction
	if g.Player.IsMoving() {
		g.queuedMove = direction
		return true
	}
	return g.startMove(direction)
}

func (g *Game) Release() { g.hold = Point{} }

func (g *Game) startMove(direction Point) bool {
	dx, dy := direction.X, direction.Y
	tx, ty := g.Player.X+dx, g.Player.Y+dy
	if !g.passable(tx, ty, true) {
		return false
	}
	g.Player.LastDirection = Point{dx, dy}
	g.Player.StartMove(tx, ty, .105)
	return true
}

func (g *Game) PlaceBomb() bool {
	if g.Finished || g.Player.ActiveBombs >= g.Player.MaxBombs {
		return false
	}
	x, y := g.Player.X, g.Player.Y
	if g.Player.IsMoving() && g.Player.MoveT/g.Player.MoveDuration > .55 {
		x, y = g.Player.ToX, g.Player.ToY
	}
	for _, b := range g.Bombs {
		if !b.Exploded && b.X == x && b.Y == y {
			return false
		}
	}
	g.Bombs = append(g.Bombs, &Bomb{ID: g.next(), X: x, Y: y, Fuse: BombFuse, Range: g.Player.Range})
	g.Player.ActiveBombs++
	g.spawnRing(float64(x)+.5, float64(y)+.5, RGB(99, 243, 255))
	return true
}

func (g *Game) Update(dt float64) {
	if dt <= 0 || g.Finished {
		return
	}
	if dt > .05 {
		dt = .05
	}
	g.Elapsed += dt
	g.Player.Step(dt)
	if !g.Player.IsMoving() && g.queuedMove != (Point{}) {
		direction := g.queuedMove
		g.queuedMove = Point{}
		g.startMove(direction)
	}
	if !g.Player.IsMoving() && g.hold != (Point{}) {
		g.startMove(g.hold)
	}
	if g.Player.Invulnerable > 0 {
		g.Player.Invulnerable -= dt
	}
	if g.Player.RespawnFlash > 0 {
		g.Player.RespawnFlash -= dt
	}
	if g.Shake > 0 {
		g.Shake = maxFloat(0, g.Shake-dt*3.2)
	}
	if g.Flash > 0 {
		g.Flash = maxFloat(0, g.Flash-dt*2.8)
	}
	for _, p := range g.Pickups {
		p.Age += dt
	}
	g.collectPickup()
	g.updateDifficulty()
	g.updateBombs(dt)
	g.updateExplosions(dt)
	g.updateEnemies(dt)
	g.updateParticles(dt)
	g.checkActiveExplosionHits()
	g.checkEnemyContact()
	g.cleanup()
	if len(g.Enemies) == 0 && g.WavePending <= 0 {
		g.Score += 250
		g.Wave++
		g.Floating = append(g.Floating, FloatingText{X: float64(GridW) / 2, Y: 1.1, Text: "+250 FALA", Color: RGB(255, 181, 71), Life: 1.5, Max: 1.5})
		g.WavePending = .8
	}
	if g.WavePending > 0 {
		g.WavePending -= dt
		if g.WavePending <= 0 && len(g.Enemies) == 0 {
			g.spawnEnemies(min(6, 1+g.Wave))
		}
	}
	if g.Elapsed >= g.RoundSeconds {
		g.Score += 500
		g.finish(true)
	}
}

func (g *Game) updateDifficulty() {
	if !g.MidWave && g.Elapsed >= g.RoundSeconds*.34 {
		g.MidWave = true
		need := 3 - len(g.Enemies)
		if need > 0 {
			g.spawnEnemies(need)
		}
		g.Floating = append(g.Floating, FloatingText{X: float64(GridW) / 2, Y: 1.2, Text: "SZYBCIEJ!", Color: RGB(255, 79, 216), Life: 1.5, Max: 1.5})
	}
	if !g.FinalWave && g.Remaining() <= 20 {
		g.FinalWave = true
		g.Last20 = true
		g.addLateCrates(12)
		g.spawnEnemies(2)
		g.Floating = append(g.Floating, FloatingText{X: float64(GridW) / 2, Y: 1.2, Text: "FINALOWE 20 SEKUND!", Color: RGB(255, 181, 71), Life: 1.8, Max: 1.8})
	}
}

func (g *Game) updateBombs(dt float64) {
	for _, b := range g.Bombs {
		if b.Exploded {
			continue
		}
		b.Fuse -= dt
		if b.Fuse <= 0 {
			g.detonate(b)
		}
	}
}

func (g *Game) detonate(b *Bomb) {
	if b == nil || b.Exploded {
		return
	}
	b.Exploded = true
	g.Player.ActiveBombs = max(0, g.Player.ActiveBombs-1)
	e := &Explosion{ID: g.next(), BombID: b.ID, TTL: .56, Duration: .56}
	e.Cells = append(e.Cells, Point{b.X, b.Y})
	dirs := []Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, d := range dirs {
		for step := 1; step <= b.Range; step++ {
			x, y := b.X+d.X*step, b.Y+d.Y*step
			if x < 0 || y < 0 || x >= GridW || y >= GridH || g.Grid[y][x] == Wall {
				break
			}
			e.Cells = append(e.Cells, Point{x, y})
			if g.Grid[y][x] == Crate {
				g.Grid[y][x] = Floor
				g.Score += 10
				g.spawnCratePickup(x, y)
				g.spawnBurst(float64(x)+.5, float64(y)+.5, RGB(255, 181, 71), 8)
				break
			}
			for _, chained := range g.Bombs {
				if !chained.Exploded && chained.X == x && chained.Y == y {
					g.detonate(chained)
				}
			}
		}
	}
	g.Explosions = append(g.Explosions, e)
	g.Shake = minFloat(1, g.Shake+.42)
	g.Flash = minFloat(1, g.Flash+.28)
	g.spawnBurst(float64(b.X)+.5, float64(b.Y)+.5, RGB(255, 117, 53), 18)
	g.applyExplosionHits(e)
}

func (g *Game) spawnCratePickup(x, y int) {
	r := g.Rng.Float64()
	switch {
	case r < .09:
		g.Pickups = append(g.Pickups, &Pickup{X: x, Y: y, Type: PickupPartner})
	case r < .16:
		g.Pickups = append(g.Pickups, &Pickup{X: x, Y: y, Type: PickupRange})
	case r < .22:
		g.Pickups = append(g.Pickups, &Pickup{X: x, Y: y, Type: PickupBomb})
	case r < .28:
		g.Pickups = append(g.Pickups, &Pickup{X: x, Y: y, Type: PickupShield})
	}
}

func (g *Game) updateExplosions(dt float64) {
	for _, e := range g.Explosions {
		e.TTL -= dt
	}
}

func (g *Game) applyExplosionHits(e *Explosion) {
	for _, enemy := range g.Enemies {
		if enemy.Alive && g.explosionContains(e, enemy.X, enemy.Y) {
			g.killEnemy(enemy, e)
		}
	}
	px, py := g.Player.X, g.Player.Y
	if g.Player.IsMoving() && g.Player.MoveT/g.Player.MoveDuration > .5 {
		px, py = g.Player.ToX, g.Player.ToY
	}
	if g.explosionContains(e, px, py) {
		g.hitPlayer()
	}
}

func (g *Game) checkActiveExplosionHits() {
	for _, e := range g.Explosions {
		if e.TTL <= 0 {
			continue
		}
		for _, enemy := range g.Enemies {
			if enemy.Alive && g.explosionContains(e, enemy.X, enemy.Y) {
				g.killEnemy(enemy, e)
			}
		}
		px, py := g.Player.X, g.Player.Y
		if g.Player.IsMoving() && g.Player.MoveT/g.Player.MoveDuration > .5 {
			px, py = g.Player.ToX, g.Player.ToY
		}
		if g.explosionContains(e, px, py) {
			g.hitPlayer()
		}
	}
}

func (g *Game) explosionContains(e *Explosion, x, y int) bool {
	for _, p := range e.Cells {
		if p.X == x && p.Y == y {
			return true
		}
	}
	return false
}

func (g *Game) killEnemy(enemy *Enemy, e *Explosion) {
	if !enemy.Alive {
		return
	}
	enemy.Alive = false
	e.EnemyKills++
	mult := min(4, e.EnemyKills)
	g.CurrentCombo = mult
	g.BestCombo = max(g.BestCombo, mult)
	points := 100 * mult
	g.Score += points
	g.Floating = append(g.Floating, FloatingText{X: float64(enemy.X) + .5, Y: float64(enemy.Y) + .3, Text: "+" + itoa(points) + " X" + itoa(mult), Color: RGB(255, 255, 255), Life: 1.1, Max: 1.1})
	g.spawnBurst(float64(enemy.X)+.5, float64(enemy.Y)+.5, RGB(255, 79, 216), 16)
}

func (g *Game) hitPlayer() {
	if g.Player.Invulnerable > 0 || g.Finished {
		return
	}
	if g.Player.Shield {
		g.Player.Shield = false
		g.Player.Invulnerable = 1.0
		g.Floating = append(g.Floating, FloatingText{X: float64(g.Player.X) + .5, Y: float64(g.Player.Y), Text: "TARCZA!", Color: RGB(99, 243, 255), Life: 1.0, Max: 1.0})
		g.spawnRing(float64(g.Player.X)+.5, float64(g.Player.Y)+.5, RGB(99, 243, 255))
		return
	}
	g.queuedMove = Point{}
	g.hold = Point{}
	g.Player.Lives--
	g.CurrentCombo = 0
	g.Player.Invulnerable = 2.0
	g.Player.RespawnFlash = 1.2
	g.Flash = 1
	g.Shake = 1
	g.spawnBurst(float64(g.Player.X)+.5, float64(g.Player.Y)+.5, RGB(255, 72, 96), 20)
	if g.Player.Lives <= 0 {
		g.finish(false)
		return
	}
	p := g.safeRespawn()
	g.Player.X, g.Player.Y = p.X, p.Y
	g.Player.ToX, g.Player.ToY = p.X, p.Y
	g.Player.FromX, g.Player.FromY = p.X, p.Y
	g.Player.MoveT = g.Player.MoveDuration
}

func (g *Game) safeRespawn() Point {
	candidates := []Point{{1, 1}, {1, GridH - 2}, {GridW - 2, 1}, {GridW - 2, GridH - 2}}
	best := candidates[0]
	bestScore := -1
	for _, p := range candidates {
		if g.Grid[p.Y][p.X] != Floor || g.isBombAt(p.X, p.Y) {
			continue
		}
		score := 100
		for _, e := range g.Enemies {
			if e.Alive {
				score += abs(e.X-p.X) + abs(e.Y-p.Y)
			}
		}
		for _, ex := range g.Explosions {
			if ex.TTL > 0 && g.explosionContains(ex, p.X, p.Y) {
				score -= 1000
			}
		}
		if score > bestScore {
			bestScore, best = score, p
		}
	}
	g.Grid[best.Y][best.X] = Floor
	return best
}

func (g *Game) updateEnemies(dt float64) {
	interval := .72
	moveDuration := .22
	if g.MidWave {
		interval, moveDuration = .50, .17
	}
	if g.FinalWave {
		interval, moveDuration = .32, .135
	}
	for _, e := range g.Enemies {
		if !e.Alive {
			continue
		}
		e.Step(dt)
		e.DecisionIn -= dt
		if !e.IsMoving() && e.DecisionIn <= 0 {
			e.DecisionIn = interval * (.72 + g.Rng.Float64()*.55)
			dirs := []Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
			g.Rng.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
			if g.Rng.Float64() < .62 {
				dx, dy := g.Player.X-e.X, g.Player.Y-e.Y
				preferred := Point{}
				if abs(dx) > abs(dy) {
					preferred.X = sign(dx)
				} else {
					preferred.Y = sign(dy)
				}
				dirs = append([]Point{preferred}, dirs...)
			}
			for _, d := range dirs {
				if d.X == 0 && d.Y == 0 {
					continue
				}
				tx, ty := e.X+d.X, e.Y+d.Y
				if g.passable(tx, ty, false) {
					e.StartMove(tx, ty, moveDuration)
					break
				}
			}
		}
	}
}

func (g *Game) checkEnemyContact() {
	if g.Player.Invulnerable > 0 {
		return
	}
	px, py := g.Player.X, g.Player.Y
	if g.Player.IsMoving() && g.Player.MoveT/g.Player.MoveDuration > .5 {
		px, py = g.Player.ToX, g.Player.ToY
	}
	for _, e := range g.Enemies {
		if !e.Alive {
			continue
		}
		ex, ey := e.X, e.Y
		if e.IsMoving() && e.MoveT/e.MoveDuration > .5 {
			ex, ey = e.ToX, e.ToY
		}
		if ex == px && ey == py {
			g.hitPlayer()
			return
		}
	}
}

func (g *Game) collectPickup() {
	px, py := g.Player.X, g.Player.Y
	if g.Player.IsMoving() && g.Player.MoveT/g.Player.MoveDuration > .55 {
		px, py = g.Player.ToX, g.Player.ToY
	}
	for i := 0; i < len(g.Pickups); i++ {
		p := g.Pickups[i]
		if p.X != px || p.Y != py {
			continue
		}
		label := ""
		switch p.Type {
		case PickupRange:
			g.Player.Range = min(5, g.Player.Range+1)
			label = "ZASIEG +1"
		case PickupBomb:
			g.Player.MaxBombs = min(4, g.Player.MaxBombs+1)
			label = "BOMBA +1"
		case PickupShield:
			g.Player.Shield = true
			label = "TARCZA"
		case PickupPartner:
			g.Score += 50
			label = "+50 PARTNER"
		}
		g.Floating = append(g.Floating, FloatingText{X: float64(px) + .5, Y: float64(py), Text: label, Color: RGB(99, 243, 255), Life: 1.2, Max: 1.2})
		g.spawnRing(float64(px)+.5, float64(py)+.5, RGB(99, 243, 255))
		g.Pickups = append(g.Pickups[:i], g.Pickups[i+1:]...)
		i--
	}
}

func (g *Game) spawnEnemies(count int) {
	for i := 0; i < count; i++ {
		p, ok := g.randomSpawnCell()
		if !ok {
			return
		}
		e := &Enemy{ID: g.next(), Mover: Mover{X: p.X, Y: p.Y, ToX: p.X, ToY: p.Y, MoveT: 1, MoveDuration: 1}, Alive: true, DecisionIn: .2 + g.Rng.Float64()*.4, PulseOffset: g.Rng.Float64() * 10}
		g.Enemies = append(g.Enemies, e)
		g.spawnRing(float64(p.X)+.5, float64(p.Y)+.5, RGB(255, 79, 216))
	}
}

func (g *Game) randomSpawnCell() (Point, bool) {
	var cells []Point
	for y := 1; y < GridH-1; y++ {
		for x := 1; x < GridW-1; x++ {
			if g.Grid[y][x] != Floor || g.isBombAt(x, y) || abs(x-g.Player.X)+abs(y-g.Player.Y) < 7 {
				continue
			}
			occupied := false
			for _, e := range g.Enemies {
				if e.Alive && e.X == x && e.Y == y {
					occupied = true
					break
				}
			}
			if !occupied {
				cells = append(cells, Point{x, y})
			}
		}
	}
	if len(cells) == 0 {
		return Point{}, false
	}
	return cells[g.Rng.Intn(len(cells))], true
}

func (g *Game) addLateCrates(count int) {
	var cells []Point
	for y := 1; y < GridH-1; y++ {
		for x := 1; x < GridW-1; x++ {
			if g.Grid[y][x] == Floor && !g.isBombAt(x, y) && abs(x-g.Player.X)+abs(y-g.Player.Y) > 3 {
				cells = append(cells, Point{x, y})
			}
		}
	}
	g.Rng.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })
	for i := 0; i < min(count, len(cells)); i++ {
		p := cells[i]
		g.Grid[p.Y][p.X] = Crate
		g.spawnRing(float64(p.X)+.5, float64(p.Y)+.5, RGB(255, 181, 71))
	}
}

func (g *Game) passable(x, y int, player bool) bool {
	if x < 0 || y < 0 || x >= GridW || y >= GridH || g.Grid[y][x] != Floor || g.isBombAt(x, y) {
		return false
	}
	if !player {
		for _, e := range g.Enemies {
			if e.Alive && e.X == x && e.Y == y {
				return false
			}
		}
	}
	return true
}

func (g *Game) isBombAt(x, y int) bool {
	for _, b := range g.Bombs {
		if !b.Exploded && b.X == x && b.Y == y {
			return true
		}
	}
	return false
}

func (g *Game) updateParticles(dt float64) {
	for i := range g.Particles {
		p := &g.Particles[i]
		p.Life -= dt
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VX *= math.Pow(.2, dt)
		p.VY *= math.Pow(.2, dt)
	}
	for i := range g.Floating {
		g.Floating[i].Life -= dt
		g.Floating[i].Y -= dt * .55
	}
}

func (g *Game) cleanup() {
	aliveEnemies := g.Enemies[:0]
	for _, e := range g.Enemies {
		if e.Alive {
			aliveEnemies = append(aliveEnemies, e)
		}
	}
	g.Enemies = aliveEnemies
	bombs := g.Bombs[:0]
	for _, b := range g.Bombs {
		if !b.Exploded {
			bombs = append(bombs, b)
		}
	}
	g.Bombs = bombs
	explosions := g.Explosions[:0]
	for _, e := range g.Explosions {
		if e.TTL > 0 {
			explosions = append(explosions, e)
		}
	}
	g.Explosions = explosions
	particles := g.Particles[:0]
	for _, p := range g.Particles {
		if p.Life > 0 {
			particles = append(particles, p)
		}
	}
	g.Particles = particles
	floating := g.Floating[:0]
	for _, f := range g.Floating {
		if f.Life > 0 {
			floating = append(floating, f)
		}
	}
	g.Floating = floating
}

func (g *Game) spawnBurst(x, y float64, c Color, count int) {
	count = min(count, 12)
	for i := 0; i < count; i++ {
		a := g.Rng.Float64() * math.Pi * 2
		speed := 1.4 + g.Rng.Float64()*4.2
		life := .35 + g.Rng.Float64()*.65
		g.Particles = append(g.Particles, Particle{X: x, Y: y, VX: math.Cos(a) * speed, VY: math.Sin(a) * speed, Life: life, Max: life, Size: .06 + g.Rng.Float64()*.11, Color: c})
	}
}

func (g *Game) spawnRing(x, y float64, c Color) {
	g.Particles = append(g.Particles, Particle{X: x, Y: y, Life: .55, Max: .55, Size: .18, Color: c, Ring: true})
}

func (g *Game) finish(completed bool) {
	if g.Finished {
		return
	}
	g.Finished = true
	survival := g.Elapsed
	if completed {
		survival = g.RoundSeconds
	}
	g.Result = GameResult{Score: g.Score, SurvivalMS: int64(survival * 1000), BestCombo: max(1, g.BestCombo), Waves: g.Wave, CompletedRun: completed}
}

func (g *Game) PrepareShowcase() {
	g.Elapsed = g.RoundSeconds - 17.4
	g.Score = 2840
	g.Wave = 4
	g.BestCombo = 3
	g.CurrentCombo = 3
	g.MidWave, g.FinalWave, g.Last20 = true, true, true
	g.Player.Range = 4
	g.Player.MaxBombs = 3
	g.Player.Shield = true
	g.Player.X, g.Player.Y, g.Player.ToX, g.Player.ToY = 5, 7, 5, 7
	g.Grid[7][5] = Floor
	g.Grid[7][6] = Floor
	g.Grid[7][7] = Floor
	g.Bombs = []*Bomb{{ID: g.next(), X: 6, Y: 7, Fuse: .45, Range: 4}, {ID: g.next(), X: 10, Y: 9, Fuse: 1.25, Range: 3}}
	g.Player.ActiveBombs = 2
	g.Pickups = append(g.Pickups, &Pickup{X: 3, Y: 9, Type: PickupPartner, Age: .4}, &Pickup{X: 11, Y: 3, Type: PickupRange, Age: .8})
	g.Explosions = append(g.Explosions, &Explosion{ID: g.next(), BombID: -1, TTL: .36, Duration: .56, Cells: []Point{{9, 5}, {8, 5}, {10, 5}, {11, 5}, {9, 4}, {9, 3}, {9, 6}, {9, 7}}})
	g.Enemies = nil
	for _, p := range []Point{{12, 9}, {11, 5}, {3, 3}, {7, 9}} {
		g.Grid[p.Y][p.X] = Floor
		g.Enemies = append(g.Enemies, &Enemy{ID: g.next(), Mover: Mover{X: p.X, Y: p.Y, ToX: p.X, ToY: p.Y, MoveT: 1, MoveDuration: 1}, Alive: true, PulseOffset: float64(p.X)})
	}
	g.spawnBurst(9.5, 5.5, RGB(255, 117, 53), 22)
}

func sign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func newScoreID(now time.Time, nick string, score int) string {
	return HashString(now.UTC().Format(time.RFC3339Nano) + "|" + nick + "|" + itoa(score))
}
