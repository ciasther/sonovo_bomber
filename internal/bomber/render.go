package bomber

import (
	"fmt"
	"math"
)

type KeyRegion struct {
	Rect  Rect
	Value string
	Label string
}

type Layout struct {
	Landscape       bool
	Margin          int
	Header          Rect
	NanoLogo        Rect
	PartnerLogo     Rect
	Content         Rect
	Board           Rect
	HUD             Rect
	Sponsor         Rect
	Title           Rect
	Subtitle        Rect
	NickDisplay     Rect
	Keyboard        Rect
	KeyboardKeys    []KeyRegion
	PinKeys         []KeyRegion
	PrimaryButton   Rect
	SecondaryButton Rect
	BackButton      Rect
	AdminButton     Rect
	TabOne          Rect
	TabTwo          Rect
	Table           Rect
	PagePrev        Rect
	PageNext        Rect
	Hint            Rect
}

func LogicalSize(screenW, screenH int) (int, int) {
	if screenW <= 0 || screenH <= 0 {
		return 1440, 810
	}
	const longEdge = 1440
	ratio := float64(screenW) / float64(screenH)
	if ratio >= 1 {
		w := longEdge
		h := int(math.Round(float64(w) / ratio))
		return w, max(640, h)
	}
	h := longEdge
	w := int(math.Round(float64(h) * ratio))
	return max(640, w), h
}

func ComputeLayout(w, h int, screen Screen) Layout {
	short := min(w, h)
	m := max(24, int(float64(short)*.032))
	headerH := clamp(int(float64(h)*.105), 90, 190)
	if screen == ScreenCountdown || screen == ScreenPlay {
		headerH = 0
	}
	land := float64(w)/float64(h) >= 1.15
	logoH := max(62, headerH-m/2)
	logoW := int(float64(logoH) * 3.5)
	logoW = min(logoW, int(float64(w)*.25))
	l := Layout{
		Landscape:   land,
		Margin:      m,
		Header:      Rect{0, 0, w, headerH},
		NanoLogo:    Rect{m, m / 2, logoW, logoH},
		PartnerLogo: Rect{w - m - min(int(float64(w)*.22), logoW), m / 2, min(int(float64(w)*.22), logoW), logoH},
		Content:     Rect{m, headerH + m/2, w - 2*m, h - headerH - 3*m/2},
	}
	if headerH == 0 {
		l.Header = Rect{}
		l.NanoLogo = Rect{}
		l.PartnerLogo = Rect{}
	}

	switch screen {
	case ScreenAttract:
		l.Title = Rect{m, headerH + int(float64(h)*.08), w - 2*m, int(float64(h) * .21)}
		l.Subtitle = Rect{int(float64(w) * .13), l.Title.Bottom(), int(float64(w) * .74), int(float64(h) * .14)}
		bw := min(int(float64(w)*.62), 980)
		bh := clamp(int(float64(h)*.13), 120, 210)
		l.PrimaryButton = Rect{(w - bw) / 2, h - bh - m*2, bw, bh}
	case ScreenNick:
		l.Title = Rect{m, headerH + m/2, w - 2*m, clamp(int(float64(h)*.10), 95, 165)}
		displayW := min(w-2*m, max(620, int(float64(w)*.72)))
		displayH := clamp(int(float64(h)*.105), 100, 170)
		l.NickDisplay = Rect{(w - displayW) / 2, l.Title.Bottom() + m/2, displayW, displayH}
		keyboardArea := Rect{m, l.NickDisplay.Bottom() + m, w - 2*m, h - l.NickDisplay.Bottom() - 2*m}
		l.Keyboard = compactKeyboardArea(keyboardArea, land)
		l.KeyboardKeys = keyboardRegions(l.Keyboard, m)
		l.BackButton = Rect{m, headerH + m/2, max(120, int(float64(short)*.11)), max(70, int(float64(short)*.065))}
	case ScreenCountdown, ScreenPlay:
		if land {
			available := l.Content
			rightW := clamp(int(float64(w)*.22), 300, 440)
			gap := m
			boardMaxW := available.W - rightW - gap
			boardH := available.H
			boardW := int(float64(boardH) * float64(GridW) / float64(GridH))
			if boardW > boardMaxW {
				boardW = boardMaxW
				boardH = int(float64(boardW) * float64(GridH) / float64(GridW))
			}
			l.Board = Rect{available.X, available.Y + (available.H-boardH)/2, boardW, boardH}
			l.HUD = Rect{l.Board.Right() + gap, available.Y + (available.H-clamp(int(float64(available.H)*.48), 250, 360))/2, rightW, clamp(int(float64(available.H)*.48), 250, 360)}
			l.Hint = Rect{l.HUD.X, l.HUD.Bottom() + m/2, l.HUD.W, min(130, available.Bottom()-l.HUD.Bottom()-m/2)}
		} else {
			available := l.Content
			hudH := clamp(int(float64(short)*.16), 104, 150)
			boardMaxH := available.H - hudH - m
			boardW := available.W
			boardH := int(float64(boardW) * float64(GridH) / float64(GridW))
			if boardH > boardMaxH {
				boardH = boardMaxH
				boardW = int(float64(boardH) * float64(GridW) / float64(GridH))
			}
			l.HUD = Rect{available.X, available.Y, available.W, hudH}
			l.Board = Rect{available.X + (available.W-boardW)/2, l.HUD.Bottom() + (available.Bottom()-l.HUD.Bottom()-boardH)/2, boardW, boardH}
			l.Hint = Rect{available.X, l.HUD.Bottom() + m/2, available.W, max(70, l.Board.Y-l.HUD.Bottom()-m)}
		}
	case ScreenSummary:
		if land {
			leftW := int(float64(l.Content.W) * .50)
			l.Table = Rect{l.Content.X, l.Content.Y, leftW, l.Content.H - 150}
			l.Sponsor = Rect{l.Content.X + leftW + m, l.Content.Y, l.Content.W - leftW - m, l.Content.H - 150}
			buttonY := l.Content.Bottom() - 120
			l.PrimaryButton = Rect{l.Content.X, buttonY, leftW, 110}
			l.SecondaryButton = Rect{l.Sponsor.X, buttonY, l.Sponsor.W, 110}
		} else {
			l.Table = Rect{l.Content.X, l.Content.Y, l.Content.W, int(float64(l.Content.H) * .44)}
			l.Sponsor = Rect{l.Content.X, l.Table.Bottom() + m/2, l.Content.W, int(float64(l.Content.H) * .34)}
			buttonY := l.Sponsor.Bottom() + m/2
			l.PrimaryButton = Rect{l.Content.X, buttonY, l.Content.W, 112}
			l.SecondaryButton = Rect{l.Content.X, l.PrimaryButton.Bottom() + m/2, l.Content.W, 96}
		}
	case ScreenRanking:
		tabH := clamp(int(float64(h)*.07), 80, 120)
		controlGap := max(10, m/2)
		adminW := clamp(int(float64(short)*.18), 130, 240)
		tabW := max(150, (l.Content.W-adminW-2*controlGap)/2)
		l.TabOne = Rect{l.Content.X, l.Content.Y, tabW, tabH}
		l.TabTwo = Rect{l.TabOne.Right() + controlGap, l.Content.Y, tabW, tabH}
		l.AdminButton = Rect{l.TabTwo.Right() + controlGap, l.Content.Y, l.Content.Right() - l.TabTwo.Right() - controlGap, tabH}
		buttonH := clamp(int(float64(h)*.085), 94, 132)
		if land {
			sponsorW := clamp(int(float64(w)*.25), 360, 520)
			l.Table = Rect{l.Content.X, l.TabOne.Bottom() + m/2, l.Content.W - sponsorW - m, l.Content.H - tabH - buttonH - 2*m}
			l.Sponsor = Rect{l.Table.Right() + m, l.Table.Y, sponsorW, l.Table.H}
			l.PrimaryButton = Rect{l.Content.X, l.Content.Bottom() - buttonH, l.Content.W, buttonH}
		} else {
			sponsorH := clamp(int(float64(h)*.15), 220, 300)
			l.Table = Rect{l.Content.X, l.TabOne.Bottom() + m/2, l.Content.W, l.Content.H - tabH - sponsorH - buttonH - 2*m}
			l.Sponsor = Rect{l.Content.X, l.Table.Bottom() + m/2, l.Content.W, sponsorH}
			l.PrimaryButton = Rect{l.Content.X, l.Content.Bottom() - buttonH, l.Content.W, buttonH}
		}
	case ScreenAdminPIN:
		boxW := min(l.Content.W, 760)
		boxH := min(l.Content.H, 1040)
		l.Table = Rect{(w - boxW) / 2, l.Content.Y + (l.Content.H-boxH)/2, boxW, boxH}
		l.BackButton = Rect{m, headerH + m/2, max(120, int(float64(short)*.11)), max(70, int(float64(short)*.065))}
		l.PinKeys = pinRegions(l.Table.Inset(m), m)
	case ScreenHistory:
		buttonH := clamp(int(float64(h)*.075), 86, 120)
		l.BackButton = Rect{l.Content.X, l.Content.Y, clamp(int(float64(w)*.16), 150, 280), buttonH}
		l.PagePrev = Rect{l.Content.Right() - 2*clamp(int(float64(w)*.13), 140, 230) - m/2, l.Content.Y, clamp(int(float64(w)*.13), 140, 230), buttonH}
		l.PageNext = Rect{l.PagePrev.Right() + m/2, l.Content.Y, l.PagePrev.W, buttonH}
		l.Table = Rect{l.Content.X, l.BackButton.Bottom() + m/2, l.Content.W, l.Content.Bottom() - l.BackButton.Bottom() - m/2}
	}
	return l
}

