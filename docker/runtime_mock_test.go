// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- API functions
// ---------------------------------------------------------------------------------------------------------------------

func (m *mockDockerAPI) Info(ctx context.Context, options client.InfoOptions) (client.SystemInfoResult, error) {
	if m.infoFunc == nil {
		panic("Info function not mocked")
	}
	return m.infoFunc(ctx, options)
}

func (m *mockDockerAPI) RegistryLogin(ctx context.Context, auth client.RegistryLoginOptions) (client.RegistryLoginResult, error) {
	if m.registryLoginFunc == nil {
		panic("RegistryLogin function not mocked")
	}
	return m.registryLoginFunc(ctx, auth)
}

func (m *mockDockerAPI) Close() error {
	if m.closeFunc == nil {
		panic("Close function not mocked")
	}
	return m.closeFunc()
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestInfo(t *testing.T) {
	tests := []struct {
		name         string
		mockResp     client.SystemInfoResult
		mockErr      error
		expectedResp types.InfoResult
		expectedErr  error
	}{
		{
			name: "Happy path",
			mockResp: client.SystemInfoResult{
				Info: system.Info{ID: "system-id"},
			},
			mockErr:      nil,
			expectedResp: types.InfoResult{ID: "system-id"},
			expectedErr:  nil,
		},
		{
			name:         "Error path - Daemon unreachable",
			mockResp:     client.SystemInfoResult{},
			mockErr:      generateRealConnectionError(),
			expectedResp: types.InfoResult{},
			expectedErr:  types.ErrConnectionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				infoFunc: func(ctx context.Context, options client.InfoOptions) (client.SystemInfoResult, error) {
					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			resp, err := rt.Info(context.Background())

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expectedRes error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expectedRes no error, got %v", err)
			}

			if resp.ID != tc.expectedResp.ID {
				t.Errorf("expectedRes returned ID to be %q, got %q", tc.expectedResp.ID, resp.ID)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name        string
		mockOpts    types.AuthConfig
		mockResp    client.RegistryLoginResult
		mockErr     error
		expectedErr error
	}{
		{
			name: "Happy path",
			mockOpts: types.AuthConfig{
				Username: "test",
				Password: "test",
			},
			mockResp: client.RegistryLoginResult{
				Auth: registry.AuthResponse{IdentityToken: "token"},
			},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name: "Error path - Daemon unreachable",
			mockOpts: types.AuthConfig{
				Username: "test",
				Password: "test",
			},
			mockResp:    client.RegistryLoginResult{},
			mockErr:     generateRealConnectionError(),
			expectedErr: types.ErrConnectionFailed,
		},
		{
			name: "Error path - Invalid server address",
			mockOpts: types.AuthConfig{
				Username:      "test",
				Password:      "test",
				ServerAddress: "invalid.io",
			},
			mockResp:    client.RegistryLoginResult{},
			mockErr:     mockErrInternal{"Error response from daemon: Get \"https://invalid.io/v2/\""},
			expectedErr: types.ErrInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				registryLoginFunc: func(ctx context.Context, auth client.RegistryLoginOptions) (client.RegistryLoginResult, error) {
					return tc.mockResp, tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			err := rt.Login(context.Background(), tc.mockOpts)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expectedRes error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expectedRes no error, got %v", err)
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name        string
		mockErr     error
		expectedErr error
	}{
		{
			name:        "Happy path - Closes successfully",
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Error path - Daemon socket error on close",
			mockErr:     fmt.Errorf("failed to close idle connections"),
			expectedErr: types.ErrInternal,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI := &mockDockerAPI{
				closeFunc: func() error {
					return tc.mockErr
				},
			}

			rt := &runtime{api: mockAPI}
			err := rt.Close()

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expectedRes error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expectedRes no error, got %v", err)
			}
		})
	}
}
