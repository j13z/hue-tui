package screens

import (
	"encoding/json"
	"testing"

	"github.com/angristan/hue-tui/internal/config"
)

func TestStepConfig(t *testing.T) {
	t.Setenv("HUE_TUI_BRIGHTNESS_STEP", "3")
	t.Setenv("HUE_TUI_FINE_BRIGHTNESS_STEP", "1")
	t.Setenv("HUE_TUI_TEMPERATURE_STEP", "invalid")
	t.Setenv("HUE_TUI_FINE_TEMPERATURE_STEP", "-2")
	t.Setenv("HUE_TUI_HUE_STEP", "15")
	t.Setenv("HUE_TUI_FINE_HUE_STEP", "1")
	t.Setenv("HUE_TUI_SATURATION_STEP", "5")
	t.Setenv("HUE_TUI_FINE_SATURATION_STEP", "0")
	cfg := loadStepConfig()
	if cfg.Brightness != 3 || cfg.FineBrightness != 1 || cfg.Temperature != 25 || cfg.FineTemperature != 5 || cfg.Hue != 15 || cfg.FineHue != 1 || cfg.Saturation != 5 || cfg.FineSaturation != 1 {
		t.Fatalf("unexpected steps: %+v", cfg)
	}
	if pickStep(5, 1, false) != 5 || pickStep(5, 1, true) != 1 {
		t.Fatal("fine step selection broken")
	}
	if hueUnits(20) != 3641 || satUnits(1) != 3 {
		t.Fatal("color step conversion broken")
	}
}

func TestJSONTuningAndEnvironmentPriority(t *testing.T) {
	t.Setenv("HUE_TUI_BRIGHTNESS_STEP", "")
	t.Setenv("HUE_TUI_FINE_BRIGHTNESS_STEP", "")
	var cfg config.Config
	if err := json.Unmarshal([]byte(`{"tuning":{"brightness_step":7,"fine_brightness_step":2,"temperature_step":10}}`), &cfg); err != nil {
		t.Fatal(err)
	}
	steps := loadStepConfigWithTuning(cfg.Tuning)
	if steps.Brightness != 7 || steps.FineBrightness != 2 || steps.Temperature != 10 {
		t.Fatalf("JSON tuning not applied: %+v", steps)
	}
	t.Setenv("HUE_TUI_BRIGHTNESS_STEP", "4")
	if got := loadStepConfigWithTuning(cfg.Tuning).Brightness; got != 4 {
		t.Fatalf("environment override failed: got %d", got)
	}
	t.Setenv("HUE_TUI_BRIGHTNESS_STEP", "")
	cfg.Tuning.BrightnessStep = -3
	if got := loadStepConfigWithTuning(cfg.Tuning).Brightness; got != 3 {
		t.Fatalf("invalid JSON value must use default 3, got %d", got)
	}
}

func TestFineAdjustmentKeys(t *testing.T) {
	for _, key := range []string{"alt+left", "alt+right", "shift+left", "shift+right", "ctrl+left", "ctrl+right", "H", "L", "W", "C", "{", "}", "_", "+"} {
		if !fineAdjustmentKey(key) {
			t.Errorf("key %q should be fine adjust", key)
		}
	}
	for _, key := range []string{"left", "right", "h", "l", "w", "c", "[", "]", "-", "="} {
		if fineAdjustmentKey(key) {
			t.Errorf("key %q should use standard adjustment", key)
		}
	}
}