func compactKeyboardArea(area Rect, landscape bool) Rect {
	maxW, maxH := 720, 560
	if landscape {
		maxW, maxH = 1080, 360
	}
	w := min(area.W, maxW)
	h := min(area.H, maxH)
	return Rect{area.X + (area.W-w)/2, area.Y + (area.H-h)/2, w, h}
}

func keyboardRegions(area Rect, margin int) []KeyRegion {
	rows := []string{"QWERTYUIOP", "ASDFGHJKL", "ZXCVBNM"}
	gap := max(8, margin/3)
	rowH := (area.H - gap*4) / 4
	var out []KeyRegion
	for ri, row := range rows {
		count := len(row)
		keyW := (area.W - gap*(count-1)) / count
		totalW := keyW*count + gap*(count-1)
		x := area.X + (area.W-totalW)/2
		y := area.Y + ri*(rowH+gap)
		for _, r := range row {
			out = append(out, KeyRegion{Rect: Rect{x, y, keyW, rowH}, Value: string(r), Label: string(r)})
			x += keyW + gap
		}
	}
	y := area.Y + 3*(rowH+gap)
	backW := int(float64(area.W) * .32)
	playW := area.W - backW - gap
	out = append(out,
		KeyRegion{Rect: Rect{area.X, y, backW, rowH}, Value: "BACK", Label: "USUN"},
		KeyRegion{Rect: Rect{area.X + backW + gap, y, playW, rowH}, Value: "PLAY", Label: "GRAJ"},
	)
	return out
}

func pinRegions(area Rect, margin int) []KeyRegion {
	gap := max(10, margin/3)
	top := area.Y + int(float64(area.H)*.24)
	gridH := area.Bottom() - top
	rowH := (gridH - gap*3) / 4
	colW := (area.W - gap*2) / 3
	labels := [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}, {"BACK", "0", "OK"}}
	var out []KeyRegion
	for y, row := range labels {
		for x, v := range row {
			label := v
			if v == "BACK" {
				label = "USUN"
			}
			out = append(out, KeyRegion{Rect: Rect{area.X + x*(colW+gap), top + y*(rowH+gap), colW, rowH}, Value: v, Label: label})
		}
	}
	return out
}

type Renderer struct {
	Surface *Surface
}

func NewRenderer(w, h int) *Renderer { return &Renderer{Surface: NewSurface(w, h)} }

func (r *Renderer) Render(a *App) *Surface {
	s := r.Surface
	a.SetViewport(s.W, s.H)
	r.drawBackground(a)
	l := ComputeLayout(s.W, s.H, a.Screen)
	switch a.Screen {
	case ScreenAttract:
		r.drawAttract(a, l)
	case ScreenNick:
		r.drawNick(a, l)
	case ScreenCountdown:
		r.drawPlay(a, l, true)
	case ScreenPlay:
		r.drawPlay(a, l, false)
	case ScreenSummary:
		r.drawSummary(a, l)
	case ScreenRanking:
		r.drawRanking(a, l)
	case ScreenAdminPIN:
		r.drawAdminPIN(a, l)
	case ScreenHistory:
		r.drawHistory(a, l)
	}
	if a.Screen != ScreenCountdown && a.Screen != ScreenPlay {
		r.drawTopBar(a, l)
	}
	r.drawTouchPulses(a)
	if a.StateAge < .28 {
		alpha := uint8(210 * (1 - a.StateAge/.28))
		s.FillRect(Rect{0, 0, s.W, s.H}, RGBA(2, 5, 18, alpha))
	}
	if msg := a.DiagnosticText(); msg != "" {
		r.drawDiagnostic(msg)
	}
	return s
}

func (r *Renderer) drawBackground(a *App) {
	s := r.Surface
	s.FillGradient(RGB(4, 9, 27), RGB(13, 17, 48))
	if a.Brand.StartBackground != nil && (a.Screen == ScreenAttract || a.Screen == ScreenSummary || a.Screen == ScreenRanking) {
		s.DrawSpriteCover(a.Brand.StartBackground, Rect{0, 0, s.W, s.H}, 92)
	}
	short := min(s.W, s.H)
	orbs := []struct {
		x, y, rr float64
		c        Color
	}{
		{.14, .26, .17, a.Brand.Primary},
		{.85, .22, .20, a.Brand.Secondary},
		{.55, .86, .24, a.Brand.Accent},
	}
	for i, o := range orbs {
		x := int(o.x*float64(s.W) + math.Sin(a.TotalTime*.35+float64(i))*float64(short)*.025)
		y := int(o.y*float64(s.H) + math.Cos(a.TotalTime*.28+float64(i))*float64(short)*.018)
		rad := int(o.rr * float64(short))
		for k := 5; k >= 1; k-- {
			s.FillCircle(x, y, rad*k/5, o.c.Alpha(uint8(4+k*3)))
		}
	}
	for i := 0; i < 42; i++ {
		x := (i*197 + int(a.TotalTime*14)*((i%3)+1)) % max(1, s.W)
		y := (i*113 + i*i*17) % max(1, s.H)
		rad := 1 + i%3
		s.FillCircle(x, y, rad, RGB(203, 245, 255).Alpha(uint8(40+i%5*12)))
	}
}

