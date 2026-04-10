// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"sync"
	"time"

	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

var _ types.Runtime = (*runtime)(nil)

type runtime struct {
	api       client.APIClient
	mu        sync.RWMutex
	authStore map[string]types.AuthConfig
}

// New initializes a new types.Runtime connected to the specified Docker host.
// The host can be specified with an IP address (TCP socket) or an address of a unix socket provided
// through the `host` argument, or from the environment variables (e.g. DOCKER_HOST).
// It returns any errors that may have occurred.
func New(host string) (types.Runtime, error) {
	var opts []client.Opt

	if host != "" {
		opts = append(opts, client.WithHost(host))
	} else {
		opts = append(opts, client.FromEnv)
	}

	cli, err := client.New(opts...)
	if err != nil {
		return nil, mapFromMobyError(err)
	}

	authStore := make(map[string]types.AuthConfig)

	return &runtime{api: cli, authStore: authStore}, nil
}

func (rt *runtime) Info(ctx context.Context) (types.InfoResult, error) {
	resp, err := rt.api.Info(ctx, client.InfoOptions{})
	if err != nil {
		return types.InfoResult{}, mapFromMobyError(err)
	}

	return fromMobySystemInfo(resp.Info), nil
}

func (rt *runtime) Login(ctx context.Context, auth types.AuthConfig) error {
	mobyAuth := client.RegistryLoginOptions{
		Username:      auth.Username,
		Password:      auth.Password,
		ServerAddress: auth.ServerAddress,
		IdentityToken: auth.IdentityToken,
		RegistryToken: auth.RegistryToken,
	}

	resp, err := rt.api.RegistryLogin(ctx, mobyAuth)
	if err != nil {
		return mapFromMobyError(err)
	}

	rt.mu.Lock()
	if resp.Auth.IdentityToken != "" {
		auth.IdentityToken = resp.Auth.IdentityToken
		auth.Password = ""
	}

	registryDomain := auth.ServerAddress
	if registryDomain == "" {
		auth.ServerAddress = "docker.io"
	}

	if rt.authStore == nil {
		rt.authStore = make(map[string]types.AuthConfig)
	}
	rt.authStore[registryDomain] = auth
	rt.mu.Unlock()

	return nil
}

func (rt *runtime) Close() error {
	err := rt.api.Close()
	return mapFromMobyError(err)
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Helper functions
// ---------------------------------------------------------------------------------------------------------------------

func fromMobySystemInfo(resp system.Info) types.InfoResult {
	t, err := time.Parse(time.RFC3339Nano, resp.SystemTime)
	if err != nil {
		t = time.Time{}
	}
	return types.InfoResult{
		ID:                resp.ID,
		Containers:        resp.Containers,
		ContainersRunning: resp.ContainersRunning,
		ContainersPaused:  resp.ContainersPaused,
		ContainersStopped: resp.ContainersStopped,
		Images:            resp.Images,
		Driver:            resp.Driver,
		MemoryLimit:       resp.MemoryLimit,
		SwapLimit:         resp.SwapLimit,
		SystemTime:        t,
		KernelVersion:     resp.KernelVersion,
		OperatingSystem:   resp.OperatingSystem,
		OSVersion:         resp.OSVersion,
		OSType:            resp.OSType,
		Architecture:      resp.Architecture,
		NCPU:              resp.NCPU,
		MemTotal:          resp.MemTotal,
		Name:              resp.Name,
		ServerVersion:     resp.ServerVersion,
		Warnings:          resp.Warnings,
	}
}
