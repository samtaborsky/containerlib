// SPDX-License-Identifier: MIT

package types

import "time"

// InfoResult represents information about the host server.
type InfoResult struct {
	// ID specifies the system ID.
	ID string

	// Containers specifies the number of containers on the system.
	Containers int
	// ContainersRunning specifies the number of running containers on the system.
	ContainersRunning int
	// ContainersPaused specifies the number of paused containers on the system.
	ContainersPaused int
	// ContainersStopped specifies the number of stopped containers on the system.
	ContainersStopped int

	// Images specifies the number of images on the system.
	Images int

	// Driver specifies the storage driver used by the system.
	Driver string

	// MemoryLimit indicates whether the host kernel supports setting memory limits on containers.
	MemoryLimit bool
	// SwapLimit indicates whether the host kernel supports swap memory limiting.
	SwapLimit bool

	// SystemTime specifies the system time.
	SystemTime time.Time

	// KernelVersion specifies the kernel version.
	KernelVersion string
	// OperatingSystem specifies the host operating system.
	OperatingSystem string
	// OSVersion specifies the host operating system version.
	OSVersion string
	// OSType indicates the operating system type.
	OSType string
	// Architecture specifies the host architecture.
	Architecture string

	// NCPU specifies the number of CPUs on the host.
	NCPU int

	// MemTotal specifies the total amount of memory on the host.
	MemTotal int64

	// Name specifies the host system name.
	Name string
	// ServerVersion specifies the host server version.
	ServerVersion string

	// Warnings contains warnings which occurred during the collection of information about the system.
	Warnings []string
}
