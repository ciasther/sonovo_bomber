package bomber

import (
	"math"
	"strings"
	"sync"
	"unicode"
)

// Font wektorowy: glif to zbiór łamanych w siatce 5x7 jednostek, rasteryzowany
// raz na rozmiar i trzymany w cache jako maska krycia.
const (
	glyphRows    = 7.0
	glyphPad     = 0.9
	glyphAdvance = 6.0
)

type glyph struct{ strokes [][]Vec2 }

func seg(v ...float64) []Vec2 {
	pts := make([]Vec2, 0, len(v)/2)
	for i := 0; i+1 < len(v); i += 2 {
		pts = append(pts, Vec2{v[i], v[i+1]})
	}
	return pts
}

func arc(cx, cy, rx, ry, a0, a1 float64) []Vec2 {
	steps := clamp(int(math.Abs(a1-a0)/12)+1, 2, 64)
	pts := make([]Vec2, 0, steps+1)
	for i := 0; i <= steps; i++ {
		a := (a0 + (a1-a0)*float64(i)/float64(steps)) * math.Pi / 180
		pts = append(pts, Vec2{cx + rx*math.Cos(a), cy + ry*math.Sin(a)})
	}
	return pts
}

var glyphs = map[rune]glyph{
	' ': {},
	'A': {[][]Vec2{seg(0.75, 6.25, 2.5, 0.75), seg(2.5, 0.75, 4.25, 6.25), seg(1.4, 4.35, 3.6, 4.35)}},
	'B': {[][]Vec2{seg(0.9, 0.75, 0.9, 6.25), arc(0.9, 2.125, 3.0, 1.375, -90, 90), arc(0.9, 4.875, 3.35, 1.375, -90, 90)}},
	'C': {[][]Vec2{arc(2.5, 3.5, 1.75, 2.75, 55, 305)}},
	'D': {[][]Vec2{seg(0.9, 0.75, 0.9, 6.25), arc(0.9, 3.5, 3.35, 2.75, -90, 90)}},
	'E': {[][]Vec2{seg(0.9, 0.75, 4.1, 0.75), seg(0.9, 0.75, 0.9, 6.25), seg(0.9, 6.25, 4.1, 6.25), seg(0.9, 3.5, 3.6, 3.5)}},
	'F': {[][]Vec2{seg(0.9, 0.75, 4.1, 0.75), seg(0.9, 0.75, 0.9, 6.25), seg(0.9, 3.5, 3.5, 3.5)}},
	'G': {[][]Vec2{arc(2.5, 3.5, 1.75, 2.75, 55, 305), seg(3.5, 5.75, 3.5, 3.45), seg(2.9, 3.45, 4.25, 3.45)}},
	'H': {[][]Vec2{seg(0.9, 0.75, 0.9, 6.25), seg(4.1, 0.75, 4.1, 6.25), seg(0.9, 3.5, 4.1, 3.5)}},
	'I': {[][]Vec2{seg(2.5, 0.75, 2.5, 6.25), seg(1.2, 0.75, 3.8, 0.75), seg(1.2, 6.25, 3.8, 6.25)}},
	'J': {[][]Vec2{seg(3.85, 0.75, 3.85, 4.85), arc(2.45, 4.85, 1.4, 1.4, 0, 180)}},
	'K': {[][]Vec2{seg(0.9, 0.75, 0.9, 6.25), seg(4.15, 0.75, 1.15, 3.55), seg(1.15, 3.45, 4.15, 6.25)}},
	'L': {[][]Vec2{seg(0.9, 0.75, 0.9, 6.25), seg(0.9, 6.25, 4.1, 6.25)}},
	'M': {[][]Vec2{seg(0.8, 6.25, 0.8, 0.75), seg(0.8, 0.75, 2.5, 3.9), seg(2.5, 3.9, 4.2, 0.75), seg(4.2, 0.75, 4.2, 6.25)}},
	'N': {[][]Vec2{seg(0.9, 6.25, 0.9, 0.75), seg(0.9, 0.75, 4.1, 6.25), seg(4.1, 6.25, 4.1, 0.75)}},
	'O': {[][]Vec2{arc(2.5, 3.5, 1.75, 2.75, 0, 360)}},
	'P': {[][]Vec2{seg(0.9, 0.75, 0.9, 6.25), arc(0.9, 2.3, 3.1, 1.55, -90, 90)}},
	'Q': {[][]Vec2{arc(2.5, 3.5, 1.75, 2.75, 0, 360), seg(2.95, 5.15, 4.2, 6.4)}},
	'R': {[][]Vec2{seg(0.9, 0.75, 0.9, 6.25), arc(0.9, 2.3, 3.1, 1.55, -90, 90), seg(1.35, 3.85, 4.15, 6.25)}},
	'S': {[][]Vec2{seg(3.95, 1.55, 3.35, 0.8, 2.2, 0.72, 1.15, 1.25, 0.95, 2.3, 1.6, 3.1, 2.9, 3.55, 3.95, 4.15, 4.15, 5.2, 3.5, 6.1, 2.2, 6.3, 1.05, 5.75)}},
	'T': {[][]Vec2{seg(0.8, 0.75, 4.2, 0.75), seg(2.5, 0.75, 2.5, 6.25)}},
	'U': {[][]Vec2{seg(0.9, 0.75, 0.9, 4.1), arc(2.5, 4.1, 1.6, 2.15, 180, 0), seg(4.1, 0.75, 4.1, 4.1)}},
	'V': {[][]Vec2{seg(0.85, 0.75, 2.5, 6.25), seg(2.5, 6.25, 4.15, 0.75)}},
	'W': {[][]Vec2{seg(0.7, 0.75, 1.55, 6.25), seg(1.55, 6.25, 2.5, 2.8), seg(2.5, 2.8, 3.45, 6.25), seg(3.45, 6.25, 4.3, 0.75)}},
	'X': {[][]Vec2{seg(0.9, 0.75, 4.1, 6.25), seg(4.1, 0.75, 0.9, 6.25)}},
	'Y': {[][]Vec2{seg(0.9, 0.75, 2.5, 3.4), seg(4.1, 0.75, 2.5, 3.4), seg(2.5, 3.4, 2.5, 6.25)}},
	'Z': {[][]Vec2{seg(0.9, 0.75, 4.1, 0.75), seg(4.1, 0.75, 0.9, 6.25), seg(0.9, 6.25, 4.1, 6.25)}},
	'0': {[][]Vec2{arc(2.5, 3.5, 1.7, 2.75, 0, 360)}},
	'1': {[][]Vec2{seg(1.35, 1.75, 2.5, 0.75), seg(2.5, 0.75, 2.5, 6.25), seg(1.3, 6.25, 3.7, 6.25)}},
	'2': {[][]Vec2{seg(0.95, 1.85, 1.45, 0.95, 2.6, 0.72, 3.75, 1.05, 4.1, 2.05, 3.55, 3.1, 0.95, 6.25, 4.15, 6.25)}},
	'3': {[][]Vec2{seg(1.0, 1.45, 1.85, 0.75, 3.05, 0.78, 3.95, 1.5, 3.85, 2.5, 2.85, 3.35, 3.95, 4.2, 4.05, 5.35, 3.1, 6.25, 1.75, 6.25, 0.95, 5.6)}},
	'4': {[][]Vec2{seg(3.35, 0.75, 0.7, 4.65), seg(0.7, 4.65, 4.3, 4.65), seg(3.35, 0.75, 3.35, 6.25)}},
	'5': {[][]Vec2{seg(4.0, 0.8, 1.15, 0.8), seg(1.15, 0.8, 1.0, 3.05), seg(1.0, 3.05, 2.15, 2.6, 3.35, 2.9, 4.1, 3.9, 4.0, 5.2, 3.0, 6.25, 1.65, 6.25, 0.9, 5.7)}},
	'6': {[][]Vec2{arc(2.55, 4.35, 1.65, 1.85, 0, 360), seg(3.6, 1.05, 2.55, 0.72, 1.45, 1.5, 1.0, 2.9, 0.9, 4.2)}},
	'7': {[][]Vec2{seg(0.85, 0.8, 4.15, 0.8), seg(4.15, 0.8, 1.7, 6.25)}},
	'8': {[][]Vec2{arc(2.5, 2.05, 1.4, 1.3, 0, 360), arc(2.5, 4.85, 1.7, 1.4, 0, 360)}},
	'9': {[][]Vec2{arc(2.45, 2.6, 1.65, 1.85, 0, 360), seg(1.4, 5.45, 2.45, 6.28, 3.55, 5.5, 4.0, 4.1, 4.1, 2.8)}},
	'.': {[][]Vec2{seg(2.5, 6.05, 2.5, 6.05)}},
	',': {[][]Vec2{seg(2.7, 5.6, 2.35, 6.4)}},
	':': {[][]Vec2{seg(2.5, 2.6, 2.5, 2.6), seg(2.5, 5.4, 2.5, 5.4)}},
	'-': {[][]Vec2{seg(1.2, 3.5, 3.8, 3.5)}},
	'+': {[][]Vec2{seg(1.05, 3.5, 3.95, 3.5), seg(2.5, 2.05, 2.5, 4.95)}},
	'/': {[][]Vec2{seg(1.0, 6.25, 4.0, 0.75)}},
	'!': {[][]Vec2{seg(2.5, 0.75, 2.5, 4.55), seg(2.5, 6.1, 2.5, 6.1)}},
	'?': {[][]Vec2{seg(1.15, 1.75, 1.75, 0.85, 2.95, 0.72, 3.9, 1.5, 3.75, 2.6, 2.9, 3.3, 2.5, 4.35), seg(2.5, 6.1, 2.5, 6.1)}},
	'(': {[][]Vec2{arc(3.7, 3.5, 2.6, 2.95, 118, 242)}},
	')': {[][]Vec2{arc(1.3, 3.5, 2.6, 2.95, -62, 62)}},
	'#': {[][]Vec2{seg(1.75, 1.1, 1.15, 5.9), seg(3.45, 1.1, 2.85, 5.9), seg(0.75, 2.6, 4.25, 2.6), seg(0.6, 4.4, 4.1, 4.4)}},
	'=': {[][]Vec2{seg(1.0, 2.6, 4.0, 2.6), seg(1.0, 4.4, 4.0, 4.4)}},
	'|': {[][]Vec2{seg(2.5, 0.6, 2.5, 6.4)}},
}

