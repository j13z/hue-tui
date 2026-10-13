package screens

import (
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMouseWritesSkipSupersededCommands(t *testing.T) {
	q := newMouseWriteCoordinator()
	var mu sync.Mutex
	var values []int
	write := func(value int) []mouseLightWrite {
		return []mouseLightWrite{{
			id: "desk",
			commands: []tea.Cmd{func() tea.Msg {
				mu.Lock()
				values = append(values, value)
				mu.Unlock()
				return nil
			}},
		}}
	}

	older := q.schedule(write(20))
	latest := q.schedule(write(80))
	older() // Previous command may only start after the new drag value.
	latest()

	if len(values) != 1 || values[0] != 80 {
		t.Fatalf("stale mouse update was applied: %v", values)
	}
}

func TestMouseUpIsLastEvenWhenOlderWriteIsInFlight(t *testing.T) {
	q := newMouseWriteCoordinator()
	started := make(chan struct{})
	finishFirst := make(chan struct{})
	result := make(chan []int, 1)

	var mu sync.Mutex
	var values []int
	old := q.schedule([]mouseLightWrite{{
		id: "play-bar",
		commands: []tea.Cmd{func() tea.Msg {
			close(started)
			<-finishFirst
			mu.Lock()
			values = append(values, 20)
			mu.Unlock()
			return nil
		}},
	}})
	done := make(chan struct{})
	go func() {
		defer close(done)
		old()
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		close(finishFirst)
		t.Fatal("first simulated Hue request did not start")
	}

	// New input must never block the TUI, even during a slow bridge call.
	scheduled := make(chan tea.Cmd, 1)
	go func() {
		scheduled <- q.schedule([]mouseLightWrite{{
			id: "play-bar",
			commands: []tea.Cmd{func() tea.Msg {
				mu.Lock()
				values = append(values, 90)
				mu.Unlock()
				return nil
			}},
		}})
	}()
	var final tea.Cmd
	select {
	case final = <-scheduled:
	case <-time.After(500 * time.Millisecond):
		close(finishFirst)
		t.Fatal("scheduling mouse-up blocked on a slow Hue request")
	}
	go func() {
		final()
		<-done
		mu.Lock()
		result <- append([]int(nil), values...)
		mu.Unlock()
	}()
	close(finishFirst)
	select {
	case got := <-result:
		if len(got) != 2 || got[0] != 20 || got[1] != 90 {
			t.Fatalf("final mouse-up request not applied last: %v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("final mouse-up request was not flushed")
	}
}
