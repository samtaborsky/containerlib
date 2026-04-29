// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"reflect"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- API functions
// ---------------------------------------------------------------------------------------------------------------------

func (m *mockDockerAPI) ImagePull(ctx context.Context, ref string, options client.ImagePullOptions) (client.ImagePullResponse, error) {
	if m.imagePullFunc == nil {
		panic("ImagePull function not mocked")
	}
	return m.imagePullFunc(ctx, ref, options)
}

func (m *mockDockerAPI) ImageInspect(ctx context.Context, image string, _ ...client.ImageInspectOption) (client.ImageInspectResult, error) {
	if m.imageInspectFunc == nil {
		panic("ImageInspect function not mocked")
	}
	return m.imageInspectFunc(ctx, image)
}

func (m *mockDockerAPI) ImageList(ctx context.Context, options client.ImageListOptions) (client.ImageListResult, error) {
	if m.imageListFunc == nil {
		panic("ImageList function not mocked")
	}
	return m.imageListFunc(ctx, options)
}

func (m *mockDockerAPI) ImageRemove(ctx context.Context, image string, options client.ImageRemoveOptions) (client.ImageRemoveResult, error) {
	if m.imageRemoveFunc == nil {
		panic("ImageRemove function not mocked")
	}
	return m.imageRemoveFunc(ctx, image, options)
}

func (m *mockDockerAPI) ImagePrune(ctx context.Context, opts client.ImagePruneOptions) (client.ImagePruneResult, error) {
	if m.imagePruneFunc == nil {
		panic("ImagePrune function not mocked")
	}
	return m.imagePruneFunc(ctx, opts)
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Helper functions
// ---------------------------------------------------------------------------------------------------------------------

type mockImagePullResponse struct {
	io.ReadCloser
	messages []jsonstream.Message
	err      error
}

func (m *mockImagePullResponse) JSONMessages(_ context.Context) iter.Seq2[jsonstream.Message, error] {
	return func(yield func(jsonstream.Message, error) bool) {
		for _, msg := range m.messages {
			if !yield(msg, nil) {
				return
			}
		}
		if m.err != nil {
			yield(jsonstream.Message{}, m.err)
		}
	}
}

func (m *mockImagePullResponse) Wait(_ context.Context) error {
	return m.err
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestImagePull(t *testing.T) {
	tests := []struct {
		name           string
		imageRef       string
		noOpts         bool
		progressChan   chan types.ImagePullProgress
		mockMsgs       []jsonstream.Message
		mockErr        error
		expectedEvents int
		expectedErr    error
	}{
		{
			name:         "Happy path",
			imageRef:     "nginx:latest",
			progressChan: make(chan types.ImagePullProgress, 3),
			mockMsgs: []jsonstream.Message{
				{Status: "Pulling from library/nginx", ID: "latest"},
				{Status: "Download complete", ID: "12345"},
			},
			mockErr:        nil,
			expectedEvents: 2,
			expectedErr:    nil,
		},
		{
			name:           "Happy path - No progress",
			imageRef:       "nginx:latest",
			progressChan:   nil,
			mockMsgs:       nil,
			mockErr:        nil,
			expectedEvents: 0,
			expectedErr:    nil,
		},
		{
			name:           "Happy path - Nil options",
			imageRef:       "nginx:latest",
			noOpts:         true,
			progressChan:   nil,
			mockMsgs:       nil,
			mockErr:        nil,
			expectedEvents: 0,
			expectedErr:    nil,
		},
		{
			name:           "Error path - Image not found",
			imageRef:       "invalid-image:latest",
			progressChan:   nil,
			mockMsgs:       nil,
			mockErr:        mockErrNotFound{"Error response from daemon: No such image: invalid-image:latest"},
			expectedEvents: 0,
			expectedErr:    types.ErrImageNotFound,
		},
		{
			name:           "Error path - Invalid image name",
			imageRef:       "DOCKER-image:latest",
			progressChan:   nil,
			mockMsgs:       nil,
			mockErr:        nil,
			expectedEvents: 0,
			expectedErr:    types.ErrInvalidInput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				imagePullFunc: func(ctx context.Context, ref string, options client.ImagePullOptions) (client.ImagePullResponse, error) {
					if ref != fmt.Sprintf("docker.io/library/%s", tc.imageRef) {
						t.Errorf("expected Docker to receive image %q, got %q", fmt.Sprintf("docker.io/library/%s", tc.imageRef), ref)
					}

					return &mockImagePullResponse{
						ReadCloser: io.NopCloser(strings.NewReader("")),
						messages:   tc.mockMsgs,
					}, tc.mockErr
				},
			}

			opts := &types.ImagePullOptions{Progress: tc.progressChan}
			if tc.noOpts == true {
				opts = nil
			}

			rt := &runtime{api: mockAPI}
			err := rt.ImagePull(context.Background(), tc.imageRef, opts)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tc.progressChan != nil && tc.noOpts == false {
				count := 0
				for range tc.progressChan {
					count++
				}
				if count != tc.expectedEvents {
					t.Errorf("expected %d events, got %d", tc.expectedEvents, count)
				}
			}
		})
	}
}