func (r *Renderer) drawTopBar(a *App, l Layout) {
	s := r.Surface
	s.FillRect(l.Header, RGB(3, 7, 22).Alpha(238))
	s.FillRect(Rect{0, l.Header.Bottom() - 2, s.W, 2}, a.Brand.Primary.Alpha(150))
	card := l.NanoLogo.Inset(-8)
	s.FillRoundRect(card, max(14, card.H/7), RGB(250, 252, 255))
	s.OutlineRoundRect(card, max(14, card.H/7), 2, a.Brand.Primary.Alpha(180))
	if a.Logo != nil {
		s.DrawSprite(a.Logo, l.NanoLogo.Inset(7), 255, true)
	} else {
		s.DrawTextCentered(l.NanoLogo, 28, "LOGO BRAK", RGB(12, 18, 35))
	}
	partnerCard := l.PartnerLogo.Inset(-7)
	s.FillRoundRect(partnerCard, max(14, partnerCard.H/7), RGB(9, 17, 41).Alpha(245))
	s.OutlineRoundRect(partnerCard, max(14, partnerCard.H/7), 2, a.Brand.Secondary.Alpha(180))
	if a.Brand.PartnerLogo != nil {
		s.DrawSprite(a.Brand.PartnerLogo, l.PartnerLogo.Inset(5), 255, true)
	} else {
		s.DrawTextCentered(l.PartnerLogo, 25, a.Brand.Config.Partner.Name, RGB(240, 248, 255))
	}
	center := Rect{l.NanoLogo.Right() + l.Margin, l.Header.Y, l.PartnerLogo.X - l.Margin - l.NanoLogo.Right() - l.Margin, l.Header.H}
	if center.W > 40 {
		s.DrawTextCentered(center.Inset(8), clamp(l.Header.H/5, 14, 25), a.Brand.Config.EventName, RGB(154, 178, 207))
	}
}

func (r *Renderer) drawAttract(a *App, l Layout) {
	s := r.Surface
	pulse := .5 + .5*math.Sin(a.TotalTime*2.3)
	titleH := clamp(int(float64(min(s.W, s.H))*.16), 112, 190)
	s.DrawTextCentered(l.Title, titleH, "BOMBER RUSH", RGB(246, 251, 255))
	accentW := min(l.Title.W, TextWidth("BOMBER RUSH", titleH)+80)
	x := l.Title.X + (l.Title.W-accentW)/2
	s.FillRoundRect(Rect{x, l.Title.Bottom() - 10, accentW, 7}, 4, Mix(a.Brand.Primary, a.Brand.Secondary, pulse))
	sub := fmt.Sprintf("%d SEKUND. NAJWYZSZY WYNIK.", a.Brand.Config.RoundSeconds)
	s.DrawTextCentered(Rect{l.Subtitle.X, l.Subtitle.Y, l.Subtitle.W, l.Subtitle.H / 2}, clamp(int(float64(min(s.W, s.H))*.045), 38, 64), sub, RGB(198, 218, 239))
	chipY := l.Subtitle.Y + l.Subtitle.H/2
	chipH := clamp(int(float64(min(s.W, s.H))*.08), 72, 118)
	gap := l.Margin
	chipW := min(520, (l.Subtitle.W-gap)/2)
	left := Rect{l.Subtitle.X + (l.Subtitle.W-(chipW*2+gap))/2, chipY, chipW, chipH}
	right := Rect{left.Right() + gap, chipY, chipW, chipH}
	r.drawInfoChip(left, "PRZESUN PALCEM", "RUCH O JEDNO POLE", a.Brand.Primary)
	r.drawInfoChip(right, "DOTKNIJ EKRANU", "USTAW BOMBE", a.Brand.Secondary)
	r.drawButton(l.PrimaryButton, "DOTKNIJ, ABY ZAGRAC", a.Brand.Primary, true, pulse)
	headline := a.Brand.Config.Partner.Headline
	s.DrawTextCentered(Rect{l.PrimaryButton.X, l.PrimaryButton.Y - l.Margin*2, l.PrimaryButton.W, l.Margin * 2}, clamp(l.Margin, 22, 38), headline, RGB(206, 222, 242))
}

func (r *Renderer) drawInfoChip(rect Rect, title, sub string, c Color) {
	s := r.Surface
	s.FillRoundRect(rect, rect.H/4, RGB(9, 17, 42).Alpha(235))
	s.OutlineRoundRect(rect, rect.H/4, 2, c.Alpha(190))
	iconR := rect.H / 5
	s.GlowCircle(rect.X+rect.H/2, rect.Y+rect.H/2, iconR, c.Alpha(220))
	textX := rect.X + rect.H
	s.DrawText(textX, rect.Y+rect.H/4, clamp(rect.H/4, 18, 32), title, RGB(246, 251, 255))
	s.DrawText(textX, rect.Y+rect.H*3/5, clamp(rect.H/7, 14, 24), sub, RGB(163, 188, 216))
}

func (r *Renderer) drawNick(a *App, l Layout) {
	s := r.Surface
	s.DrawTextCentered(l.Title, clamp(int(float64(min(s.W, s.H))*.075), 64, 105), "TWOJ NICK", RGB(246, 251, 255))
	r.drawCard(l.NickDisplay, a.Brand.Primary)
	nick := a.Nick
	if nick == "" {
		nick = "_"
	}
	s.DrawTextCentered(l.NickDisplay.Inset(l.Margin/2), clamp(l.NickDisplay.H/2, 45, 92), nick, a.Brand.Primary)
	for _, key := range l.KeyboardKeys {
		accent := RGB(24, 37, 66)
		if key.Value == "PLAY" {
			accent = a.Brand.Primary
		} else if key.Value == "BACK" {
			accent = a.Brand.Secondary
		}
		r.drawKey(key.Rect, key.Label, accent, key.Value == "PLAY")
	}
	r.drawSmallButton(l.BackButton, "WROC", RGB(116, 137, 165))
}

func (r *Renderer) drawPlay(a *App, l Layout, countdown bool) {
	if a.Game == nil {
		return
	}
	r.drawBoard(a, l.Board)
	r.drawHUD(a, l.HUD)
	if a.TutorialVisible && a.StateAge < 6 {
		r.drawPlayHint(a, l.Hint)
	}
	if countdown {
		s := r.Surface
		s.FillRoundRect(l.Board, max(22, l.Board.W/50), RGB(3, 7, 22).Alpha(152))
		value := "3"
		if a.StateAge >= 1.05 {
			value = "2"
		}
		if a.StateAge >= 2.10 {
			value = "1"
		}
		if a.StateAge >= 3.02 {
			value = "START"
		}
		pulse := .5 + .5*math.Sin(a.StateAge*8)
		s.GlowCircle(l.Board.X+l.Board.W/2, l.Board.Y+l.Board.H/2, min(l.Board.W, l.Board.H)/7, Mix(a.Brand.Primary, a.Brand.Secondary, pulse).Alpha(80))
		s.DrawTextCentered(l.Board.Inset(l.Board.W/7), min(l.Board.W, l.Board.H)/5, value, RGB(255, 255, 255))
	}
}

