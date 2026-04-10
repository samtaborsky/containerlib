// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

func (rt *runtime) ContainerCreate(ctx context.Context, cfg *types.ContainerCreateConfig) (types.ContainerCreateResult, error) {
	mobyOpts, err := toMobyContainerConfig(cfg)
	if err != nil {
		return types.ContainerCreateResult{}, fmt.Errorf("%w: %w", types.ErrInvalidInput, err)
	}
	resp, err := rt.api.ContainerCreate(ctx, mobyOpts)
	if err != nil {
		return types.ContainerCreateResult{}, mapFromMobyError(err, types.ErrImageNotFound)
	}

	return types.ContainerCreateResult{ID: resp.ID}, nil
}

func (rt *runtime) ContainerStart(ctx context.Context, id string, _ *types.ContainerStartOptions) error {
	_, err := rt.api.ContainerStart(ctx, id, client.ContainerStartOptions{})
	return mapFromMobyError(err, types.ErrContainerNotFound)
}

func (rt *runtime) ContainerStop(ctx context.Context, id string, opts *types.ContainerStopOptions) error {
	_, err := rt.api.ContainerStop(ctx, id, toMobyContainerStopOpts(opts))
	return mapFromMobyError(err, types.ErrContainerNotFound)
}

func (rt *runtime) ContainerPause(ctx context.Context, id string, _ *types.ContainerPauseOptions) error {
	_, err := rt.api.ContainerPause(ctx, id, client.ContainerPauseOptions{})
	return mapFromMobyError(err, types.ErrContainerNotFound)
}

func (rt *runtime) ContainerUnpause(ctx context.Context, id string, _ *types.ContainerUnpauseOptions) error {
	_, err := rt.api.ContainerUnpause(ctx, id, client.ContainerUnpauseOptions{})
	return mapFromMobyError(err, types.ErrContainerNotFound)
}

func (rt *runtime) ContainerRestart(ctx context.Context, id string, opts *types.ContainerRestartOptions) error {
	_, err := rt.api.ContainerRestart(ctx, id, toMobyContainerRestartOpts(opts))
	return mapFromMobyError(err, types.ErrContainerNotFound)
}

func (rt *runtime) ContainerRemove(ctx context.Context, id string, opts *types.ContainerRemoveOptions) error {
	_, err := rt.api.ContainerRemove(ctx, id, toMobyContainerRemoveOpts(opts))
	return mapFromMobyError(err, types.ErrContainerNotFound)
}

func (rt *runtime) ContainerStatus(ctx context.Context, id string, opts *types.ContainerStatusOptions) (types.ContainerStatusResult, error) {
	resp, err := rt.api.ContainerInspect(ctx, id, toMobyContainerInspectOpts(opts))
	if err != nil {
		return types.ContainerStatusResult{}, mapFromMobyError(err, types.ErrContainerNotFound)
	}

	return fromMobyInspectResponse(resp.Container), nil
}

func (rt *runtime) ContainerList(ctx context.Context, opts *types.ContainerListOptions) (types.ContainerListResult, error) {
	resp, err := rt.api.ContainerList(ctx, toMobyContainerListOpts(opts))
	if err != nil {
		return types.ContainerListResult{}, mapFromMobyError(err)
	}

	return fromMobyContainerList(resp), nil
}

func (rt *runtime) ContainerWait(ctx context.Context, id string, opts *types.ContainerWaitOptions) (types.ContainerWaitResult, error) {
	mobyOpts, err := toMobyContainerWaitOpts(opts)
	if err != nil {
		return types.ContainerWaitResult{}, fmt.Errorf("%w: %w", types.ErrInvalidInput, err)
	}
	resp := rt.api.ContainerWait(ctx, id, mobyOpts)
	errCh, resCh := resp.Error, resp.Result

	select {
	case err := <-errCh:
		return types.ContainerWaitResult{}, mapFromMobyError(err, types.ErrContainerNotFound)
	case res := <-resCh:
		if res.Error != nil {
			err := mapFromMobyError(errors.New(res.Error.Message))
			return types.ContainerWaitResult{ExitCode: res.StatusCode}, fmt.Errorf("wait failed: %w", err)
		}
		return types.ContainerWaitResult{ExitCode: res.StatusCode}, nil
	case <-ctx.Done():
		return types.ContainerWaitResult{}, mapFromMobyError(ctx.Err())
	}
}

