package bomber

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed coreassets/Logo-nanoVo.png
var nanoVoLogoPNG []byte

type Screen int

const (
	ScreenAttract Screen = iota
	ScreenNick
	ScreenCountdown
	ScreenPlay
	ScreenSummary
	ScreenRanking
	ScreenAdminPIN
	ScreenHistory
)

type RankingMode int

const (
	RankingEvent RankingMode = iota
	RankingToday
)

type TouchPulse struct {
	X, Y float64
	Life float64
	Max  float64
	Good bool
}

type SoundCue uint8

const (
	SoundBombPlaced SoundCue = iota + 1
	SoundBombBlocked
	SoundExplosion
	SoundPickup
	SoundDamage
	SoundTimeWarning
	SoundRoundEnd
)

type App struct {
	Root             string
	Brand            *Branding
	Logo             *Sprite
	LogoMissing      bool
	Store            *ScoreStore
	Screen           Screen
	StateAge         float64
	TotalTime        float64
	ViewW, ViewH     int
	Nick             string
	Game             *Game
	LastEntry        ScoreEntry
	LastPosition     int
	Entries          []ScoreEntry
	RankingMode      RankingMode
	AdminPINInput    string
	AdminPINError    bool
	HistoryPage      int
	Pulses           []TouchPulse
	Sounds           []SoundCue
	TutorialVisible  bool
	tutorialShown    bool
	timeWarningSent  bool
	lastBrandCheck   float64
	lastFingerprint  string
	LastBrandError   string
	LastStorageError string
	VerificationMode bool
	Seed             int64
}

func NewApp(root string, width, height int) (*App, error) {
	brand, err := LoadBranding(root)
	if brand == nil {
		return nil, err
	}
	logo, logoErr := decodeEmbeddedLogo()
	app := &App{
		Root:            root,
		Brand:           brand,
		Logo:            logo,
		LogoMissing:     logoErr != nil,
		Store:           NewScoreStore(root),
		Screen:          ScreenAttract,
		ViewW:           width,
		ViewH:           height,
		Seed:            time.Now().UnixNano(),
		lastFingerprint: brand.Fingerprint,
	}
	if err != nil {
		app.LastBrandError = err.Error()
	}
	app.reloadScores()
	if storageErr := app.Store.CheckWritable(); storageErr != nil {
		app.LastStorageError = storageErr.Error()
	}
	return app, nil
}

func decodeEmbeddedLogo() (*Sprite, error) {
	img, err := png.Decode(bytes.NewReader(nanoVoLogoPNG))
	if err != nil {
		return nil, err
	}
	return SpriteFromImage(img), nil
}

func (a *App) SetViewport(w, h int) {
	if w > 0 {
		a.ViewW = w
	}
	if h > 0 {
		a.ViewH = h
	}
}

func (a *App) Update(dt float64) {
	if dt <= 0 {
		return
	}
	if dt > .1 {
		dt = .1
	}
	a.TotalTime += dt
	a.StateAge += dt
	for i := range a.Pulses {
		a.Pulses[i].Life -= dt
	}
	pulses := a.Pulses[:0]
	for _, p := range a.Pulses {
		if p.Life > 0 {
			pulses = append(pulses, p)
		}
	}
	a.Pulses = pulses

	if a.TotalTime-a.lastBrandCheck >= 1.0 {
		a.lastBrandCheck = a.TotalTime
		a.reloadBranding()
	}

	switch a.Screen {
	case ScreenCountdown:
		if a.StateAge >= 3.25 {
			a.changeScreen(ScreenPlay)
		}
	case ScreenPlay:
		if a.Game != nil {
			remaining := a.Game.Remaining()
			lives := a.Game.Player.Lives
			pickups := len(a.Game.Pickups)
			explosions := len(a.Game.Explosions)
			a.Game.Update(dt)
			if !a.timeWarningSent && remaining > 10 && a.Game.Remaining() <= 10 {
				a.Sounds = append(a.Sounds, SoundTimeWarning)
				a.timeWarningSent = true
			}
			if a.Game.Player.Lives < lives {
				a.Sounds = append(a.Sounds, SoundDamage)
			}
			if len(a.Game.Pickups) < pickups {
				a.Sounds = append(a.Sounds, SoundPickup)
			}
			if len(a.Game.Explosions) > explosions {
				a.Sounds = append(a.Sounds, SoundExplosion)
			}
			if a.Game.Finished {
				a.Sounds = append(a.Sounds, SoundRoundEnd)
				a.completeRun()
			}
		}
	case ScreenRanking:
		if a.StateAge >= float64(a.Brand.Config.IdleReturnSeconds) {
			a.resetToAttract()
		}
	case ScreenAdminPIN:
		if a.StateAge >= 30 {
			a.changeScreen(ScreenRanking)
		}
	case ScreenHistory:
		if a.StateAge >= 45 {
			a.resetToAttract()
		}
	}
}

