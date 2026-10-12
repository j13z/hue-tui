package models

import "testing"

func TestBrightnessPercentRoundTrip(t *testing.T) {
	for pct := 0; pct <= 100; pct++ {
		light := Light{}
		light.SetBrightnessPct(pct)
		if got := light.BrightnessPct(); got != pct {
			t.Fatalf("brightness %d%% rounded to %d%%", pct, got)
		}
	}
}
