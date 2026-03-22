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

	go func() {
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
	}()

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
		types.EventActionOOM:
		return true
	default:
		return false
	}
}
