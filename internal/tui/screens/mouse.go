package screens

import (
	"time"

	"github.com/angristan/hue-tui/internal/api"
	"github.com/angristan/hue-tui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// Keep the list's brightness hitboxes in sync with the rendering geometry.
func lightBarLayout(width int) (nameWidth, barWidth int) {
	available := width - 13
	barWidth = max(8, min(20, available*35/100))
	nameWidth = max(10, min(45, available-barWidth))
	return nameWidth, barWidth
}

func (m MainModel) contentStartY() int {
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

// itemAtY follows the list's blank lines and scroll indicators in View().
func (m MainModel) itemAtY(y int) int {
	y0 := m.contentStartY()
	if y < y0 || y >= y0+m.contentHeight() {
		return -1
	}
	row := y0
	if m.scrollOffset > 0 {
		row++
	}
	end := min(len(m.items), m.scrollOffset+m.visibleLines())
	for i := m.scrollOffset; i < end; i++ {
		if m.items[i].isRoom && i > m.scrollOffset {
			row++
		}
		if y == row {
			return i
		}
		row++
	}
	return -1
}

// panelBarHit returns the brightness slider in either the light or group panel.
func (m MainModel) panelBarHit(x, y int) (start, width int, ok bool) {
	if !m.showPanel || m.width < 80 || m.loading || m.SelectedItem() == nil {
		return 0, 0, false
	}
	panelWidth := max(30, min(45, m.width*30/100))
	start = m.contentWidth() + 5
	width = max(10, min(25, panelWidth-10))
	return start, width, y == m.contentStartY()+7 && x >= start && x < start+width
}

// temperatureBarHit matches the preferred ambiance panel, not the color gadget UI.
// The Mode row in the light details panel switches between ambiance and
// optional color controls without sending a command to the bridge.
func (m MainModel) modeToggleHit(x, y int) bool {
	light := m.SelectedLight()
	if light == nil || !light.SupportsColorTemp || !light.SupportsColor || !m.showPanel || m.width < 80 || m.loading {
		return false
	}
	start := m.contentWidth() + 5
	panelWidth := max(30, min(45, m.width*30/100))
	return y == m.contentStartY()+9 && x >= start && x < start+panelWidth-8
}

func (m MainModel) temperatureBarHit(x, y int) (start, width int, ok bool) {
	light := m.SelectedLight()
	if light == nil || !light.SupportsColorTemp || m.showColorControls {
		return 0, 0, false
	}
	if !m.showPanel || m.width < 80 || m.loading {
		return 0, 0, false
	}
	panelWidth := max(30, min(45, m.width*30/100))
	start = m.contentWidth() + 5
	width = max(10, min(25, panelWidth-10))
	return start, width, y == m.contentStartY()+12 && x >= start && x < start+width
}

func (m MainModel) brightnessBarHit(x, y int) (item, start, width int, ok bool) {
	if m.loading {
		return -1, 0, 0, false
	}
	idx := m.itemAtY(y)
	if idx >= 0 {
		nameWidth, barWidth := lightBarLayout(m.contentWidth())
		barStart := 6 + nameWidth
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
	return max(1, min(100, pct))
}

func mouseBarMirek(x, start, width int) int {
	if width <= 1 {
		return 153
	}
	pos := max(0, min(width-1, x-start))
	return 153 + (pos*347+(width-1)/2)/(width-1)
}

func (m MainModel) dragLights() []*models.Light {
	if m.mouseDragItem < 0 || m.mouseDragItem >= len(m.items) {
		return nil
	}
	item := m.items[m.mouseDragItem]
	if item.isRoom {
		return item.room.Lights
	}
	if item.light != nil {
		return []*models.Light{item.light}
	}
	return nil
}

// Preview every motion immediately; send throttled Hue commands during drag
// and always commit the latest value on mouse-up.
func (m MainModel) changeMouseBrightness(bridge api.BridgeClient, addPending PendingAdder, pct int, send bool) (MainModel, tea.Cmd) {
	var writes []mouseLightWrite
	for _, light := range m.dragLights() {
		var cmds []tea.Cmd
		if send && m.mouseDragWasOff[light.ID] {
			if addPending != nil {
				addPending(light.ID, "on", true, DirExact)
			}
			cmds = append(cmds, m.toggleLightCmd(bridge, light.ID, true))
		}
		light.On = true
		light.SetBrightnessPct(pct)
		if send {
			if addPending != nil {
				addPending(light.ID, "brightness", pct, DirExact)
			}
			cmds = append(cmds, m.setBrightnessCmd(bridge, light.ID, pct))
			writes = append(writes, mouseLightWrite{id: light.ID, commands: cmds})
		}
	}
	for _, room := range m.rooms {
		room.UpdateState() // zones can share light pointers with rooms
	}
	if !send || m.mouseWrites == nil {
		return m, nil
	}
	return m, m.mouseWrites.schedule(writes)
}

func (m MainModel) changeMouseTemperature(bridge api.BridgeClient, addPending PendingAdder, mirek int, send bool) (MainModel, tea.Cmd) {
	lights := m.dragLights()
	if len(lights) != 1 || !lights[0].SupportsColorTemp {
		return m, nil
	}
	light := lights[0]
	if light.Color == nil {
		light.Color = &models.Color{Brightness: light.Brightness}
	}
	light.Color.Mirek = uint16(mirek)
	light.Color.Mode = models.ColorModeColorTemp
	light.Color.InvalidateCache()
	light.On = true
	if !send {
		return m, nil
	}
	var cmds []tea.Cmd
	if m.mouseDragWasOff[light.ID] {
		if addPending != nil {
			addPending(light.ID, "on", true, DirExact)
		}
		cmds = append(cmds, m.toggleLightCmd(bridge, light.ID, true))
	}
	if addPending != nil {
		addPending(light.ID, "color_temp", mirek, DirExact)
	}
	cmds = append(cmds, m.setColorTempCmd(bridge, light.ID, mirek))
	if m.mouseWrites == nil {
		return m, nil
	}
	return m, m.mouseWrites.schedule([]mouseLightWrite{{id: light.ID, commands: cmds}})
}

func (m MainModel) previewDrag(bridge api.BridgeClient, addPending PendingAdder, value int, send bool) (MainModel, tea.Cmd) {
	if m.mouseDragKind == "temperature" {
		return m.changeMouseTemperature(bridge, addPending, value, send)
	}
	return m.changeMouseBrightness(bridge, addPending, value, send)
}

func (m MainModel) dragValueForX(x int) int {
	if m.mouseDragKind == "temperature" {
		return mouseBarMirek(x, m.mouseDragX, m.mouseDragWidth)
	}
	return mouseBarPercent(x, m.mouseDragX, m.mouseDragWidth)
}

func (m MainModel) startDrag(idx, x, start, width int, kind string, bridge api.BridgeClient, addPending PendingAdder) (MainModel, tea.Cmd) {
	m.selectedIndex = idx
	m.mouseDragging = true
	m.mouseDragKind = kind
	m.mouseDragItem = idx
	m.mouseDragX, m.mouseDragWidth = start, width
	m.mouseDragWasOff = make(map[string]bool)
	for _, light := range m.dragLights() {
		m.mouseDragWasOff[light.ID] = !light.On
	}
	m.mouseDragValue = m.dragValueForX(x)
	m.mouseDragSent = m.mouseDragValue
	m.mouseDragSentAt = time.Now()
	// Initial press sends immediately, so a simple click is also responsive.
	return m.previewDrag(bridge, addPending, m.mouseDragValue, true)
}

func (m MainModel) handleMouse(msg tea.MouseMsg, bridge api.BridgeClient, addPending PendingAdder) (MainModel, tea.Cmd) {
	if m.searchMode {
		return m, nil
	}
	if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
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
		// Use the release coordinate even if the terminal skipped motion events.
		m.mouseDragValue = m.dragValueForX(msg.X)
		m.mouseDragging = false
		m.mouseDragSent = m.mouseDragValue
		m, cmd := m.previewDrag(bridge, addPending, m.mouseDragValue, true)
		m.mouseDragWasOff = nil
		return m, cmd

	case tea.MouseActionMotion:
		if !m.mouseDragging || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		value := m.dragValueForX(msg.X)
		if value == m.mouseDragValue {
			return m, nil
		}
		m.mouseDragValue = value
		send := time.Since(m.mouseDragSentAt) >= 120*time.Millisecond
		if send {
			m.mouseDragSentAt = time.Now()
			m.mouseDragSent = value
		}
		return m.previewDrag(bridge, addPending, value, send)

	case tea.MouseActionPress:
		if msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		if m.modeToggleHit(msg.X, msg.Y) {
			m.showColorControls = !m.showColorControls
			return m, nil
		}
		// The temperature slider is actionable even when the light is still in
		// XY color mode; touching it switches to ambiance on mouse-up.
		if start, width, hit := m.temperatureBarHit(msg.X, msg.Y); hit {
			return m.startDrag(m.selectedIndex, msg.X, start, width, "temperature", bridge, addPending)
		}
		if idx, start, width, hit := m.brightnessBarHit(msg.X, msg.Y); hit {
			return m.startDrag(idx, msg.X, start, width, "brightness", bridge, addPending)
		}

		idx := m.itemAtY(msg.Y)
		if idx >= 0 && msg.X >= 0 && msg.X < m.contentWidth() {
			m.selectedIndex = idx
			m.ensureVisible()
			if !m.items[idx].isRoom && msg.X == 2 {
				return m.Update(tea.KeyMsg{Type: tea.KeySpace}, bridge, addPending)
			}
		}
	}
	return m, nil
}