func TestImageInspect(t *testing.T) {
	tests := []struct {
		name         string
		imageID      string
		mockResp     client.ImageInspectResult
		mockErr      error
		expectedResp types.ImageInspectResult
		expectedErr  error
	}{
		{
			name:    "Happy path",
			imageID: "valid-image-id",
			mockResp: client.ImageInspectResult{
				InspectResponse: image.InspectResponse{ID: "valid-image-id"},
			},
			mockErr:      nil,
			expectedResp: types.ImageInspectResult{ID: "valid-image-id"},
			expectedErr:  nil,
		},
		{
			name:         "Error path - Invalid image ID",
			imageID:      "invalid-image-id",
			mockResp:     client.ImageInspectResult{},
			mockErr:      mockErrNotFound{"Error response from daemon: No such image: invalid-image-id"},
			expectedResp: types.ImageInspectResult{},
			expectedErr:  types.ErrImageNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				imageInspectFunc: func(ctx context.Context, image string, _ ...client.ImageInspectOption) (client.ImageInspectResult, error) {
					if image != tc.imageID {
						t.Errorf("expected Docker to receive image %q, got %q", tc.imageID, image)
					}

					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ImageInspect(context.Background(), tc.imageID, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if resp.ID != tc.expectedResp.ID {
				t.Errorf("expected returned ID to be %q, got %q", tc.expectedResp.ID, resp.ID)
			}
		})
	}
}

func TestImageList(t *testing.T) {
	tests := []struct {
		name         string
		mockResp     client.ImageListResult
		mockErr      error
		expectedResp types.ImageListResult
		expectedErr  error
	}{
		{
			name: "Happy path",
			mockResp: client.ImageListResult{
				Items: []image.Summary{
					{Containers: 1, ID: "image-id-1"},
					{Containers: 2, ID: "image-id-2"},
				},
			},
			mockErr: nil,
			expectedResp: types.ImageListResult{
				Images: []types.ImageSummary{
					{ID: "image-id-1"},
					{ID: "image-id-2"},
				},
			},
			expectedErr: nil,
		},
		{
			name:         "Error path - Daemon unreachable",
			mockResp:     client.ImageListResult{},
			mockErr:      generateRealConnectionError(),
			expectedResp: types.ImageListResult{},
			expectedErr:  types.ErrConnectionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				imageListFunc: func(ctx context.Context, options client.ImageListOptions) (client.ImageListResult, error) {
					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ImageList(context.Background(), nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(resp.Images, tc.expectedResp.Images) {
				t.Errorf("expected config %+v, got %+v", tc.expectedResp.Images, resp.Images)
			}
		})
	}
}

func TestImageRemove(t *testing.T) {
	tests := []struct {
		name         string
		inputID      string
		mockResp     client.ImageRemoveResult
		mockErr      error
		expectedResp types.ImageRemoveResult
		expectedErr  error
	}{
		{
			name:    "Happy path",
			inputID: "valid-image-id",
			mockResp: client.ImageRemoveResult{
				Items: []image.DeleteResponse{
					{Deleted: "tag-1", Untagged: "tag-1"},
					{Deleted: "tag-2", Untagged: "tag-2"},
				},
			},
			mockErr: nil,
			expectedResp: types.ImageRemoveResult{
				ImagesRemoved: []types.ImageRemoveSummary{
					{Deleted: "tag-1", Untagged: "tag-1"},
					{Deleted: "tag-2", Untagged: "tag-2"},
				},
			},
			expectedErr: nil,
		},
		{
			name:         "Error path - Image not found",
			inputID:      "invalid-image-id",
			mockResp:     client.ImageRemoveResult{},
			mockErr:      mockErrNotFound{"Error response from daemon: No such image: invalid-image-id"},
			expectedResp: types.ImageRemoveResult{},
			expectedErr:  types.ErrImageNotFound,
		},
		{
			name:         "Error path - Daemon unreachable",
			inputID:      "valid-image-id",
			mockResp:     client.ImageRemoveResult{},
			mockErr:      generateRealConnectionError(),
			expectedResp: types.ImageRemoveResult{},
			expectedErr:  types.ErrConnectionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				imageRemoveFunc: func(ctx context.Context, image string, options client.ImageRemoveOptions) (client.ImageRemoveResult, error) {
					if image != tc.inputID {
						t.Errorf("expected Docker to receive image %q, got %q", tc.inputID, image)
					}

					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ImageRemove(context.Background(), tc.inputID, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(resp.ImagesRemoved, tc.expectedResp.ImagesRemoved) {
				t.Errorf("expected config %+v, got %+v", tc.expectedResp.ImagesRemoved, resp.ImagesRemoved)
			}
		})
	}
}

func TestImagePrune(t *testing.T) {
	tests := []struct {
		name         string
		mockResp     client.ImagePruneResult
		mockErr      error
		expectedResp types.ImagePruneResult
		expectedErr  error
	}{
		{
			name: "Happy path",
			mockResp: client.ImagePruneResult{
				Report: image.PruneReport{
					ImagesDeleted: []image.DeleteResponse{
						{Deleted: "tag-1", Untagged: "tag-1"},
					},
				},
			},
			mockErr: nil,
			expectedResp: types.ImagePruneResult{
				ImagesRemoved: []types.ImageRemoveSummary{
					{Deleted: "tag-1", Untagged: "tag-1"},
				},
			},
			expectedErr: nil,
		},
		{
			name:         "Error path - Daemon unreachable",
			mockResp:     client.ImagePruneResult{},
			mockErr:      generateRealConnectionError(),
			expectedResp: types.ImagePruneResult{},
			expectedErr:  types.ErrConnectionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				imagePruneFunc: func(ctx context.Context, options client.ImagePruneOptions) (client.ImagePruneResult, error) {
					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ImagePrune(context.Background(), nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(resp.ImagesRemoved, tc.expectedResp.ImagesRemoved) {
				t.Errorf("expected config %+v, got %+v", tc.expectedResp.ImagesRemoved, resp.ImagesRemoved)
			}
		})
	}
}