func (a *App) Tap(x, y int) {
	a.Pulses = append(a.Pulses, TouchPulse{X: float64(x), Y: float64(y), Life: .45, Max: .45, Good: true})
	a.StateAge = maxFloat(0, a.StateAge)
	layout := ComputeLayout(a.ViewW, a.ViewH, a.Screen)
	switch a.Screen {
	case ScreenAttract:
		a.Nick = ""
		a.changeScreen(ScreenNick)
	case ScreenNick:
		a.handleNickTap(layout, x, y)
	case ScreenCountdown:
		return
	case ScreenPlay:
		if a.Game != nil {
			good := a.Game.PlaceBomb()
			a.Pulses[len(a.Pulses)-1].Good = good
			if good {
				a.Sounds = append(a.Sounds, SoundBombPlaced)
			} else {
				a.Sounds = append(a.Sounds, SoundBombBlocked)
			}
		}
	case ScreenSummary:
		if layout.PrimaryButton.Contains(x, y) {
			a.changeScreen(ScreenRanking)
		} else if layout.SecondaryButton.Contains(x, y) {
			a.Nick = ""
			a.changeScreen(ScreenNick)
		}
	case ScreenRanking:
		switch {
		case layout.TabOne.Contains(x, y):
			a.RankingMode = RankingEvent
			a.StateAge = 0
		case layout.TabTwo.Contains(x, y):
			a.RankingMode = RankingToday
			a.StateAge = 0
		case layout.AdminButton.Contains(x, y):
			a.AdminPINInput = ""
			a.AdminPINError = false
			a.changeScreen(ScreenAdminPIN)
		case layout.PrimaryButton.Contains(x, y):
			a.Nick = ""
			a.changeScreen(ScreenNick)
		}
	case ScreenAdminPIN:
		a.handleAdminPINTap(layout, x, y)
	case ScreenHistory:
		switch {
		case layout.BackButton.Contains(x, y):
			a.changeScreen(ScreenRanking)
		case layout.PagePrev.Contains(x, y):
			a.HistoryPage = max(0, a.HistoryPage-1)
			a.StateAge = 0
		case layout.PageNext.Contains(x, y):
			maxPage := max(0, (len(a.Entries)-1)/12)
			a.HistoryPage = min(maxPage, a.HistoryPage+1)
			a.StateAge = 0
		}
	}
}

func (a *App) Swipe(dx, dy int) {
	if a.Screen != ScreenPlay || a.Game == nil {
		return
	}
	if abs(dx) > abs(dy) {
		if dx < 0 {
			a.Game.Swipe(-1, 0)
		} else {
			a.Game.Swipe(1, 0)
		}
	} else {
		if dy < 0 {
			a.Game.Swipe(0, -1)
		} else {
			a.Game.Swipe(0, 1)
		}
	}
}

func (a *App) Release() {
	if a.Game != nil {
		a.Game.Release()
	}
}

