package bomber

import "testing"

func TestFontSmallGlyphsAreVisible(t *testing.T) {
	s := NewSurface(120, 40)
	s.Clear(RGB(0, 0, 0))
	s.DrawText(2, 2, 18, "TEST", RGB(255, 255, 255))
	visible := 0
	for _, p := range s.Pix {
		if p.R() > 0 || p.G() > 0 || p.B() > 0 {
			visible++
		}
	}
	if visible == 0 {
		t.Fatal("maly tekst nie zostal narysowany")
	}
}

func TestResponsiveLayoutsStayInsideViewport(t *testing.T) {
	for _, size := range [][2]int{{1440, 810}, {810, 1440}} {
		for screen := ScreenAttract; screen <= ScreenHistory; screen++ {
			l := ComputeLayout(size[0], size[1], screen)
			for _, r := range []Rect{l.Header, l.NanoLogo, l.Content, l.Board, l.HUD, l.Sponsor, l.Table, l.PrimaryButton, l.SecondaryButton, l.BackButton, l.AdminButton, l.TabOne, l.TabTwo, l.PagePrev, l.PageNext, l.Hint} {
				if r.W == 0 && r.H == 0 {
					continue
				}
				if r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0 || r.Right() > size[0] || r.Bottom() > size[1] {
					t.Fatalf("screen=%d size=%v rect=%+v", screen, size, r)
				}
			}
		}
	}
}

func TestPlayLayoutKeepsBoardClearOfSecondaryContent(t *testing.T) {
	for _, size := range [][2]int{{1440, 810}, {810, 1440}} {
		l := ComputeLayout(size[0], size[1], ScreenPlay)
		if l.Header.H != 0 {
			t.Fatalf("size=%v naglowek zajmuje miejsce podczas gry: %+v", size, l.Header)
		}
		if l.Sponsor.W != 0 || l.Sponsor.H != 0 {
			t.Fatalf("size=%v panel sponsora zajmuje miejsce podczas gry: %+v", size, l.Sponsor)
		}
		if l.Board.W*l.Board.H <= l.HUD.W*l.HUD.H*3 {
			t.Fatalf("size=%v plansza jest za mala wzgledem HUD: board=%+v hud=%+v", size, l.Board, l.HUD)
		}
	}
}

func TestNickKeyboardUsesCompactResponsiveBounds(t *testing.T) {
	for _, size := range [][2]int{{1440, 810}, {810, 1440}} {
		l := ComputeLayout(size[0], size[1], ScreenNick)
		maxW, maxH := 1080, 360
		if !l.Landscape {
			maxW, maxH = 720, 560
		}
		if l.Keyboard.W > maxW || l.Keyboard.H > maxH {
			t.Fatalf("size=%v klawiatura jest za duza: %+v", size, l.Keyboard)
		}
		for _, key := range l.KeyboardKeys {
			if key.Rect.W <= 0 || key.Rect.H <= 0 || key.Rect.X < l.Keyboard.X || key.Rect.Y < l.Keyboard.Y || key.Rect.Right() > l.Keyboard.Right() || key.Rect.Bottom() > l.Keyboard.Bottom() {
				t.Fatalf("size=%v klawisz poza klawiatura: %+v area=%+v", size, key.Rect, l.Keyboard)
			}
		}
	}
}

func TestKeyboardAcceptsTouchNearKeyEdge(t *testing.T) {
	a := &App{Screen: ScreenNick, ViewW: 1440, ViewH: 810}
	l := ComputeLayout(a.ViewW, a.ViewH, a.Screen)
	key := l.KeyboardKeys[0]
	a.Tap(key.Rect.Right()+1, key.Rect.Y+key.Rect.H/2)
	if a.Nick != key.Value {
		t.Fatalf("nick=%q, oczekiwano %q po dotknieciu przy krawedzi", a.Nick, key.Value)
	}
}

func TestCachedRenderMatchesFreshRender(t *testing.T) {
	a, err := NewApp("../..", 1440, 810)
	if err != nil {
		t.Fatal(err)
	}
	a.PrepareVerificationData()
	warm := NewRenderer(1440, 810)
	for i := 0; i < 3; i++ {
		a.Update(.016)
		warm.Render(a)
	}
	a.Update(.12)
	a.Game.Grid[3][3] = Floor
	fresh := NewRenderer(1440, 810).Render(a)
	cached := warm.Render(a)
	for i := range fresh.Pix {
		if fresh.Pix[i] != cached.Pix[i] {
			t.Fatalf("piksel %d rozni sie: %x vs %x", i, fresh.Pix[i], cached.Pix[i])
		}
	}
}