func (r *Renderer) drawBoard(a *App, rect Rect) {
	g := a.Game
	s := r.Surface
	shakeX, shakeY := 0, 0
	if g.Shake > 0 {
		shakeX = int(math.Sin(a.TotalTime*61) * g.Shake * 8)
		shakeY = int(math.Cos(a.TotalTime*53) * g.Shake * 6)
	}
	rect.X += shakeX
	rect.Y += shakeY
	corner := max(18, rect.W/55)
	s.FillRoundRect(rect.Inset(-8), corner+8, a.Brand.Primary.Alpha(45))
	s.FillRoundRect(rect, corner, RGB(5, 12, 31))
	s.OutlineRoundRect(rect, corner, 3, a.Brand.Primary.Alpha(180))
	cellW := float64(rect.W) / GridW
	cellH := float64(rect.H) / GridH
	cell := math.Min(cellW, cellH)
	gridW := int(cell * GridW)
	gridH := int(cell * GridH)
	ox := rect.X + (rect.W-gridW)/2
	oy := rect.Y + (rect.H-gridH)/2
	cellRect := func(x, y int) Rect {
		x0 := ox + int(float64(x)*cell)
		y0 := oy + int(float64(y)*cell)
		x1 := ox + int(float64(x+1)*cell)
		y1 := oy + int(float64(y+1)*cell)
		return Rect{x0, y0, x1 - x0, y1 - y0}
	}
	for y := 0; y < GridH; y++ {
		for x := 0; x < GridW; x++ {
			cr := cellRect(x, y)
			floor := RGB(10, 23, 50)
			if (x+y)%2 == 0 {
				floor = RGB(12, 28, 58)
			}
			s.FillRect(cr, floor)
			s.FillRect(Rect{cr.X + 2, cr.Y + 2, max(1, cr.W-4), 1}, RGB(55, 91, 126).Alpha(45))
			s.FillRect(Rect{cr.X, cr.Bottom() - 1, cr.W, 1}, RGB(37, 65, 93).Alpha(90))
			s.FillRect(Rect{cr.Right() - 1, cr.Y, 1, cr.H}, RGB(37, 65, 93).Alpha(70))
			switch g.Grid[y][x] {
			case Wall:
				in := max(3, cr.W/12)
				rr := cr.Inset(in)
				s.FillRoundRect(Rect{rr.X + in/2, rr.Y + in, rr.W, rr.H}, max(5, rr.W/6), RGB(2, 7, 20).Alpha(100))
				s.FillRoundRect(rr, max(5, rr.W/6), RGB(24, 46, 80))
				s.FillRoundRect(Rect{rr.X + rr.W/8, rr.Y + rr.H/9, rr.W * 3 / 4, rr.H / 5}, rr.H/10, RGB(70, 111, 151).Alpha(150))
				s.OutlineRoundRect(rr, max(5, rr.W/6), max(1, rr.W/40), a.Brand.Primary.Alpha(80))
			case Crate:
				in := max(4, cr.W/10)
				rr := cr.Inset(in)
				s.FillRoundRect(Rect{rr.X + in/2, rr.Y + in, rr.W, rr.H}, max(5, rr.W/7), RGB(2, 7, 20).Alpha(120))
				s.FillRoundRect(rr, max(5, rr.W/7), RGB(94, 50, 37))
				s.OutlineRoundRect(rr, max(5, rr.W/7), max(2, rr.W/28), a.Brand.Accent.Alpha(230))
				s.Line(rr.X+rr.W/5, rr.Y+rr.H/5, rr.Right()-rr.W/5, rr.Bottom()-rr.H/5, max(2, rr.W/18), a.Brand.Accent.Alpha(125))
				s.Line(rr.Right()-rr.W/5, rr.Y+rr.H/5, rr.X+rr.W/5, rr.Bottom()-rr.H/5, max(2, rr.W/18), a.Brand.Accent.Alpha(125))
			}
		}
	}
	// pickups
	for _, p := range g.Pickups {
		cr := cellRect(p.X, p.Y)
		cx, cy := cr.X+cr.W/2, cr.Y+cr.H/2
		pulse := .82 + .14*math.Sin(a.TotalTime*5+p.Age)
		rad := int(float64(min(cr.W, cr.H)) * .27 * pulse)
		s.GlowCircle(cx, cy, rad, a.Brand.Primary.Alpha(58))
		s.Ring(cx, cy, rad+max(3, rad/3), 1, a.Brand.Primary.Alpha(35))
		switch p.Type {
		case PickupPartner:
			if a.Brand.ProductItem != nil {
				s.DrawSprite(a.Brand.ProductItem, Rect{cx - rad, cy - rad, rad * 2, rad * 2}, 255, true)
			} else {
				s.FillRoundRect(Rect{cx - rad, cy - rad, rad * 2, rad * 2}, rad/4, a.Brand.Secondary)
			}
		case PickupRange:
			s.FillCircle(cx, cy, rad, a.Brand.Accent)
			s.DrawTextCentered(Rect{cx - rad, cy - rad, rad * 2, rad * 2}, rad, "+R", RGB(8, 15, 31))
		case PickupBomb:
			s.FillCircle(cx, cy, rad, RGB(19, 25, 42))
			s.Ring(cx, cy, rad, max(2, rad/5), a.Brand.Primary)
			s.DrawTextCentered(Rect{cx - rad, cy - rad, rad * 2, rad * 2}, rad, "+B", RGB(255, 255, 255))
		case PickupShield:
			s.Ring(cx, cy, rad, max(3, rad/4), a.Brand.Primary)
			s.Ring(cx, cy, rad/2, max(2, rad/5), RGB(255, 255, 255))
		}
	}
	// bombs
	for _, b := range g.Bombs {
		cr := cellRect(b.X, b.Y)
		cx, cy := cr.X+cr.W/2, cr.Y+cr.H/2
		fraction := clampFloat(b.Fuse/BombFuse, 0, 1)
		pulse := 1 + .08*math.Sin((BombFuse-b.Fuse)*15)
		rad := int(float64(min(cr.W, cr.H)) * .29 * pulse)
		s.GlowCircle(cx, cy, rad, RGB(255, 72, 96).Alpha(uint8(30+int((1-fraction)*70))))
		s.FillCircle(cx, cy, rad, RGB(8, 11, 23))
		s.Ring(cx, cy, rad, max(2, rad/7), Mix(a.Brand.Primary, RGB(255, 72, 96), 1-fraction))
		s.FillCircle(cx-rad/3, cy-rad/3, max(2, rad/8), RGB(255, 255, 255).Alpha(190))
		if fraction < .45 {
			s.Ring(cx, cy, rad+max(4, rad/3), max(2, rad/10), RGB(255, 72, 96).Alpha(150))
		}
		fuseX := cx + rad*2/3
		fuseY := cy - rad*2/3
		s.Line(cx+rad/3, cy-rad/2, fuseX, fuseY, max(2, rad/7), RGB(255, 181, 71))
		s.FillCircle(fuseX, fuseY, max(3, rad/7), RGB(255, 238, 165))
	}
	// explosions
	for _, e := range g.Explosions {
		progress := 1 - e.TTL/e.Duration
		alpha := uint8(230 * (1 - progress))
		for _, p := range e.Cells {
			cr := cellRect(p.X, p.Y).Inset(max(1, int(cell*.04)))
			s.FillRoundRect(cr, max(4, cr.W/4), RGB(255, 108, 43).Alpha(alpha))
			inner := cr.Inset(max(2, cr.W/5))
			s.FillRoundRect(inner, max(2, inner.W/4), RGB(255, 244, 171).Alpha(uint8(min(255, int(alpha)+25))))
		}
	}
	// enemies
	for _, e := range g.Enemies {
		if !e.Alive {
			continue
		}
		x, y := e.Visual()
		cx := ox + int((x+.5)*cell)
		cy := oy + int((y+.5)*cell)
		cy -= int(math.Abs(math.Sin(a.TotalTime*7+e.PulseOffset)) * cell * .045)
		rad := int(cell * (.27 + .025*math.Sin(a.TotalTime*6+e.PulseOffset)))
		s.GlowCircle(cx, cy, rad, a.Brand.Secondary.Alpha(70))
		unit := max(2, rad/6)
		bodyW, bodyH := rad*2, rad*2
		body := Rect{cx - bodyW/2, cy - rad/3, bodyW, bodyH}
		footW, footH := max(3, rad/2), max(3, rad/3)
		s.FillRoundRect(Rect{body.X + bodyW/8, body.Bottom() - footH/2, footW, footH}, unit, RGB(55, 13, 65))
		s.FillRoundRect(Rect{body.Right() - bodyW/8 - footW, body.Bottom() - footH/2, footW, footH}, unit, RGB(55, 13, 65))
		s.FillRoundRect(body, max(3, rad/2), RGB(86, 19, 85))
		s.FillRoundRect(Rect{body.X + bodyW/5, body.Y + bodyH/5, bodyW * 3 / 5, bodyH / 2}, unit, a.Brand.Secondary)
		s.FillRect(Rect{body.X + bodyW/3, body.Y + bodyH/5, max(1, bodyW/8), bodyH / 2}, RGB(255, 160, 236).Alpha(145))
		s.FillRoundRect(Rect{body.X - unit, body.Y + bodyH/2, unit * 2, max(3, rad/2)}, unit, RGB(67, 15, 74))
		s.FillRoundRect(Rect{body.Right() - unit, body.Y + bodyH/2, unit * 2, max(3, rad/2)}, unit, RGB(67, 15, 74))
		// Rogi i szerokie oczy dają przeciwnikowi czytelną sylwetkę potworka.
		s.Line(cx-rad/2, body.Y+unit, cx-rad/3, body.Y-rad/2, max(2, unit), a.Brand.Secondary)
		s.Line(cx+rad/2, body.Y+unit, cx+rad/3, body.Y-rad/2, max(2, unit), a.Brand.Secondary)
		eyeR := max(2, rad/6)
		s.FillCircle(cx-rad/3, cy-rad/10, eyeR, RGB(255, 255, 255))
		s.FillCircle(cx+rad/3, cy-rad/10, eyeR, RGB(255, 255, 255))
		s.FillCircle(cx-rad/3, cy-rad/10, max(1, eyeR/2), RGB(25, 10, 45))
		s.FillCircle(cx+rad/3, cy-rad/10, max(1, eyeR/2), RGB(25, 10, 45))
		s.FillRoundRect(Rect{cx - rad/3, cy + rad/3, rad * 2 / 3, max(2, rad/8)}, unit, RGB(255, 181, 71))
		s.Ring(cx, cy, rad, max(2, rad/8), a.Brand.Secondary.Alpha(190))
	}
	// player
	px, py := g.Player.Visual()
	pcx := ox + int((px+.5)*cell)
	pcy := oy + int((py+.5)*cell)
	if g.Player.IsMoving() {
		pcy -= int(math.Abs(math.Sin(a.TotalTime*30)) * cell * .06)
	}
	pr := int(cell * .31)
	visible := g.Player.Invulnerable <= 0 || int(g.Player.Invulnerable*10)%2 == 0
	if visible {
		s.GlowCircle(pcx, pcy, pr, a.Brand.Primary.Alpha(76))
		// Charakter ma czytelny profil Bombermana nawet przy małym rozmiarze pola.
		unit := max(2, pr/7)
		bodyW, bodyH := pr*4/3, pr*5/3
		body := Rect{pcx - bodyW/2, pcy - pr/8, bodyW, bodyH}
		legW, legH := max(3, bodyW/3), max(3, pr/2)
		s.FillRoundRect(Rect{body.X + bodyW/8, body.Bottom() - legH/2, legW, legH}, unit, RGB(18, 35, 70))
		s.FillRoundRect(Rect{body.Right() - bodyW/8 - legW, body.Bottom() - legH/2, legW, legH}, unit, RGB(18, 35, 70))
		s.FillRoundRect(body, max(2, pr/4), RGB(28, 91, 132))
		s.FillRoundRect(Rect{body.X + bodyW/5, body.Y + bodyH/8, bodyW * 3 / 5, bodyH / 3}, unit, a.Brand.Primary)
		s.FillRect(Rect{body.X + bodyW/3, body.Y + bodyH/8, max(1, bodyW/7), bodyH / 3}, RGB(210, 255, 255).Alpha(150))
		armY := body.Y + bodyH/3
		s.FillRoundRect(Rect{body.X - bodyW/5, armY, max(3, bodyW/4), max(3, pr/2)}, unit, RGB(23, 68, 105))
		s.FillRoundRect(Rect{body.Right() - bodyW/20, armY, max(3, bodyW/4), max(3, pr/2)}, unit, RGB(23, 68, 105))

		headR := pr * 3 / 4
		head := Rect{pcx - headR, pcy - pr - headR/2, headR * 2, headR * 2}
		s.FillRoundRect(head, max(2, pr/4), RGB(255, 214, 166))
		// Hełm, daszek i boczne nauszniki budują charakterystyczną sylwetkę.
		s.FillRoundRect(Rect{head.X - unit, head.Y - unit, head.W + unit*2, head.H * 3 / 5}, max(2, pr/3), RGB(245, 181, 55))
		s.FillRoundRect(Rect{head.X + head.W/5, head.Y + head.H/5, head.W * 3 / 5, max(2, head.H/5)}, unit, RGB(255, 235, 143))
		s.FillRoundRect(Rect{head.X - unit*2, head.Y + head.H/3, unit * 3, head.H / 3}, unit, RGB(222, 145, 34))
		s.FillRoundRect(Rect{head.Right() - unit, head.Y + head.H/3, unit * 3, head.H / 3}, unit, RGB(222, 145, 34))
		faceY := head.Y + head.H*3/5
		lookX := 0
		if g.Player.LastDirection.X != 0 {
			lookX = g.Player.LastDirection.X * max(1, pr/10)
		}
		eyeR := max(2, pr/10)
		s.FillCircle(pcx-headR/2+lookX, faceY, eyeR, RGB(17, 24, 44))
		s.FillCircle(pcx+headR/2+lookX, faceY, eyeR, RGB(17, 24, 44))
		s.FillRoundRect(Rect{pcx - headR/3, faceY + headR/3, headR * 2 / 3, max(2, pr/10)}, unit, RGB(206, 74, 85))
		s.Ring(pcx, pcy, pr, max(2, pr/10), a.Brand.Primary.Alpha(170))
		if g.Player.Shield {
			s.Ring(pcx, pcy, pr+max(5, pr/4), max(2, pr/10), a.Brand.Primary.Alpha(210))
		}
	}
	// particles
	for _, p := range g.Particles {
		alpha := uint8(255 * clampFloat(p.Life/p.Max, 0, 1))
		cx := ox + int(p.X*cell)
		cy := oy + int(p.Y*cell)
		size := max(2, int(p.Size*cell*(1+(1-p.Life/p.Max)*2)))
		if p.Ring {
			s.Ring(cx, cy, size*3, max(2, size/2), p.Color.Alpha(alpha))
		} else {
			s.FillCircle(cx, cy, size, p.Color.Alpha(alpha))
		}
	}
	for _, f := range g.Floating {
		alpha := uint8(255 * clampFloat(f.Life/f.Max, 0, 1))
		x := ox + int(f.X*cell)
		y := oy + int(f.Y*cell)
		h := clamp(int(cell*.25), 16, 32)
		w := TextWidth(f.Text, h)
		s.DrawTextShadow(x-w/2, y, h, f.Text, f.Color.Alpha(alpha))
	}
	if g.Flash > 0 {
		s.FillRoundRect(rect, corner, RGB(255, 255, 255).Alpha(uint8(g.Flash*70)))
	}
}