func (a *App) handleNickTap(layout Layout, x, y int) {
	for _, key := range layout.KeyboardKeys {
		if !containsTouchKey(key.Rect, x, y) {
			continue
		}
		switch key.Value {
		case "BACK":
			if len(a.Nick) > 0 {
				runes := []rune(a.Nick)
				a.Nick = string(runes[:len(runes)-1])
			}
		case "PLAY":
			if strings.TrimSpace(a.Nick) != "" {
				a.startRun()
			} else {
				a.Pulses[len(a.Pulses)-1].Good = false
			}
		default:
			if len([]rune(a.Nick)) < 10 {
				a.Nick += key.Value
			}
		}
		return
	}
	if layout.BackButton.Contains(x, y) {
		a.resetToAttract()
	}
}

func (a *App) handleAdminPINTap(layout Layout, x, y int) {
	for _, key := range layout.PinKeys {
		if !containsTouchKey(key.Rect, x, y) {
			continue
		}
		switch key.Value {
		case "BACK":
			if len(a.AdminPINInput) > 0 {
				a.AdminPINInput = a.AdminPINInput[:len(a.AdminPINInput)-1]
			}
		case "OK":
			if a.AdminPINInput == a.Brand.Config.AdminPIN {
				a.HistoryPage = 0
				a.reloadScores()
				a.changeScreen(ScreenHistory)
			} else {
				a.AdminPINError = true
				a.AdminPINInput = ""
				a.StateAge = 0
			}
		default:
			if len(a.AdminPINInput) < 12 {
				a.AdminPINInput += key.Value
			}
		}
		return
	}
	if layout.BackButton.Contains(x, y) {
		a.changeScreen(ScreenRanking)
	}
}

func containsTouchKey(rect Rect, x, y int) bool {
	padding := max(4, min(rect.W, rect.H)/14)
	return rect.Inset(-padding).Contains(x, y)
}

func (a *App) startRun() {
	a.Game = NewGame(a.Brand.Config.RoundSeconds, a.Seed+time.Now().UnixNano())
	a.TutorialVisible = !a.tutorialShown
	a.tutorialShown = true
	a.timeWarningSent = false
	a.changeScreen(ScreenCountdown)
}

func (a *App) DrainSounds() []SoundCue {
	sounds := a.Sounds
	a.Sounds = nil
	return sounds
}

func (a *App) completeRun() {
	if a.Game == nil {
		return
	}
	now := time.Now()
	result := a.Game.Result
	entry := ScoreEntry{
		ID:           newScoreID(now, a.Nick, result.Score),
		EventID:      a.Brand.Config.EventID,
		Nick:         normalizeNick(a.Nick),
		Score:        result.Score,
		SurvivalMS:   result.SurvivalMS,
		BestCombo:    result.BestCombo,
		Waves:        result.Waves,
		CreatedAt:    now,
		CompletedRun: result.CompletedRun,
	}
	if err := a.Store.Append(entry); err != nil {
		a.LastStorageError = err.Error()
	} else {
		a.LastStorageError = ""
	}
	a.LastEntry = entry
	a.reloadScores()
	ranked := RankEntries(a.Entries, a.Brand.Config.EventID, nil)
	a.LastPosition = PositionOf(ranked, entry.ID)
	a.changeScreen(ScreenSummary)
}

func normalizeNick(nick string) string {
	nick = strings.TrimSpace(NormalizeText(nick))
	if nick == "" {
		return "GRACZ"
	}
	runes := []rune(nick)
	if len(runes) > 10 {
		runes = runes[:10]
	}
	return string(runes)
}

func (a *App) changeScreen(screen Screen) {
	a.Screen = screen
	a.StateAge = 0
}

func (a *App) resetToAttract() {
	a.Game = nil
	a.AdminPINInput = ""
	a.AdminPINError = false
	a.HistoryPage = 0
	a.TutorialVisible = false
	a.changeScreen(ScreenAttract)
}

