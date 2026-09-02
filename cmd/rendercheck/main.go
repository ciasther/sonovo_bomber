package main

import (
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"bomberrush/internal/bomber"
)

type check struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

type report struct {
	GeneratedAt         string            `json:"generated_at"`
	Checks              []check           `json:"checks"`
	Screenshots         map[string]string `json:"screenshots"`
	RenderAverageMS     float64           `json:"render_average_ms"`
	RenderP95MS         float64           `json:"render_p95_ms"`
	BrandFingerprintOld string            `json:"brand_fingerprint_before"`
	BrandFingerprintNew string            `json:"brand_fingerprint_after"`
	PartnerLogoSHAOld   string            `json:"partner_logo_sha256_before"`
	PartnerLogoSHANew   string            `json:"partner_logo_sha256_after"`
	ExecutableSHAOld    string            `json:"exe_sha256_before"`
	ExecutableSHANew    string            `json:"exe_sha256_after"`
	ExecutableSizeBytes int64             `json:"exe_size_bytes"`
}

func main() {
	rootFlag := flag.String("root", ".", "katalog projektu")
	outFlag := flag.String("out", "evidence", "katalog dowodow")
	flag.Parse()
	root, err := filepath.Abs(*rootFlag)
	must(err)
	out, err := filepath.Abs(*outFlag)
	must(err)
	must(os.RemoveAll(out))
	must(os.MkdirAll(out, 0o755))

	temp, err := os.MkdirTemp("", "bomber-rush-rendercheck-")
	must(err)
	defer os.RemoveAll(temp)
	must(copyDir(filepath.Join(root, "assets"), filepath.Join(temp, "assets")))
	must(os.MkdirAll(filepath.Join(temp, "data"), 0o755))

	r := report{GeneratedAt: time.Now().Format(time.RFC3339), Screenshots: map[string]string{}}
	add := func(name string, passed bool, detail string) {
		r.Checks = append(r.Checks, check{Name: name, Passed: passed, Detail: detail})
		if !passed {
			fmt.Fprintf(os.Stderr, "FAIL: %s: %s\n", name, detail)
		}
	}

	landW, landH := bomber.LogicalSize(3840, 2160)
	portraitW, portraitH := bomber.LogicalSize(2160, 3840)
	add("logical-landscape", landW == 1440 && landH == 810, fmt.Sprintf("%dx%d", landW, landH))
	add("logical-portrait", portraitW == 810 && portraitH == 1440, fmt.Sprintf("%dx%d", portraitW, portraitH))

	app, err := bomber.NewApp(temp, landW, landH)
	must(err)
	add("nanovo-logo-embedded", app.Logo != nil && !app.LogoMissing, "logo jest dekodowane z zasobu wbudowanego")
	add("external-branding-loaded", app.Brand.PartnerLogo != nil && app.Brand.StartBackground != nil && app.Brand.SummaryAd != nil && app.Brand.QR != nil && app.Brand.ProductItem != nil, "wszystkie zasoby partnera wczytane z katalogu assets")
	add("layouts-in-bounds", validateLayouts(), "wszystkie kluczowe prostokaty mieszcza sie w pionie i poziomie")
	exePath := filepath.Join(root, "dist", "BomberRush.exe")
	peOK, peDetail := validateWindowsPE(exePath)
	add("windows-pe64-gui", peOK, peDetail)
	if st, statErr := os.Stat(exePath); statErr == nil {
		r.ExecutableSizeBytes = st.Size()
	}

	for i := 0; i < 4; i++ {
		app.Update(.1)
	}
	must(render4K(app, landW, landH, 3840, 2160, filepath.Join(out, "01-start-landscape-4k.png")))
	r.Screenshots["start_landscape_4k"] = "01-start-landscape-4k.png"

	// Pełny przepływ dotykowy: ekran startowy -> nick -> countdown -> gra -> ruch i bomba.
	app.Tap(landW/2, landH/2)
	add("touch-start", app.Screen == bomber.ScreenNick, fmt.Sprintf("screen=%d", app.Screen))
	layout := bomber.ComputeLayout(landW, landH, app.Screen)
	for _, letter := range []string{"N", "A", "N", "O", "V", "O"} {
		if !tapKey(app, layout, letter) {
			add("touch-keyboard", false, "brak klawisza "+letter)
			break
		}
	}
	add("touch-keyboard", app.Nick == "NANOVO", "nick="+app.Nick)
	for i := 0; i < 4; i++ {
		app.Update(.1)
	}
	must(render4K(app, landW, landH, 3840, 2160, filepath.Join(out, "02-nick-landscape-4k.png")))
	r.Screenshots["nick_landscape_4k"] = "02-nick-landscape-4k.png"
	_ = tapKey(app, layout, "PLAY")
	for i := 0; i < 4; i++ {
		app.Update(.1)
	}
	must(render4K(app, landW, landH, 3840, 2160, filepath.Join(out, "03-countdown-landscape-4k.png")))
	r.Screenshots["countdown_landscape_4k"] = "03-countdown-landscape-4k.png"
	add("touch-play", app.Screen == bomber.ScreenCountdown && app.Game != nil, fmt.Sprintf("screen=%d", app.Screen))
	for i := 0; i < 34; i++ {
		app.Update(.1)
	}
	add("countdown", app.Screen == bomber.ScreenPlay, fmt.Sprintf("screen=%d", app.Screen))
	beforeX := app.Game.Player.X
	app.Swipe(400, 0)
	app.Release()
	for i := 0; i < 3; i++ {
		app.Update(.06)
	}
	moved := app.Game.Player.X == beforeX+1
	app.Tap(landW/2, landH/2)
	add("touch-swipe", moved, fmt.Sprintf("x=%d->%d", beforeX, app.Game.Player.X))
	add("touch-bomb", len(app.Game.Bombs) == 1 && app.Game.Player.ActiveBombs == 1, fmt.Sprintf("bombs=%d", len(app.Game.Bombs)))

	app.PrepareVerificationData()
	must(render4K(app, landW, landH, 3840, 2160, filepath.Join(out, "04-game-landscape-4k.png")))
	r.Screenshots["game_landscape_4k"] = "04-game-landscape-4k.png"

	portrait, err := bomber.NewApp(temp, portraitW, portraitH)
	must(err)
	portrait.PrepareVerificationData()
	must(render4K(portrait, portraitW, portraitH, 2160, 3840, filepath.Join(out, "05-game-portrait-4k.png")))
	r.Screenshots["game_portrait_4k"] = "05-game-portrait-4k.png"

	portrait.PrepareVerificationSummary()
	must(render4K(portrait, portraitW, portraitH, 2160, 3840, filepath.Join(out, "06-summary-portrait-4k.png")))
	r.Screenshots["summary_portrait_4k"] = "06-summary-portrait-4k.png"

	must(portrait.SeedVerificationRanking())
	must(render4K(portrait, portraitW, portraitH, 2160, 3840, filepath.Join(out, "07-ranking-portrait-4k.png")))
	r.Screenshots["ranking_portrait_4k"] = "07-ranking-portrait-4k.png"
	add("ranking-persistence", len(portrait.Entries) == 10, fmt.Sprintf("entries=%d", len(portrait.Entries)))
	portrait.Screen = bomber.ScreenAdminPIN
	portrait.StateAge = 1
	portrait.AdminPINInput = "24"
	must(render4K(portrait, portraitW, portraitH, 2160, 3840, filepath.Join(out, "08-admin-pin-portrait-4k.png")))
	r.Screenshots["admin_pin_portrait_4k"] = "08-admin-pin-portrait-4k.png"
	history, err := bomber.NewApp(temp, landW, landH)
	must(err)
	history.Screen = bomber.ScreenHistory
	history.StateAge = 1
	must(render4K(history, landW, landH, 3840, 2160, filepath.Join(out, "09-history-landscape-4k.png")))
	r.Screenshots["history_landscape_4k"] = "09-history-landscape-4k.png"

	// Podmiana grafiki partnera w czasie działania, bez budowania binarium.
	swapApp, err := bomber.NewApp(temp, landW, landH)
	must(err)
	for i := 0; i < 4; i++ {
		swapApp.Update(.1)
	}
	beforeFingerprint := swapApp.Brand.Fingerprint
	beforeLogo := filepath.Join(temp, "assets", "partner-logo.png")
	altLogo := filepath.Join(root, "testdata", "verification-alt-partner-logo.png")
	r.PartnerLogoSHAOld = fileSHA(beforeLogo)
	r.ExecutableSHAOld = fileSHA(exePath)
	must(render4K(swapApp, landW, landH, 3840, 2160, filepath.Join(out, "10-brand-before-4k.png")))
	must(copyFile(altLogo, beforeLogo))
	now := time.Now().Add(2 * time.Second)
	must(os.Chtimes(beforeLogo, now, now))
	for i := 0; i < 12; i++ {
		swapApp.Update(.1)
	}
	r.BrandFingerprintOld = beforeFingerprint
	r.BrandFingerprintNew = swapApp.Brand.Fingerprint
	r.PartnerLogoSHANew = fileSHA(beforeLogo)
	r.ExecutableSHANew = fileSHA(exePath)
	must(render4K(swapApp, landW, landH, 3840, 2160, filepath.Join(out, "11-brand-after-4k.png")))
	r.Screenshots["brand_before_4k"] = "10-brand-before-4k.png"
	r.Screenshots["brand_after_4k"] = "11-brand-after-4k.png"
	add("branding-hot-reload", r.BrandFingerprintOld != r.BrandFingerprintNew && r.PartnerLogoSHAOld != r.PartnerLogoSHANew, "fingerprint oraz grafika zmienily sie bez restartu")
	add("exe-unchanged-during-brand-swap", r.ExecutableSHAOld != "" && r.ExecutableSHAOld == r.ExecutableSHANew, "sha256="+r.ExecutableSHAOld)

	// Pomiar pełnego renderowania gry w rozdzielczości logicznej 1440x810.
	benchApp, err := bomber.NewApp(temp, landW, landH)
	must(err)
	benchApp.PrepareVerificationData()
	benchRenderer := bomber.NewRenderer(landW, landH)
	for i := 0; i < 3; i++ {
		benchApp.Update(1.0 / 60.0)
		benchRenderer.Render(benchApp)
	}
	var durations []float64
	for i := 0; i < 30; i++ {
		started := time.Now()
		benchApp.Update(1.0 / 60.0)
		benchRenderer.Render(benchApp)
		durations = append(durations, float64(time.Since(started).Microseconds())/1000)
	}
	sort.Float64s(durations)
	for _, d := range durations {
		r.RenderAverageMS += d
	}
	r.RenderAverageMS /= float64(len(durations))
	r.RenderP95MS = durations[int(float64(len(durations)-1)*.95)]
	add("renderer-performance", r.RenderAverageMS < 16.7 && r.RenderP95MS < 22.0, fmt.Sprintf("avg=%.2fms p95=%.2fms", r.RenderAverageMS, r.RenderP95MS))

	payload, err := json.MarshalIndent(r, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(out, "verification.json"), append(payload, '\n'), 0o644))

	failed := 0
	for _, c := range r.Checks {
		if !c.Passed {
			failed++
		}
	}
	fmt.Printf("checks=%d failed=%d render_avg=%.2fms render_p95=%.2fms\n", len(r.Checks), failed, r.RenderAverageMS, r.RenderP95MS)
	if failed > 0 {
		os.Exit(1)
	}
}

