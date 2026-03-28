// SPDX-License-Identifier: MIT

package types

import "time"

// InfoResult represents information about the host server.
type InfoResult struct {
	// ID specifies the system ID.
	ID string `json:"id"`

	// Containers specifies the number of containers on the system.
	Containers int `json:"containers"`
	// ContainersRunning specifies the number of running containers on the system.
	ContainersRunning int `json:"containersRunning"`
	// ContainersPaused specifies the number of paused containers on the system.
	ContainersPaused int `json:"containersPaused"`
	// ContainersStopped specifies the number of stopped containers on the system.
	ContainersStopped int `json:"containersStopped"`

	// Images specifies the number of images on the system.
	Images int `json:"images"`

	// Driver specifies the storage driver used by the system.
	Driver string `json:"driver"`

	// MemoryLimit indicates whether the host kernel supports setting memory limits on containers.
	MemoryLimit bool `json:"memoryLimit"`
	// SwapLimit indicates whether the host kernel supports swap memory limiting.
	SwapLimit bool `json:"swapLimit"`

	// SystemTime specifies the system time.
	SystemTime time.Time `json:"systemTime"`

	// KernelVersion specifies the kernel version.
	KernelVersion string `json:"kernelVersion"`
	// OperatingSystem specifies the host operating system.
	OperatingSystem string `json:"os"`
	// OSVersion specifies the host operating system version.
	OSVersion string `json:"osVersion"`
	// OSType indicates the operating system type.
	OSType string `json:"osType"`
	// Architecture specifies the host architecture.
	Architecture string `json:"architecture"`

	// NCPU specifies the number of CPUs on the host.
	NCPU int `json:"nCPU"`

	// MemTotal specifies the total amount of memory on the host.
	MemTotal int64 `json:"memTotal"`

	// Name specifies the host system name.
	Name string `json:"name"`
	// ServerVersion specifies the host server version.
	ServerVersion string `json:"serverVersion"`

	// Warnings contains warnings which occurred during the collection of information about the system.
	Warnings []string `json:"warnings,omitempty"`
}
