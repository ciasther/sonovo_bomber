package bomber

import "testing"

func filled(s *Surface) int {
	n := 0
	for _, p := range s.Pix {
		if p.A() > 0 {
			n++
		}
	}
	return n
}

func TestFillCircleHasSoftEdge(t *testing.T) {
	s := NewSurface(41, 41)
	white := RGB(255, 255, 255)
	s.FillCircle(20, 20, 15, white)
	if got := s.Pix[20*41+20]; got != white {
		t.Fatalf("srodek nie jest pelny: %x", got)
	}
	edge := s.Pix[20*41+35]
	if edge.A() == 0 || edge.A() == 255 {
		t.Fatalf("krawedz nie jest wygladzona: alpha=%d", edge.A())
	}
	if outside := s.Pix[20*41+39]; outside.A() != 0 {
		t.Fatalf("piksel poza okregiem zostal zamalowany: %x", outside)
	}
	for dy := -15; dy <= 15; dy++ {
		l := s.Pix[(20+dy)*41+20-15]
		r := s.Pix[(20+dy)*41+20+15]
		if l != r {
			t.Fatalf("okrag nie jest symetryczny w wierszu %d: %x vs %x", dy, l, r)
		}
	}
}

func TestFillCircleIgnoresNonPositiveRadiusAndClips(t *testing.T) {
	s := NewSurface(20, 20)
	s.FillCircle(10, 10, 0, RGB(255, 0, 0))
	s.FillCircle(10, 10, -3, RGB(255, 0, 0))
	if n := filled(s); n != 0 {
		t.Fatalf("promien <= 0 narysowal %d pikseli", n)
	}
	s.FillCircle(-40, -40, 8, RGB(255, 0, 0))
	s.FillCircle(200, 200, 8, RGB(255, 0, 0))
	if n := filled(s); n != 0 {
		t.Fatalf("okrag poza powierzchnia narysowal %d pikseli", n)
	}
	s.FillCircle(0, 0, 6, RGB(255, 0, 0))
	if s.Pix[0].A() == 0 {
		t.Fatal("okrag przyciety do rogu nie zostal narysowany")
	}
}

func TestRingWiderThanRadiusFillsCircle(t *testing.T) {
	c := RGB(0, 255, 0)
	ring := NewSurface(31, 31)
	ring.Ring(15, 15, 10, 10, c)
	full := NewSurface(31, 31)
	full.FillCircle(15, 15, 10, c)
	for i := range ring.Pix {
		if ring.Pix[i] != full.Pix[i] {
			t.Fatalf("piksel %d: ring=%x circle=%x", i, ring.Pix[i], full.Pix[i])
		}
	}
}

func TestRingLeavesHoleAndBlendsCenterOnce(t *testing.T) {
	s := NewSurface(41, 41)
	s.Ring(20, 20, 15, 3, RGB(255, 255, 255))
	if got := s.Pix[20*41+20]; got.A() != 0 {
		t.Fatalf("srodek pierscienia zostal zamalowany: %x", got)
	}
	if got := s.Pix[20*41+20-14]; got.A() == 0 {
		t.Fatal("pasmo pierscienia nie zostalo narysowane")
	}
	for dy := -20; dy <= 20; dy++ {
		y := 20 + dy
		if y < 0 || y > 40 {
			continue
		}
		l := s.Pix[y*41+20-15]
		r := s.Pix[y*41+20+15]
		if l != r {
			t.Fatalf("pierscien nie jest symetryczny w wierszu %d: %x vs %x", dy, l, r)
		}
	}
}

func TestFillRoundRectMatchesFillRectWithoutRadius(t *testing.T) {
	c := RGB(10, 20, 30)
	a := NewSurface(20, 20)
	a.FillRoundRect(Rect{2, 3, 10, 8}, 0, c)
	b := NewSurface(20, 20)
	b.FillRect(Rect{2, 3, 10, 8}, c)
	for i := range a.Pix {
		if a.Pix[i] != b.Pix[i] {
			t.Fatalf("piksel %d: round=%x rect=%x", i, a.Pix[i], b.Pix[i])
		}
	}
}

