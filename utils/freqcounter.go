package utils

import (
	"fmt"
	"time"
)

type FrequencyCounter struct {
	Interval        int
	ExtremeInterval int
	count           int
	frequency       int
	startTime       time.Time
}

func NewFrequencyCounter() *FrequencyCounter {
	var c FrequencyCounter
	c.Interval = 1
	c.frequency = -1
	c.count = 0
	c.startTime = time.Now()
	return &c
}

func (c *FrequencyCounter) Count() {
	c.count++

	if c.count >= c.Interval {
		endTime := time.Now()
		secs := endTime.Sub(c.startTime).Seconds()
		c.frequency = int(float64(c.count) / secs)
		c.startTime = endTime
		c.count = 0
	}
}

func (c *FrequencyCounter) Frequency() int {
	return c.frequency // counts per second
}

func (c *FrequencyCounter) Period() float32 {
	return 1 / float32(c.Frequency())
}

func (c *FrequencyCounter) String() string {
	return fmt.Sprintf("%d Hz", c.Frequency())
}
