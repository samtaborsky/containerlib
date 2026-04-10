// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestToMobyEventsOpts(t *testing.T) {
	curTime := time.Now().Truncate(time.Second)
	mobyTime := curTime.Format(time.RFC3339Nano)
	mobyFilters := make(client.Filters).Add("foo", "bar").Add("type", "container", "image", "network", "volume")

	tests := []struct {
		name     string
		input    *types.EventsStreamOptions
		expected client.EventsListOptions
	}{
		{
			name: "Happy path",
			input: &types.EventsStreamOptions{
				Filters: make(types.Filters).Add("foo", "bar"),
				Since:   curTime,
			},
			expected: client.EventsListOptions{
				Since:   mobyTime,
				Filters: mobyFilters,
			},
		},
		{
			name:     "Happy path - Empty options",
			input:    nil,
			expected: client.EventsListOptions{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := toMobyEventsOpts(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func TestStreamMobyEvents(t *testing.T) {
	tests := []struct {
		name           string
		mockMsg        *events.Message
		mockErr        error
		cancelCtx      bool
		expectedEvents int
		expectedErrors int
	}{
		{
			name: "Happy path - Supported event",
			mockMsg: &events.Message{
				Type:   "container",
				Action: "start",
				Actor:  events.Actor{ID: "123"},
			},
			expectedEvents: 1,
		},
		{
			name: "Happy path - Unsupported Type",
			mockMsg: &events.Message{
				Type:   "secret",
				Action: "delete",
			},
			expectedEvents: 0,
		},
		{
			name: "Happy path - Unsupported Action",
			mockMsg: &events.Message{
				Type:   "image",
				Action: "push",
			},
			expectedEvents: 0,
		},
		{
			name:           "Error path - Stream error",
			mockErr:        errors.New("stream interrupted"),
			expectedErrors: 1,
		},
		{
			name:           "Error path - Context cancelled",
			cancelCtx:      true,
			expectedErrors: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mobyMsgs := make(chan events.Message, 1)
			mobyErrs := make(chan error, 1)
			outEvents := make(chan types.Event, 1)
			outErrors := make(chan error, 1)

			ctx, cancel := context.WithCancel(context.Background())

			if tc.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}

			if tc.mockMsg != nil {
				mobyMsgs <- *tc.mockMsg
			}
			if tc.mockErr != nil {
				mobyErrs <- tc.mockErr
			}

			resp := client.EventsResult{
				Messages: (<-chan events.Message)(mobyMsgs),
				Err:      (<-chan error)(mobyErrs),
			}

			go func() {
				time.Sleep(10 * time.Millisecond)
				close(mobyMsgs)
				close(mobyErrs)
			}()

			streamMobyEvents(ctx, resp, outEvents, outErrors)

			receivedEvents := 0
			for range outEvents {
				receivedEvents++
			}
			if receivedEvents != tc.expectedEvents {
				t.Errorf("expected %d events, got %d", tc.expectedEvents, receivedEvents)
			}

			receivedErrors := 0
			for range outErrors {
				receivedErrors++
			}
			if receivedErrors != tc.expectedErrors {
				t.Errorf("expected %d errors, got %d", tc.expectedErrors, receivedErrors)
			}
		})
	}
}
