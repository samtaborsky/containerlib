// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"encoding/json"
	"io"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

func (rt *runtime) ContainerStats(ctx context.Context, id string, opts *types.ContainerStatsOptions) (types.ContainerStatsResult, error) {
	resp, err := rt.api.ContainerStats(ctx, id, toMobyContainerStatsOpts(opts))
	if err != nil {
		return types.ContainerStatsResult{}, mapFromMobyError(err, types.ErrContainerNotFound)
	}

	outStats := make(chan types.ContainerStats)
	outErrors := make(chan error)

	go decodeStatsResponse(ctx, resp.Body, outStats, outErrors)

	return types.ContainerStatsResult{
		Stats:  outStats,
		Errors: outErrors,
	}, nil
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Helper functions
// ---------------------------------------------------------------------------------------------------------------------

// toMobyContainerStatsOpts transforms types.ContainerStatsOptions into a generic type required by the Docker SDK.
func toMobyContainerStatsOpts(opts *types.ContainerStatsOptions) client.ContainerStatsOptions {
	if opts == nil {
		return client.ContainerStatsOptions{IncludePreviousSample: true}
	}

	return client.ContainerStatsOptions{
		Stream:                opts.Stream,
		IncludePreviousSample: true,
	}
}

// decodeStatsResponse continuously reads and decodes the raw telemetry JSON stream from the Docker daemon,
// mapping it into types.ContainerStats.
//
// This function is designed to run as a background goroutine. It assumes full ownership of the provided io.ReadCloser
// and output channels, ensuring they are all properly closed when the stream terminates.
//
// The decoding loop will run indefinitely until one of three things happens:
//
// 1. The provided context is canceled or times out.
//
// 2. The Docker daemon closes the connection (io.EOF).
//
// 3. The JSON stream becomes corrupted or unreadable.
func decodeStatsResponse(ctx context.Context, reader io.ReadCloser, outStats chan<- types.ContainerStats, outErrors chan<- error) {
	defer close(outStats)
	defer close(outErrors)
	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			outErrors <- err
		}
	}(reader)

	decoder := json.NewDecoder(reader)
	for {
		select {
		case <-ctx.Done():
			outErrors <- mapFromMobyError(ctx.Err())
			return
		default:
			var raw container.StatsResponse
			if err := decoder.Decode(&raw); err != nil {
				if err != io.EOF {
					outErrors <- mapFromMobyError(err)
				}
				return
			}

			stats := fromMobyContainerStats(&raw)
			select {
			case outStats <- stats:
			case <-ctx.Done():
				outErrors <- mapFromMobyError(ctx.Err())
				return
			}
		}
	}
}

// fromMobyContainerStats transforms container.StatsResponse into types.ContainerStats.
func fromMobyContainerStats(stats *container.StatsResponse) types.ContainerStats {
	return types.ContainerStats{
		ID:              stats.ID,
		Name:            stats.Name,
		OperatingSystem: stats.OSType,
		ReadTime:        stats.Read,
		CPU:             mapToCPUStats(stats),
		Memory:          mapToMemoryStats(stats),
	}
}

// getProcessorCount identifies the processor count based on the OS type.
func getProcessorCount(stats *container.StatsResponse) float64 {
	if stats.OSType == "windows" {
		return float64(stats.NumProcs)
	}
	return float64(stats.CPUStats.OnlineCPUs)
}

// mapToCPUStats calculates CPU related statistics and returns them in types.CPUStats.
func mapToCPUStats(stats *container.StatsResponse) types.CPUStats {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)

	var systemDelta float64
	onlineCPUs := getProcessorCount(stats)

	if stats.OSType == "windows" {
		systemDelta = float64(stats.Read.Sub(stats.PreRead).Nanoseconds())
	} else {
		systemDelta = float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	}

	var cpuPercent float64
	if systemDelta > 0.0 && cpuDelta > 0.0 && onlineCPUs > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}

	usageNano := stats.CPUStats.CPUUsage.TotalUsage
	if stats.OSType == "windows" {
		usageNano *= 100
	}

	return types.CPUStats{
		UsedPercent: cpuPercent,
		UsageNano:   usageNano,
	}
}

// mapToMemoryStats calculates memory related statistics and returns them in types.MemoryStats.
func mapToMemoryStats(stats *container.StatsResponse) types.MemoryStats {
	usage := stats.MemoryStats.Usage
	limit := stats.MemoryStats.Limit
	percentage := 0.0

	if limit > 0 && usage <= limit {
		percentage = float64(usage) / float64(limit) * 100
	}

	return types.MemoryStats{
		UsedPercent: percentage,
		UsedMb:      bytesToMegabytes(usage),
	}
}