func (rt *runtime) ContainerExec(ctx context.Context, id string, opts *types.ContainerExecOptions) (types.ContainerExecResult, error) {
	execCfg, err := toMobyExecCreateOpts(opts)
	if err != nil {
		return types.ContainerExecResult{}, fmt.Errorf("%w: %w", types.ErrInvalidInput, err)
	}

	execRes, err := rt.api.ExecCreate(ctx, id, execCfg)
	if err != nil {
		return types.ContainerExecResult{}, mapFromMobyError(err)
	}

	resp, err := rt.api.ExecAttach(ctx, execRes.ID, client.ExecAttachOptions{TTY: opts.TTY})
	if err != nil {
		return types.ContainerExecResult{}, mapFromMobyError(err)
	}
	defer resp.Close()

	outWriter := opts.Stdout
	if outWriter == nil {
		outWriter = io.Discard
	}
	errWriter := opts.Stderr
	if errWriter == nil {
		errWriter = io.Discard
	}

	if opts.TTY {
		_, err = io.Copy(outWriter, resp.Reader)
	} else {
		_, err = stdcopy.StdCopy(outWriter, errWriter, resp.Reader)
	}
	if err != nil {
		return types.ContainerExecResult{}, fmt.Errorf("stream execution failed: %w", err)
	}

	inspect, err := rt.api.ExecInspect(ctx, execRes.ID, client.ExecInspectOptions{})
	if err != nil {
		return types.ContainerExecResult{}, mapFromMobyError(err)
	}

	return types.ContainerExecResult{ExitCode: int64(inspect.ExitCode)}, nil
}

func (rt *runtime) ContainerLogs(ctx context.Context, id string, opts *types.ContainerLogsOptions) (types.ContainerLogsResult, error) {
	ret := types.ContainerLogsResult{}
	mobyOpts, err := toMobyContainerLogsOpts(opts)
	if err != nil {
		return ret, fmt.Errorf("%w: %w", types.ErrInvalidInput, err)
	}
	resp, err := rt.api.ContainerLogs(ctx, id, mobyOpts)
	if err != nil {
		return ret, mapFromMobyError(err)
	}

	outWriter := opts.Stdout
	if outWriter == nil {
		outWriter = io.Discard
	}
	errWriter := opts.Stderr
	if errWriter == nil {
		errWriter = io.Discard
	}

	_, err = stdcopy.StdCopy(outWriter, errWriter, resp)
	if err != nil && err != io.EOF {
		return ret, fmt.Errorf("failed to stream logs: %w", err)
	}
	return ret, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Helper functions
// ---------------------------------------------------------------------------------------------------------------------

// toMobyContainerConfig parses and transforms types.ContainerCreateConfig into generic config types required by the Docker SDK.
func toMobyContainerConfig(cfg *types.ContainerCreateConfig) (client.ContainerCreateOptions, error) {
	if cfg == nil {
		return client.ContainerCreateOptions{}, fmt.Errorf("config cannot be nil")
	}
	if cfg.Image == "" {
		return client.ContainerCreateOptions{}, fmt.Errorf("image name is required")
	}

	containerCfg := &container.Config{
		Image:      cfg.Image,
		User:       cfg.User,
		Env:        mapToEnv(cfg.Env),
		Labels:     cfg.Labels,
		Cmd:        cfg.Cmd,
		Entrypoint: cfg.Entrypoint,
	}

	restartPolicy := cfg.Restart
	if !restartPolicy.IsValid() {
		return client.ContainerCreateOptions{}, fmt.Errorf("invalid restart policy '%s'", restartPolicy)
	}
	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyMode(restartPolicy),
		},
		Mounts: toMobyMounts(cfg.Mounts),
		Resources: container.Resources{
			NanoCPUs: toNanoCPUs(cfg.CPUs),
			Memory:   megabytesToBytes(cfg.MemoryMb),
		},
		Privileged: cfg.Privileged,
	}

	portMap, err := toMobyPortMap(cfg.Ports)
	if err != nil {
		return client.ContainerCreateOptions{}, err
	}
	hostCfg.PortBindings = portMap

	return client.ContainerCreateOptions{
		Name:       cfg.Name,
		Config:     containerCfg,
		HostConfig: hostCfg,
	}, nil
}

