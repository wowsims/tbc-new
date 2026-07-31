//go:build !(js && wasm)

package reforgeoptimizer

import (
	"strconv"
	"testing"
)

func TestHiGHSWasmRuntimeConcurrencyUsesEnvOverride(t *testing.T) {
	t.Setenv("WOWSIMS_HIGHS_WASM_RUNTIME_CONCURRENCY", "7")

	if got := getHiGHSWasmRuntimeConcurrency(); got != 7 {
		t.Fatalf("runtime concurrency = %d, want env override 7", got)
	}
}

func TestDefaultHiGHSWasmRuntimeConcurrency(t *testing.T) {
	testCases := []struct {
		numCPU int
		want   int
	}{
		{numCPU: 1, want: 1},
		{numCPU: 2, want: 2},
		{numCPU: 4, want: 4},
		{numCPU: 18, want: 18},
	}

	for _, testCase := range testCases {
		t.Run(strconv.Itoa(testCase.numCPU), func(t *testing.T) {
			if got := defaultHiGHSWasmRuntimeConcurrency(testCase.numCPU); got != testCase.want {
				t.Fatalf("default concurrency for %d CPUs = %d, want %d", testCase.numCPU, got, testCase.want)
			}
		})
	}
}
