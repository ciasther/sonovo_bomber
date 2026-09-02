package bomber

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"math"
	"strconv"
	"strings"
)

type Color uint32

func RGBA(r, g, b, a uint8) Color {
	return Color(uint32(a)<<24 | uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

func RGB(r, g, b uint8) Color { return RGBA(r, g, b, 255) }

func (c Color) R() uint8 { return uint8(uint32(c) >> 16) }
func (c Color) G() uint8 { return uint8(uint32(c) >> 8) }
func (c Color) B() uint8 { return uint8(c) }
func (c Color) A() uint8 { return uint8(uint32(c) >> 24) }

func (c Color) Alpha(a uint8) Color { return RGBA(c.R(), c.G(), c.B(), a) }

func Mix(a, b Color, t float64) Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	lerp := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + .5) }
	return RGBA(lerp(a.R(), b.R()), lerp(a.G(), b.G()), lerp(a.B(), b.B()), lerp(a.A(), b.A()))
}

func ParseHexColor(s string, fallback Color) Color {
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
	if len(s) != 6 && len(s) != 8 {
		return fallback
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return fallback
	}
	if len(s) == 6 {
		return RGBA(uint8(v>>16), uint8(v>>8), uint8(v), 255)
	}
	return RGBA(uint8(v>>24), uint8(v>>16), uint8(v>>8), uint8(v))
}

type Rect struct{ X, Y, W, H int }

func (r Rect) Right() int  { return r.X + r.W }
func (r Rect) Bottom() int { return r.Y + r.H }
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && y >= r.Y && x < r.Right() && y < r.Bottom()
}
func (r Rect) Inset(v int) Rect { return Rect{r.X + v, r.Y + v, r.W - 2*v, r.H - 2*v} }

type Surface struct {
	W, H int
	Pix  []Color
}

func NewSurface(w, h int) *Surface {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return &Surface{W: w, H: h, Pix: make([]Color, w*h)}
}

func (s *Surface) Clear(c Color) {
	for i := range s.Pix {
		s.Pix[i] = c
	}
}

func (s *Surface) Set(x, y int, c Color) {
	if x < 0 || y < 0 || x >= s.W || y >= s.H {
		return
	}
	if c.A() == 255 {
		s.Pix[y*s.W+x] = c
		return
	}
	s.Pix[y*s.W+x] = blendOver(s.Pix[y*s.W+x], c)
}

func blendOver(dst, src Color) Color {
	sa := uint32(src.A())
	if sa == 255 {
		return src
	}
	if sa == 0 {
		return dst
	}
	inv := 255 - sa
	r := (uint32(src.R())*sa + uint32(dst.R())*inv + 127) / 255
	g := (uint32(src.G())*sa + uint32(dst.G())*inv + 127) / 255
	b := (uint32(src.B())*sa + uint32(dst.B())*inv + 127) / 255
	da := uint32(dst.A())
	a := sa + (da*inv+127)/255
	return RGBA(uint8(r), uint8(g), uint8(b), uint8(a))
}

func (s *Surface) FillRect(r Rect, c Color) {
	x0, y0 := max(0, r.X), max(0, r.Y)
	x1, y1 := min(s.W, r.Right()), min(s.H, r.Bottom())
	if x0 >= x1 || y0 >= y1 {
		return
	}
	if c.A() == 255 {
		for y := y0; y < y1; y++ {
			row := s.Pix[y*s.W+x0 : y*s.W+x1]
			for i := range row {
				row[i] = c
			}
		}
		return
	}
	for y := y0; y < y1; y++ {
		base := y * s.W
		for x := x0; x < x1; x++ {
			i := base + x
			s.Pix[i] = blendOver(s.Pix[i], c)
		}
	}
}

func (s *Surface) FillGradient(top, bottom Color) {
	for y := 0; y < s.H; y++ {
		t := float64(y) / float64(max(1, s.H-1))
		c := Mix(top, bottom, t)
		row := s.Pix[y*s.W : (y+1)*s.W]
		for i := range row {
			row[i] = c
		}
	}
}