func (r *Renderer) drawHUD(a *App, rect Rect) {
	g := a.Game
	if g == nil {
		return
	}
	r.drawCard(rect, a.Brand.Primary)
	pad := max(14, rect.W/24)
	inner := rect.Inset(pad)
	if rect.W > rect.H {
		gap := pad / 2
		cardW := (inner.W - gap*2) / 3
		items := []struct {
			label, value string
			c            Color
		}{
			{"WYNIK", formatScore(g.Score), a.Brand.Primary},
			{"CZAS", fmt.Sprintf("%02d", int(math.Ceil(g.Remaining()))), a.Brand.Accent},
			{"ZYCIA", itoa(g.Player.Lives), RGB(255, 86, 112)},
		}
		for i, it := range items {
			r.drawStatCard(Rect{inner.X + i*(cardW+gap), inner.Y, cardW, inner.H}, it.label, it.value, it.c)
		}
		r.drawTimeBar(a, Rect{inner.X, inner.Bottom() - 5, inner.W, 5})
		return
	}
	rowH := (inner.H - pad) / 3
	r.drawStatCard(Rect{inner.X, inner.Y, inner.W, rowH}, "WYNIK", formatScore(g.Score), a.Brand.Primary)
	r.drawStatCard(Rect{inner.X, inner.Y + rowH + pad/2, inner.W, rowH}, "CZAS", fmt.Sprintf("%02d", int(math.Ceil(g.Remaining()))), a.Brand.Accent)
	r.drawStatCard(Rect{inner.X, inner.Y + 2*(rowH+pad/2), inner.W, rowH}, "ZYCIA", itoa(g.Player.Lives), RGB(255, 86, 112))
	r.drawTimeBar(a, Rect{inner.X, inner.Bottom() - 5, inner.W, 5})
}

