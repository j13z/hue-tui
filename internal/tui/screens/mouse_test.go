package screens

import (
	"testing"

	"github.com/angristan/hue-tui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

func mouseTestModel() MainModel {
	m := NewMainModel(nil)
	m.SetSize(120, 30)
	m.SetData([]*models.Room{
		{ID: "a", Name: "Arbeitszimmer", Lights: []*models.Light{
			{ID: "desk", Name: "Desk", On: true, Brightness: 127},
			{ID: "play", Name: "Play Bar", On: true, Brightness: 127},
		}},
		{ID: "b", Name: "Play Bars", Lights: []*models.Light{
			{ID: "other", Name: "Other", On: true, Brightness: 127},
		}},
	}, nil)
	return m
}

func TestMouseRowMapping(t *testing.T) {
	m := mouseTestModel()
	// Header at 0; blank at 1; first room at 2, lights at 3 and 4.
	for _, test := range []struct{ y, want int }{
		{0, -1}, {1, -1}, {2, 0}, {3, 1}, {4, 2},
		{5, -1}, {6, 3}, {7, 4},
	} {
		if got := m.itemAtY(test.y); got != test.want {
			t.Errorf("row %d maps to %d, expected %d", test.y, got, test.want)
		}
	}
	m.scrollOffset = 2
	// With a scroll indicator, row 2 is no longer a selectable item.
	if got := m.itemAtY(2); got != -1 {
		t.Fatalf("scroll indicator must not be selectable, got %d", got)
	}
	if got := m.itemAtY(3); got != 2 {
		t.Fatalf("first visible item wrong after scroll: %d", got)
	}
}

func TestMouseBrightnessDrag(t *testing.T) {
	m := mouseTestModel()
	nameWidth, barWidth := lightBarLayout(m.contentWidth())
	start := 6 + nameWidth
	idx, gotStart, gotWidth, hit := m.brightnessBarHit(start+barWidth/2, 3)
	if !hit || idx != 1 || gotStart != start || gotWidth != barWidth {
		t.Fatalf("wrong brightness hit: item=%d start=%d width=%d hit=%v", idx, gotStart, gotWidth, hit)
	}
	press := tea.MouseMsg{X: start, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	updated, _ := m.Update(press, nil, nil)
	m = updated
	if !m.mouseDragging || m.selectedIndex != 1 || m.items[1].light.BrightnessPct() != 1 {
		t.Fatalf("press did not select slider and set low brightness")
	}
	motion := tea.MouseMsg{X: start + barWidth - 1, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion}
	updated, _ = m.Update(motion, nil, nil)
	m = updated
	if m.items[1].light.BrightnessPct() != 100 {
		t.Fatalf("drag should preview 100%%, got %d", m.items[1].light.BrightnessPct())
	}
	release := tea.MouseMsg{X: start + barWidth - 1, Y: 3, Action: tea.MouseActionRelease}
	updated, _ = m.Update(release, nil, nil)
	m = updated
	if m.mouseDragging || m.mouseDragSent != 100 {
		t.Fatalf("release did not finalize drag")
	}
	if m.items[2].light.BrightnessPct() != 50 {
		t.Fatalf("drag must not adjust another light")
	}
}

func TestMousePanelAndSelection(t *testing.T) {
	m := mouseTestModel()
	x, width, hit := m.panelBarHit(m.contentWidth()+5, m.contentStartY()+7)
	if !hit || width < 10 || x != m.contentWidth()+5 {
		t.Fatalf("room details slider geometry wrong")
	}
	updated, _ := m.Update(tea.MouseMsg{X: 8, Y: 4, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}, nil, nil)
	m = updated
	if m.selectedIndex != 2 {
		t.Fatalf("clicking a row must select it, got %d", m.selectedIndex)
	}
	if !m.items[2].light.On {
		t.Fatal("clicking light name must not toggle it")
	}
	updated, _ = m.Update(tea.MouseMsg{X: 2, Y: 4, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}, nil, nil)
	m = updated
	if m.items[2].light.On {
		t.Fatal("clicking the status bullet must toggle the light off")
	}
}

func TestMouseBarPercent(t *testing.T) {
	if got := mouseBarPercent(10, 10, 20); got != 1 {
		t.Fatalf("left end should be minimum 1%%, got %d", got)
	}
	if got := mouseBarPercent(29, 10, 20); got != 100 {
		t.Fatalf("right end should be 100%%, got %d", got)
	}
	if got := mouseBarPercent(-100, 10, 20); got != 1 {
		t.Fatalf("drag beyond left should clamp, got %d", got)
	}
	if got := mouseBarPercent(100, 10, 20); got != 100 {
		t.Fatalf("drag beyond right should clamp, got %d", got)
	}
}

func TestGroupSliderMouseUpAlwaysCommits(t *testing.T) {
	m := mouseTestModel()
	nameWidth, barWidth := lightBarLayout(m.contentWidth())
	start := 6 + nameWidth
	if idx, _, _, hit := m.brightnessBarHit(start, 2); !hit || idx != 0 {
		t.Fatalf("group header must have an actionable brightness slider")
	}
	var sent []string
	pending := func(id, field string, value interface{}, dir Direction) {
		if field == "brightness" {
			sent = append(sent, id)
		}
	}
	press := tea.MouseMsg{X: start + barWidth - 1, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	m, _ = m.Update(press, nil, pending)
	if len(sent) != 2 {
		t.Fatalf("press should immediately send live group brightness: %v", sent)
	}
	// No motion: even if press and release are identical, send final value.
	release := tea.MouseMsg{X: start + barWidth - 1, Y: 2, Action: tea.MouseActionRelease}
	m, cmd := m.Update(release, nil, pending)
	if cmd == nil || len(sent) != 4 {
		t.Fatalf("release must send brightness again for both group lights: %v", sent)
	}
	for _, light := range m.rooms[0].Lights {
		if light.BrightnessPct() != 100 {
			t.Fatalf("group drag didn't update %s", light.ID)
		}
	}
}

func TestTemperatureSliderFromXYCommitsOnRelease(t *testing.T) {
	m := mouseTestModel()
	m.selectedIndex = 1
	light := m.items[1].light
	light.SupportsColorTemp = true
	light.Color = &models.Color{Mode: models.ColorModeXY, Brightness: light.Brightness}
	start := m.contentWidth() + 5
	y := m.contentStartY() + 12
	if _, _, hit := m.temperatureBarHit(start, y); !hit {
		t.Fatalf("expected temperature slider in default ambiance panel")
	}
	var sent int
	pending := func(id, field string, value interface{}, dir Direction) {
		if field == "color_temp" {
			sent++
		}
	}
	m, _ = m.Update(tea.MouseMsg{X: start, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}, nil, pending)
	if sent != 1 {
		t.Fatal("temperature press must send an immediate live update")
	}
	if light.Color.Mode != models.ColorModeColorTemp {
		t.Fatal("temperature slider must preview ambiance mode")
	}
	m, cmd := m.Update(tea.MouseMsg{X: start + 24, Y: y, Action: tea.MouseActionRelease}, nil, pending)
	if cmd == nil || sent != 2 || light.Color.Mirek != 500 {
		t.Fatalf("mouse up must send final temperature, sent=%d mirek=%d", sent, light.Color.Mirek)
	}
	if m.mouseDragging {
		t.Fatal("dragging must end on release")
	}
}

func TestMouseBrightnessReleaseUsesFinalPosition(t *testing.T) {
	m := mouseTestModel()
	nameWidth, barWidth := lightBarLayout(m.contentWidth())
	start := 6 + nameWidth
	var sent int
	pending := func(id, field string, value interface{}, dir Direction) {
		if field == "brightness" {
			sent = value.(int)
		}
	}
	m, _ = m.Update(tea.MouseMsg{X: start, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}, nil, pending)
	if sent != 1 {
		t.Fatalf("press must send the first live value (1%%): %d", sent)
	}
	m, _ = m.Update(tea.MouseMsg{X: start + barWidth - 1, Y: 3, Action: tea.MouseActionRelease}, nil, pending)
	if sent != 100 || m.items[1].light.BrightnessPct() != 100 {
		t.Fatalf("release position lost: sent=%d", sent)
	}
}

func TestMouseModeToggleWithoutSendingHueCommand(t *testing.T) {
	m := mouseTestModel()
	m.selectedIndex = 1
	light := m.items[1].light
	light.SupportsColor = true
	light.SupportsColorTemp = true
	light.Color = &models.Color{Mode: models.ColorModeXY, Brightness: light.Brightness}

	x := m.contentWidth() + 5
	y := m.contentStartY() + 9
	if !m.modeToggleHit(x, y) {
		t.Fatal("ambiance/color mode label must be clickable")
	}
	m, cmd := m.Update(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}, nil, nil)
	if cmd != nil || !m.showColorControls || light.Color.Mode != models.ColorModeXY {
		t.Fatal("clicking Mode should only change UI, not bridge color mode")
	}
	m, _ = m.Update(tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}, nil, nil)
	if m.showColorControls {
		t.Fatal("second click must return to temperature panel")
	}
}
