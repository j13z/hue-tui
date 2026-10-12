package screens

import (
	"os"
	"strconv"
	"strings"

	"github.com/angristan/hue-tui/internal/config"
)

// StepConfig holds normal/fine adjustment increments, in percentage points,
// mirek (temperature), and degrees (hue).
type StepConfig struct {
	Brightness, FineBrightness   int
	Temperature, FineTemperature int
	Hue, FineHue                 int
	Saturation, FineSaturation   int
}

func readStep(name string, defaultValue, maximum int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maximum {
		return defaultValue
	}
	return n
}

// If a JSON step is omitted or invalid, use the built-in default.
func configuredStep(n, defaultValue, maxValue int) int {
	if n < 1 || n > maxValue {
		return defaultValue
	}
	return n
}

func loadStepConfig() StepConfig { return loadStepConfigWithTuning(nil) }

func loadStepConfigWithTuning(tuning *config.TuningConfig) StepConfig {
	var t config.TuningConfig
	if tuning != nil {
		t = *tuning
	}
	// Priority: environment variable > JSON config > built-in default.
	return StepConfig{
		Brightness:      readStep("HUE_TUI_BRIGHTNESS_STEP", configuredStep(t.BrightnessStep, 3, 100), 100),
		FineBrightness:  readStep("HUE_TUI_FINE_BRIGHTNESS_STEP", configuredStep(t.FineBrightnessStep, 1, 100), 100),
		Temperature:     readStep("HUE_TUI_TEMPERATURE_STEP", configuredStep(t.TemperatureStep, 25, 347), 347),
		FineTemperature: readStep("HUE_TUI_FINE_TEMPERATURE_STEP", configuredStep(t.FineTemperatureStep, 5, 347), 347),
		Hue:             readStep("HUE_TUI_HUE_STEP", configuredStep(t.HueStep, 20, 360), 360),
		FineHue:         readStep("HUE_TUI_FINE_HUE_STEP", configuredStep(t.FineHueStep, 2, 360), 360),
		Saturation:      readStep("HUE_TUI_SATURATION_STEP", configuredStep(t.SaturationStep, 10, 100), 100),
		FineSaturation:  readStep("HUE_TUI_FINE_SATURATION_STEP", configuredStep(t.FineSaturationStep, 1, 100), 100),
	}
}

func fineAdjustmentKey(key string) bool {
	if strings.HasPrefix(key, "alt+") {
		return true
	}
	switch key {
	case "H", "L", "W", "C", "{", "}", "_", "+":
		return true
	default:
		return false
	}
}

func pickStep(normal, fine int, useFine bool) int {
	if useFine {
		return fine
	}
	return normal
}

func hueUnits(degrees int) int {
	return (degrees*65536 + 180) / 360
}

func satUnits(percent int) int {
	return (percent*254 + 50) / 100
}
