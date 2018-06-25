package main

import (
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Event interface{}

type Direction int
const (
	DirectionForward Direction = iota
	DirectionBackward
	DirectionLeft
	DirectionRight
)

type MoveEvent struct {
}

type ResizeEvent struct {
	Width, Height int
}

type EventQueue struct {
	events []Event
	head, tail int
	length int
}

func NewEventQueue(w *Window) *EventQueue {
	var q EventQueue
	q.events = make([]Event, 50)
	q.head = 0
	q.tail = 0
	q.length = 0

	w.glfwWin.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		q.AddResizeEvent(width, height)
	})

	return &q
}

func (q *EventQueue) grow() {
	eventsBehindHead := q.events[q.head:]

	// use Go's automatic slice resizing to grow queue
	dummyEvent := ResizeEvent{}
	q.events = append(q.events, dummyEvent)

	eventsBack := q.events[len(q.events) - len(eventsBehindHead):]
	copy(eventsBack, eventsBehindHead)

	q.head = len(q.events) - len(eventsBehindHead)
}

func (q *EventQueue) full() bool {
	return q.length == cap(q.events)
}

func (q *EventQueue) empty() bool {
	return q.length == 0
}

func (q *EventQueue) PopEvent() Event {
	if q.empty() {
		return nil
	} else {
		e := q.events[q.head]
		q.head = (q.head + 1) % cap(q.events)
		q.length = q.length - 1
		return e
	}
}

func (q *EventQueue) AddEvent(e Event) {
	if q.full() {
		q.grow()
	}
	q.events[q.tail] = e
	q.tail = (q.tail + 1) % cap(q.events)
	q.length = q.length + 1
}

func (q *EventQueue) AddResizeEvent(width, height int) {
	var e ResizeEvent
	e.Width = width
	e.Height = height
	q.AddEvent(&e)
}
