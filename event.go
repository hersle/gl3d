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
	DirectionUp
	DirectionDown
)

type MoveEvent struct {
	dir Direction
	start bool
}

type LookEvent struct {
	dir Direction
	start bool
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

	w.glfwWin.SetKeyCallback(func(_ *glfw.Window, k glfw.Key, _ int, a glfw.Action, _ glfw.ModifierKey) {
		var start bool
		switch a {
			case glfw.Release:
				start = false
			case glfw.Press:
				start = true
			default:
				return
		}

		switch k {
			case glfw.KeyW:
				q.AddMoveEvent(DirectionForward, start)
			case glfw.KeyA:
				q.AddMoveEvent(DirectionLeft, start)
			case glfw.KeyS:
				q.AddMoveEvent(DirectionBackward, start)
			case glfw.KeyD:
				q.AddMoveEvent(DirectionRight, start)
			case glfw.KeyUp:
				q.AddLookEvent(DirectionUp, start)
			case glfw.KeyLeft:
				q.AddLookEvent(DirectionLeft, start)
			case glfw.KeyDown:
				q.AddLookEvent(DirectionDown, start)
			case glfw.KeyRight:
				q.AddLookEvent(DirectionRight, start)
		}
	})

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

func (q *EventQueue) AddMoveEvent(dir Direction, start bool) {
	var e MoveEvent
	e.dir = dir
	e.start = start
	q.AddEvent(&e)
}
func (q *EventQueue) AddLookEvent(dir Direction, start bool) {
	var e LookEvent
	e.dir = dir
	e.start = start
	q.AddEvent(&e)
}
