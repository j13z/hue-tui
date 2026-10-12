package screens

import "testing"

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
