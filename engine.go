package main

import (
	"sync"
	"time"
)

// a tiny elm-style runtime: one goroutine owns the model, everything else talks to it with messages

// Msg is anything that can be fed into Update
type Msg any

// Cmd is a side effect - it runs on its own goroutine and reports back with a Msg (or nil for nothing)
type Cmd func() Msg

type Engine struct {
	msgs     chan Msg
	done     chan struct{}
	stopOnce sync.Once
}

func NewEngine() *Engine {
	return &Engine{
		msgs: make(chan Msg, 256),
		done: make(chan struct{}),
	}
}

// Send feeds a message into the update loop. safe to call from any goroutine
func (e *Engine) Send(msg Msg) {
	select {
	case e.msgs <- msg:
	case <-e.done:
	}
}

// Stop ends the update loop. safe to call more than once
func (e *Engine) Stop() {
	e.stopOnce.Do(func() { close(e.done) })
}

// Run blocks until Stop is called, returning the final model
func (e *Engine) Run(m model) model {
	e.run(m.Init())
	for {
		select {
		case <-e.done:
			return m
		case msg := <-e.msgs:
			var cmds []Cmd
			m, cmds = m.Update(msg)
			e.run(cmds)
		}
	}
}

func (e *Engine) run(cmds []Cmd) {
	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}
		go func() {
			if msg := cmd(); msg != nil {
				e.Send(msg)
			}
		}()
	}
}

// Tick waits d and then reports back with fn's message
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		return fn(<-time.After(d))
	}
}
