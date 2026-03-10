package adapters

import (
	"fmt"
	"math/rand/v2"
)

type FlakyAdapter struct {
	FailureRate float64 // 0.0 to 1.0 (e.g., 0.7 = 70% failure rate)
}

func NewFlakyAdapter(failureRate float64) *FlakyAdapter {
	return &FlakyAdapter{FailureRate: failureRate}
}

func (f *FlakyAdapter) Send(recipient, message string) error {
	// Random failure simulation
	if rand.Float64() < f.FailureRate {
		return fmt.Errorf("network timeout (simulated)")
	}

	fmt.Printf("✅ [FLAKY] Successfully sent to %s: %s\n", recipient, message)
	return nil
}
