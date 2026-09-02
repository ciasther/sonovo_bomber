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
