// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- API functions
// ---------------------------------------------------------------------------------------------------------------------

func (m *mockDockerAPI) Events(ctx context.Context, options client.EventsListOptions) client.EventsResult {
	if m.eventsFunc == nil {
		panic("Events function not mocked")
	}
	return m.eventsFunc(ctx, options)
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestEventsStream(t *testing.T) {
	tests := []struct {
		name           string
		input          *types.EventsStreamOptions
		mockMsgs       []events.Message
		mockErr        error
		cancelCtx      bool
		expectedEvents int
		expectedErrors int
		expectedErr    error
	}{
		{
			name: "Happy path",
			input: &types.EventsStreamOptions{
				Filters: map[string][]string{"type": {"container"}},
			},
			mockMsgs: []events.Message{
				{Type: "container", Action: "start", Actor: events.Actor{ID: "cont-1"}, TimeNano: 1000},
				{Type: "container", Action: "stop", Actor: events.Actor{ID: "cont-1"}, TimeNano: 2000},
			},
			expectedEvents: 2,
			expectedErrors: 0,
		},
		{
			name:           "Error path - Stream error",
			mockErr:        errors.New("stream interrupted"),
			expectedEvents: 0,
			expectedErrors: 1,
			expectedErr:    types.ErrInternal,
		},
		{
			name:           "Error path - Context cancelled",
			mockErr:        errors.New("context canceled"),
			cancelCtx:      true,
			expectedEvents: 0,
			expectedErrors: 1,
			expectedErr:    types.ErrInternal,
		},
		{
			name:           "Error path - Daemon unreachable",
			mockErr:        generateRealConnectionError(),
			expectedEvents: 0,
			expectedErrors: 1,
			expectedErr:    types.ErrConnectionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mobyMsgs := make(chan events.Message, len(tc.mockMsgs))
			mobyErrs := make(chan error, 1)

			for _, m := range tc.mockMsgs {
				mobyMsgs <- m
			}
			if tc.mockErr != nil {
				mobyErrs <- tc.mockErr
			}

			mockAPI := &mockDockerAPI{
				eventsFunc: func(ctx context.Context, options client.EventsListOptions) client.EventsResult {
					return client.EventsResult{
						Messages: mobyMsgs,
						Err:      mobyErrs,
					}
				},
			}

			ctx, cancel := context.WithCancel(context.Background())

			if tc.cancelCtx {
				cancel()
			} else {
				defer cancel()
			}

			rt := &runtime{api: mockAPI}
			res := rt.EventsStream(ctx, tc.input)

			var receivedEvents []types.Event
			var receivedErrors []error

			go func() {
				time.Sleep(10 * time.Millisecond)
				close(mobyMsgs)
				close(mobyErrs)
			}()

			done := make(chan struct{})
			go func() {
				for e := range res.Events {
					receivedEvents = append(receivedEvents, e)
				}
				for err := range res.Errors {
					receivedErrors = append(receivedErrors, err)
				}
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(200 * time.Millisecond):
				t.Fatal("test timed out waiting for streamMobyEvents to finish")
			}

			if len(receivedEvents) != tc.expectedEvents {
				t.Errorf("expected %d events, got %d", tc.expectedEvents, len(receivedEvents))
			}
			if len(receivedErrors) != tc.expectedErrors {
				t.Errorf("expected %d errors, got %d", tc.expectedErrors, len(receivedErrors))
			}

			if tc.expectedErr != nil {
				if len(receivedErrors) == 0 {
					t.Errorf("expected error %v, but got none", tc.expectedErr)
				} else {
					actualErr := receivedErrors[0]
					if !errors.Is(actualErr, tc.expectedErr) {
						t.Errorf("expected error %v, got %v", tc.expectedErr, actualErr)
					}
				}
			}
		})
	}
}