// toMobyPortMap converts []types.PortBinding to a Docker network.PortMap.
func toMobyPortMap(ports []types.PortBinding) (network.PortMap, error) {
	bindingMap := make(network.PortMap)
	for _, p := range ports {
		dockerPort, err := network.ParsePort(fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))
		if err != nil {
			return nil, fmt.Errorf("invalid container port '%d/%s': %w", p.ContainerPort, p.Protocol, err)
		}

		binding := network.PortBinding{
			HostPort: fmt.Sprintf("%d", p.HostPort),
			HostIP:   p.HostIP,
		}

		bindingMap[dockerPort] = append(bindingMap[dockerPort], binding)
	}
	return bindingMap, nil
}

// toMobyMounts converts []types.Mount to a []mount.Mount.
func toMobyMounts(mounts []types.Mount) []mount.Mount {
	var res []mount.Mount
	for _, m := range mounts {
		mnt := mount.Mount{
			Type:     mount.Type(m.Type),
			Source:   m.Source,
			Target:   m.Destination,
			ReadOnly: m.ReadOnly,
		}
		res = append(res, mnt)
	}
	return res
}

// toMobyContainerStopOpts transforms types.ContainerStopOptions into a generic type required by the Docker SDK.
func toMobyContainerStopOpts(opts *types.ContainerStopOptions) client.ContainerStopOptions {
	if opts == nil {
		return client.ContainerStopOptions{}
	}

	return client.ContainerStopOptions{
		Signal:  opts.Signal,
		Timeout: opts.Timeout,
	}
}

// toMobyContainerRestartOpts transfroms types.ContainerRestartOptions into a generic type required by the Docker SDK.
func toMobyContainerRestartOpts(opts *types.ContainerRestartOptions) client.ContainerRestartOptions {
	if opts == nil {
		return client.ContainerRestartOptions{}
	}

	return client.ContainerRestartOptions{
		Signal:  opts.Signal,
		Timeout: opts.Timeout,
	}
}

// toMobyContainerRemoveOpts transforms types.ContainerRemoveOptions into a generic type required by the Docker SDK.
func toMobyContainerRemoveOpts(opts *types.ContainerRemoveOptions) client.ContainerRemoveOptions {
	if opts == nil {
		return client.ContainerRemoveOptions{}
	}

	return client.ContainerRemoveOptions{
		RemoveVolumes: opts.RemoveVolumes,
		Force:         opts.Force,
	}
}

// toMobyContainerInspectOpts transforms types.ContainerStatusOptions into a generic type required by the Docker SDK.
func toMobyContainerInspectOpts(opts *types.ContainerStatusOptions) client.ContainerInspectOptions {
	if opts == nil {
		return client.ContainerInspectOptions{}
	}

	return client.ContainerInspectOptions{
		Size: opts.Size,
	}
}

// toMobyContainerListOpts transforms types.ContainerListOptions into a generic type required by the Docker SDK.
func toMobyContainerListOpts(opts *types.ContainerListOptions) client.ContainerListOptions {
	if opts == nil {
		return client.ContainerListOptions{}
	}

	return client.ContainerListOptions{
		All:     opts.All,
		Filters: mapToMobyFilters(opts.Filters),
	}
}

// toMobyContainerWaitOpts transforms types.ContainerWaitOptions into a generic type required by the Docker SDK.
func toMobyContainerWaitOpts(opts *types.ContainerWaitOptions) (client.ContainerWaitOptions, error) {
	if opts == nil {
		return client.ContainerWaitOptions{}, fmt.Errorf("options cannot be nil")
	}
	if !opts.Condition.IsValid() {
		return client.ContainerWaitOptions{}, fmt.Errorf("invalid condition: %v", opts.Condition)
	}

	return client.ContainerWaitOptions{
		Condition: container.WaitCondition(opts.Condition),
	}, nil
}

