package utils

import (
	"strconv"
	"golang.org/x/exp/constraints"
	"github.com/veandco/go-sdl2/sdl"
	"math"
)

type V2 struct { X, Y float32 }

var NaN32 = float32(math.NaN())

func (v *V2) ManhattanLength() float32 {
	return Abs(v.X) + Abs(v.Y)
}

func (v *V2) String() string {
	return "V2{ " + FloatStr(v.X) + ", " + FloatStr(v.Y) + " }"
}

func Max[A constraints.Ordered](a, b A) A {
	if a > b { return a } else { return b }
}

func Min[A constraints.Ordered](a, b A) A {
	if a < b { return a } else { return b }
}

func Clamp[A constraints.Ordered](v, min, max A) A {
	if v < min { return min }
	if v > max { return max }
	return v
}

func Abs[A constraints.Float | constraints.Integer](a A) A {
	if a < 0 { return -a } else { return a }
}

func Btof(b bool) float32 {
	if b { return 1 } else { return 0 }
}

func Die(err error) {
	if err != nil { panic(err) }
}

func FloatStr[F constraints.Float](f F) string {
	return strconv.FormatFloat(float64(f), 'f', 3, 32)
}

func ColStr(c sdl.Color) string {
	return strconv.Itoa(int(c.R)) + "," +
		strconv.Itoa(int(c.G)) + "," +
		strconv.Itoa(int(c.B)) + "," +
		strconv.Itoa(int(c.A))
}