func validateWindowsPE(path string) (bool, string) {
	f, err := pe.Open(path)
	if err != nil {
		return false, err.Error()
	}
	defer f.Close()
	if f.FileHeader.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
		return false, fmt.Sprintf("machine=0x%x", f.FileHeader.Machine)
	}
	header, ok := f.OptionalHeader.(*pe.OptionalHeader64)
	if !ok {
		return false, "brak naglowka PE32+"
	}
	if header.Subsystem != 2 {
		return false, fmt.Sprintf("subsystem=%d", header.Subsystem)
	}
	return true, "PE32+ x86-64, Windows GUI"
}

func validateLayouts() bool {
	for _, size := range [][2]int{{1440, 810}, {810, 1440}} {
		for screen := bomber.ScreenAttract; screen <= bomber.ScreenHistory; screen++ {
			l := bomber.ComputeLayout(size[0], size[1], screen)
			for _, r := range []bomber.Rect{l.Header, l.NanoLogo, l.Content, l.Board, l.HUD, l.Sponsor, l.Table, l.PrimaryButton, l.SecondaryButton, l.BackButton, l.AdminButton, l.TabOne, l.TabTwo, l.PagePrev, l.PageNext} {
				if r.W == 0 && r.H == 0 {
					continue
				}
				if r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0 || r.Right() > size[0] || r.Bottom() > size[1] {
					return false
				}
			}
			for _, key := range append(append([]bomber.KeyRegion(nil), l.KeyboardKeys...), l.PinKeys...) {
				if key.Rect.W <= 0 || key.Rect.H <= 0 || key.Rect.X < 0 || key.Rect.Y < 0 || key.Rect.Right() > size[0] || key.Rect.Bottom() > size[1] {
					return false
				}
			}
			if screen == bomber.ScreenRanking && (overlaps(l.TabOne, l.TabTwo) || overlaps(l.TabOne, l.AdminButton) || overlaps(l.TabTwo, l.AdminButton)) {
				return false
			}
		}
	}
	return true
}

func overlaps(a, b bomber.Rect) bool {
	return a.X < b.Right() && a.Right() > b.X && a.Y < b.Bottom() && a.Bottom() > b.Y
}

func tapKey(app *bomber.App, layout bomber.Layout, value string) bool {
	for _, key := range layout.KeyboardKeys {
		if key.Value == value {
			app.Tap(key.Rect.X+key.Rect.W/2, key.Rect.Y+key.Rect.H/2)
			return true
		}
	}
	return false
}

func render4K(app *bomber.App, logicalW, logicalH, outW, outH int, path string) error {
	r := bomber.NewRenderer(logicalW, logicalH)
	s := r.Render(app)
	return bomber.SavePNG(path, bomber.ScaleImageNearest(s, outW, outH))
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := copyDir(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func fileSHA(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
