package platform

import (
	"math"
	"strings"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/img"
	. "github.com/glupi-borna/soko/internal/utils"
)

func (p *platform) SetColor(c sdl.Color) {
	p.Renderer.SetDrawColor(c.R, c.G, c.B, c.A)
}

func (p *platform) RawSetFont(font *Font) {
	p.Font = font
}

func (p *platform) SetFont(font_name string, font_size int) {
	font := GetFont(font_name, font_size)
	p.Font = font
}

func (p *platform) DrawRectOutlined(x, y, w, h float32) {
	p.Renderer.DrawRectF(&sdl.FRect{x, y, w, h})
}

func (p *platform) DrawRectFilled(x, y, w, h float32) {
	p.Renderer.FillRectF(&sdl.FRect{x, y, w, h})
}

func (p *platform) DrawText(text string, x, y float32) {
	r, g, b, a, _ := p.Renderer.GetDrawColor()
	c := sdl.Color{r, g, b, a}

	tex := p.GetTextTexture(p.Font, text, c)
	_, _, w, h, _ := tex.Query()

	p.Renderer.CopyF(tex, nil, &sdl.FRect{
		float32(math.Round(float64(x))),
		float32(math.Round(float64(y))),
		float32(w), float32(h),
	})
}

var ARCPOINTS = make([]sdl.FPoint, 0)

func maybe_resize[K any](arr *[]K, size int) {
	if cap(*arr) < size {
		*arr = make([]K, size*2)
	} else {
		*arr = (*arr)[:size]
	}
}

func ArcPoints(x, y, r, startAngle, endAngle float32) []sdl.FPoint {
	CORNER_POINTS := cornerPointsCount(r)

	maybe_resize(&ARCPOINTS, CORNER_POINTS)

	angleRange := float64(Abs(startAngle - endAngle))
	start := float64(startAngle)
	rad := float64(r)

	imax := float64(CORNER_POINTS-1)

	for i:= 0; i<CORNER_POINTS; i++ {
		frac := float64(i)/imax
		angle := frac*angleRange + start
		xoff := float32(math.Cos(angle) * rad)
		yoff := float32(math.Sin(angle) * rad)
		ARCPOINTS[i].X = x+xoff
		ARCPOINTS[i].Y = y+yoff
	}

	return ARCPOINTS
}

var RRPOINTS = make([]sdl.FPoint, 0)

func RoundRectPoints(x, y, w, h, r float32) []sdl.FPoint {
	r = Clamp(r, 0, Min(w/2, h/2))

	// CORNER_POINTS := cornerPointsCount(r)
	// TOTAL_POINTS := CORNER_POINTS * 4

	x1 := x+r
	y1 := y+r
	x2 := x+w-r
	y2 := y+h-r

	RRPOINTS = RRPOINTS[:0]
	const pihalf = math.Pi/2
	RRPOINTS = append(RRPOINTS, ArcPoints(x1, y1, r, pihalf*2, pihalf*3)...)
	RRPOINTS = append(RRPOINTS, ArcPoints(x2, y1, r, pihalf*3, pihalf*4)...)
	RRPOINTS = append(RRPOINTS, ArcPoints(x2, y2, r, 0,        pihalf)...)
	RRPOINTS = append(RRPOINTS, ArcPoints(x1, y2, r, pihalf,   pihalf*2)...)

	return RRPOINTS
}

// Calculates ideal number of points for 90 degree
// arc depending on the radius.
// Used for drawing rounded rectangles.
func cornerPointsCount(r float32) int {
	const MIN_SEGMENT_LENGTH = 0.5
	// The formula is
	//     2*pi*r*(angle/360)
	// Since the angle is always 90, we can
	// simplify (2 * (90/360)) to 0.5
	arc_length := 0.5 * math.Pi * r
	return int(Max(arc_length / MIN_SEGMENT_LENGTH, 2))
}

func (p *platform) DrawPoints(pts []sdl.FPoint) {
	p.Renderer.DrawPointsF(pts)
}

func (p *platform) DrawRoundRectOutlined(x, y, w, h, r float32) {
	points := RoundRectPoints(x, y, Max(w-1, 0), Max(h-1, 0), r)
	p.Renderer.DrawLinesF(points)
	l := len(points)-1
	p.Renderer.DrawLineF(points[0].X, points[0].Y, points[l].X, points[l].Y)
}

var RRVERTS = make([]sdl.Vertex, 0)

func (p *platform) DrawRoundRectFilled(x, y, w, h, r float32) {
	R,G,B,A,_ := p.Renderer.GetDrawColor()
	c := sdl.Color{R,G,B,A}

	points := RoundRectPoints(x, y, w, h, r)
	point_count := len(points)

	maybe_resize(&RRVERTS, point_count*3)
	for i := range RRVERTS { RRVERTS[i] = sdl.Vertex{Color: c} }

	midx := x+w/2
	midy := y+h/2

	for i:=0 ; i<point_count ; i++ {
		idx := i*3
		RRVERTS[idx].Position.X = points[i].X
		RRVERTS[idx].Position.Y = points[i].Y

		RRVERTS[idx+1].Position.X = midx
		RRVERTS[idx+1].Position.Y = midy

		next_i := (i+1)%point_count
		RRVERTS[idx+2].Position.X = points[next_i].X
		RRVERTS[idx+2].Position.Y = points[next_i].Y
	}

	p.Renderer.RenderGeometry(nil, RRVERTS, nil)
}

var tex_cache map[string]*sdl.Texture = make(map[string]*sdl.Texture)

func (p *platform) DrawImage(x, y, w, h float32, url string) bool {
	tex, ok := tex_cache[url]
	if !ok {
		realurl := url
		if strings.HasPrefix(realurl, "file://") {
			realurl = realurl[7:]
		}

		tex, err := img.LoadTexture(p.Renderer, realurl)
		tex_cache[url] = tex

		if err != nil {
			println(url, err.Error())
			return false
		}
	}

	if tex == nil { return false }
	_, _, tiw, tih, _ := tex.Query()
	tw, th := float32(tiw), float32(tih)

	ratio := tw / th

	if tw > w {
		tw = w
		th = w / ratio
	}

	if th > h {
		th = h
		tw = h * ratio
	}

	ox, oy := Max(w-tw, 0)/2, Max(h-th, 0)/2

	p.Renderer.CopyF(tex, nil, &sdl.FRect{
		X: x + ox, Y: y + oy, W: tw, H: th,
	})

	return true
}

func (p *platform) ImageSize(url string) (int32, int32) {
	tex, ok := tex_cache[url]
	if !ok { return 0, 0 }
	_, _, w, h, _ := tex.Query()
	return w, h
}
