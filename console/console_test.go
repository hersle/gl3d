package console

import (
	"testing"
	"fmt"
)

func TestFull(t *testing.T) {
	c := NewConsole()

	for i := 0; i < c.capacity + 1; i++ {
		line := fmt.Sprintf("line %d", i+1)
		c.AddLine(line)
	}

	println(c.String())
}

func TestNotFull(t *testing.T) {
	c := NewConsole()

	for i := 0; i < c.capacity; i++ {
		line := fmt.Sprintf("line %d", i+1)
		c.AddLine(line)
	}

	println(c.String())
}
