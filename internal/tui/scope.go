package tui

import (
	"fmt"

	"github.com/angristan/hue-tui/internal/models"
)

// scopeGroup restricts the room and scene lists to the requested Hue room.
// It accepts a v1 group ID (such as "3") or a v2 room UUID.
func scopeGroup(rooms []*models.Room, scenes []*models.Scene, group string) ([]*models.Room, []*models.Scene, error) {
	if group == "" {
		return rooms, scenes, nil
	}

	legacyID := "/groups/" + group
	for _, room := range rooms {
		if room.ID == group || room.IDV1 == legacyID {
			var filteredScenes []*models.Scene
			for _, scene := range scenes {
				if scene.RoomID == room.ID {
					filteredScenes = append(filteredScenes, scene)
				}
			}
			return []*models.Room{room}, filteredScenes, nil
		}
	}

	return nil, nil, fmt.Errorf("Hue room for group %q not found (v1 group ID or v2 room UUID expected)", group)
}
