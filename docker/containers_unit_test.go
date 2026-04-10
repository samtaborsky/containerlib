// SPDX-License-Identifier: MIT

package docker

import (
	"io"
	"net/netip"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	cont "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestToMobyContainerConfig(t *testing.T) {
	tests := []struct {
		name         string
		inputCfg     *types.ContainerCreateConfig
		expectedOpts client.ContainerCreateOptions
		expectedErr  string
	}{
		{
			name: "Happy path",
			inputCfg: &types.ContainerCreateConfig{
				Name:  "test-container",
				Image: "alpine:latest",
				User:  "root",
				Cmd:   []string{"echo", "hello"},
			},
			expectedOpts: client.ContainerCreateOptions{
				Name: "test-container",
				Config: &cont.Config{
					Image: "alpine:latest",
					User:  "root",
					Cmd:   []string{"echo", "hello"},
				},
				HostConfig: &cont.HostConfig{
					RestartPolicy: cont.RestartPolicy{
						Name: "",
					},
				},
			},
			expectedErr: "",
		},
		{
			name:         "Error path - Nil config",
			inputCfg:     nil,
			expectedOpts: client.ContainerCreateOptions{},
			expectedErr:  "config cannot be nil",
		},
		{
			name:         "Error path - Missing image",
			inputCfg:     &types.ContainerCreateConfig{Name: "my-container"},
			expectedOpts: client.ContainerCreateOptions{},
			expectedErr:  "image name is required",
		},
		{
			name: "Error path - Invalid restart policy",
			inputCfg: &types.ContainerCreateConfig{
				Image:   "nginx:latest",
				Restart: "invalid-policy",
			},
			expectedOpts: client.ContainerCreateOptions{},
			expectedErr:  "invalid restart policy",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := toMobyContainerConfig(tc.inputCfg)

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

			if res.Name != tc.expectedOpts.Name {
				t.Errorf("expected name %q, got %q", tc.expectedOpts.Name, res.Name)
			}
			if !reflect.DeepEqual(res.Config, tc.expectedOpts.Config) {
				t.Errorf("expected config %+v, got %+v", tc.expectedOpts.Config, res.Config)
			}
		})
	}
}

func TestToMobyPortMap(t *testing.T) {
	port, _ := network.ParsePort("5678/tcp")

	tests := []struct {
		name        string
		input       []types.PortBinding
		expectedRes network.PortMap
		expectedErr string
	}{
		{
			name: "Happy path",
			input: []types.PortBinding{
				{HostPort: 1234, ContainerPort: 5678, Protocol: "tcp"},
			},
			expectedRes: map[network.Port][]network.PortBinding{
				port: {{HostIP: netip.Addr{}, HostPort: "1234"}},
			},
			expectedErr: "",
		},
		{
			name: "Error path - Invalid port",
			input: []types.PortBinding{
				{HostPort: 1234, ContainerPort: -1, Protocol: "tcp"},
			},
			expectedRes: map[network.Port][]network.PortBinding{},
			expectedErr: "invalid container port",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := toMobyPortMap(tc.input)

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

			if !reflect.DeepEqual(res, tc.expectedRes) {
				t.Errorf("expected portmap %+v, got %+v", tc.expectedRes, res)
			}
		})
	}
}

func TestToMobyMounts(t *testing.T) {
	tests := []struct {
		name        string
		input       []types.Mount
		expectedRes []mount.Mount
	}{
		{
			name: "Happy path",
			input: []types.Mount{
				{Source: "/var/run", Destination: "/var/run", Type: "bind"},
				{Source: "/home/test", Destination: "/www", Type: "bind"},
			},
			expectedRes: []mount.Mount{
				{Source: "/var/run", Target: "/var/run", Type: "bind"},
				{Source: "/home/test", Target: "/www", Type: "bind"},
			},
		},
		{
			name:        "Happy path - Empty inputCfg",
			input:       nil,
			expectedRes: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyMounts(tc.input)

			if !reflect.DeepEqual(res, tc.expectedRes) {
				t.Errorf("expected %+v, got %+v", tc.expectedRes, res)
			}
		})
	}
}

