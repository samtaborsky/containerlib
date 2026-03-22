// SPDX-License-Identifier: MIT

package types

import (
	"fmt"
	"io"
	"time"
)

// AuthConfig contains authentication information for a container registry.
type AuthConfig struct {
	Username string
	Password string
	// ServerAddress is the image registry address.
	// If no address is provided, docker.io is used.
	ServerAddress string

	// IdentityToken is a specialized session token.
	// If provided, it is used instead of a password to negotiate an
	// access token with the container registry authentication server.
	IdentityToken string

	// RegistryToken is a raw bearer-token (usually a JWT) used mainly in
	// CI/CD environments to bypass username/password negotiation and provide
	// immediate, temporary access.
	RegistryToken string
}

// ContainerCreateConfig represents the configuration needed for the creation of a new container.
type ContainerCreateConfig struct {
	// Name is an optional name of the container.
	Name string
	// Image specifies an image to be used and a version (e.g. "alpine:latest").
	// This is the only mandatory field.
	Image string

	// Env is a map of desired environment variables and their values.
	Env map[string]string
	// Labels specifies metadata for the container.
	Labels map[string]string

	// Ports is a list of port bindings for the container.
	Ports []PortBinding
	// Mounts is a list of bind mounts to volumes or directories for the container.
	Mounts []Mount

	// CPUs specifies the maximum number of CPUs available to the container.
	CPUs float64
	// MemoryMb specifies the maximum memory in megabytes available to the container.
	// If set to 0, the container will have unlimited memory.
	MemoryMb int64

	// RestartPolicy specifies the restart policy used for the container.
	Restart RestartPolicy
}

// ContainerCreateResult holds the ID of the created container.
type ContainerCreateResult string

// ContainerStartOptions may hold future optional arguments used for starting containers.
type ContainerStartOptions struct {
	// Future options can be added here
}

// ContainerStopOptions holds optional arguments used for stopping containers.
type ContainerStopOptions struct {
	// Signal specifies a signal to be (gracefully) sent to the container (default is "SIGTERM").
	Signal string
	// Timeout specifies a time (in seconds) to wait for the container to stop gracefully
	// before forcibly terminating it with "SIGKILL".
	Timeout *int
}

// ContainerRemoveOptions holds optional arguments used for removing containers.
type ContainerRemoveOptions struct {
	// RemoveVolumes specifies whether volumes mounted to the container should also be removed
	// (default is false).
	RemoveVolumes bool
	// Force specifies whether the container should be forcibly terminated (if running) before
	// removing it (default is false).
	Force bool
}

// ContainerStatusResult represents the current status of a container.
type ContainerStatusResult struct {
	// ID represents the container ID.
	ID string

	// Status specifies the current state of the container (e.g. "running").
	Status string
	// IPAddress specifies the IP address of the container.
	IPAddress string

	// ExitCode contains the exit code of the container, if it is stopped.
	ExitCode int
}

// ContainerListResult holds a list of containers on the host system and basic information about them.
type ContainerListResult struct {
	Containers []ContainerSummary
}

// ContainerSummary represents basic information about a container.
type ContainerSummary struct {
	// ID represents the container ID.
	ID string

	// Names is a list of names associated with the container.
	Names []string
	// Image represents the image running on the container.
	Image string

	// State represents a human-readable state of the container (e.g. "Up 1 minute").
	State string
	// Status represent the current state of the container (e.g. "running").
	Status string

	// Created is a timestamp of container creation.
	Created time.Time

	// Labels contains metadata for the container.
	Labels map[string]string
}

// ContainerWaitResult holds the exit code of the exited container process.
type ContainerWaitResult int64

// ContainerStatusOptions holds optional arguments used for inspecting containers.
type ContainerStatusOptions struct {
	// Size specifies whether to calculate the size of the container's filesystem (costly operation).
	// It should not be used unless needed (default is false).
	Size bool
}