func TestFillRoundRectSoftensCornersAndStaysInBounds(t *testing.T) {
	s := NewSurface(40, 40)
	r := Rect{5, 5, 30, 30}
	s.FillRoundRect(r, 10, RGB(255, 255, 255))
	if got := s.Pix[20*40+20]; got.A() != 255 {
		t.Fatalf("wnetrze nie jest pelne: %x", got)
	}
	if got := s.Pix[5*40+5]; got.A() == 255 {
		t.Fatalf("rog nie zostal zaokraglony: %x", got)
	}
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			if r.Contains(x, y) {
				continue
			}
			if s.Pix[y*40+x].A() != 0 {
				t.Fatalf("piksel %d,%d poza prostokatem: %x", x, y, s.Pix[y*40+x])
			}
		}
	}
}

func TestFillRoundRectClipsAndHandlesTinyRects(t *testing.T) {
	s := NewSurface(20, 20)
	s.FillRoundRect(Rect{-10, -10, 8, 8}, 4, RGB(255, 0, 0))
	s.FillRoundRect(Rect{0, 0, 0, 5}, 3, RGB(255, 0, 0))
	s.FillRoundRect(Rect{0, 0, 5, -2}, 3, RGB(255, 0, 0))
	if n := filled(s); n != 0 {
		t.Fatalf("puste lub zewnetrzne prostokaty narysowaly %d pikseli", n)
	}
	s.FillRoundRect(Rect{4, 4, 1, 1}, 6, RGB(255, 0, 0))
	if s.Pix[4*20+4].A() != 255 {
		t.Fatal("prostokat 1x1 nie zostal narysowany")
	}
}

func TestOutlineRoundRectDrawsBorderAndKeepsInterior(t *testing.T) {
	s := NewSurface(40, 40)
	r := Rect{5, 5, 30, 30}
	s.OutlineRoundRect(r, 8, 2, RGB(255, 255, 255))
	if got := s.Pix[20*40+5]; got.A() != 255 {
		t.Fatalf("lewy bok ramki nie jest pelny: %x", got)
	}
	if got := s.Pix[20*40+20]; got.A() != 0 {
		t.Fatalf("wnetrze ramki zostalo zamalowane: %x", got)
	}
	if got := s.Pix[7*40+7]; got.A() == 0 {
		t.Fatal("luk narozny ramki nie zostal narysowany")
	}
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			if r.Contains(x, y) {
				continue
			}
			if s.Pix[y*40+x].A() != 0 {
				t.Fatalf("piksel %d,%d poza ramka: %x", x, y, s.Pix[y*40+x])
			}
		}
	}
}

func TestOutlineRoundRectDrawsThinRects(t *testing.T) {
	s := NewSurface(20, 20)
	s.OutlineRoundRect(Rect{2, 10, 12, 1}, 4, 2, RGB(255, 255, 255))
	if n := filled(s); n != 12 {
		t.Fatalf("ramka o wysokosci 1 narysowala %d pikseli, oczekiwano 12", n)
	}
	s.Clear(RGBA(0, 0, 0, 0))
	s.OutlineRoundRect(Rect{2, 2, 10, 10}, 4, 0, RGB(255, 255, 255))
	if n := filled(s); n != 0 {
		t.Fatalf("zerowa grubosc narysowala %d pikseli", n)
	}
}

func TestSoftShadowFadesOutward(t *testing.T) {
	for _, radius := range []int{3, 12} {
		s := NewSurface(81, 81)
		s.SoftShadow(40, 40, radius, 160)
		inner := s.Pix[40*81+40].A()
		outer := s.Pix[40*81+40+radius+max(1, radius/5)+1].A()
		if inner == 0 {
			t.Fatalf("radius=%d: cien nie zostal narysowany", radius)
		}
		if outer == 0 || outer >= inner {
			t.Fatalf("radius=%d: brak gradientu cienia: inner=%d outer=%d", radius, inner, outer)
		}
	}
	s := NewSurface(20, 20)
	s.SoftShadow(10, 10, 0, 160)
	s.SoftShadow(10, 10, 5, 0)
	if n := filled(s); n != 0 {
		t.Fatalf("pusty cien narysowal %d pikseli", n)
	}
}
