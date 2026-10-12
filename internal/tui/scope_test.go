package tui

import (
	"testing"

	"github.com/angristan/hue-tui/internal/models"
)

func TestScopeGroup(t *testing.T) {
	office := &models.Room{ID: "office-uuid", IDV1: "/groups/3", Name: "Office"}
	living := &models.Room{ID: "living-uuid", IDV1: "/groups/30", Name: "Living Room"}
	zone := &models.Room{ID: "zone-uuid", IDV1: "/groups/81", Name: "Play Bars", IsZone: true}
	groups := []*models.Room{office, living, zone}
	scenes := []*models.Scene{
		{ID: "office-scene", RoomID: office.ID},
		{ID: "living-scene", RoomID: living.ID},
		{ID: "zone-scene", RoomID: zone.ID},
	}
	for _, tc := range []struct{
		scope string
		wantGroups, wantScenes int
	}{
		{"", 2, 3},
		{"3", 1, 1},
		{"office-uuid", 1, 1},
		{"81", 1, 1},
		{"zone-uuid", 1, 1},
		{"3,81", 2, 2},
		{" 81, 3,81 ", 2, 2},
		{"/groups/81", 1, 1},
	} {
		gotGroups, gotScenes, err := scopeGroup(groups, scenes, tc.scope)
		if err != nil {
			t.Errorf("scope %q: %v", tc.scope, err)
			continue
		}
		if len(gotGroups) != tc.wantGroups || len(gotScenes) != tc.wantScenes {
			t.Errorf("scope %q: got %d groups, %d scenes", tc.scope, len(gotGroups), len(gotScenes))
		}
	}
	for _, scope := range []string{"99", "3,99", "3,", ",81", "3,,81", "300"} {
		g, s, err := scopeGroup(groups, scenes, scope)
		if err == nil || len(g) != 0 || len(s) != 0 {
			t.Errorf("scope %q must fail closed, groups=%v scenes=%v err=%v", scope, g, s, err)
		}
	}
}