// ContainerListOptions holds optional arguments used for listing containers present on the host system.
type ContainerListOptions struct {
	// All specifies whether to also return non-running containers in the list.
	All bool

	// Filters contain predicates for filtering the request.
	Filters Filters
}

// ContainerWaitOptions holds optional arguments used for waiting on containers.
type ContainerWaitOptions struct {
	Condition WaitCondition
}

// ContainerExecOptions holds the command to be executed inside the container.
type ContainerExecOptions struct {
	// User specifies the user which will run the command.
	User string
	// Cmd holds the command and all its arguments. Must not be empty.
	Cmd []string

	// AttachStdout specifies whether the content of STDOUT will be returned.
	AttachStdout bool
	// AttachStderr specifies whether the content of STDERR will be returned.
	AttachStderr bool

	// WorkingDir specifies the working directory to run the command.
	WorkingDir string
	// Env holds the environment variables.
	Env []string
}

// ContainerExecResult holds the output from executing a command inside a container.
type ContainerExecResult struct {
	// ExitCode holds the exit code of the executed command.
	ExitCode int64

	// Stdout holds the contents of STDOUT if AttachStdout was set to true.
	Stdout string
	// Stderr holds the contents of STDERR if AttachStderr was set to true.
	Stderr string
}

func (opts *ContainerExecOptions) Validate() error {
	if opts == nil {
		return fmt.Errorf("ContainerExecOptions cannot be nil")
	}
	if len(opts.Cmd) == 0 {
		return fmt.Errorf("ContainerExecOptions Command cannot be empty")
	}
	return nil
}

// ContainerLogsOptions holds optional parameters which affect the type and structure of returned logs.
// One of ShowStdout or ShowStderr fields must be set to true.
type ContainerLogsOptions struct {
	// ShowStdout specifies whether the STDOUT stream will be captured.
	//
	// One of ShowStdout or ShowStderr must be set to true.
	ShowStdout bool
	// ShowStderr specifies whether the STDERR stream will be captured.
	//
	// One of ShowStdout or ShowStderr must be set to true.
	ShowStderr bool

	// Since represents a time from which onward logs should be returned.
	Since time.Time
	// Until represents a time up to which logs should be returned.
	Until time.Time

	// Timestamps specifies whether each line of logs will have its timestamp at the beginning.
	Timestamps bool

	// Follow keeps the connection open and continously stream new logs as they are produced.
	Follow bool

	// Tail specifies how many lines from the end should be returned.
	// It should be a stringified number or "all".
	Tail string

	// Details specifies whether to include detailed information and metadata which were passed
	// to the containers logging driver.
	Details bool
}

// ContainerLogsResult holds a stream of logs from the container.
type ContainerLogsResult io.ReadCloser

// WaitCondition represents a state of a container which has to be reached for the condition to be met.
type WaitCondition string

const (
	WaitConditionNotRunning WaitCondition = "not-running"
	WaitConditionNextExit   WaitCondition = "next-exit"
	WaitConditionRemoved    WaitCondition = "removed"
)

func (w WaitCondition) String() string {
	return string(w)
}

func (w WaitCondition) IsValid() bool {
	switch w {
	case "", WaitConditionNotRunning, WaitConditionNextExit, WaitConditionRemoved:
		return true
	default:
		return false
	}
}

// RestartPolicy represents the restart policy used for a container.
type RestartPolicy string

const (
	RestartDisable       RestartPolicy = "no"
	RestartAlways        RestartPolicy = "always"
	RestartOnFailure     RestartPolicy = "on-failure"
	RestartUnlessStopped RestartPolicy = "unless-stopped"
)

func (r RestartPolicy) String() string {
	return string(r)
}

func (r RestartPolicy) IsValid() bool {
	switch r {
	case "", RestartAlways, RestartDisable, RestartOnFailure, RestartUnlessStopped:
		return true
	default:
		return false
	}
}