type glyphKey struct {
	r    rune
	unit int
}

type glyphMask struct {
	mask       *Surface
	offX, offY int
}

var (
	glyphCacheMu sync.Mutex
	glyphCache   = map[glyphKey]*glyphMask{}
)

// Jedno źródło metryki: TextWidth, DrawText i klucz cache liczą tę samą jednostkę.
func glyphUnit(height int) int { return max(1, height/7) }

func glyphStrokeWidth(unit int) float64 { return math.Max(1.4, float64(unit)*1.05) }

func rasterizeGlyph(r rune, unit int) *glyphMask {
	g, ok := glyphs[r]
	if !ok {
		g = glyphs['?']
	}
	width := glyphStrokeWidth(unit)
	half := width/2 + 1
	fu := float64(unit)
	pad := glyphPad * fu
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	scaled := make([][]Vec2, 0, len(g.strokes))
	for _, stroke := range g.strokes {
		pts := make([]Vec2, len(stroke))
		for i, p := range stroke {
			pts[i] = Vec2{p.X*fu + pad, p.Y*fu + pad}
			minX, maxX = math.Min(minX, pts[i].X), math.Max(maxX, pts[i].X)
			minY, maxY = math.Min(minY, pts[i].Y), math.Max(maxY, pts[i].Y)
		}
		scaled = append(scaled, pts)
	}
	if len(scaled) == 0 {
		return &glyphMask{}
	}
	x0 := int(math.Floor(minX - half))
	y0 := int(math.Floor(minY - half))
	x1 := int(math.Ceil(maxX + half))
	y1 := int(math.Ceil(maxY + half))
	m := NewSurface(x1-x0+1, y1-y0+1)
	white := RGB(255, 255, 255)
	for _, pts := range scaled {
		shifted := make([]Vec2, len(pts))
		for i, p := range pts {
			shifted[i] = Vec2{p.X - float64(x0), p.Y - float64(y0)}
		}
		m.StrokePath(shifted, width, white)
	}
	return &glyphMask{mask: m, offX: x0 - int(math.Round(pad)), offY: y0 - int(math.Round(pad))}
}