func TestLogoCardsFitInsideHeader(t *testing.T) {
	sizes := [][2]int{{1440, 810}, {810, 1440}, {1920, 1080}, {1080, 1920}, {1280, 800}, {800, 1280}}
	screens := []Screen{ScreenAttract, ScreenNick, ScreenRanking, ScreenSummary, ScreenAdminPIN, ScreenHistory}
	logos := []*Sprite{nil, {W: 1200, H: 300}, {W: 400, H: 400}, {W: 300, H: 1200}}
	for _, sz := range sizes {
		for _, sc := range screens {
			l := ComputeLayout(sz[0], sz[1], sc)
			for _, sp := range logos {
				for _, card := range []Rect{logoCard(l.NanoLogo, sp, false), logoCard(l.PartnerLogo, sp, true)} {
					if card.W < 1 || card.H < 1 {
						t.Fatalf("%v %d: pusta karta %+v", sz, sc, card)
					}
					if card.X < 0 || card.Y < 0 || card.Right() > sz[0] || card.Bottom() > l.Header.Bottom() {
						t.Fatalf("%v %d: karta %+v poza naglowkiem %+v", sz, sc, card, l.Header)
					}
				}
			}
			if l.NanoLogo.Right() >= l.PartnerLogo.X {
				t.Fatalf("%v %d: logo nachodza na siebie", sz, sc)
			}
		}
	}
}

func TestKeyboardKeysStayInsideAreaAndAreTouchSized(t *testing.T) {
	for _, sz := range [][2]int{{1440, 810}, {810, 1440}, {1920, 1080}} {
		l := ComputeLayout(sz[0], sz[1], ScreenNick)
		if len(l.KeyboardKeys) == 0 {
			t.Fatalf("%v: brak klawiszy", sz)
		}
		for _, k := range l.KeyboardKeys {
			if k.Rect.W < 24 || k.Rect.H < 24 {
				t.Fatalf("%v: klawisz za maly: %+v", sz, k.Rect)
			}
			if k.Rect.X < l.Keyboard.X || k.Rect.Right() > l.Keyboard.Right() || k.Rect.Y < l.Keyboard.Y || k.Rect.Bottom() > l.Keyboard.Bottom() {
				t.Fatalf("%v: klawisz %q poza obszarem: %+v w %+v", sz, k.Value, k.Rect, l.Keyboard)
			}
		}
	}
}

func TestPressedKeyIsClearedAfterShortTime(t *testing.T) {
	a := &App{Screen: ScreenNick, ViewW: 1440, ViewH: 810, Brand: &Branding{}, PressedKey: "A"}
	a.Update(.05)
	if a.PressedKey != "A" {
		t.Fatal("podswietlenie znika za szybko")
	}
	a.Update(.1)
	a.Update(.1)
	if a.PressedKey != "" {
		t.Fatal("podswietlenie nie znika")
	}
}

func TestTouchInKeyGapPicksNearestKey(t *testing.T) {
	l := ComputeLayout(1440, 810, ScreenNick)
	var q, w KeyRegion
	for _, k := range l.KeyboardKeys {
		if k.Value == "Q" {
			q = k
		}
		if k.Value == "W" {
			w = k
		}
	}
	if q.Value == "" || w.Value == "" {
		t.Fatal("brak klawiszy Q i W")
	}
	y := q.Rect.Y + q.Rect.H/2
	for _, tc := range []struct {
		x    int
		want string
	}{
		{q.Rect.Right() + 1, "Q"},
		{w.Rect.X - 1, "W"},
		{q.Rect.X + q.Rect.W/2, "Q"},
		{w.Rect.X + w.Rect.W/2, "W"},
	} {
		got, ok := findTouchKey(l.KeyboardKeys, tc.x, y)
		if !ok || got.Value != tc.want {
			t.Fatalf("x=%d: %q (ok=%v), oczekiwano %q", tc.x, got.Value, ok, tc.want)
		}
	}
	if _, ok := findTouchKey(l.KeyboardKeys, 5, 5); ok {
		t.Fatal("dotyk daleko od klawiatury trafil w klawisz")
	}
}