func (s *Surface) FillRoundRect(r Rect, radius int, c Color) {
	if radius <= 0 {
		s.FillRect(r, c)
		return
	}
	radius = min(radius, min(r.W, r.H)/2)
	for yy := 0; yy < r.H; yy++ {
		y := r.Y + yy
		if y < 0 || y >= s.H {
			continue
		}
		inset := 0
		if yy < radius {
			dy := float64(radius-yy) - .5
			inset = int(math.Ceil(float64(radius) - math.Sqrt(maxFloat(0, float64(radius*radius)-dy*dy))))
		} else if yy >= r.H-radius {
			dy := float64(yy-(r.H-radius-1)) - .5
			inset = int(math.Ceil(float64(radius) - math.Sqrt(maxFloat(0, float64(radius*radius)-dy*dy))))
		}
		s.FillRect(Rect{r.X + inset, y, r.W - 2*inset, 1}, c)
	}
}

func (s *Surface) OutlineRoundRect(r Rect, radius, width int, c Color) {
	if width <= 0 {
		return
	}
	for i := 0; i < width; i++ {
		rr := r.Inset(i)
		rad := max(0, radius-i)
		for x := rr.X + rad; x < rr.Right()-rad; x++ {
			s.Set(x, rr.Y, c)
			s.Set(x, rr.Bottom()-1, c)
		}
		for y := rr.Y + rad; y < rr.Bottom()-rad; y++ {
			s.Set(rr.X, y, c)
			s.Set(rr.Right()-1, y, c)
		}
		cx1, cy1 := rr.X+rad, rr.Y+rad
		cx2, cy2 := rr.Right()-rad-1, rr.Bottom()-rad-1
		for a := 0; a <= 90; a++ {
			t := float64(a) * math.Pi / 180
			dx := int(math.Round(float64(rad) * math.Cos(t)))
			dy := int(math.Round(float64(rad) * math.Sin(t)))
			s.Set(cx2+dx, cy1-dy, c)
			s.Set(cx1-dx, cy1-dy, c)
			s.Set(cx1-dx, cy2+dy, c)
			s.Set(cx2+dx, cy2+dy, c)
		}
	}
}

func (s *Surface) FillCircle(cx, cy, radius int, c Color) {
	if radius <= 0 {
		return
	}
	r2 := radius * radius
	for dy := -radius; dy <= radius; dy++ {
		dx := int(math.Sqrt(float64(max(0, r2-dy*dy))))
		s.FillRect(Rect{cx - dx, cy + dy, dx*2 + 1, 1}, c)
	}
}

func (s *Surface) Ring(cx, cy, radius, width int, c Color) {
	if width <= 0 || radius <= 0 {
		return
	}
	rOuter2 := radius * radius
	rInner := max(0, radius-width)
	rInner2 := rInner * rInner
	for dy := -radius; dy <= radius; dy++ {
		outer := int(math.Sqrt(float64(max(0, rOuter2-dy*dy))))
		inner := -1
		if abs(dy) <= rInner {
			inner = int(math.Sqrt(float64(max(0, rInner2-dy*dy))))
		}
		if inner < 0 {
			s.FillRect(Rect{cx - outer, cy + dy, outer*2 + 1, 1}, c)
		} else {
			s.FillRect(Rect{cx - outer, cy + dy, outer - inner, 1}, c)
			s.FillRect(Rect{cx + inner + 1, cy + dy, outer - inner, 1}, c)
		}
	}
}

func (s *Surface) Line(x0, y0, x1, y1, width int, c Color) {
	dx, dy := x1-x0, y1-y0
	steps := max(abs(dx), abs(dy))
	if steps == 0 {
		s.FillCircle(x0, y0, max(1, width/2), c)
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(float64(x0) + float64(dx)*t))
		y := int(math.Round(float64(y0) + float64(dy)*t))
		s.FillCircle(x, y, max(1, width/2), c)
	}
}

