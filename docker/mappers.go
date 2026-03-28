// SPDX-License-Identifier: MIT

package docker

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// mapToMobyFilters transforms types.Filters into a generic type required by the Docker SDK.
func mapToMobyFilters(f types.Filters) client.Filters {
	mobyFilters := make(client.Filters)

	if f == nil {
		return mobyFilters
	}
	for key, values := range f {
		for _, val := range values {
			mobyFilters.Add(key, val)
		}
	}

	return mobyFilters
}

// mapToMobyTime transforms time.Time into a RFC3339Nano time string used by the Docker SDK.
func mapToMobyTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

// mapFromMobyError maps an error returned from the Docker SDK to one of the library defined error types.
func mapFromMobyError(err error, override ...error) error {
	if err == nil {
		return nil
	}

	// Because of dependency version problems, github.com/moby/moby/errdefs could not be used.
	// github.com/containerd/errdefs works for most Moby SDK errors except for conflict errors,
	// so this custom logic was needed. Other error types work as expected.
	var conflictErr interface{ Conflict() bool }
	if errors.As(err, &conflictErr) && conflictErr.Conflict() {
		return fmt.Errorf("%w: %w", types.ErrConflict, err)
	}
	if strings.Contains(err.Error(), "Conflict") || strings.Contains(err.Error(), "already in use") {
		return fmt.Errorf("%w: %w", types.ErrConflict, err)
	}

	if errdefs.IsNotFound(err) {
		finalError := types.ErrNotFound
		if len(override) > 0 {
			finalError = override[0]
		}
		return fmt.Errorf("%w: %w", finalError, err)
	}
	if errdefs.IsPermissionDenied(err) {
		return fmt.Errorf("%w: %w", types.ErrPermissionsDenied, err)
	}
	if errdefs.IsUnavailable(err) {
		return fmt.Errorf("%w: %w", types.ErrConnectionFailed, err)
	}
	if errdefs.IsInvalidArgument(err) {
		return fmt.Errorf("%w: %w", types.ErrInvalidInput, err)
	}
	if errdefs.IsUnauthorized(err) {
		return fmt.Errorf("%w: %w", types.ErrUnauthorized, err)
	}

	return fmt.Errorf("%w: %w", types.ErrInternal, err)
}

// toNanoCPUs converts regular allocation of CPUs (e.g. 0.5 cores) to NanoCPUs used by Docker SDK.
func toNanoCPUs(cpus float64) int64 {
	return int64(cpus * 1e9)
}

// megabytesToBytes converts megabytes to bytes.
func megabytesToBytes(mb int64) int64 {
	return mb * (1024 * 1024)
}

// bytesToMegabytes converts bytes to megabytes.
func bytesToMegabytes(b uint64) uint64 {
	return b / (1024 * 1024)
}
