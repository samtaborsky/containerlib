// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestToMobyImagePullOpts(t *testing.T) {
	tests := []struct {
		name     string
		opts     *types.ImagePullOptions
		expected client.ImagePullOptions
	}{
		{
			name: "Happy path",
			opts: &types.ImagePullOptions{
				All: true,
			},
			expected: client.ImagePullOptions{
				All: true,
			},
		},
		{
			name:     "Happy path - Empty input",
			opts:     nil,
			expected: client.ImagePullOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyImagePullOpts(tc.opts)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestToMobyImageListOpts(t *testing.T) {
	tests := []struct {
		name     string
		opts     *types.ImageListOptions
		expected client.ImageListOptions
	}{
		{
			name: "Happy path",
			opts: &types.ImageListOptions{
				All: true,
			},
			expected: client.ImageListOptions{
				All: true,
			},
		},
		{
			name:     "Happy path - Empty input",
			opts:     nil,
			expected: client.ImageListOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyImageListOpts(tc.opts)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestToMobyImageRemoveOpts(t *testing.T) {
	tests := []struct {
		name     string
		opts     *types.ImageRemoveOptions
		expected client.ImageRemoveOptions
	}{
		{
			name: "Happy path",
			opts: &types.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			},
			expected: client.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			},
		},
		{
			name:     "Happy path - Empty input",
			opts:     nil,
			expected: client.ImageRemoveOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyImageRemoveOpts(tc.opts)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestToMobyImagePruneOpts(t *testing.T) {
	tests := []struct {
		name     string
		opts     *types.ImagePruneOptions
		expected client.ImagePruneOptions
	}{
		{
			name: "Happy path",
			opts: &types.ImagePruneOptions{
				Filters: make(types.Filters).Add("foo", "bar"),
			},
			expected: client.ImagePruneOptions{
				Filters: make(client.Filters).Add("foo", "bar"),
			},
		},
		{
			name:     "Happy path - Empty input",
			opts:     nil,
			expected: client.ImagePruneOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyImagePruneOpts(tc.opts)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestFromMobyImageInspectResult(t *testing.T) {
	tests := []struct {
		name     string
		input    client.ImageInspectResult
		expected types.ImageInspectResult
	}{
		{
			name: "Happy path",
			input: client.ImageInspectResult{
				InspectResponse: image.InspectResponse{ID: "valid-id"},
			},
			expected: types.ImageInspectResult{
				ID: "valid-id",
			},
		},
		{
			name: "Edge case path - Invalid time string",
			input: client.ImageInspectResult{
				InspectResponse: image.InspectResponse{Created: "invalid-time"},
			},
			expected: types.ImageInspectResult{
				Created: time.Time{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := fromMobyImageInspectResult(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestFromMobyImageListResult(t *testing.T) {
	tests := []struct {
		name     string
		input    client.ImageListResult
		expected types.ImageListResult
	}{
		{
			name: "Happy path",
			input: client.ImageListResult{
				Items: []image.Summary{
					{ID: "valid-id-1", RepoTags: []string{"foo:bar"}},
					{ID: "valid-id-2", Created: 20},
				},
			},
			expected: types.ImageListResult{
				Images: []types.ImageSummary{
					{ID: "valid-id-1", Tags: []string{"foo:bar"}},
					{ID: "valid-id-2", Created: time.Unix(20, 0)},
				},
			},
		},
		{
			name: "Edge case path - Zero time returned",
			input: client.ImageListResult{
				Items: []image.Summary{
					{ID: "valid-id-1", Created: 0},
				},
			},
			expected: types.ImageListResult{
				Images: []types.ImageSummary{
					{ID: "valid-id-1", Created: time.Time{}},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := fromMobyImageListResult(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestFromMobyImageRemoveResult(t *testing.T) {
	tests := []struct {
		name     string
		input    client.ImageRemoveResult
		expected types.ImageRemoveResult
	}{
		{
			name: "Happy path",
			input: client.ImageRemoveResult{
				Items: []image.DeleteResponse{
					{Deleted: "tag-1", Untagged: "tag-1"},
					{Deleted: "tag-2", Untagged: "tag-2"},
				},
			},
			expected: types.ImageRemoveResult{
				ImagesRemoved: []types.ImageRemoveSummary{
					{Deleted: "tag-1", Untagged: "tag-1"},
					{Deleted: "tag-2", Untagged: "tag-2"},
				},
			},
		},
		{
			name:     "Happy path - No removed images",
			input:    client.ImageRemoveResult{},
			expected: types.ImageRemoveResult{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := fromMobyImageRemoveResult(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestFromMobyImagePruneReport(t *testing.T) {
	tests := []struct {
		name     string
		input    client.ImagePruneResult
		expected types.ImagePruneResult
	}{
		{
			name: "Happy path",
			input: client.ImagePruneResult{
				Report: image.PruneReport{
					ImagesDeleted: []image.DeleteResponse{
						{Deleted: "tag-1", Untagged: "tag-1"},
						{Deleted: "tag-2", Untagged: "tag-2"}},
					SpaceReclaimed: 20,
				},
			},
			expected: types.ImagePruneResult{
				ImagesRemoved: []types.ImageRemoveSummary{
					{Deleted: "tag-1", Untagged: "tag-1"},
					{Deleted: "tag-2", Untagged: "tag-2"},
				},
				SpaceReclaimed: 20,
			},
		},
		{
			name:     "Happy path - No pruned images",
			input:    client.ImagePruneResult{},
			expected: types.ImagePruneResult{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := fromMobyImagePruneResult(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestEncodeRegistryAuth(t *testing.T) {
	tests := []struct {
		name        string
		input       types.AuthConfig
		expected    string
		expectedErr error
	}{
		{
			name: "Happy path",
			input: types.AuthConfig{
				Username:      "foo",
				Password:      "bar",
				ServerAddress: "docker.io",
			},
			expected:    "eyJ1c2VybmFtZSI6ImZvbyIsInBhc3N3b3JkIjoiYmFyIiwic2VydmVyYWRkcmVzcyI6ImRvY2tlci5pbyJ9",
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, _ := encodeRegistryAuth(tc.input)

			if res != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestDecodePullStream(t *testing.T) {
	tests := []struct {
		name          string
		outChan       chan types.ImagePullProgress
		mockMsgs      []jsonstream.Message
		mockErr       error
		cancelCtx     bool
		expectedCount int
		expectedErr   string
	}{
		{
			name:    "Happy path - Nil channel",
			outChan: nil,
			mockMsgs: []jsonstream.Message{
				{Status: "Already done"},
			},
			mockErr:       nil,
			expectedCount: 0,
		},
		{
			name:    "Happy path - Full stream success",
			outChan: make(chan types.ImagePullProgress, 5),
			mockMsgs: []jsonstream.Message{
				{Status: "Downloading"},
				{Status: "Extracting"},
			},
			mockErr:       nil,
			expectedCount: 2,
		},
		{
			name:          "Error path - Failure in Wait",
			outChan:       nil,
			mockMsgs:      nil,
			mockErr:       errors.New("connection lost"),
			expectedCount: 0,
			expectedErr:   "image pull failed",
		},
		{
			name:    "Error path - Stream contains JSON error",
			outChan: make(chan types.ImagePullProgress, 5),
			mockMsgs: []jsonstream.Message{
				{Error: &jsonstream.Error{
					Message: "no space left", Code: 1},
				},
			},
			mockErr:       nil,
			expectedCount: 0,
			expectedErr:   "no space left",
		},
		{
			name:    "Error path - Context cancelled",
			outChan: make(chan types.ImagePullProgress, 5),
			mockMsgs: []jsonstream.Message{
				{Status: "Downloading"},
			},
			cancelCtx:   true,
			expectedErr: "context canceled",
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

			resp := &mockImagePullResponse{
				messages: tc.mockMsgs,
				err:      tc.mockErr,
			}

			err := decodePullStream(ctx, resp, tc.outChan)

			if tc.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.expectedErr)
				}
				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Fatalf("expected error containing %q, got %q", tc.expectedErr, err.Error())
				}
				return
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tc.outChan != nil {
				count := 0
				for range tc.outChan {
					count++
				}
				if count != tc.expectedCount && !tc.cancelCtx {
					t.Errorf("expected %d events, got %d", tc.expectedCount, count)
				}
			}
		})
	}
}

func TestMapToPullProgress(t *testing.T) {
	auxMsg := json.RawMessage(`{"ID": "sha256:valid-digest"}`)

	tests := []struct {
		name     string
		input    jsonstream.Message
		expected types.ImagePullProgress
	}{
		{
			name: "Happy path",
			input: jsonstream.Message{
				Status: "Downloading",
				ID:     "valid-id",
				Progress: &jsonstream.Progress{
					Current: 2,
					Total:   2,
				},
			},
			expected: types.ImagePullProgress{
				Status:  "Downloading",
				ID:      "valid-id",
				Current: 2,
				Total:   2,
			},
		},
		{
			name: "Happy path - Digest",
			input: jsonstream.Message{
				Status: "Downloading",
				ID:     "valid-id",
				Aux:    &auxMsg,
			},
			expected: types.ImagePullProgress{
				Status: "Downloading",
				ID:     "valid-id",
				Digest: "sha256:valid-digest",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := mapToPullProgress(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}
