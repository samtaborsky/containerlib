// SPDX-License-Identifier: MIT

package types

import "time"

// ContainerStatsOptions holds optional arguments used for retrieving resource usage information.
type ContainerStatsOptions struct {
	Stream bool
}

// ContainerStatsResult contains resource usage information of a container.
type ContainerStatsResult struct {
	Stats  <-chan ContainerStats
	Errors <-chan error
}

// ContainerStats contains information about the resource usage of a container at a given time.
type ContainerStats struct {
	// ID represents the container ID.
	ID string `json:"id"`
	// Name is the name of the queried container.
	Name string `json:"name"`
	// OperatingSystem specifies the operating system of the container ("windows" or "linux").
	OperatingSystem string `json:"os"`
	// ReadTime specifies the time at which the stats were obtained.
	ReadTime time.Time `json:"readTime"`

	// CPU holds information about CPU usage of the queried container.
	CPU CPUStats `json:"cpu"`
	// Memory holds information about memory usage of the queried container.
	Memory MemoryStats `json:"memory"`
}

// CPUStats holds information about CPU usage at a single point in time.
type CPUStats struct {
	// UsedPercent specifies the current CPU usage as a calculated percentage of the total host capacity.
	//
	// A value of 50.0 represents using half of a single core's capacity.
	//
	// A value of 300.0 represents three cores being used at their maximum capacity.
	//
	// This value is normalized across Windows and Linux runtimes.
	UsedPercent float64 `json:"usedPercent"`
	// UsageNano represents the total absolute CPU time used by the container.
	UsageNano uint64 `json:"usageNano"`
}

// MemoryStats holds information about memory usage at a single point in time.
type MemoryStats struct {
	// UsedPercent specifies the percentage of memory used.
	UsedPercent float64 `json:"usedPercent"`
	// UsedMb specifies the amount of memory used by the container in megabytes.
	UsedMb uint64 `json:"usedMb"`
	// MaxUsedMb specifies the maximum recorded memory usage in megabytes.
	MaxUsedMb uint64 `json:"maxUsedMb,omitempty"`
	// LimitMb specifies the memory limit in megabytes.
	LimitMb uint64 `json:"limitMb,omitempty"`
}
