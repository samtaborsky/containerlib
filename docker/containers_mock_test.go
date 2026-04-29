// SPDX-License-Identifier: MIT

package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"

	cont "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- API functions
// ---------------------------------------------------------------------------------------------------------------------

func (m *mockDockerAPI) ContainerCreate(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
	if m.containerCreateFunc == nil {
		panic("ContainerCreate function not mocked")
	}
	return m.containerCreateFunc(ctx, options)
}

func (m *mockDockerAPI) ContainerStart(ctx context.Context, container string, options client.ContainerStartOptions) (client.ContainerStartResult, error) {
	if m.containerStartFunc == nil {
		panic("ContainerStart function not mocked")
	}
	return m.containerStartFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerStop(ctx context.Context, container string, options client.ContainerStopOptions) (client.ContainerStopResult, error) {
	if m.containerStopFunc == nil {
		panic("ContainerStop function not mocked")
	}
	return m.containerStopFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerPause(ctx context.Context, container string, options client.ContainerPauseOptions) (client.ContainerPauseResult, error) {
	if m.containerPauseFunc == nil {
		panic("ContainerPause function not mocked")
	}
	return m.containerPauseFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerUnpause(ctx context.Context, container string, options client.ContainerUnpauseOptions) (client.ContainerUnpauseResult, error) {
	if m.containerUnpauseFunc == nil {
		panic("ContainerUnpause function not mocked")
	}
	return m.containerUnpauseFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerRestart(ctx context.Context, container string, options client.ContainerRestartOptions) (client.ContainerRestartResult, error) {
	if m.containerRestartFunc == nil {
		panic("ContainerRestart function not mocked")
	}
	return m.containerRestartFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerRemove(ctx context.Context, container string, options client.ContainerRemoveOptions) (client.ContainerRemoveResult, error) {
	if m.containerRemoveFunc == nil {
		panic("ContainerRemove function not mocked")
	}
	return m.containerRemoveFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerInspect(ctx context.Context, container string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	if m.containerInspectFunc == nil {
		panic("ContainerInspect function not mocked")
	}
	return m.containerInspectFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
	if m.containerListFunc == nil {
		panic("ContainerList function not mocked")
	}
	return m.containerListFunc(ctx, options)
}

func (m *mockDockerAPI) ContainerWait(ctx context.Context, container string, options client.ContainerWaitOptions) client.ContainerWaitResult {
	if m.containerWaitFunc == nil {
		panic("ContainerWait function not mocked")
	}
	return m.containerWaitFunc(ctx, container, options)
}

func (m *mockDockerAPI) ContainerLogs(ctx context.Context, container string, options client.ContainerLogsOptions) (client.ContainerLogsResult, error) {
	if m.containerLogsFunc == nil {
		panic("ContainerLogs function not mocked")
	}
	return m.containerLogsFunc(ctx, container, options)
}

func (m *mockDockerAPI) ExecCreate(ctx context.Context, container string, options client.ExecCreateOptions) (client.ExecCreateResult, error) {
	if m.execCreateFunc == nil {
		panic("ExecCreate function not mocked")
	}
	return m.execCreateFunc(ctx, container, options)
}

func (m *mockDockerAPI) ExecAttach(ctx context.Context, execID string, options client.ExecAttachOptions) (client.ExecAttachResult, error) {
	if m.execAttachFunc == nil {
		panic("ExecAttach function not mocked")
	}
	return m.execAttachFunc(ctx, execID, options)
}

func (m *mockDockerAPI) ExecInspect(ctx context.Context, execID string, options client.ExecInspectOptions) (client.ExecInspectResult, error) {
	if m.execInspectFunc == nil {
		panic("ExecInspect function not mocked")
	}
	return m.execInspectFunc(ctx, execID, options)
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Helper functions
// ---------------------------------------------------------------------------------------------------------------------

// makeLogStream generates an 8 byte multiplex header used by the Docker API.
// The header is used by stdcopy.StdCopy to separate Stdout (streamType 1) and Stderr (streamType 2)
// when TTY is disabled.
func makeLogStream(streamType byte, payload string) []byte {
	header := make([]byte, 8)
	header[0] = streamType

	binary.BigEndian.PutUint32(header[4:8], uint32(len(payload)))
	return append(header, []byte(payload)...)
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestContainerCreate(t *testing.T) {
	tests := []struct {
		name         string
		inputCfg     types.ContainerCreateConfig
		mockResp     client.ContainerCreateResult
		mockErr      error
		expectedResp types.ContainerCreateResult
		expectedErr  error
	}{
		{
			name: "Happy path",
			inputCfg: types.ContainerCreateConfig{
				Image: "nginx:latest",
			},
			mockResp: client.ContainerCreateResult{
				ID: "new-container-id",
			},
			mockErr:      nil,
			expectedResp: types.ContainerCreateResult{ID: "new-container-id"},
			expectedErr:  nil,
		},
		{
			name: "Error path - Invalid image",
			inputCfg: types.ContainerCreateConfig{
				Image: "",
			},
			mockResp:     client.ContainerCreateResult{},
			mockErr:      nil,
			expectedResp: types.ContainerCreateResult{},
			expectedErr:  types.ErrInvalidInput,
		},
		{
			name: "Error path - Invalid port",
			inputCfg: types.ContainerCreateConfig{
				Image: "nginx:latest",
				Ports: []types.PortBinding{{ContainerPort: -1}},
			},
			mockResp:     client.ContainerCreateResult{},
			mockErr:      nil,
			expectedResp: types.ContainerCreateResult{},
			expectedErr:  types.ErrInvalidInput,
		},
		{
			name: "Error path - Image not found",
			inputCfg: types.ContainerCreateConfig{
				Image: "missing-image:latest",
			},
			mockResp:     client.ContainerCreateResult{},
			mockErr:      mockErrNotFound{"Error response from daemon: No such image: missing-image:latest"},
			expectedResp: types.ContainerCreateResult{},
			expectedErr:  types.ErrImageNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerCreateFunc: func(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
					if options.Config == nil || options.Config.Image != tc.inputCfg.Image {
						t.Errorf("expected Docker to receive image %q, got %q", tc.inputCfg.Image, options.Config.Image)
					}
					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ContainerCreate(context.Background(), &tc.inputCfg)

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

func TestContainerStart(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "Happy path",
			inputID:     "valid-container-id",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Error path - Container not found",
			inputID:     "missing-container-id",
			mockErr:     mockErrNotFound{"Error response from daemon: No such container: missing-container-id"},
			expectedErr: types.ErrContainerNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerStartFunc: func(ctx context.Context, container string, options client.ContainerStartOptions) (client.ContainerStartResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					return client.ContainerStartResult{}, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			err := rt.ContainerStart(context.Background(), tc.inputID, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestContainerStop(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "Happy path",
			inputID:     "valid-container-id",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Error path - Container not found",
			inputID:     "missing-container-id",
			mockErr:     mockErrNotFound{"Error response from daemon: No such container: missing-container-id"},
			expectedErr: types.ErrContainerNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerStopFunc: func(ctx context.Context, container string, options client.ContainerStopOptions) (client.ContainerStopResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					return client.ContainerStopResult{}, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			err := rt.ContainerStop(context.Background(), tc.inputID, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestContainerPause(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "Happy path",
			inputID:     "valid-container-id",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Error path - Container not found",
			inputID:     "missing-container-id",
			mockErr:     mockErrNotFound{"Error response from daemon: No such container: missing-container-id"},
			expectedErr: types.ErrContainerNotFound,
		},
		{
			name:        "Error path - Container not running",
			inputID:     "not-running-id",
			mockErr:     mockErrConflict{"Error response from daemon: container not-running-id is not running"},
			expectedErr: types.ErrConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerPauseFunc: func(ctx context.Context, container string, options client.ContainerPauseOptions) (client.ContainerPauseResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					return client.ContainerPauseResult{}, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			err := rt.ContainerPause(context.Background(), tc.inputID, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestContainerUnpause(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "Happy path",
			inputID:     "valid-container-id",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Error path - Container not found",
			inputID:     "missing-container-id",
			mockErr:     mockErrNotFound{"Error response from daemon: No such container: missing-container-id"},
			expectedErr: types.ErrContainerNotFound,
		},
		{
			name:        "Error path - Container not paused",
			inputID:     "not-paused-id",
			mockErr:     mockErrConflict{"Error response from daemon: container not-running-id is not paused"},
			expectedErr: types.ErrConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerUnpauseFunc: func(ctx context.Context, container string, options client.ContainerUnpauseOptions) (client.ContainerUnpauseResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					return client.ContainerUnpauseResult{}, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			err := rt.ContainerUnpause(context.Background(), tc.inputID, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestContainerRestart(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "Happy path",
			inputID:     "valid-container-id",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Error path - Container not found",
			inputID:     "missing-container-id",
			mockErr:     mockErrNotFound{"Error response from daemon: No such container: missing-container-id"},
			expectedErr: types.ErrContainerNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerRestartFunc: func(ctx context.Context, container string, options client.ContainerRestartOptions) (client.ContainerRestartResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					return client.ContainerRestartResult{}, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			err := rt.ContainerRestart(context.Background(), tc.inputID, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestContainerRemove(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "Happy path",
			inputID:     "valid-container-id",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Error path - Container not found",
			inputID:     "missing-container-id",
			mockErr:     mockErrNotFound{"Error response from daemon: No such container: missing-container-id"},
			expectedErr: types.ErrContainerNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerRemoveFunc: func(ctx context.Context, container string, options client.ContainerRemoveOptions) (client.ContainerRemoveResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					return client.ContainerRemoveResult{}, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			err := rt.ContainerRemove(context.Background(), tc.inputID, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestContainerStatus(t *testing.T) {
	tests := []struct {
		name         string
		inputID      string
		mockResp     client.ContainerInspectResult
		mockErr      error
		expectedResp types.ContainerStatusResult
		expectedErr  error
	}{
		{
			name:    "Happy path",
			inputID: "valid-container-id",
			mockResp: client.ContainerInspectResult{
				Container: cont.InspectResponse{
					ID:    "valid-container-id",
					State: &cont.State{},
				},
			},
			mockErr:      nil,
			expectedResp: types.ContainerStatusResult{ID: "valid-container-id"},
			expectedErr:  nil,
		},
		{
			name:         "Error path - Container not found",
			inputID:      "missing-container-id",
			mockResp:     client.ContainerInspectResult{},
			mockErr:      mockErrNotFound{"Error response from daemon: No such container: missing-container-id"},
			expectedResp: types.ContainerStatusResult{},
			expectedErr:  types.ErrContainerNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerInspectFunc: func(ctx context.Context, container string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ContainerStatus(context.Background(), tc.inputID, nil)

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

func TestContainerList(t *testing.T) {
	tests := []struct {
		name         string
		mockResp     client.ContainerListResult
		mockErr      error
		expectedResp types.ContainerListResult
		expectedErr  error
	}{
		{
			name: "Happy path",
			mockResp: client.ContainerListResult{
				Items: []cont.Summary{{ID: "container-1"}, {ID: "container-2"}},
			},
			mockErr: nil,
			expectedResp: types.ContainerListResult{
				Containers: []types.ContainerSummary{{ID: "container-1"}, {ID: "container-2"}},
			},
			expectedErr: nil,
		},
		{
			name:         "Error path - Daemon unreachable",
			mockResp:     client.ContainerListResult{},
			mockErr:      generateRealConnectionError(),
			expectedResp: types.ContainerListResult{},
			expectedErr:  types.ErrConnectionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerListFunc: func(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ContainerList(context.Background(), nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(resp.Containers) != len(tc.expectedResp.Containers) {
				t.Fatalf("expected %d containers, got %d", len(tc.expectedResp.Containers), len(resp.Containers))
			}

			if !reflect.DeepEqual(resp.Containers, tc.expectedResp.Containers) {
				t.Errorf("expected config %+v, got %+v", tc.expectedResp.Containers, resp.Containers)
			}
		})
	}
}

func TestContainerWait(t *testing.T) {
	tests := []struct {
		name         string
		inputID      string
		waitCond     string
		cancelCtx    bool
		mockWaitResp cont.WaitResponse
		mockWaitErr  error
		expectedResp types.ContainerWaitResult
		expectedErr  error
	}{
		{
			name:         "Happy path",
			inputID:      "valid-container-id",
			waitCond:     "not-running",
			mockWaitResp: cont.WaitResponse{StatusCode: 0},
			mockWaitErr:  nil,
			expectedResp: types.ContainerWaitResult{ExitCode: 0},
			expectedErr:  nil,
		},
		{
			name:     "Error path - Error while waiting for container",
			inputID:  "valid-container-id",
			waitCond: "not-running",
			mockWaitResp: cont.WaitResponse{
				StatusCode: 0,
				Error:      &cont.WaitExitError{Message: "Wait exit error occurred"},
			},
			mockWaitErr:  nil,
			expectedResp: types.ContainerWaitResult{ExitCode: 0},
			expectedErr:  types.ErrInternal,
		},
		{
			name:         "Error path - Context timeout",
			inputID:      "valid-container-id",
			waitCond:     "not-running",
			cancelCtx:    true,
			mockWaitResp: cont.WaitResponse{},
			mockWaitErr:  nil,
			expectedResp: types.ContainerWaitResult{},
			expectedErr:  types.ErrInternal,
		},
		{
			name:         "Error path - Container not found",
			inputID:      "missing-container-id",
			waitCond:     "not-running",
			mockWaitResp: cont.WaitResponse{},
			mockWaitErr:  mockErrNotFound{"Error response from daemon: No such container: missing-container-id"},
			expectedResp: types.ContainerWaitResult{},
			expectedErr:  types.ErrContainerNotFound,
		},
		{
			name:         "Error path - Invalid wait condition",
			inputID:      "valid-container-id",
			waitCond:     "invalid",
			mockWaitResp: cont.WaitResponse{},
			mockWaitErr:  nil,
			expectedResp: types.ContainerWaitResult{},
			expectedErr:  types.ErrInvalidInput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerWaitFunc: func(ctx context.Context, container string, options client.ContainerWaitOptions) client.ContainerWaitResult {
					respCh := make(chan cont.WaitResponse, 1)
					errCh := make(chan error, 1)

					if !tc.cancelCtx {
						if tc.mockWaitErr != nil {
							errCh <- tc.mockWaitErr
						} else {
							respCh <- tc.mockWaitResp
						}
					}

					return client.ContainerWaitResult{
						Result: respCh,
						Error:  errCh,
					}
				},
			}

			opts := types.ContainerWaitOptions{
				Condition: types.WaitCondition(tc.waitCond),
			}

			ctx, cancel := context.WithCancel(context.Background())

			if tc.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ContainerWait(ctx, tc.inputID, &opts)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if resp.ExitCode != tc.expectedResp.ExitCode {
				t.Errorf("expected exit code to be %d, got %d", tc.expectedResp.ExitCode, resp.ExitCode)
			}
		})
	}
}

func TestContainerExec(t *testing.T) {
	tests := []struct {
		name          string
		inputID       string
		inputCmd      []string
		mockOutput    string
		noInput       bool
		mockExitCode  int
		mockCreateErr error
		expectedResp  types.ContainerExecResult
		expectedErr   error
	}{
		{
			name:          "Happy path",
			inputID:       "running-container-id",
			inputCmd:      []string{"echo", "hello"},
			mockOutput:    "hello\n",
			mockExitCode:  0,
			mockCreateErr: nil,
			expectedResp:  types.ContainerExecResult{ExitCode: 0},
			expectedErr:   nil,
		},
		{
			name:          "Error path - Input empty",
			inputID:       "running-container-id",
			inputCmd:      nil,
			mockOutput:    "",
			noInput:       true,
			mockExitCode:  0,
			mockCreateErr: nil,
			expectedResp:  types.ContainerExecResult{},
			expectedErr:   types.ErrInvalidInput,
		},
		{
			name:          "Error path - Container not running",
			inputID:       "stopped-container-id",
			inputCmd:      []string{"echo", "hello"},
			mockOutput:    "",
			mockExitCode:  0,
			mockCreateErr: mockErrConflict{"Error response from daemon: container stopped-container-id is not running"},
			expectedResp:  types.ContainerExecResult{},
			expectedErr:   types.ErrConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				// ExecCreate
				execCreateFunc: func(ctx context.Context, container string, options client.ExecCreateOptions) (client.ExecCreateResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					return client.ExecCreateResult{ID: "mock-exec-id"}, tc.mockCreateErr
				},

				// ExecAttach
				execAttachFunc: func(ctx context.Context, execID string, options client.ExecAttachOptions) (client.ExecAttachResult, error) {
					fakeStream := bufio.NewReader(strings.NewReader(tc.mockOutput))
					_, conn := net.Pipe()
					response := client.HijackedResponse{
						Reader: fakeStream,
						Conn:   conn,
					}
					return client.ExecAttachResult{HijackedResponse: response}, nil
				},

				// ExecInspect
				execInspectFunc: func(ctx context.Context, execID string, options client.ExecInspectOptions) (client.ExecInspectResult, error) {
					return client.ExecInspectResult{ExitCode: tc.mockExitCode}, nil
				},
			}

			var opts *types.ContainerExecOptions
			var outBuf bytes.Buffer
			if !tc.noInput {
				opts = &types.ContainerExecOptions{
					Cmd:    tc.inputCmd,
					Stdout: &outBuf,
					TTY:    true,
				}
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.ContainerExec(context.Background(), tc.inputID, opts)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if resp.ExitCode != tc.expectedResp.ExitCode {
				t.Errorf("expected exit code to be %d, got %d", tc.expectedResp.ExitCode, resp.ExitCode)
			}
			if outBuf.String() != tc.mockOutput {
				t.Errorf("expected output %q, got %q", tc.mockOutput, outBuf.String())
			}
		})
	}
}

func TestContainerLogs(t *testing.T) {
	tests := []struct {
		name        string
		inputID     string
		mockStdout  string
		mockStderr  string
		noInput     bool
		mockErr     error
		expectedErr error
	}{
		{
			name:        "Happy path",
			inputID:     "running-container-id",
			mockStdout:  "this is a message in Stdout\n",
			mockStderr:  "this is a message in Stderr\n",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Error path - Input empty",
			inputID:     "stopped-container-id",
			mockStdout:  "",
			mockStderr:  "",
			noInput:     true,
			mockErr:     nil,
			expectedErr: types.ErrInvalidInput,
		},
		{
			name:        "Error path - Container not running",
			inputID:     "stopped-container-id",
			mockStdout:  "",
			mockStderr:  "",
			mockErr:     mockErrConflict{"Error response from daemon: container stopped-container-id is not running"},
			expectedErr: types.ErrConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				containerLogsFunc: func(ctx context.Context, container string, options client.ContainerLogsOptions) (client.ContainerLogsResult, error) {
					if container != tc.inputID {
						t.Errorf("expected Docker to receive container ID %q, got %q", tc.inputID, container)
					}
					if tc.mockErr != nil {
						return nil, tc.mockErr
					}

					var streamData []byte
					if tc.mockStdout != "" {
						streamData = append(streamData, makeLogStream(1, tc.mockStdout)...)
					}
					if tc.mockStderr != "" {
						streamData = append(streamData, makeLogStream(2, tc.mockStderr)...)
					}

					fakeStream := io.NopCloser(bytes.NewReader(streamData))
					return fakeStream, nil
				},
			}

			var opts *types.ContainerLogsOptions
			var outStdBuf, outErrBuf bytes.Buffer
			if !tc.noInput {
				opts = &types.ContainerLogsOptions{
					Stdout: &outStdBuf,
					Stderr: &outErrBuf,
				}
			}

			rt := &runtime{api: mockAPI}
			_, err := rt.ContainerLogs(context.Background(), tc.inputID, opts)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if outStdBuf.String() != tc.mockStdout {
				t.Errorf("expected output %q, got %q", tc.mockStdout, outStdBuf.String())
			}
			if outErrBuf.String() != tc.mockStderr {
				t.Errorf("expected output %q, got %q", tc.mockStderr, outErrBuf.String())
			}
		})
	}
}
