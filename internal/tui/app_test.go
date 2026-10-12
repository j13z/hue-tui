package tui

import (
	"testing"

	"github.com/angristan/hue-tui/internal/config"
	"github.com/angristan/hue-tui/internal/tui/messages"
)

func TestDemoModeInit(t *testing.T) {
	// Create a demo mode model
	cfg := &config.Config{}
	model := NewModel(cfg, true)

	t.Logf("Initial state: screen=%d, demoMode=%v, bridge=%v", model.screen, model.demoMode, model.bridge != nil)

	if model.screen != ScreenMain {
		t.Errorf("Expected ScreenMain, got %d", model.screen)
	}
	if !model.demoMode {
		t.Error("Expected demoMode to be true")
	}
	if model.bridge == nil {
		t.Fatal("Expected bridge to be set")
	}

	// Simulate the fetchDataCmd directly
	fetchCmd := model.fetchDataCmd()
	fetchMsg := fetchCmd()
	t.Logf("fetchDataCmd returned: %T", fetchMsg)

	if dataMsg, ok := fetchMsg.(messages.DataFetchedMsg); ok {
		t.Logf("DataFetchedMsg: %d rooms, %d scenes", len(dataMsg.Rooms), len(dataMsg.Scenes))

		if len(dataMsg.Rooms) == 0 {
			t.Fatal("DataFetchedMsg.Rooms is empty!")
		}
		if len(dataMsg.Scenes) == 0 {
			t.Fatal("DataFetchedMsg.Scenes is empty!")
		}

		// Now send this message to Update
		newModel, _ := model.Update(dataMsg)
		updatedModel := newModel.(Model)

		// Check via View output - it should show "Connected" not "Loading"
		view := updatedModel.View()
		t.Logf("View output (first 200 chars): %s", view[:min(200, len(view))])

		if contains(view, "Loading") {
			t.Error("View should not contain 'Loading' after DataFetchedMsg")
		}
		if !contains(view, "Connected") {
			t.Error("View should contain 'Connected' after DataFetchedMsg")
		}
	} else if errMsg, ok := fetchMsg.(messages.ErrorMsg); ok {
		t.Fatalf("fetchDataCmd returned error: %v", errMsg.Err)
	} else {
		t.Fatalf("fetchDataCmd returned unexpected type: %T", fetchMsg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestFetchDataCmdGroupScope(t *testing.T) {
	model := NewModel(&config.Config{}, true)
	unfiltered := model.fetchDataCmd()()
	data, ok := unfiltered.(messages.DataFetchedMsg)
	if !ok || len(data.Rooms) == 0 {
		t.Fatalf("expected demo rooms, got %T", unfiltered)
	}

	want := data.Rooms[0].ID
	model.SetGroupScope(want)
	filtered := model.fetchDataCmd()()
	scoped, ok := filtered.(messages.DataFetchedMsg)
	if !ok {
		t.Fatalf("expected scoped DataFetchedMsg, got %T", filtered)
	}
	if len(scoped.Rooms) != 1 || scoped.Rooms[0].ID != want {
		t.Fatalf("wrong scoped rooms: %+v", scoped.Rooms)
	}
	for _, scene := range scoped.Scenes {
		if scene.RoomID != want {
			t.Fatalf("out-of-scope scene: %+v", scene)
		}
	}

	model.SetGroupScope("unknown-group")
	unknown := model.fetchDataCmd()()
	if _, ok := unknown.(messages.ErrorMsg); !ok {
		t.Fatalf("unknown group must fail closed, got %T", unknown)
	}
}