func TestToMobyContainerStopOpts(t *testing.T) {
	tests := []struct {
		name        string
		inputOpts   *types.ContainerStopOptions
		expectedRes client.ContainerStopOptions
	}{
		{
			name: "Happy path",
			inputOpts: &types.ContainerStopOptions{
				Signal: "SIGKILL",
			},
			expectedRes: client.ContainerStopOptions{
				Signal: "SIGKILL",
			},
		},
		{
			name:        "Happy path - Empty inputCfg",
			inputOpts:   nil,
			expectedRes: client.ContainerStopOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyContainerStopOpts(tc.inputOpts)

			if !reflect.DeepEqual(res, tc.expectedRes) {
				t.Errorf("expected %+v, got %+v", tc.expectedRes, res)
			}
		})
	}
}

func TestToMobyContainerRestartOpts(t *testing.T) {
	tests := []struct {
		name        string
		input       *types.ContainerRestartOptions
		expectedRes client.ContainerRestartOptions
	}{
		{
			name: "Happy path",
			input: &types.ContainerRestartOptions{
				Signal: "SIGKILL",
			},
			expectedRes: client.ContainerRestartOptions{
				Signal: "SIGKILL",
			},
		},
		{
			name:        "Happy path - Empty inputCfg",
			input:       nil,
			expectedRes: client.ContainerRestartOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyContainerRestartOpts(tc.input)

			if !reflect.DeepEqual(res, tc.expectedRes) {
				t.Errorf("expected %+v, got %+v", tc.expectedRes, res)
			}
		})
	}
}

func TestToMobyContainerRemoveOpts(t *testing.T) {
	tests := []struct {
		name        string
		input       *types.ContainerRemoveOptions
		expectedRes client.ContainerRemoveOptions
	}{
		{
			name: "Happy path",
			input: &types.ContainerRemoveOptions{
				Force:         true,
				RemoveVolumes: true,
			},
			expectedRes: client.ContainerRemoveOptions{
				Force:         true,
				RemoveVolumes: true,
			},
		},
		{
			name:        "Happy path - Empty inputCfg",
			input:       nil,
			expectedRes: client.ContainerRemoveOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyContainerRemoveOpts(tc.input)

			if !reflect.DeepEqual(res, tc.expectedRes) {
				t.Errorf("expected %+v, got %+v", tc.expectedRes, res)
			}
		})
	}
}

func TestToMobyContainerInspectOpts(t *testing.T) {
	tests := []struct {
		name         string
		inputOpts    *types.ContainerStatusOptions
		expectedOpts client.ContainerInspectOptions
	}{
		{
			name: "Happy path",
			inputOpts: &types.ContainerStatusOptions{
				Size: true,
			},
			expectedOpts: client.ContainerInspectOptions{
				Size: true,
			},
		},
		{
			name:         "Happy path - Empty inputCfg",
			inputOpts:    nil,
			expectedOpts: client.ContainerInspectOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyContainerInspectOpts(tc.inputOpts)

			if !reflect.DeepEqual(res, tc.expectedOpts) {
				t.Errorf("expected %+v, got %+v", tc.expectedOpts, res)
			}
		})
	}
}

func TestToMobyContainerListOpts(t *testing.T) {
	tests := []struct {
		name         string
		inputOpts    *types.ContainerListOptions
		expectedOpts client.ContainerListOptions
	}{
		{
			name: "Happy path",
			inputOpts: &types.ContainerListOptions{
				All:     true,
				Filters: make(types.Filters).Add("test", "true"),
			},
			expectedOpts: client.ContainerListOptions{
				All:     true,
				Filters: make(client.Filters).Add("test", "true"),
			},
		},
		{
			name:         "Happy path - Empty inputCfg",
			inputOpts:    nil,
			expectedOpts: client.ContainerListOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyContainerListOpts(tc.inputOpts)

			if !reflect.DeepEqual(res, tc.expectedOpts) {
				t.Errorf("expected %+v, got %+v", tc.expectedOpts, res)
			}
		})
	}
}

