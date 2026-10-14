package screens

import (
	"testing"

	"github.com/angristan/hue-tui/internal/models"
)

func TestGroupDisplayOrder(t *testing.T) {
	makeGroups := func() []*models.Room {
		return []*models.Room{
			{ID: "zone-81", Name: "Play Bars", IsZone: true, Lights: []*models.Light{{ID: "play-2", Name: "Zwei", On: true}, {ID: "play-1", Name: "Eins", On: true}}},
			{ID: "room-3", Name: "Arbeitszimmer", Lights: []*models.Light{{ID: "desk", Name: "Desk", On: true}}},
		}
	}
	for _, tc := range []struct {
		name string
		explicit bool
		wantFirst string
	}{
		{"explicit CLI order", true, "Play Bars"},
		{"default alphabetical order", false, "Arbeitszimmer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMainModel(nil)
			m.SetSize(120, 30)
			m.SetExplicitGroupOrder(tc.explicit)
			m.SetData(makeGroups(), nil)
			if got := m.rooms[0].Name; got != tc.wantFirst {
				t.Fatalf("first group is %q; want %q", got, tc.wantFirst)
			}
			if len(m.items) < 1 || !m.items[0].isRoom || m.items[0].room.Name != tc.wantFirst {
				t.Fatalf("displayed group order does not match model: %#v", m.items)
			}
		})
	}
}

func TestGroupDisplayOrderPersistsAcrossRefresh(t *testing.T) {
	m := NewMainModel(nil)
	m.SetSize(120, 30)
	m.SetExplicitGroupOrder(true)
	groupA := &models.Room{ID: "zone-81", Name: "Play Bars", Lights: []*models.Light{{ID: "play", Name: "Play", On: true}}}
	groupB := &models.Room{ID: "room-3", Name: "Arbeitszimmer", Lights: []*models.Light{{ID: "desk", Name: "Desk", On: true}}}
	for i := 0; i < 2; i++ {
		m.SetData([]*models.Room{groupA, groupB}, nil)
		if m.rooms[0] != groupA || m.rooms[1] != groupB {
			t.Fatalf("iteration %d: explicit order lost", i)
		}
	}
	m.SetExplicitGroupOrder(false)
	m.SetData([]*models.Room{groupA, groupB}, nil)
	if m.rooms[0] != groupB {
		t.Fatal("switching back to unscoped mode should restore alphabetical sorting")
	}
}
