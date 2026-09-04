package bomber

import "testing"

func maskCoverage(s *Surface) int {
	n := 0
	for _, p := range s.Pix {
		if Color(p).A() > 0 {
			n++
		}
	}
	return n
}

func TestEveryGlyphDrawsPixels(t *testing.T) {
	for r := range glyphs {
		if r == ' ' {
			continue
		}
		s := NewSurface(60, 60)
		s.DrawText(10, 10, 35, string(r), RGB(255, 255, 255))
		if maskCoverage(s) == 0 {
			t.Errorf("glif %q nie narysowal zadnego piksela", string(r))
		}
	}
}

func TestSpaceDrawsNothingButAdvances(t *testing.T) {
	s := NewSurface(60, 60)
	got := s.DrawText(5, 5, 35, " ", RGB(255, 255, 255))
	if maskCoverage(s) != 0 {
		t.Fatal("spacja narysowala piksele")
	}
	if got.W != TextWidth(" ", 35) {
		t.Fatalf("advance spacji %d != TextWidth %d", got.W, TextWidth(" ", 35))
	}
}

func TestTextWidthMatchesDrawnAdvance(t *testing.T) {
	for _, text := range []string{"A", "BOMBER", "RUSH 2026", "GRAJ"} {
		for _, h := range []int{7, 14, 33, 96, 190} {
			s := NewSurface(4000, 400)
			got := s.DrawText(0, 0, h, text, RGB(255, 255, 255))
			if got.W != TextWidth(text, h) {
				t.Fatalf("%q h=%d: DrawText.W=%d TextWidth=%d", text, h, got.W, TextWidth(text, h))
			}
			if got.H != 7*glyphUnit(h) {
				t.Fatalf("%q h=%d: wysokosc %d", text, h, got.H)
			}
		}
	}
}

// Zasięg narysowanych pikseli, a nie deklarowana metryka: (minX, minY, maxX, maxY) względem pióra.
func inkBounds(s *Surface, penX, penY int) (int, int, int, int, bool) {
	minX, minY, maxX, maxY := 1<<30, 1<<30, -1<<30, -1<<30
	found := false
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			if s.Pix[y*s.W+x].A() == 0 {
				continue
			}
			found = true
			minX, minY = min(minX, x-penX), min(minY, y-penY)
			maxX, maxY = max(maxX, x-penX), max(maxY, y-penY)
		}
	}
	return minX, minY, maxX, maxY, found
}

func TestDrawnGlyphInkStaysInsideAdvanceBox(t *testing.T) {
	const height = 70
	unit := glyphUnit(height)
	for r := range glyphs {
		if r == ' ' {
			continue
		}
		s := NewSurface(400, 400)
		s.DrawText(150, 150, height, string(r), RGB(255, 255, 255))
		minX, minY, maxX, maxY, found := inkBounds(s, 150, 150)
		if !found {
			t.Fatalf("glif %q nic nie narysowal", string(r))
		}
		// Jeden piksel zapasu to obwodka antyaliasingu, nie rozjazd metryki.
		if minX < -1 || maxX > int(glyphAdvance)*unit {
			t.Errorf("glif %q wychodzi w poziomie poza pole 0..%d: %d..%d", string(r), int(glyphAdvance)*unit, minX, maxX)
		}
		if minY < -1 || maxY > int(glyphRows)*unit {
			t.Errorf("glif %q wychodzi w pionie poza pole 0..%d: %d..%d", string(r), int(glyphRows)*unit, minY, maxY)
		}
	}
}

func TestDrawnTextInkMatchesReportedWidth(t *testing.T) {
	for _, text := range []string{"BOMBER RUSH", "NANOVO 2026", "GRAJ"} {
		for _, h := range []int{14, 35, 96, 190} {
			s := NewSurface(3000, 400)
			s.DrawText(60, 60, h, text, RGB(255, 255, 255))
			minX, minY, maxX, maxY, found := inkBounds(s, 60, 60)
			if !found {
				t.Fatalf("%q h=%d nic nie narysowalo", text, h)
			}
			w := TextWidth(text, h)
			if minX < -1 || maxX > w+1 {
				t.Fatalf("%q h=%d: piksele %d..%d poza szerokoscia %d", text, h, minX, maxX, w)
			}
			if minY < -1 || maxY > int(glyphRows)*glyphUnit(h) {
				t.Fatalf("%q h=%d: piksele %d..%d poza wysokoscia %d", text, h, minY, maxY, int(glyphRows)*glyphUnit(h))
			}
			// Metryka nie może być też wyraźnie za szeroka.
			if maxX < w-2*glyphUnit(h) {
				t.Fatalf("%q h=%d: metryka %d znacznie szersza niz piksele do %d", text, h, w, maxX)
			}
		}
	}
}

func TestCachedGlyphMatchesFreshRaster(t *testing.T) {
	glyphCacheMu.Lock()
	glyphCache = map[glyphKey]*glyphMask{}
	glyphCacheMu.Unlock()
	a := NewSurface(120, 120)
	a.DrawText(10, 10, 70, "SO", RGB(200, 30, 40))
	b := NewSurface(120, 120)
	b.DrawText(10, 10, 70, "SO", RGB(200, 30, 40))
	for i := range a.Pix {
		if a.Pix[i] != b.Pix[i] {
			t.Fatalf("cache zmienil wynik na pikselu %d", i)
		}
	}
}

func TestNormalizeTextMapsDiacriticsAndUnknown(t *testing.T) {
	if got := NormalizeText("Zażółć gęślą"); got != "ZAZOLC GESLA" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeText("aéb"); got != "A?B" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeText("li\nnia"); got != "LI\nNIA" {
		t.Fatalf("got %q", got)
	}
}

func TestFitTextHeightNeverExceedsWidth(t *testing.T) {
	h := FitTextHeight("BOMBER RUSH", 200, 300)
	if TextWidth("BOMBER RUSH", h) > 300 {
		t.Fatalf("h=%d szerokosc %d > 300", h, TextWidth("BOMBER RUSH", h))
	}
	if h > 200 {
		t.Fatalf("h=%d > maxHeight", h)
	}
	if FitTextHeight("BOMBER RUSH", 200, 1) != 7 {
		t.Fatal("brak dolnego ograniczenia 7")
	}
}

func TestDrawTextStopsAtNewline(t *testing.T) {
	s := NewSurface(400, 100)
	got := s.DrawText(0, 0, 35, "AB\nCD", RGB(255, 255, 255))
	if got.W != TextWidth("AB", 35) {
		t.Fatalf("W=%d, oczekiwano %d", got.W, TextWidth("AB", 35))
	}
}

func TestDrawParagraphWrapsAndClips(t *testing.T) {
	s := NewSurface(400, 400)
	used := s.DrawParagraph(Rect{0, 0, 200, 400}, 21, 6, "jeden dwa trzy cztery piec szesc", RGB(255, 255, 255))
	if used <= 0 {
		t.Fatal("brak linii")
	}
	if s.DrawParagraph(Rect{0, 0, 200, 400}, 21, 6, "   ", RGB(255, 255, 255)) != 0 {
		t.Fatal("pusty tekst powinien dac 0")
	}
}