func TestToMobyContainerWaitOpts(t *testing.T) {
	tests := []struct {
		name         string
		inputOpts    *types.ContainerWaitOptions
		expectedOpts client.ContainerWaitOptions
		expectedErr  string
	}{
		{
			name: "Happy path",
			inputOpts: &types.ContainerWaitOptions{
				Condition: types.WaitConditionRemoved,
			},
			expectedOpts: client.ContainerWaitOptions{
				Condition: cont.WaitConditionRemoved,
			},
			expectedErr: "",
		},
		{
			name:         "Error path - Empty inputCfg",
			inputOpts:    nil,
			expectedOpts: client.ContainerWaitOptions{},
			expectedErr:  "options cannot be nil",
		},
		{
			name: "Error path - Invalid wait condition",
			inputOpts: &types.ContainerWaitOptions{
				Condition: "invalid",
			},
			expectedOpts: client.ContainerWaitOptions{},
			expectedErr:  "invalid condition: invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := toMobyContainerWaitOpts(tc.inputOpts)

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

			if !reflect.DeepEqual(res, tc.expectedOpts) {
				t.Errorf("expected %+v, got %+v", tc.expectedOpts, res)
			}
		})
	}
}

func TestToMobyExecCreateOpts(t *testing.T) {
	tests := []struct {
		name         string
		inputOpts    *types.ContainerExecOptions
		expectedOpts client.ExecCreateOptions
		expectedErr  string
	}{
		{
			name: "Happy path",
			inputOpts: &types.ContainerExecOptions{
				User:   "root",
				Cmd:    []string{"echo", "hello"},
				Stdout: io.Discard,
			},
			expectedOpts: client.ExecCreateOptions{
				User:         "root",
				Cmd:          []string{"echo", "hello"},
				AttachStdout: true,
			},
			expectedErr: "",
		},
		{
			name:         "Error path - Empty inputCfg",
			inputOpts:    nil,
			expectedOpts: client.ExecCreateOptions{},
			expectedErr:  "options cannot be nil",
		},
		{
			name: "Error path - Empty command",
			inputOpts: &types.ContainerExecOptions{
				Cmd: []string{},
			},
			expectedOpts: client.ExecCreateOptions{},
			expectedErr:  "command cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := toMobyExecCreateOpts(tc.inputOpts)

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

			if !reflect.DeepEqual(res, tc.expectedOpts) {
				t.Errorf("expected %+v, got %+v", tc.expectedOpts, res)
			}
		})
	}
}

func TestToMobyContainerLogsOpts(t *testing.T) {
	curTime := time.Now()

	tests := []struct {
		name         string
		inputOpts    *types.ContainerLogsOptions
		expectedOpts client.ContainerLogsOptions
		expectedErr  string
	}{
		{
			name: "Happy path - One stream",
			inputOpts: &types.ContainerLogsOptions{
				Stdout:  io.Discard,
				Stderr:  nil,
				Details: true,
				Since:   curTime,
			},
			expectedOpts: client.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: false,
				Details:    true,
				Since:      curTime.Format(time.RFC3339Nano),
			},
			expectedErr: "",
		},
		{
			name: "Happy path - Both streams",
			inputOpts: &types.ContainerLogsOptions{
				Stdout: io.Discard,
				Stderr: io.Discard,
			},
			expectedOpts: client.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
			},
			expectedErr: "",
		},
		{
			name:         "Error path - Empty inputCfg",
			inputOpts:    nil,
			expectedOpts: client.ContainerLogsOptions{},
			expectedErr:  "options cannot be nil",
		},
		{
			name: "Error path - No streams",
			inputOpts: &types.ContainerLogsOptions{
				Details:    true,
				Timestamps: true,
			},
			expectedOpts: client.ContainerLogsOptions{},
			expectedErr:  "one of Stdout, Stderr must not be nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := toMobyContainerLogsOpts(tc.inputOpts)

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

			if !reflect.DeepEqual(res, tc.expectedOpts) {
				t.Errorf("expected %+v, got %+v", tc.expectedOpts, res)
			}
		})
	}
}

