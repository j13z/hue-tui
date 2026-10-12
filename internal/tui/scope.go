package tui

import (
	"fmt"
	"strings"

	"github.com/angristan/hue-tui/internal/models"
)

// scopeGroup accepts comma-separated Hue v1 group IDs or v2 room/zone UUIDs.
// An invalid member fails closed: no controls are returned.
func scopeGroup(groups []*models.Room, scenes []*models.Scene, scope string) ([]*models.Room, []*models.Scene, error) {
	if strings.TrimSpace(scope) == "" {
		// Preserve the pre-zone room-only default view.
		var rooms []*models.Room
		for _, group := range groups {
			if !group.IsZone {
				rooms = append(rooms, group)
			}
		}
		return rooms, scenes, nil
	}

	byID := make(map[string]*models.Room)
	for _, group := range groups {
		byID[group.ID] = group
		if group.IDV1 != "" {
			byID[group.IDV1] = group
		}
	}

	var selected []*models.Room
	wanted := make(map[string]bool)
	for _, raw := range strings.Split(scope, ",") {
		id := strings.TrimSpace(raw)
		if id == "" {
			return nil, nil, fmt.Errorf("empty group ID in %q", scope)
		}
		group := byID[id]
		if group == nil {
			group = byID["/groups/"+id]
		}
		if group == nil {
			return nil, nil, fmt.Errorf("Hue group %q not found (v1 group ID or v2 room/zone UUID expected)", id)
		}
		if !wanted[group.ID] {
			wanted[group.ID] = true
			selected = append(selected, group)
		}
	}

	var selectedScenes []*models.Scene
	for _, scene := range scenes {
		if wanted[scene.RoomID] {
			selectedScenes = append(selectedScenes, scene)
		}
	}
	return selected, selectedScenes, nil
}