func (r *Renderer) drawTimeBar(a *App, rect Rect) {
	remaining := clampFloat(a.Game.Remaining()/a.Game.RoundSeconds, 0, 1)
	c := a.Brand.Primary
	if a.Game.Remaining() <= 10 {
		c = Mix(RGB(255, 86, 112), a.Brand.Accent, .25+.25*math.Sin(a.TotalTime*8))
	}
	r.Surface.FillRoundRect(rect, rect.H/2, RGB(31, 43, 68))
	fill := rect
	fill.W = max(1, int(float64(rect.W)*remaining))
	r.Surface.FillRoundRect(fill, rect.H/2, c)
}

func (r *Renderer) drawPlayHint(a *App, rect Rect) {
	if rect.W <= 0 || rect.H <= 0 {
		return
	}
	s := r.Surface
	s.FillRoundRect(rect, max(14, rect.H/5), RGB(8, 16, 38).Alpha(225))
	s.OutlineRoundRect(rect, max(14, rect.H/5), 2, a.Brand.Primary.Alpha(150))
	s.DrawTextCentered(rect.Inset(10), clamp(rect.H/4, 16, 28), "PRZESUN: RUCH   |   DOTKNIJ: BOMBA", RGB(226, 241, 255))
}

func (r *Renderer) drawStatCard(rect Rect, label, value string, c Color) {
	s := r.Surface
	s.FillRoundRect(rect, max(12, rect.H/5), RGB(13, 24, 50))
	s.OutlineRoundRect(rect, max(12, rect.H/5), 2, c.Alpha(130))
	s.FillRoundRect(Rect{rect.X + rect.W/8, rect.Y + 3, rect.W * 3 / 4, 3}, 2, c.Alpha(180))
	s.DrawTextCentered(Rect{rect.X + 8, rect.Y + rect.H/10, rect.W - 16, rect.H / 3}, clamp(rect.H/6, 15, 28), label, RGB(151, 178, 207))
	valueH := rect.H * 3 / 5
	if label == "CZAS" {
		valueH = rect.H * 2 / 3
	}
	s.DrawTextCentered(Rect{rect.X + 8, rect.Y + rect.H/3, rect.W - 16, valueH}, clamp(rect.H/3, 30, 58), value, c)
}

func (r *Renderer) drawSponsorPanel(a *App, rect Rect, summary bool) {
	s := r.Surface
	r.drawCard(rect, a.Brand.Secondary)
	pad := max(14, min(rect.W, rect.H)/18)
	inner := rect.Inset(pad)
	if summary {
		qrSize := clamp(min(inner.W, inner.H)/3, 120, 320)
		gap := max(12, pad/2)
		adRect := inner
		if inner.W*10 > inner.H*12 {
			adRect.W = max(1, inner.W-qrSize-gap)
			qrRect := Rect{adRect.Right() + gap, inner.Y + (inner.H-qrSize)/2, qrSize, qrSize}
			if a.Brand.SummaryAd != nil {
				s.DrawSprite(a.Brand.SummaryAd, adRect, 255, true)
			} else {
				s.DrawTextCentered(Rect{adRect.X, adRect.Y, adRect.W, adRect.H / 2}, clamp(adRect.H/8, 24, 52), a.Brand.Config.Partner.SummaryTitle, RGB(255, 255, 255))
				s.DrawParagraph(Rect{adRect.X + pad, adRect.Y + adRect.H/2, adRect.W - 2*pad, adRect.H/2 - pad}, clamp(adRect.H/14, 16, 28), 8, a.Brand.Config.Partner.SummaryText, RGB(193, 213, 235))
			}
			if a.Brand.QR != nil {
				s.FillRoundRect(qrRect.Inset(-8), 18, RGB(255, 255, 255))
				s.DrawSprite(a.Brand.QR, qrRect, 255, true)
			}
		} else {
			qrSize = min(qrSize, inner.H/3)
			adRect.H = max(1, inner.H-qrSize-gap)
			qrRect := Rect{inner.X + (inner.W-qrSize)/2, adRect.Bottom() + gap, qrSize, qrSize}
			if a.Brand.SummaryAd != nil {
				s.DrawSprite(a.Brand.SummaryAd, adRect, 255, true)
			} else {
				s.DrawTextCentered(Rect{adRect.X, adRect.Y, adRect.W, adRect.H / 2}, clamp(adRect.H/8, 24, 52), a.Brand.Config.Partner.SummaryTitle, RGB(255, 255, 255))
				s.DrawParagraph(Rect{adRect.X + pad, adRect.Y + adRect.H/2, adRect.W - 2*pad, adRect.H/2 - pad}, clamp(adRect.H/14, 16, 28), 8, a.Brand.Config.Partner.SummaryText, RGB(193, 213, 235))
			}
			if a.Brand.QR != nil {
				s.FillRoundRect(qrRect.Inset(-8), 18, RGB(255, 255, 255))
				s.DrawSprite(a.Brand.QR, qrRect, 255, true)
			}
		}
		return
	}
	logoH := int(float64(inner.H) * .42)
	if rect.W*10 > rect.H*18 {
		logoH = inner.H
	}
	if a.Brand.PartnerLogo != nil {
		s.DrawSprite(a.Brand.PartnerLogo, Rect{inner.X, inner.Y, inner.W, logoH}, 255, true)
	} else {
		s.DrawTextCentered(Rect{inner.X, inner.Y, inner.W, logoH}, clamp(logoH/4, 24, 50), a.Brand.Config.Partner.Name, RGB(255, 255, 255))
	}
	if logoH < inner.H {
		s.DrawParagraph(Rect{inner.X, inner.Y + logoH + pad/2, inner.W, inner.H - logoH - pad/2}, clamp(pad, 18, 30), 8, a.Brand.Config.Partner.Headline, RGB(200, 219, 239))
	}
}

