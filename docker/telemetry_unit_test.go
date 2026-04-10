// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// trackingReadCloser wraps an io.Reader to simulate an io.ReadCloser
// while tracking if Close() was called and optionally injecting an error.
type trackingReadCloser struct {
	io.Reader
	closeErr    error
	closeCalled bool
}

func (m *trackingReadCloser) Close() error {
	m.closeCalled = true
	return m.closeErr
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestToMobyContainerStatsOpts(t *testing.T) {
	tests := []struct {
		name     string
		input    *types.ContainerStatsOptions
		expected client.ContainerStatsOptions
	}{
		{
			name: "Happy path",
			input: &types.ContainerStatsOptions{
				Stream: true,
			},
			expected: client.ContainerStatsOptions{
				Stream:                true,
				IncludePreviousSample: true,
			},
		},
		{
			name:  "Happy path - Empty input",
			input: nil,
			expected: client.ContainerStatsOptions{
				IncludePreviousSample: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyContainerStatsOpts(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestDecodeStatsStream(t *testing.T) {
	tests := []struct {
		name           string
		mockJSON       string
		cancelCtx      bool
		readerCloseErr error
		expectedStats  int
		expectedErr    string
	}{
		{
			name:          "Happy Path - Single stat",
			mockJSON:      `{"read": "2026-04-04T00:00:00Z", "networks": {}}` + "\n",
			expectedStats: 1,
		},
		{
			name:          "Happy Path - Multiple stats stream",
			mockJSON:      `{"read": "2026-04-04T00:00:00Z"}` + "\n" + `{"read": "2026-04-04T00:00:01Z"}` + "\n",
			expectedStats: 2,
		},
		{
			name:          "Error Path - Malformed JSON",
			mockJSON:      `{"read": "2026-04-04T00:00:00Z", this is not valid json`,
			expectedStats: 0,
			expectedErr:   "invalid character",
		},
		{
			name:          "Error Path - Context Canceled",
			mockJSON:      `{"read": "2026-04-04T00:00:00Z"}`,
			cancelCtx:     true,
			expectedStats: 0,
			expectedErr:   "context canceled",
		},
		{
			name:           "Error Path - Reader Close failure",
			mockJSON:       `{"read": "2026-04-04T00:00:00Z"}`,
			cancelCtx:      false,
			readerCloseErr: fmt.Errorf("simulated close failure"),
			expectedStats:  1,
			expectedErr:    "simulated close failure",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if tc.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}

			mockReader := &trackingReadCloser{
				Reader:   strings.NewReader(tc.mockJSON),
				closeErr: tc.readerCloseErr,
			}

			outStats := make(chan types.ContainerStats, 3)
			outErrors := make(chan error, 3)

			decodeStatsResponse(ctx, mockReader, outStats, outErrors)

			if !mockReader.closeCalled {
				t.Error("expected reader.Close() to be called, but it was not")
			}

			var receivedStats []types.ContainerStats
			for stat := range outStats {
				receivedStats = append(receivedStats, stat)
			}
			if len(receivedStats) != tc.expectedStats {
				t.Errorf("expected %d stats, got %d", tc.expectedStats, len(receivedStats))
			}

			var receivedErrs []error
			for err := range outErrors {
				receivedErrs = append(receivedErrs, err)
			}

			if tc.expectedErr != "" {
				if len(receivedErrs) == 0 {
					t.Fatalf("expected error containing %q, got none", tc.expectedErr)
				}
				found := false
				for _, err := range receivedErrs {
					if strings.Contains(err.Error(), tc.expectedErr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, but got %v", tc.expectedErr, receivedErrs)
				}
			} else if len(receivedErrs) > 0 {
				t.Errorf("expected no errors, got %v", receivedErrs)
			}
		})
	}
}

func TestFromMobyContainerStats(t *testing.T) {
	curTime := time.Now()

	tests := []struct {
		name     string
		input    *container.StatsResponse
		expected types.ContainerStats
	}{
		{
			name: "Happy path",
			input: &container.StatsResponse{
				ID:     "container-id",
				OSType: "linux",
				Read:   curTime,
			},
			expected: types.ContainerStats{
				ID:              "container-id",
				OperatingSystem: "linux",
				ReadTime:        curTime,
			},
		},
		{
			name:     "Happy path - Empty response",
			input:    nil,
			expected: types.ContainerStats{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := fromMobyContainerStats(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestMapToCPUStats(t *testing.T) {
	now := time.Now()
	pre := now.Add(-1 * time.Second)

	mockStatsLinux := &container.StatsResponse{
		Read:    now,
		PreRead: pre,
		OSType:  "linux",
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 500000000,
			},
			SystemUsage: 2000000000,
			OnlineCPUs:  2,
		},
		PreCPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 250000000,
			},
			SystemUsage: 1000000000,
		},
	}

	mockStatsWindows := &container.StatsResponse{
		Read:    now,
		PreRead: pre,
		OSType:  "windows",
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 5000000,
			},
		},
		NumProcs: 4,
		PreCPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 2500000,
			},
		},
	}

	tests := []struct {
		name     string
		input    *container.StatsResponse
		expected types.CPUStats
	}{
		{
			name:  "Happy path - Linux",
			input: mockStatsLinux,
			expected: types.CPUStats{
				UsedPercent: 50.0,
				UsageNano:   500000000,
			},
		},
		{
			name:  "Happy path - Windows",
			input: mockStatsWindows,
			expected: types.CPUStats{
				UsedPercent: 1.0,
				UsageNano:   500000000,
			},
		},
		{
			name:  "Happy path - Zero values",
			input: &container.StatsResponse{OSType: "linux"},
			expected: types.CPUStats{
				UsedPercent: 0.0,
				UsageNano:   0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := mapToCPUStats(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestMapToMemoryStats(t *testing.T) {
	tests := []struct {
		name     string
		input    *container.StatsResponse
		expected types.MemoryStats
	}{
		{
			name: "Happy path",
			input: &container.StatsResponse{
				MemoryStats: container.MemoryStats{
					Usage: uint64(megabytesToBytes(50)),
					Limit: uint64(megabytesToBytes(100)),
				},
			},
			expected: types.MemoryStats{
				UsedPercent: 50.0,
				UsedMb:      50,
				LimitMb:     100,
			},
		},
		{
			name:  "Happy path - Zero values",
			input: &container.StatsResponse{},
			expected: types.MemoryStats{
				UsedPercent: 0,
				UsedMb:      0,
				LimitMb:     0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := mapToMemoryStats(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}
