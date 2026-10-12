package screens

import (
	"time"

	"github.com/angristan/hue-tui/internal/api"
	"github.com/angristan/hue-tui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

const mouseDragInterval = 120 * time.Millisecond

// lightBarLayout is shared by the renderer and mouse hit testing.
func lightBarLayout(width int) (nameWidth, barWidth int) {
	available := width - 13
	barWidth = max(8, min(20, available*35/100))
	nameWidth = max(10, min(45, available-barWidth))
	return nameWidth, barWidth
}

func (m MainModel) contentStartY() int {
	// Header (row 0) and a blank line. A search field adds one more row.
	if m.searchMode || m.searchQuery != "" {
		return 3
	}
	return 2
}

func (m MainModel) contentWidth() int {
	if m.showPanel && m.width >= 80 {
		panelWidth := max(30, min(45, m.width*30/100))
		return m.width - panelWidth - 3
	}
	return m.width
}

func (m MainModel) contentHeight() int {
	height := m.height - 5
	if m.searchMode || m.searchQuery != "" {
		height--
	}
	return max(3, height)
}

// itemAtY follows the exact blank lines and scroll indicators in View().
// It returns -1 for empty rows, the header and the status/footer.
func (m MainModel) itemAtY(y int) int {
	y0 := m.contentStartY()
	if y < y0 || y >= y0+m.contentHeight() {
		return -1
	}
	row := y0
	if m.scrollOffset > 0 {
		row++ // "more above" indicator
	}
	end := min(len(m.items), m.scrollOffset+m.visibleLines())
	for i := m.scrollOffset; i < end; i++ {
		if m.items[i].isRoom && i > m.scrollOffset {
			row++ // blank line preceding subsequent room headers
		}
		if y == row {
			return i
		}
		row++
	}
	return -1
}

// panelBarHit returns the brightness slider geometry on the details panel.
func (m MainModel) panelBarHit(x, y int) (start, width int, ok bool) {
	if !m.showPanel || m.width < 80 || m.loading || m.SelectedItem() == nil {
		return 0, 0, false
	}
	panelWidth := max(30, min(45, m.width*30/100))
	// Content + 2 spaces, then rounded border and 2 cells of padding.
	start = m.contentWidth() + 5
	width = max(10, min(25, panelWidth-10))
	// Panel top border + padding + title/blank/status/blank/label.
	barY := m.contentStartY() + 7
	return start, width, y == barY && x >= start && x < start+width
}

func (m MainModel) brightnessBarHit(x, y int) (item, start, width int, ok bool) {
	if m.loading {
		return -1, 0, 0, false
	}
	idx := m.itemAtY(y)
	if idx >= 0 && !m.items[idx].isRoom {
		nameWidth, barWidth := lightBarLayout(m.contentWidth())
		barStart := 6 + nameWidth // 2 cursor + 1 icon + space + name + 2 spaces
		if x >= barStart && x < barStart+barWidth {
			return idx, barStart, barWidth, true
		}
	}
	if barStart, barWidth, hit := m.panelBarHit(x, y); hit {
		return m.selectedIndex, barStart, barWidth, true
	}
	return -1, 0, 0, false
}

func mouseBarPercent(x, start, width int) int {
	if width <= 1 {
		return 1
	}
	pct := (x - start) * 100 / (width - 1)
	return max(1, min(100, pct)) // turning lights off remains a separate action
}

func (m MainModel) changeMouseBrightness(bridge api.BridgeClient, addPending PendingAdder, pct int, send bool) (MainModel, tea.Cmd) {
	if m.mouseDragItem < 0 || m.mouseDragItem >= len(m.items) {
		return m, nil
	}
	item := m.items[m.mouseDragItem]
	var lights []*models.Light
	if item.isRoom {
		lights = item.room.Lights
	} else if item.light != nil {
		lights = []*models.Light{item.light}
	}
	var cmds []tea.Cmd
	for _, light := range lights {
		if !light.On {
			light.On = true
			if send {
				if addPending != nil {
					addPending(light.ID, "on", true, DirExact)
				}
				cmds = append(cmds, m.toggleLightCmd(bridge, light.ID, true))
			}
		}
		oldPct := light.BrightnessPct()
		light.SetBrightnessPct(pct)
		if send {
			if addPending != nil {
				direction := DirExact
				if pct > oldPct {
					direction = DirUp
				} else if pct < oldPct {
					direction = DirDown
				}
				addPending(light.ID, "brightness", pct, direction)
			}
			cmds = append(cmds, m.setBrightnessCmd(bridge, light.ID, pct))
		}
	}
	if item.room != nil {
		item.room.UpdateState()
	}
	return m, tea.Batch(cmds...)
}

func (m MainModel) handleMouse(msg tea.MouseMsg, bridge api.BridgeClient, addPending PendingAdder) (MainModel, tea.Cmd) {
	if m.searchMode {
		return m, nil
	}

	if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
		// Scroll the viewport without changing the active selection.
		step := 3
		if msg.Button == tea.MouseButtonWheelUp {
			step = -step
		}
		m.scrollOffset = max(0, min(max(0, len(m.items)-m.visibleLines()), m.scrollOffset+step))
		return m, nil
	}

	switch msg.Action {
	case tea.MouseActionRelease:
		if !m.mouseDragging {
			return m, nil
		}
		m.mouseDragging = false
		if m.mouseDragValue != m.mouseDragSent {
			m.mouseDragSent = m.mouseDragValue
			return m.changeMouseBrightness(bridge, addPending, m.mouseDragValue, true)
		}
		return m, nil

	case tea.MouseActionMotion:
		if !m.mouseDragging || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		pct := mouseBarPercent(msg.X, m.mouseDragX, m.mouseDragWidth)
		if pct == m.mouseDragValue {
			return m, nil
		}
		m.mouseDragValue = pct
		send := time.Since(m.mouseDragSentAt) >= mouseDragInterval
		if send {
			m.mouseDragSent = pct
			m.mouseDragSentAt = time.Now()
		}
		return m.changeMouseBrightness(bridge, addPending, pct, send)

	case tea.MouseActionPress:
		if msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		idx, start, width, onBar := m.brightnessBarHit(msg.X, msg.Y)
		if onBar {
			m.selectedIndex = idx
			m.mouseDragging = true
			m.mouseDragItem = idx
			m.mouseDragX, m.mouseDragWidth = start, width
			m.mouseDragValue = mouseBarPercent(msg.X, start, width)
			m.mouseDragSent = m.mouseDragValue
			m.mouseDragSentAt = time.Now()
			return m.changeMouseBrightness(bridge, addPending, m.mouseDragValue, true)
		}

		idx = m.itemAtY(msg.Y)
		if idx >= 0 && msg.X >= 0 && msg.X < m.contentWidth() {
			m.selectedIndex = idx
			m.ensureVisible()
			if !m.items[idx].isRoom && msg.X == 2 {
				// Click the status dot to toggle the selected light.
				return m.Update(tea.KeyMsg{Type: tea.KeySpace}, bridge, addPending)
			}
		}
	}
	return m, nil
}
