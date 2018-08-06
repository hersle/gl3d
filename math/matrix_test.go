package math

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestDeterminant(t *testing.T) {
	var n int = 1000
	var dev float32
	var devMax float32 = 0
	var devMaxAllowed float32 = 1e-2
	a := NewMat4Zero()
	b := NewMat4Zero()
	ident := NewMat4Identity()

	rand.Seed(time.Now().UnixNano())

	println("testing", n, "random matrices")

	for n > 0 {
		n--

		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				a.Set(i, j, rand.Float32()*100)
				b.Set(i, j, a.At(i, j))
			}
		}

		err := b.Invert()
		if err == nil {
			println("skipped inverting a singular matrix")
			continue
		}
		a.Mult(b)
		// a should now be the identity

		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				dev = float32(math.Abs(float64(a.At(i, j) - ident.At(i, j))))
				devMax = float32(math.Max(float64(dev), float64(devMax)))
			}
		}
	}

	println("max deviation from identity matrix:", devMax)
	println("max allowed deviation from identity matrix:", devMaxAllowed)
	if devMax > devMaxAllowed {
		t.Fail()
	}

	return
}
