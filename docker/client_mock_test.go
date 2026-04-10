// SPDX-License-Identifier: MIT

package docker

import (
	"context"

	"github.com/moby/moby/client"
)

var _ client.APIClient = (*mockDockerAPI)(nil)

type mockDockerAPI struct {
	client.APIClient

	// docker/runtime.go
	infoFunc          func(ctx context.Context, options client.InfoOptions) (client.SystemInfoResult, error)
	registryLoginFunc func(ctx context.Context, auth client.RegistryLoginOptions) (client.RegistryLoginResult, error)
	closeFunc         func() error

	// docker/containers.go
	containerCreateFunc  func(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error)
	containerStartFunc   func(ctx context.Context, container string, options client.ContainerStartOptions) (client.ContainerStartResult, error)
	containerStopFunc    func(ctx context.Context, container string, options client.ContainerStopOptions) (client.ContainerStopResult, error)
	containerPauseFunc   func(ctx context.Context, container string, options client.ContainerPauseOptions) (client.ContainerPauseResult, error)
	containerUnpauseFunc func(ctx context.Context, container string, options client.ContainerUnpauseOptions) (client.ContainerUnpauseResult, error)
	containerRestartFunc func(ctx context.Context, container string, options client.ContainerRestartOptions) (client.ContainerRestartResult, error)
	containerRemoveFunc  func(ctx context.Context, container string, options client.ContainerRemoveOptions) (client.ContainerRemoveResult, error)
	containerInspectFunc func(ctx context.Context, container string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error)
	containerListFunc    func(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error)
	containerWaitFunc    func(ctx context.Context, container string, options client.ContainerWaitOptions) client.ContainerWaitResult
	containerLogsFunc    func(ctx context.Context, container string, options client.ContainerLogsOptions) (client.ContainerLogsResult, error)

	execCreateFunc  func(ctx context.Context, container string, options client.ExecCreateOptions) (client.ExecCreateResult, error)
	execAttachFunc  func(ctx context.Context, execID string, options client.ExecAttachOptions) (client.ExecAttachResult, error)
	execInspectFunc func(ctx context.Context, execID string, options client.ExecInspectOptions) (client.ExecInspectResult, error)

	// docker/images.go
	imagePullFunc    func(ctx context.Context, ref string, options client.ImagePullOptions) (client.ImagePullResponse, error)
	imageInspectFunc func(ctx context.Context, image string, _ ...client.ImageInspectOption) (client.ImageInspectResult, error)
	imageListFunc    func(ctx context.Context, options client.ImageListOptions) (client.ImageListResult, error)
	imageRemoveFunc  func(ctx context.Context, image string, options client.ImageRemoveOptions) (client.ImageRemoveResult, error)
	imagePruneFunc   func(ctx context.Context, opts client.ImagePruneOptions) (client.ImagePruneResult, error)

	// docker/events.go
	eventsFunc func(ctx context.Context, options client.EventsListOptions) client.EventsResult

	// docker/telemetry.go
	containerStatsFunc func(ctx context.Context, container string, options client.ContainerStatsOptions) (client.ContainerStatsResult, error)
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Errors
// ---------------------------------------------------------------------------------------------------------------------

type mockErrInternal struct {
	msg string
}

func (e mockErrInternal) Error() string {
	return e.msg
}

func (e mockErrInternal) System() {}

type mockErrNotFound struct {
	msg string
}

func (e mockErrNotFound) Error() string {
	return e.msg
}

func (e mockErrNotFound) NotFound() {}

type mockErrConflict struct {
	msg string
}

func (e mockErrConflict) Error() string {
	return e.msg
}

func (e mockErrConflict) Conflict() {}

// ---------------------------------------------------------------------------------------------------------------------
// --- Helper functions
// ---------------------------------------------------------------------------------------------------------------------

// generateRealConnectionError generates a genuine Moby connection error.
// This is needed as Moby uses a client.IsErrConnectionFailed identification method
// and a private errConnectionFailed type. This type cannot be mocked by duck typing.
func generateRealConnectionError() error {
	api, _ := client.New(client.WithHost("unix:///var/run/fake-docker.sock"))
	_, err := api.Ping(context.Background(), client.PingOptions{})
	return err
}