func glyphFor(r rune, unit int) *glyphMask {
	key := glyphKey{r, unit}
	glyphCacheMu.Lock()
	defer glyphCacheMu.Unlock()
	if m, ok := glyphCache[key]; ok {
		return m
	}
	if len(glyphCache) > 4096 {
		glyphCache = map[glyphKey]*glyphMask{}
	}
	m := rasterizeGlyph(r, unit)
	glyphCache[key] = m
	return m
}

func NormalizeText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'ą', 'Ą':
			r = 'A'
		case 'ć', 'Ć':
			r = 'C'
		case 'ę', 'Ę':
			r = 'E'
		case 'ł', 'Ł':
			r = 'L'
		case 'ń', 'Ń':
			r = 'N'
		case 'ó', 'Ó':
			r = 'O'
		case 'ś', 'Ś':
			r = 'S'
		case 'ź', 'Ź', 'ż', 'Ż':
			r = 'Z'
		default:
			r = unicode.ToUpper(r)
		}
		if r == '\n' || r == '\t' {
			b.WriteRune(r)
		} else if _, ok := glyphs[r]; ok {
			b.WriteRune(r)
		} else {
			b.WriteRune('?')
		}
	}
	return b.String()
}

func TextWidth(text string, height int) int {
	text = NormalizeText(text)
	unit := glyphUnit(height)
	count := 0
	for _, r := range text {
		if r == '\n' {
			break
		}
		count++
	}
	if count == 0 {
		return 0
	}
	return count*int(glyphAdvance)*unit - unit
}

