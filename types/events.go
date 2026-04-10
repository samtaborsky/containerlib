// SPDX-License-Identifier: MIT

package types

import (
	"time"
)

// EventsStreamOptions holds optional arguments for filtering which events get sent from the daemon.
type EventsStreamOptions struct {
	// Since represents a time from which onward events should be returned.
	Since time.Time `json:"since"`
	// Until represents a time up to which events should be returned.
	Until time.Time `json:"until"`

	// Filters contain predicates for filtering the request.
	Filters Filters `json:"filters,omitempty"`
}

// EventsStreamResult holds streams of errors and events from the daemon.
type EventsStreamResult struct {
	Events <-chan Event
	Errors <-chan error
}

// Event contains data about the event.
type Event struct {
	// Type represens the type of the event.
	Type EventType `json:"type"`

	// Action represents the event action.
	Action EventAction `json:"action"`

	// Actor represents the actor of the event.
	Actor EventActor `json:"actor"`

	// Time specifies the time of the event happening.
	Time time.Time `json:"time"`
}

// EventType represents the type of event (the type of resource that generated it).
type EventType string

const (
	// EventTypeContainer is a type that a container generates.
	EventTypeContainer EventType = "container"

	// EventTypeImage is a type that an image generates.
	EventTypeImage EventType = "image"

	// EventTypeVolume is a type that a volume generates.
	EventTypeVolume EventType = "volume"

	// EventTypeNetwork is a type that a network generates.
	EventTypeNetwork EventType = "network"
)

// EventAction represents the action that is connected to the event.
type EventAction string

const (
	// EventActionCreate indicates that a container or resource was initialized but not yet started.
	EventActionCreate EventAction = "create"

	// EventActionStart indicates that a container's main process has started running.
	EventActionStart EventAction = "start"

	// EventActionRestart indicates that a container was restarted.
	EventActionRestart EventAction = "restart"

	// EventActionStop indicates that a container was gracefully stopped.
	EventActionStop EventAction = "stop"

	// EventActionPause indicates that a running container's processes were frozen.
	EventActionPause EventAction = "pause"

	// EventActionUnPause indicates that a paused container's processes were resumed.
	EventActionUnPause EventAction = "unpause"

	// EventActionKill indicates that a container was forcefully terminated using a system signal (e.g., SIGKILL).
	EventActionKill EventAction = "kill"

	// EventActionDie indicates that a container's main process has exited, regardless of the reason.
	EventActionDie EventAction = "die"

	// EventActionOOM indicates that a container was terminated by the host kernel for exceeding its memory limit.
	EventActionOOM EventAction = "oom"

	// EventActionDestroy indicates that a container's filesystem and resources were completely destroyed.
	EventActionDestroy EventAction = "destroy"

	// EventActionRemove indicates that a resource (like a container or volume) was removed from the daemon.
	EventActionRemove EventAction = "remove"

	// EventActionDelete indicates that a resource (typically an image) was deleted from the host.
	EventActionDelete EventAction = "delete"

	// EventActionPull indicates that an image was pulled from a registry.
	EventActionPull EventAction = "pull"
)

// EventActor contains information about the resource that generated the event.
type EventActor struct {
	ID string
}