func (s *Surface) GlowCircle(cx, cy, radius int, c Color) {
	for i := 4; i >= 1; i-- {
		r := radius + i*radius/3
		a := uint8(max(2, int(c.A())/(i*5)))
		s.FillCircle(cx, cy, r, c.Alpha(a))
	}
	s.FillCircle(cx, cy, radius, c)
}

type Sprite struct {
	W, H int
	Pix  []Color
}

func SpriteFromImage(img image.Image) *Sprite {
	b := img.Bounds()
	sp := &Sprite{W: b.Dx(), H: b.Dy(), Pix: make([]Color, b.Dx()*b.Dy())}
	for y := 0; y < sp.H; y++ {
		for x := 0; x < sp.W; x++ {
			r, g, b, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			sp.Pix[y*sp.W+x] = RGBA(uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
		}
	}
	return sp
}

func (s *Surface) DrawSprite(sp *Sprite, dst Rect, opacity uint8, contain bool) {
	if sp == nil || dst.W <= 0 || dst.H <= 0 || sp.W <= 0 || sp.H <= 0 {
		return
	}
	d := dst
	if contain {
		scale := math.Min(float64(dst.W)/float64(sp.W), float64(dst.H)/float64(sp.H))
		d.W = max(1, int(float64(sp.W)*scale))
		d.H = max(1, int(float64(sp.H)*scale))
		d.X = dst.X + (dst.W-d.W)/2
		d.Y = dst.Y + (dst.H-d.H)/2
	}
	x0, y0 := max(0, d.X), max(0, d.Y)
	x1, y1 := min(s.W, d.Right()), min(s.H, d.Bottom())
	for y := y0; y < y1; y++ {
		sy := (y - d.Y) * sp.H / d.H
		for x := x0; x < x1; x++ {
			sx := (x - d.X) * sp.W / d.W
			c := sp.Pix[sy*sp.W+sx]
			if opacity != 255 {
				c = c.Alpha(uint8(uint16(c.A()) * uint16(opacity) / 255))
			}
			s.Set(x, y, c)
		}
	}
}

func (s *Surface) DrawSpriteCover(sp *Sprite, dst Rect, opacity uint8) {
	if sp == nil || dst.W <= 0 || dst.H <= 0 {
		return
	}
	scale := math.Max(float64(dst.W)/float64(sp.W), float64(dst.H)/float64(sp.H))
	sw, sh := float64(dst.W)/scale, float64(dst.H)/scale
	sx0 := (float64(sp.W) - sw) / 2
	sy0 := (float64(sp.H) - sh) / 2
	x0, y0 := max(0, dst.X), max(0, dst.Y)
	x1, y1 := min(s.W, dst.Right()), min(s.H, dst.Bottom())
	for y := y0; y < y1; y++ {
		sy := int(sy0 + float64(y-dst.Y)*sh/float64(dst.H))
		sy = clamp(sy, 0, sp.H-1)
		for x := x0; x < x1; x++ {
			sx := int(sx0 + float64(x-dst.X)*sw/float64(dst.W))
			sx = clamp(sx, 0, sp.W-1)
			c := sp.Pix[sy*sp.W+sx]
			if opacity != 255 {
				c = c.Alpha(uint8(uint16(c.A()) * uint16(opacity) / 255))
			}
			s.Set(x, y, c)
		}
	}
}

func (s *Surface) Image() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, s.W, s.H))
	for i, c := range s.Pix {
		j := i * 4
		img.Pix[j] = c.R()
		img.Pix[j+1] = c.G()
		img.Pix[j+2] = c.B()
		img.Pix[j+3] = c.A()
	}
	return img
}

func ScaleImageNearest(src *Surface, w, h int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := y * src.H / h
		for x := 0; x < w; x++ {
			sx := x * src.W / w
			c := src.Pix[sy*src.W+sx]
			i := (y*w + x) * 4
			img.Pix[i] = c.R()
			img.Pix[i+1] = c.G()
			img.Pix[i+2] = c.B()
			img.Pix[i+3] = c.A()
		}
	}
	return img
}

func HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