func FitTextHeight(text string, maxHeight, maxWidth int) int {
	h := maxHeight
	for h > 7 && TextWidth(text, h) > maxWidth {
		h--
	}
	return max(7, h)
}

func (s *Surface) DrawText(x, y, height int, text string, c Color) Rect {
	text = NormalizeText(text)
	unit := glyphUnit(height)
	cursor := x
	for _, r := range text {
		if r == '\n' {
			break
		}
		if g := glyphFor(r, unit); g.mask != nil {
			s.DrawMask(g.mask, cursor+g.offX, y+g.offY, c)
		}
		cursor += int(glyphAdvance) * unit
	}
	return Rect{x, y, max(0, cursor-x-unit), int(glyphRows) * unit}
}

func (s *Surface) DrawTextShadow(x, y, height int, text string, c Color) Rect {
	unit := glyphUnit(height)
	s.DrawText(x+max(1, unit/2), y+max(1, unit/2), height, text, RGBA(0, 0, 0, 110))
	return s.DrawText(x, y, height, text, c)
}

func (s *Surface) DrawTextCentered(r Rect, height int, text string, c Color) Rect {
	h := FitTextHeight(text, height, max(1, r.W))
	w := TextWidth(text, h)
	unit := glyphUnit(h)
	actualH := unit * int(glyphRows)
	return s.DrawTextShadow(r.X+(r.W-w)/2, r.Y+(r.H-actualH)/2, h, text, c)
}

func (s *Surface) DrawParagraph(r Rect, height, lineGap int, text string, c Color) int {
	text = NormalizeText(text)
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}
	var lines []string
	line := ""
	for _, word := range words {
		candidate := word
		if line != "" {
			candidate = line + " " + word
		}
		if TextWidth(candidate, height) <= r.W || line == "" {
			line = candidate
		} else {
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	unit := glyphUnit(height)
	lineH := int(glyphRows)*unit + lineGap
	for i, ln := range lines {
		if r.Y+i*lineH+int(glyphRows)*unit > r.Bottom() {
			break
		}
		s.DrawText(r.X, r.Y+i*lineH, height, ln, c)
	}
	return len(lines) * lineH
}
