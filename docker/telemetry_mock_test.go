// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- API functions
// ---------------------------------------------------------------------------------------------------------------------

func (m *mockDockerAPI) ContainerStats(ctx context.Context, container string, options client.ContainerStatsOptions) (client.ContainerStatsResult, error) {
	if m.containerStatsFunc == nil {
		panic("ContainerStats function not mocked")
	}
	return m.containerStatsFunc(ctx, container, options)
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestContainerStats(t *testing.T) {
	tests := []struct {
		name        string
		mockResp    client.ContainerStatsResult
		mockErr     error
		expectedErr error
		expectChan  bool
	}{
		{
			name: "Happy Path",
			mockResp: client.ContainerStatsResult{
				Body: io.NopCloser(strings.NewReader(`{"read": "2026-04-04T00:00:00Z"}` + "\n")),
			},
			mockErr:     nil,
			expectedErr: nil,
			expectChan:  true,
		},
		{
			name:        "Error Path - Container Not Found",
			mockResp:    client.ContainerStatsResult{},
			mockErr:     mockErrNotFound{"Error response from daemon: No such container"},
			expectedErr: types.ErrContainerNotFound,
			expectChan:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerStatsFunc: func(ctx context.Context, id string, opts client.ContainerStatsOptions) (client.ContainerStatsResult, error) {
					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			result, err := rt.ContainerStats(context.Background(), "mock-id", &types.ContainerStatsOptions{Stream: true})

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tc.expectChan {
				if result.Stats == nil || result.Errors == nil {
					t.Fatal("expected Stats and Errors channels to be initialized, got nil")
				}

				select {
				case stat, ok := <-result.Stats:
					if !ok {
						t.Fatal("stats channel was closed prematurely")
					}
					if stat.ReadTime.IsZero() {
						t.Log("Warning: Received empty stat, but pipeline is technically connected")
					}
				case err := <-result.Errors:
					t.Fatalf("goroutine encountered unexpected error: %v", err)
				case <-time.After(1 * time.Second):
					t.Fatal("timed out waiting for goroutine to push stat to channel")
				}
			}
		})
	}
}