func (a *App) reloadScores() {
	entries, err := a.Store.Load()
	if err != nil {
		a.LastStorageError = err.Error()
		return
	}
	a.Entries = entries
}

func (a *App) reloadBranding() {
	fingerprint, err := BrandingFingerprint(a.Root)
	if err != nil {
		a.LastBrandError = err.Error()
		return
	}
	if fingerprint == a.lastFingerprint {
		return
	}
	brand, loadErr := LoadBranding(a.Root)
	if brand != nil {
		a.Brand = brand
		a.lastFingerprint = brand.Fingerprint
	}
	if loadErr != nil {
		a.LastBrandError = loadErr.Error()
	} else {
		a.LastBrandError = ""
	}
}

func (a *App) RankingEntries(limit int) []ScoreEntry {
	var ranked []ScoreEntry
	if a.RankingMode == RankingToday {
		now := time.Now()
		ranked = RankEntries(a.Entries, a.Brand.Config.EventID, &now)
	} else {
		ranked = RankEntries(a.Entries, a.Brand.Config.EventID, nil)
	}
	if limit > 0 && len(ranked) > limit {
		ranked = ranked[:limit]
	}
	return ranked
}

func (a *App) HistoryEntries() []ScoreEntry {
	all := append([]ScoreEntry(nil), a.Entries...)
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	return all
}

func (a *App) PrepareVerificationData() {
	a.VerificationMode = true
	a.Nick = "NANOVO"
	a.Game = NewGame(a.Brand.Config.RoundSeconds, 20260901)
	a.Game.PrepareShowcase()
	a.Screen = ScreenPlay
	a.StateAge = 1.4
}

func (a *App) PrepareVerificationSummary() {
	a.LastEntry = ScoreEntry{ID: "verification", EventID: a.Brand.Config.EventID, Nick: "NANOVO", Score: 5240, SurvivalMS: 75000, BestCombo: 4, Waves: 6, CreatedAt: time.Date(2026, 9, 1, 10, 45, 0, 0, time.Local), CompletedRun: true}
	a.LastPosition = 1
	a.Screen = ScreenSummary
	a.StateAge = 2
}

func (a *App) SeedVerificationRanking() error {
	if err := os.MkdirAll(filepath.Dir(a.Store.Path), 0o755); err != nil {
		return err
	}
	_ = os.Remove(a.Store.Path)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local)
	rows := []struct {
		nick                          string
		score, survival, combo, waves int
	}{
		{"MILA", 6120, 75000, 4, 7}, {"NANOVO", 5240, 75000, 4, 6}, {"KUBA", 4790, 72400, 3, 5}, {"OLA", 4310, 68900, 3, 5}, {"MAX", 3990, 64000, 2, 4}, {"IZA", 3540, 60800, 3, 4}, {"LEO", 3210, 55200, 2, 4}, {"ADA", 2880, 51000, 2, 3}, {"TOM", 2470, 47000, 2, 3}, {"EWA", 2130, 42600, 1, 3},
	}
	for i, r := range rows {
		e := ScoreEntry{ID: fmt.Sprintf("v-%d", i), EventID: a.Brand.Config.EventID, Nick: r.nick, Score: r.score, SurvivalMS: int64(r.survival), BestCombo: r.combo, Waves: r.waves, CreatedAt: now.Add(-time.Duration(i) * 7 * time.Minute), CompletedRun: r.survival == 75000}
		if err := a.Store.Append(e); err != nil {
			return err
		}
	}
	a.reloadScores()
	a.Screen = ScreenRanking
	a.StateAge = 2
	return nil
}

func (a *App) DiagnosticText() string {
	parts := []string{}
	if a.LastBrandError != "" {
		parts = append(parts, "BRANDING: "+a.LastBrandError)
	}
	if a.LastStorageError != "" {
		parts = append(parts, "DANE: "+a.LastStorageError)
	}
	return strings.Join(parts, " | ")
}
