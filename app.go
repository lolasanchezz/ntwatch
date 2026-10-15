package main

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// the frontend listens for this - fired every uiInterval with a Snapshot
const snapshotEvent = "ntwatch:snapshot"

// App is the wails side - it only ever talks to the engine with messages
type App struct {
	ctx    context.Context
	engine *Engine
}

func NewApp() *App {
	return &App{engine: NewEngine()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.engine.Run(initialModel(func(s Snapshot) {
		runtime.EventsEmit(ctx, snapshotEvent, s)
	}))
}

func (a *App) shutdown(ctx context.Context) {
	a.engine.Stop()
}

// Snapshot returns the current state, for when the frontend wants it right now
// instead of waiting for the next event
func (a *App) Snapshot() Snapshot {
	reply := make(chan Snapshot, 1)
	a.engine.Send(snapshotRequestMsg{reply: reply})
	select {
	case snap := <-reply:
		return snap
	case <-time.After(time.Second):
		return Snapshot{Err: "engine didn't respond"}
	}
}

// SelectProcess picks which process gets the detailed view
func (a *App) SelectProcess(pid, processName string) {
	a.engine.Send(selectProcessMsg{Pid: pid, ProcessName: processName})
}