// toMobyExecCreateOpts transforms types.ContainerExecOptions into a generic type required by the Docker SDK.
func toMobyExecCreateOpts(opts *types.ContainerExecOptions) (client.ExecCreateOptions, error) {
	if opts == nil {
		return client.ExecCreateOptions{}, fmt.Errorf("options cannot be nil")
	}
	if len(opts.Cmd) == 0 {
		return client.ExecCreateOptions{}, fmt.Errorf("command cannot be empty")
	}

	return client.ExecCreateOptions{
		User:         opts.User,
		Cmd:          opts.Cmd,
		WorkingDir:   opts.WorkingDir,
		Env:          opts.Env,
		TTY:          opts.TTY,
		AttachStdout: opts.Stdout != nil,
		AttachStderr: opts.Stderr != nil,
	}, nil
}

// toMobyContainerLogsOpts transforms types.ContainerLogsOptions into a generic type required by the Docker SDK.
func toMobyContainerLogsOpts(opts *types.ContainerLogsOptions) (client.ContainerLogsOptions, error) {
	if opts == nil {
		return client.ContainerLogsOptions{}, fmt.Errorf("options cannot be nil")
	}

	showStdout := opts.Stdout != nil
	showStderr := opts.Stderr != nil
	if !showStdout && !showStderr {
		return client.ContainerLogsOptions{}, fmt.Errorf("one of Stdout, Stderr must not be nil")
	}

	return client.ContainerLogsOptions{
		ShowStdout: showStdout,
		ShowStderr: showStderr,
		Since:      mapToMobyTime(opts.Since),
		Until:      mapToMobyTime(opts.Until),
		Timestamps: opts.Timestamps,
		Follow:     opts.Follow,
		Tail:       opts.Tail,
		Details:    opts.Details,
	}, nil
}

// fromMobyInspectResponse transforms container.InspectResponse into types.ContainerStatusResult.
func fromMobyInspectResponse(resp container.InspectResponse) types.ContainerStatusResult {
	ipAddr := netip.Addr{}
	if resp.State.Status != "exited" {
		ipAddr = getIPAddressFromNetworkSettings(resp.NetworkSettings)
	}
	status := types.ContainerStatusResult{
		ID:        resp.ID,
		Status:    string(resp.State.Status),
		IPAddress: ipAddr,
		ExitCode:  resp.State.ExitCode,
	}
	return status
}

// fromMobyContainerList transforms client.ContainerListResult into types.ContainerListResult.
func fromMobyContainerList(resp client.ContainerListResult) types.ContainerListResult {
	var res []types.ContainerSummary
	for _, c := range resp.Items {
		t := time.Time{}
		if c.Created != 0 {
			t = time.Unix(c.Created, 0)
		}

		cont := types.ContainerSummary{
			ID:      c.ID,
			Names:   c.Names,
			Image:   c.Image,
			State:   string(c.State),
			Status:  c.Status,
			Created: t,
			Labels:  c.Labels,
		}
		res = append(res, cont)
	}
	return types.ContainerListResult{
		Containers: res,
	}
}

// getIPAddressFromNetworkSettings extracts an IP address from a container's network settings, if any exists.
func getIPAddressFromNetworkSettings(net *container.NetworkSettings) netip.Addr {
	if net == nil {
		return netip.Addr{}
	}

	networks := net.Networks
	priority := []string{"bridge", "host", "overlay"}

	for _, name := range priority {
		if settings, ok := networks[name]; ok && settings.IPAddress.IsValid() {
			return settings.IPAddress
		}
	}

	for _, settings := range networks {
		if settings.IPAddress.IsValid() {
			return settings.IPAddress
		}
	}

	return netip.Addr{}
}

// mapToEnv converts map[string]string to []string{"KEY=VAL"}.
func mapToEnv(env map[string]string) []string {
	var res []string
	for k, v := range env {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return res
}
