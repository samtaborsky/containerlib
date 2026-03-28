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
	// If no ServerAddress is provided, docker.io is used.
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
	Name string `json:"name"`
	// Image specifies an image to be used and a version (e.g. "alpine:latest").
	// This is the only mandatory field.
	Image string `json:"image"`

	// User specifies the user which will run the commands inside the container.
	// A "user:group" format is also supported.
	User string `json:"user"`

	// Env is a map of desired environment variables and their values.
	Env map[string]string `json:"env,omitempty"`
	// Labels specifies metadata for the container.
	Labels map[string]string `json:"labels,omitempty"`

	// Cmd holds the command to run when the container starts.
	Cmd []string `json:"cmd,omitempty"`
	// Entrypoint represents the entrypoint to run when the container starts.
	Entrypoint []string `json:"entrypoint,omitempty"`

	// Ports is a list of port bindings for the container.
	Ports []PortBinding `json:"ports,omitempty"`
	// Mounts is a list of bind mounts to volumes or directories for the container.
	Mounts []Mount `json:"mounts,omitempty"`

	// CPUs specifies the maximum number of CPUs available to the container.
	CPUs float64 `json:"cpus"`
	// MemoryMb specifies the maximum memory in megabytes available to the container.
	//
	// If set to 0, the container will have the default memory limit.
	// If set to -1, the container will have unlimited memory.
	MemoryMb int64 `json:"memoryMb"`

	// RestartPolicy specifies the restart policy used for the container.
	Restart RestartPolicy `json:"restartPolicy"`

	// Privileged specifies whether to run the container in privileged mode.
	Privileged bool `json:"privileged"`
}

// ContainerCreateResult holds the ID of the created container.
type ContainerCreateResult struct {
	ID string `json:"id"`
}

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
	// RemoveVolumes specifies whether volumes mounted to the container should also be removed.
	RemoveVolumes bool
	// Force specifies whether the container should be forcibly terminated (if running) before removing it.
	Force bool
}

// ContainerStatusOptions holds optional arguments used for inspecting containers.
type ContainerStatusOptions struct {
	// Size specifies whether to calculate the size of the container's filesystem (costly operation).
	// It should not be used unless needed.
	Size bool
}

// ContainerStatusResult represents the current status of a container.
type ContainerStatusResult struct {
	// ID represents the container ID.
	ID string `json:"id"`

	// Status specifies the current state of the container (e.g. "running").
	Status string `json:"status"`
	// IPAddress specifies the IP address of the container.
	IPAddress string `json:"ipAddress"`

	// ExitCode contains the exit code of the container, if it is stopped.
	ExitCode int `json:"exitCode"`
}

// ContainerListOptions holds optional arguments used for listing containers present on the host system.
type ContainerListOptions struct {
	// All specifies whether to also return non-running containers in the list.
	All bool

	// Filters contain predicates for filtering the request.
	Filters Filters
}

// ContainerListResult holds a list of containers on the host system and basic information about them.
type ContainerListResult struct {
	Containers []ContainerSummary `json:"containers"`
}

// ContainerSummary represents basic information about a container.
type ContainerSummary struct {
	// ID represents the container ID.
	ID string `json:"id"`

	// Names is a list of names associated with the container.
	Names []string `json:"names"`
	// Image represents the image running on the container.
	Image string `json:"image"`

	// State represents a human-readable state of the container (e.g. "Up 1 minute").
	State string `json:"state"`
	// Status represent the current state of the container (e.g. "running").
	Status string `json:"status"`

	// Created is a timestamp of container creation.
	Created time.Time `json:"created"`

	// Labels contains metadata for the container.
	Labels map[string]string `json:"labels,omitempty"`
}

// ContainerWaitOptions holds optional arguments used for waiting on containers.
type ContainerWaitOptions struct {
	// Condition represents the specific condition which is being waited for.
	Condition WaitCondition
}

// ContainerWaitResult holds the exit code of the exited container process.
type ContainerWaitResult struct {
	ExitCode int64 `json:"exitCode"`
}

// ContainerExecOptions holds the command to be executed inside the container.
type ContainerExecOptions struct {
	// User specifies the user which will run the command.
	User string
	// Cmd holds the command and all its arguments. Must not be empty.
	Cmd []string

	// Stdout specifies whether the content of Stdout will be returned as io.Writer.
	// If empty, Stdout will be ignored.
	Stdout io.Writer
	// Stderr specifies whether the content of Stderr will be returned as io.Writer.
	// If empty, Stderr will be ignored.
	Stderr io.Writer

	// WorkingDir specifies the working directory to run the command.
	WorkingDir string
	// Env holds the environment variables.
	Env []string

	// TTY specifies whether to allocate a pseudo-TTY, used e.g. for interactive commands or scripts.
	TTY bool
}

// ContainerExecResult holds the exit code result from executing a command inside a container.
type ContainerExecResult struct {
	ExitCode int64 `json:"exitCode"`
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
// One of Stdout or Stderr fields must be set to true.
type ContainerLogsOptions struct {
	// Stdout specifies whether the Stdout stream will be returned as io.Writer.
	//
	// One of Stdout or Stderr must be not nil.
	Stdout io.Writer
	// Stderr specifies whether the Stderr stream will be returned as io.Writer.
	//
	// One of Stdout or Stderr must be not nil.
	Stderr io.Writer

	// Since represents a time from which onward logs should be returned.
	Since time.Time
	// Until represents a time up to which logs should be returned.
	Until time.Time

	// Timestamps specifies whether each line of logs will have its timestamp at the beginning.
	Timestamps bool

	// Follow keeps the connection open and continuously stream new logs as they are produced.
	Follow bool

	// Tail specifies how many lines from the end should be returned.
	// It should be a stringified number or "all".
	Tail string

	// Details specifies whether to include detailed information and metadata which were passed
	// to the containers logging driver.
	Details bool
}

// ContainerLogsResult may hold future optional information returned from the function.
type ContainerLogsResult struct {
}

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