func (r *Renderer) drawSummary(a *App, l Layout) {
	s := r.Surface
	r.drawCard(l.Table, a.Brand.Primary)
	inner := l.Table.Inset(max(22, l.Margin))
	rowH := inner.H / 6
	s.DrawTextCentered(Rect{inner.X, inner.Y, inner.W, rowH}, clamp(rowH/2, 42, 82), "KONIEC RUNDY", RGB(246, 251, 255))
	s.DrawTextCentered(Rect{inner.X, inner.Y + rowH, inner.W, rowH * 2}, clamp(rowH, 72, 150), formatScore(a.LastEntry.Score), a.Brand.Primary)
	s.DrawTextCentered(Rect{inner.X, inner.Y + rowH*3, inner.W, rowH / 2}, clamp(rowH/4, 18, 30), "WYNIK", RGB(156, 183, 212))
	statsY := inner.Y + rowH*4
	gap := l.Margin / 2
	statW := (inner.W - gap*2) / 3
	place := "-"
	if a.LastPosition > 0 {
		place = "#" + itoa(a.LastPosition)
	}
	stats := []struct {
		label, value string
		c            Color
	}{
		{"MIEJSCE", place, a.Brand.Secondary},
		{"SERIA", "X" + itoa(max(1, a.LastEntry.BestCombo)), a.Brand.Accent},
		{"CZAS", formatDuration(a.LastEntry.SurvivalMS), RGB(255, 255, 255)},
	}
	for i, st := range stats {
		r.drawStatCard(Rect{inner.X + i*(statW+gap), statsY, statW, inner.Bottom() - statsY}, st.label, st.value, st.c)
	}
	r.drawSponsorPanel(a, l.Sponsor, true)
	r.drawButton(l.PrimaryButton, "POKAZ RANKING", a.Brand.Primary, false, 0)
	r.drawButton(l.SecondaryButton, "ZAGRAJ PONOWNIE", a.Brand.Secondary, false, 0)
}

func (r *Renderer) drawRanking(a *App, l Layout) {
	activeOne := a.RankingMode == RankingEvent
	r.drawTab(l.TabOne, "WYDARZENIE", activeOne, a.Brand.Primary)
	r.drawTab(l.TabTwo, "DZISIAJ", !activeOne, a.Brand.Secondary)
	r.drawSmallButton(l.AdminButton, "HISTORIA", RGB(116, 137, 165))
	r.drawRankingTable(a, l.Table, a.RankingEntries(10))
	r.drawSponsorPanel(a, l.Sponsor, false)
	r.drawButton(l.PrimaryButton, "NOWA GRA", a.Brand.Primary, false, 0)
}

func (r *Renderer) drawRankingTable(a *App, rect Rect, entries []ScoreEntry) {
	s := r.Surface
	r.drawCard(rect, a.Brand.Primary)
	pad := max(14, rect.W/45)
	inner := rect.Inset(pad)
	headerH := clamp(inner.H/11, 46, 74)
	cols := []int{int(float64(inner.W) * .08), int(float64(inner.W) * .25), int(float64(inner.W) * .22), int(float64(inner.W) * .20)}
	cols = append(cols, inner.W-cols[0]-cols[1]-cols[2]-cols[3])
	labels := []string{"#", "NICK", "WYNIK", "CZAS", "DATA"}
	x := inner.X
	for i, label := range labels {
		s.DrawTextCentered(Rect{x, inner.Y, cols[i], headerH}, clamp(headerH/3, 16, 25), label, RGB(147, 174, 204))
		x += cols[i]
	}
	rowH := (inner.H - headerH) / 10
	for i := 0; i < 10; i++ {
		y := inner.Y + headerH + i*rowH
		row := Rect{inner.X, y, inner.W, rowH - 3}
		bg := RGB(12, 23, 49).Alpha(210)
		if i < 3 {
			bg = Mix(a.Brand.Primary.Alpha(45), a.Brand.Secondary.Alpha(35), float64(i)/2)
		}
		s.FillRoundRect(row, max(8, row.H/5), bg)
		if i >= len(entries) {
			continue
		}
		e := entries[i]
		values := []string{itoa(i + 1), e.Nick, formatScore(e.Score), formatDuration(e.SurvivalMS), e.CreatedAt.Format("02.01 15:04")}
		x = inner.X
		for ci, v := range values {
			c := RGB(231, 241, 252)
			if ci == 0 && i < 3 {
				c = []Color{a.Brand.Accent, RGB(210, 225, 239), RGB(214, 132, 89)}[i]
			}
			s.DrawTextCentered(Rect{x + 4, y, cols[ci] - 8, rowH - 3}, clamp(rowH/3, 15, 26), v, c)
			x += cols[ci]
		}
	}
}

func (r *Renderer) drawAdminPIN(a *App, l Layout) {
	r.drawCard(l.Table, a.Brand.Secondary)
	s := r.Surface
	inner := l.Table.Inset(l.Margin)
	s.DrawTextCentered(Rect{inner.X, inner.Y, inner.W, int(float64(inner.H) * .10)}, clamp(int(float64(inner.H)*.045), 36, 64), "PANEL ADMINISTRATORA", RGB(246, 251, 255))
	msg := "WPROWADZ PIN"
	if a.AdminPINError {
		msg = "BLEDNY PIN"
	}
	c := RGB(157, 183, 210)
	if a.AdminPINError {
		c = RGB(255, 86, 112)
	}
	s.DrawTextCentered(Rect{inner.X, inner.Y + int(float64(inner.H)*.09), inner.W, int(float64(inner.H) * .07)}, clamp(int(float64(inner.H)*.025), 21, 35), msg, c)
	dots := ""
	for range a.AdminPINInput {
		dots += "#"
	}
	if dots == "" {
		dots = "----"
	}
	s.DrawTextCentered(Rect{inner.X, inner.Y + int(float64(inner.H)*.15), inner.W, int(float64(inner.H) * .09)}, clamp(int(float64(inner.H)*.055), 42, 72), dots, a.Brand.Primary)
	for _, key := range l.PinKeys {
		accent := RGB(25, 38, 67)
		if key.Value == "OK" {
			accent = a.Brand.Primary
		} else if key.Value == "BACK" {
			accent = a.Brand.Secondary
		}
		r.drawKey(key.Rect, key.Label, accent, key.Value == "OK")
	}
	r.drawSmallButton(l.BackButton, "WROC", RGB(116, 137, 165))
}

