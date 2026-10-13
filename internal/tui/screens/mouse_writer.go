package screens

import (
	"sync"

	"github.com/angristan/hue-tui/internal/tui/messages"
	tea "github.com/charmbracelet/bubbletea"
)

// mouseWriteCoordinator guarantees that the most recent mouse value wins:
// a newer write cannot overtake an older in-flight HTTP request, and older
// queued commands are discarded if newer values supersede them.
type mouseWriteCoordinator struct {
	mu sync.Mutex
	next uint64
	latest map[string]uint64
}

type mouseLightWrite struct {
	id string
	commands []tea.Cmd
}

func newMouseWriteCoordinator() *mouseWriteCoordinator {
	return &mouseWriteCoordinator{latest: make(map[string]uint64)}
}

func (q *mouseWriteCoordinator) schedule(writes []mouseLightWrite) tea.Cmd {
	if len(writes) == 0 {
		return nil
	}

	// Version reservation happens synchronously from the TUI Update method.
	q.mu.Lock()
	tickets := make([]uint64, len(writes))
	for i, w := range writes {
		q.next++
		tickets[i] = q.next
		q.latest[w.id] = q.next
	}
	q.mu.Unlock()

	return func() tea.Msg {
		for i, w := range writes {
			q.mu.Lock()
			if q.latest[w.id] != tickets[i] {
				q.mu.Unlock()
				continue // Stale request, including a delayed intermediate drag value
			}
			// Serialize power and brightness/temperature writes for one light,
			// holding the lock across the actual HTTP requests.
			var result tea.Msg
			for _, command := range w.commands {
				if command == nil {
					continue
				}
				result = command()
				if _, failed := result.(messages.ErrorMsg); failed {
					break
				}
			}
			q.mu.Unlock()
			if result != nil {
				return result
			}
		}
		return nil
	}
}
