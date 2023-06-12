package platform

import (
	"math"
	"github.com/veandco/go-sdl2/sdl"
	. "github.com/glupi-borna/wiggo/internal/utils"
)

func SetColor(c sdl.Color) {
	Platform.Renderer.SetDrawColor(c.R, c.G, c.B, c.A)
}

func DrawRectOutlined(x, y, w, h float32) {
	Platform.Renderer.DrawRectF(&sdl.FRect{x, y, w, h})
}

func DrawRectFilled(x, y, w, h float32) {
	Platform.Renderer.FillRectF(&sdl.FRect{x, y, w, h})
}


func ArcPoints(x, y, r, startAngle, endAngle float32) []sdl.FPoint {
	CORNER_POINTS := Max(5, int(r)/5)

	points := make([]sdl.FPoint, CORNER_POINTS)
	angleRange := float64(Abs(startAngle - endAngle))
	start := float64(startAngle)
	rad := float64(r)

	imax := float64(CORNER_POINTS-1)

	for i:= 0; i<CORNER_POINTS; i++ {
		frac := float64(i)/imax
		angle := frac*angleRange + start
		xoff := float32(math.Cos(angle) * rad)
		yoff := float32(math.Sin(angle) * rad)
		points[i].X = x+xoff
		points[i].Y = y+yoff
	}

	return points
}

func RoundRectPoints(x, y, w, h, r float32) []sdl.FPoint {
	r = Clamp(r, 0, Min(w/2, h/2))

	CORNER_POINTS := Max(5, int(r)/5)
	TOTAL_POINTS := CORNER_POINTS * 4

	x1 := x+r
	y1 := y+r
	x2 := x+w-r
	y2 := y+h-r

	points := make([]sdl.FPoint, 0, TOTAL_POINTS)
	const pihalf = math.Pi/2
	points = append(points, ArcPoints(x1, y1, r, pihalf*2, pihalf*3)...)
	points = append(points, ArcPoints(x2, y1, r, pihalf*3, pihalf*4)...)
	points = append(points, ArcPoints(x2, y2, r, 0,        pihalf)...)
	points = append(points, ArcPoints(x1, y2, r, pihalf,   pihalf*2)...)

	return points
}

func DrawPoints(pts []sdl.FPoint) {
	Platform.Renderer.DrawPointsF(pts)
}

func DrawRoundRectOutlined(x, y, w, h, r float32) {
	points := RoundRectPoints(x, y, w, h, r)
	Platform.Renderer.DrawLinesF(points)
	l := len(points)-1
	Platform.Renderer.DrawLineF(points[0].X, points[0].Y, points[l].X, points[l].Y)
}

func DrawRoundRectFilled(x, y, w, h, r float32) {
	R,G,B,A,_ := Platform.Renderer.GetDrawColor()
	c := sdl.Color{R,G,B,A}

	points := RoundRectPoints(x, y, w, h, r)
	point_count := len(points)

	verts := make([]sdl.Vertex, point_count*3)
	for i := range verts { verts[i] = sdl.Vertex{Color: c} }

	midx := x+w/2
	midy := y+h/2

	for i:=0 ; i<point_count ; i++ {
		idx := i*3
		verts[idx].Position.X = points[i].X
		verts[idx].Position.Y = points[i].Y

		verts[idx+1].Position.X = midx
		verts[idx+1].Position.Y = midy

		next_i := (i+1)%point_count
		verts[idx+2].Position.X = points[next_i].X
		verts[idx+2].Position.Y = points[next_i].Y
	}

	Platform.Renderer.RenderGeometry(nil, verts, nil)
}
