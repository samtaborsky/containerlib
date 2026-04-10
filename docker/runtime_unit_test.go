// SPDX-License-Identifier: MIT

package docker

import (
	"reflect"
	"testing"
	"time"

	"github.com/moby/moby/api/types/system"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestNew(t *testing.T) {
	rt, err := New("tcp://127.0.0.1:2375")

	if err != nil {
		t.Fatalf("New() failed with valid host: %v", err)
	}
	if rt == nil {
		t.Fatal("New() returned a nil Runtime interface")
	}

	rt, err = New("")

	if err != nil {
		t.Fatalf("New() failed with valid host: %v", err)
	}
	if rt == nil {
		t.Fatal("New() returned a nil Runtime interface")
	}

	internal, ok := rt.(*runtime)
	if !ok {
		t.Fatal("returned interface does not wrap the internal runtime struct")
	}

	if internal.authStore == nil {
		t.Error("New() failed to initialize the authStore map")
	}

	_, err = New("/host")
	if err == nil {
		t.Error("expectedRes error for invalid host scheme, got nil")
	}
}

func TestFromMobySystemInfo(t *testing.T) {
	validTimeStr := "2026-04-04T18:00:00.000000000Z"
	validTime, _ := time.Parse(time.RFC3339Nano, validTimeStr)

	tests := []struct {
		name     string
		input    system.Info
		expected types.InfoResult
	}{
		{
			name: "Happy Path - Valid Time and Data",
			input: system.Info{
				ID:                "mock-id",
				Containers:        10,
				ContainersRunning: 5,
				SystemTime:        validTimeStr,
				Architecture:      "x86_64",
			},
			expected: types.InfoResult{
				ID:                "mock-id",
				Containers:        10,
				ContainersRunning: 5,
				SystemTime:        validTime,
				Architecture:      "x86_64",
			},
		},
		{
			name: "Edge case path - Invalid time string",
			input: system.Info{
				ID:         "mock-id-2",
				SystemTime: "invalid-time",
			},
			expected: types.InfoResult{
				ID:         "mock-id-2",
				SystemTime: time.Time{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := fromMobySystemInfo(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expectedRes %+v, got %+v", tc.expected, res)
			}
		})
	}
}
