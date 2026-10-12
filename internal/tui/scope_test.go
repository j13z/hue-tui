package tui

import (
	"testing"

	"github.com/angristan/hue-tui/internal/models"
)

func TestScopeGroup(t *testing.T) {
	office := &models.Room{ID: "office-uuid", IDV1: "/groups/3", Name: "Office"}
	living := &models.Room{ID: "living-uuid", IDV1: "/groups/30", Name: "Living Room"}
	rooms := []*models.Room{office, living}
	officeScene := &models.Scene{ID: "office-scene", RoomID: office.ID}
	livingScene := &models.Scene{ID: "living-scene", RoomID: living.ID}
	scenes := []*models.Scene{officeScene, livingScene}

	for _, group := range []string{"3", "office-uuid"} {
		t.Run(group, func(t *testing.T) {
			gotRooms, gotScenes, err := scopeGroup(rooms, scenes, group)
			if err != nil {
				t.Fatal(err)
			}
			if len(gotRooms) != 1 || gotRooms[0] != office {
				t.Fatalf("unexpected rooms: %+v", gotRooms)
			}
			if len(gotScenes) != 1 || gotScenes[0] != officeScene {
				t.Fatalf("unexpected scenes: %+v", gotScenes)
			}
		})
	}

	t.Run("unscoped", func(t *testing.T) {
		gotRooms, gotScenes, err := scopeGroup(rooms, scenes, "")
		if err != nil || len(gotRooms) != 2 || len(gotScenes) != 2 {
			t.Fatalf("unexpected unscoped result: rooms=%v scenes=%v err=%v", gotRooms, gotScenes, err)
		}
	})

	t.Run("unknown does not expose other rooms", func(t *testing.T) {
		for _, group := range []string{"4", "300", "unknown"} {
			gotRooms, gotScenes, err := scopeGroup(rooms, scenes, group)
			if err == nil || len(gotRooms) != 0 || len(gotScenes) != 0 {
				t.Fatalf("group=%q: expected error and no rooms/scenes, got rooms=%v scenes=%v err=%v", group, gotRooms, gotScenes, err)
			}
		}
	})
}
