package screens

import (
	"os"
	"strconv"
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

func loadStepConfig() StepConfig {
	return StepConfig{
		Brightness:      readStep("HUE_TUI_BRIGHTNESS_STEP", 3, 100),
		FineBrightness:  readStep("HUE_TUI_FINE_BRIGHTNESS_STEP", 1, 100),
		Temperature:     readStep("HUE_TUI_TEMPERATURE_STEP", 25, 347),
		FineTemperature: readStep("HUE_TUI_FINE_TEMPERATURE_STEP", 5, 347),
		Hue:             readStep("HUE_TUI_HUE_STEP", 20, 360),
		FineHue:         readStep("HUE_TUI_FINE_HUE_STEP", 2, 360),
		Saturation:      readStep("HUE_TUI_SATURATION_STEP", 10, 100),
		FineSaturation:  readStep("HUE_TUI_FINE_SATURATION_STEP", 1, 100),
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
