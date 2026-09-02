package bomber

import (
	"strings"
	"unicode"
)

var glyphs = map[rune][7]uint8{
	' ': {},
	'A': {0b01110, 0b10001, 0b10001, 0b11111, 0b10001, 0b10001, 0b10001},
	'B': {0b11110, 0b10001, 0b10001, 0b11110, 0b10001, 0b10001, 0b11110},
	'C': {0b01111, 0b10000, 0b10000, 0b10000, 0b10000, 0b10000, 0b01111},
	'D': {0b11110, 0b10001, 0b10001, 0b10001, 0b10001, 0b10001, 0b11110},
	'E': {0b11111, 0b10000, 0b10000, 0b11110, 0b10000, 0b10000, 0b11111},
	'F': {0b11111, 0b10000, 0b10000, 0b11110, 0b10000, 0b10000, 0b10000},
	'G': {0b01111, 0b10000, 0b10000, 0b10111, 0b10001, 0b10001, 0b01111},
	'H': {0b10001, 0b10001, 0b10001, 0b11111, 0b10001, 0b10001, 0b10001},
	'I': {0b11111, 0b00100, 0b00100, 0b00100, 0b00100, 0b00100, 0b11111},
	'J': {0b00111, 0b00010, 0b00010, 0b00010, 0b10010, 0b10010, 0b01100},
	'K': {0b10001, 0b10010, 0b10100, 0b11000, 0b10100, 0b10010, 0b10001},
	'L': {0b10000, 0b10000, 0b10000, 0b10000, 0b10000, 0b10000, 0b11111},
	'M': {0b10001, 0b11011, 0b10101, 0b10101, 0b10001, 0b10001, 0b10001},
	'N': {0b10001, 0b11001, 0b10101, 0b10011, 0b10001, 0b10001, 0b10001},
	'O': {0b01110, 0b10001, 0b10001, 0b10001, 0b10001, 0b10001, 0b01110},
	'P': {0b11110, 0b10001, 0b10001, 0b11110, 0b10000, 0b10000, 0b10000},
	'Q': {0b01110, 0b10001, 0b10001, 0b10001, 0b10101, 0b10010, 0b01101},
	'R': {0b11110, 0b10001, 0b10001, 0b11110, 0b10100, 0b10010, 0b10001},
	'S': {0b01111, 0b10000, 0b10000, 0b01110, 0b00001, 0b00001, 0b11110},
	'T': {0b11111, 0b00100, 0b00100, 0b00100, 0b00100, 0b00100, 0b00100},
	'U': {0b10001, 0b10001, 0b10001, 0b10001, 0b10001, 0b10001, 0b01110},
	'V': {0b10001, 0b10001, 0b10001, 0b10001, 0b10001, 0b01010, 0b00100},
	'W': {0b10001, 0b10001, 0b10001, 0b10101, 0b10101, 0b10101, 0b01010},
	'X': {0b10001, 0b10001, 0b01010, 0b00100, 0b01010, 0b10001, 0b10001},
	'Y': {0b10001, 0b10001, 0b01010, 0b00100, 0b00100, 0b00100, 0b00100},
	'Z': {0b11111, 0b00001, 0b00010, 0b00100, 0b01000, 0b10000, 0b11111},
	'0': {0b01110, 0b10001, 0b10011, 0b10101, 0b11001, 0b10001, 0b01110},
	'1': {0b00100, 0b01100, 0b00100, 0b00100, 0b00100, 0b00100, 0b01110},
	'2': {0b01110, 0b10001, 0b00001, 0b00010, 0b00100, 0b01000, 0b11111},
	'3': {0b11110, 0b00001, 0b00001, 0b01110, 0b00001, 0b00001, 0b11110},
	'4': {0b00010, 0b00110, 0b01010, 0b10010, 0b11111, 0b00010, 0b00010},
	'5': {0b11111, 0b10000, 0b10000, 0b11110, 0b00001, 0b00001, 0b11110},
	'6': {0b01110, 0b10000, 0b10000, 0b11110, 0b10001, 0b10001, 0b01110},
	'7': {0b11111, 0b00001, 0b00010, 0b00100, 0b01000, 0b01000, 0b01000},
	'8': {0b01110, 0b10001, 0b10001, 0b01110, 0b10001, 0b10001, 0b01110},
	'9': {0b01110, 0b10001, 0b10001, 0b01111, 0b00001, 0b00001, 0b01110},
	'.': {0, 0, 0, 0, 0, 0b00110, 0b00110},
	',': {0, 0, 0, 0, 0b00110, 0b00110, 0b00100},
	':': {0, 0b00110, 0b00110, 0, 0b00110, 0b00110, 0},
	'-': {0, 0, 0, 0b11111, 0, 0, 0},
	'+': {0, 0b00100, 0b00100, 0b11111, 0b00100, 0b00100, 0},
	'/': {0b00001, 0b00010, 0b00010, 0b00100, 0b01000, 0b01000, 0b10000},
	'!': {0b00100, 0b00100, 0b00100, 0b00100, 0b00100, 0, 0b00100},
	'?': {0b01110, 0b10001, 0b00001, 0b00010, 0b00100, 0, 0b00100},
	'(': {0b00010, 0b00100, 0b01000, 0b01000, 0b01000, 0b00100, 0b00010},
	')': {0b01000, 0b00100, 0b00010, 0b00010, 0b00010, 0b00100, 0b01000},
	'#': {0b01010, 0b11111, 0b01010, 0b01010, 0b11111, 0b01010, 0},
	'=': {0, 0b11111, 0, 0b11111, 0, 0, 0},
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
	unit := max(1, height/7)
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
	return count*6*unit - unit
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
	unit := max(1, height/7)
	cursor := x
	for _, r := range text {
		if r == '\n' {
			break
		}
		g, ok := glyphs[r]
		if !ok {
			g = glyphs['?']
		}
		for row := 0; row < 7; row++ {
			bits := g[row]
			for col := 0; col < 5; col++ {
				if bits&(1<<(4-col)) != 0 {
					s.FillRoundRect(Rect{cursor + col*unit, y + row*unit, unit, unit}, unit/4, c)
				}
			}
		}
		cursor += 6 * unit
	}
	return Rect{x, y, max(0, cursor-x-unit), 7 * unit}
}

func (s *Surface) DrawTextShadow(x, y, height int, text string, c Color) Rect {
	unit := max(1, height/7)
	s.DrawText(x+max(1, unit/2), y+max(1, unit/2), height, text, RGBA(0, 0, 0, 110))
	return s.DrawText(x, y, height, text, c)
}

func (s *Surface) DrawTextCentered(r Rect, height int, text string, c Color) Rect {
	h := FitTextHeight(text, height, max(1, r.W))
	w := TextWidth(text, h)
	unit := max(1, h/7)
	actualH := unit * 7
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
	unit := max(1, height/7)
	lineH := 7*unit + lineGap
	for i, ln := range lines {
		if r.Y+i*lineH+7*unit > r.Bottom() {
			break
		}
		s.DrawText(r.X, r.Y+i*lineH, height, ln, c)
	}
	return len(lines) * lineH
}