func (r *Renderer) drawHistory(a *App, l Layout) {
	r.drawSmallButton(l.BackButton, "WROC", RGB(116, 137, 165))
	r.drawSmallButton(l.PagePrev, "WSTECZ", a.Brand.Secondary)
	r.drawSmallButton(l.PageNext, "DALEJ", a.Brand.Primary)
	s := r.Surface
	r.drawCard(l.Table, a.Brand.Secondary)
	entries := a.HistoryEntries()
	start := a.HistoryPage * 12
	end := min(len(entries), start+12)
	pad := max(16, l.Table.W/45)
	inner := l.Table.Inset(pad)
	headerH := clamp(inner.H/14, 44, 72)
	s.DrawTextCentered(Rect{inner.X, inner.Y, inner.W, headerH}, clamp(headerH/2, 26, 44), fmt.Sprintf("PELNA HISTORIA - STRONA %d", a.HistoryPage+1), RGB(246, 251, 255))
	rowH := (inner.H - headerH) / 12
	for i := 0; i < 12; i++ {
		y := inner.Y + headerH + i*rowH
		row := Rect{inner.X, y, inner.W, rowH - 3}
		s.FillRoundRect(row, max(6, row.H/6), RGB(13, 24, 50).Alpha(220))
		idx := start + i
		if idx >= end {
			continue
		}
		e := entries[idx]
		date := e.CreatedAt.Format("02.01 15:04")
		text := fmt.Sprintf("%s  %s  %s  %s", e.Nick, formatScore(e.Score), formatDuration(e.SurvivalMS), date)
		s.DrawText(row.X+14, row.Y+(row.H-clamp(row.H/3, 14, 25))/2, clamp(row.H/3, 14, 25), text, RGB(221, 234, 247))
	}
}

func (r *Renderer) drawCard(rect Rect, accent Color) {
	s := r.Surface
	s.FillRoundRect(rect.Inset(-4), max(20, rect.W/45), accent.Alpha(35))
	s.FillRoundRect(rect, max(18, rect.W/50), RGB(7, 15, 37).Alpha(246))
	s.OutlineRoundRect(rect, max(18, rect.W/50), 2, accent.Alpha(130))
}

func (r *Renderer) drawButton(rect Rect, label string, c Color, pulse bool, p float64) {
	s := r.Surface
	glow := uint8(40)
	if pulse {
		glow = uint8(40 + p*50)
	}
	s.FillRoundRect(rect.Inset(-max(5, rect.H/14)), max(18, rect.H/3), c.Alpha(glow))
	bg := Mix(RGB(13, 29, 57), c, .28)
	s.FillRoundRect(rect, max(18, rect.H/3), bg)
	s.OutlineRoundRect(rect, max(18, rect.H/3), max(2, rect.H/35), c)
	s.DrawTextCentered(rect.Inset(rect.H/5), clamp(rect.H/3, 26, 56), label, RGB(251, 254, 255))
}

func (r *Renderer) drawSmallButton(rect Rect, label string, c Color) {
	s := r.Surface
	s.FillRoundRect(rect, max(12, rect.H/3), RGB(15, 27, 52).Alpha(245))
	s.OutlineRoundRect(rect, max(12, rect.H/3), 2, c.Alpha(190))
	s.DrawTextCentered(rect.Inset(8), clamp(rect.H/3, 17, 30), label, RGB(231, 241, 252))
}

func (r *Renderer) drawKey(rect Rect, label string, c Color, active bool) {
	s := r.Surface
	bg := RGB(17, 29, 55)
	outline := c.Alpha(130)
	if active {
		bg = Mix(bg, c, .28)
		outline = c
	}
	s.FillRoundRect(rect, max(10, rect.H/5), bg)
	s.OutlineRoundRect(rect, max(10, rect.H/5), max(1, rect.H/40), outline)
	fg := RGB(235, 244, 253)
	if active {
		fg = RGB(255, 255, 255)
	}
	s.DrawTextCentered(rect.Inset(max(5, rect.H/8)), clamp(rect.H/3, 20, 42), label, fg)
}

func (r *Renderer) drawTab(rect Rect, label string, active bool, c Color) {
	s := r.Surface
	bg := RGB(15, 27, 52)
	fg := RGB(145, 170, 199)
	if active {
		bg = Mix(bg, c, .24)
		fg = RGB(251, 254, 255)
	}
	s.FillRoundRect(rect, max(12, rect.H/3), bg)
	s.OutlineRoundRect(rect, max(12, rect.H/3), 2, c.Alpha(func() uint8 {
		if active {
			return 255
		}
		return 90
	}()))
	s.DrawTextCentered(rect.Inset(8), clamp(rect.H/3, 17, 31), label, fg)
}

func (r *Renderer) drawTouchPulses(a *App) {
	s := r.Surface
	for _, p := range a.Pulses {
		progress := 1 - p.Life/p.Max
		rad := int(float64(min(s.W, s.H)) * (.012 + progress*.045))
		alpha := uint8(180 * (1 - progress))
		c := a.Brand.Primary
		if !p.Good {
			c = RGB(255, 86, 112)
		}
		s.Ring(int(p.X), int(p.Y), rad, max(2, rad/12), c.Alpha(alpha))
		if p.Good {
			s.Ring(int(p.X), int(p.Y), max(2, rad/2), max(2, rad/14), c.Alpha(alpha/2))
		} else {
			arm := max(5, rad/3)
			s.Line(int(p.X)-arm, int(p.Y)-arm, int(p.X)+arm, int(p.Y)+arm, max(2, rad/10), c.Alpha(alpha))
			s.Line(int(p.X)+arm, int(p.Y)-arm, int(p.X)-arm, int(p.Y)+arm, max(2, rad/10), c.Alpha(alpha))
		}
	}
}

func (r *Renderer) drawDiagnostic(msg string) {
	s := r.Surface
	h := max(34, min(s.W, s.H)/35)
	box := Rect{10, s.H - h - 10, s.W - 20, h}
	s.FillRoundRect(box, h/3, RGB(75, 19, 30).Alpha(230))
	s.DrawTextCentered(box.Inset(8), clamp(h/3, 12, 22), msg, RGB(255, 206, 214))
}

func formatScore(v int) string {
	if v < 1000 {
		return itoa(v)
	}
	s := itoa(v)
	out := ""
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += " "
		}
		out += string(r)
	}
	return out
}

func formatDuration(ms int64) string {
	seconds := ms / 1000
	tenths := (ms % 1000) / 100
	return fmt.Sprintf("%02d.%d S", seconds, tenths)
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
