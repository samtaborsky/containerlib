// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

func (rt *runtime) EventsStream(ctx context.Context, opts *types.EventsStreamOptions) types.EventsStreamResult {
	resp := rt.api.Events(ctx, toMobyEventsOpts(opts))

	outEvents := make(chan types.Event)
	outErrors := make(chan error, 1)

	go streamMobyEvents(ctx, resp, outEvents, outErrors)

	return types.EventsStreamResult{
		Events: outEvents,
		Errors: outErrors,
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Helper functions
// ---------------------------------------------------------------------------------------------------------------------

// toMobyEventsOpts transforms types.EventsStreamOptions into a generic type required by the Docker SDK.
func toMobyEventsOpts(opts *types.EventsStreamOptions) client.EventsListOptions {
	if opts == nil {
		return client.EventsListOptions{}
	}

	filters := mapToMobyFilters(opts.Filters).Add("type", "container", "image", "network", "volume")

	return client.EventsListOptions{
		Since:   mapToMobyTime(opts.Since),
		Until:   mapToMobyTime(opts.Until),
		Filters: filters,
	}
}

// streamMobyEvents continuously listens to the Docker daemon's event stream, filtering and mapping generic Moby event
// structs into types.Event structs.
//
// This function is designed to run as a background goroutine. It assumes full ownership of the provided output
// channels, ensuring they are all properly closed when the stream terminates.
//
// The streaming loop will run indefinitely until one of three things happens:
//
// 1. The provided context is canceled or times out.
//
// 2. The underlying Docker messages channel is closed.
//
// 3. A fatal error is received from the underlying Docker error channel.
func streamMobyEvents(ctx context.Context, resp client.EventsResult, outEvents chan<- types.Event, outErrors chan<- error) {
	defer close(outEvents)
	defer close(outErrors)

	for {
		select {
		case <-ctx.Done():
			outErrors <- mapFromMobyError(ctx.Err())
			return
		case event, ok := <-resp.Messages:
			if !ok {
				return
			}
			if !isSupportedEventType(event.Type) || !isSupportedEventAction(event.Action) {
				continue
			}

			outEvent := types.Event{
				Type:   types.EventType(event.Type),
				Action: types.EventAction(event.Action),
				Actor:  types.EventActor{ID: event.Actor.ID},
				Time:   time.Unix(0, event.TimeNano),
			}
			outEvents <- outEvent
		case err, ok := <-resp.Err:
			if !ok {
				return
			}
			if err != nil {
				outErrors <- mapFromMobyError(err)
			}
			return
		}
	}
}

// isSupportedEventType determines whether the provided events.Type is supported by the library.
func isSupportedEventType(eventType events.Type) bool {
	switch types.EventType(eventType) {
	case types.EventTypeContainer,
		types.EventTypeImage,
		types.EventTypeVolume,
		types.EventTypeNetwork:
		return true
	default:
		return false
	}
}

// isSupportedEventAction determines whether the provided events.Action is supported by the library.
func isSupportedEventAction(action events.Action) bool {
	switch types.EventAction(action) {
	case types.EventActionCreate,
		types.EventActionStart,
		types.EventActionStop,
		types.EventActionRestart,
		types.EventActionPause,
		types.EventActionUnPause,
		types.EventActionRemove,
		types.EventActionDelete,
		types.EventActionDestroy,
		types.EventActionDie,
		types.EventActionKill,
		types.EventActionOOM,
		types.EventActionPull:
		return true
	default:
		return false
	}
}