func TestFromMobyInspectResponse(t *testing.T) {
	addr := netip.MustParseAddr("127.0.0.1")
	net := &cont.NetworkSettings{
		Networks: map[string]*network.EndpointSettings{
			"bridge": {IPAddress: addr},
		},
	}

	tests := []struct {
		name        string
		inputResp   cont.InspectResponse
		expectedRes types.ContainerStatusResult
	}{
		{
			name: "Happy path - Not exited",
			inputResp: cont.InspectResponse{
				ID: "container-id",
				State: &cont.State{
					Status:   cont.StateCreated,
					ExitCode: 0,
				},
				NetworkSettings: net,
			},
			expectedRes: types.ContainerStatusResult{
				ID:        "container-id",
				Status:    "created",
				ExitCode:  0,
				IPAddress: addr,
			},
		},
		{
			name: "Happy path - Exited",
			inputResp: cont.InspectResponse{
				ID: "container-id",
				State: &cont.State{
					Status:   cont.StateExited,
					ExitCode: 0,
				},
				NetworkSettings: net,
			},
			expectedRes: types.ContainerStatusResult{
				ID:        "container-id",
				Status:    "exited",
				ExitCode:  0,
				IPAddress: netip.Addr{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := fromMobyInspectResponse(tc.inputResp)

			if !reflect.DeepEqual(res, tc.expectedRes) {
				t.Errorf("expected %+v, got %+v", tc.expectedRes, res)
			}
		})
	}
}

func TestFromMobyContainerList(t *testing.T) {
	curTime := time.Now().Truncate(time.Second)

	tests := []struct {
		name        string
		inputRes    client.ContainerListResult
		expectedRes types.ContainerListResult
	}{
		{
			name: "Happy path",
			inputRes: client.ContainerListResult{
				Items: []cont.Summary{
					{ID: "container-1", Image: "alpine:latest", Created: curTime.Unix()},
					{ID: "container-2", Image: "alpine:latest"},
				},
			},
			expectedRes: types.ContainerListResult{
				Containers: []types.ContainerSummary{
					{ID: "container-1", Image: "alpine:latest", Created: curTime},
					{ID: "container-2", Image: "alpine:latest"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := fromMobyContainerList(tc.inputRes)

			for i, container := range res.Containers {
				if !reflect.DeepEqual(container, tc.expectedRes.Containers[i]) {
					t.Errorf("expected %+v, got %+v", tc.expectedRes.Containers[i], container)
				}
			}
		})
	}
}

func TestGetIPAddressFromNetworkSettings(t *testing.T) {
	addr1 := netip.MustParseAddr("127.0.0.1")
	addr2 := netip.MustParseAddr("127.0.0.2")

	tests := []struct {
		name     string
		input    *cont.NetworkSettings
		expected netip.Addr
	}{
		{
			name: "Happy path - Bridge valid",
			input: &cont.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"bridge": {IPAddress: addr1},
				},
			},
			expected: addr1,
		},
		{
			name: "Happy path - Prioritize bridge",
			input: &cont.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"overlay": {IPAddress: addr2},
					"bridge":  {IPAddress: addr1},
				},
			},
			expected: addr1,
		},
		{
			name: "Happy path - Bridge not valid",
			input: &cont.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"net1":   {IPAddress: addr2},
					"bridge": {IPAddress: netip.Addr{}},
				},
			},
			expected: addr2,
		},
		{
			name: "Happy path - No valid IP address",
			input: &cont.NetworkSettings{
				Networks: map[string]*network.EndpointSettings{
					"net1":   {IPAddress: netip.Addr{}},
					"bridge": {IPAddress: netip.Addr{}},
				},
			},
			expected: netip.Addr{},
		},
		{
			name:     "Happy path - No networks",
			input:    nil,
			expected: netip.Addr{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := getIPAddressFromNetworkSettings(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, res)
			}
		})
	}
}

func TestMapToEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected []string
	}{
		{
			name: "Happy path",
			input: map[string]string{
				"FOO": "BAR",
				"key": "value",
			},
			expected: []string{
				"key=value",
				"FOO=BAR",
			},
		},
		{
			name:     "Happy path - empty env",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := mapToEnv(tc.input)

			sort.Strings(res)
			sort.Strings(tc.expected)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, res)
			}
		})
	}
}
