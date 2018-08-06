package math

import (
	"math"
)

func Min(a, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}

func Max(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

func Radians(degrees float32) float32 {
	return degrees / 360 * 2 * math.Pi
}

func Degrees(radians float32) float32 {
	return radians / (2 * math.Pi) * 360
}
