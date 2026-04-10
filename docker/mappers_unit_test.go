// SPDX-License-Identifier: MIT

package docker

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// --- Errors
// ---------------------------------------------------------------------------------------------------------------------

type mockConnectionFailed struct{}

func (e mockConnectionFailed) Error() string { return "simulated connection failure" }

// As intercepts the errors.As() call made inside client.IsErrConnectionFailed.
func (e mockConnectionFailed) As(target interface{}) bool {
	if reflect.TypeOf(target).Elem().Name() == "errConnectionFailed" {
		return true
	}
	return false
}

type mockErrdefs struct {
	keyword string
}

func (e mockErrdefs) Error() string { return "mock " + e.keyword }

// Is intercepts the errors.Is() call made inside errdefs.Is... functions.
func (e mockErrdefs) Is(target error) bool {
	targetStr := strings.ToLower(target.Error())
	return strings.Contains(targetStr, e.keyword)
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Tests
// ---------------------------------------------------------------------------------------------------------------------

func TestToMobyFilters(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Filters
		expected client.Filters
	}{
		{
			name:     "Happy path",
			input:    make(types.Filters).Add("foo", "bar"),
			expected: make(client.Filters).Add("foo", "bar"),
		},
		{
			name:     "Happy path - Empty filters",
			input:    types.Filters{},
			expected: client.Filters{},
		},
		{
			name:     "Happy path - Nil filters",
			input:    nil,
			expected: client.Filters{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := mapToMobyFilters(tc.input)

			if !reflect.DeepEqual(res, tc.expected) {
				t.Errorf("expected filters %+v, got %+v", tc.expected, res)
			}
		})
	}
}

func TestMapFromMobyError(t *testing.T) {
	tests := []struct {
		name        string
		input       error
		override    []error
		expectedErr error
	}{
		{
			name:        "Happy path - Nil error",
			input:       nil,
			expectedErr: nil,
		},
		{
			name:        "Happy path - Conflict (string match 'is not')",
			input:       fmt.Errorf("the resource is not available"),
			expectedErr: types.ErrConflict,
		},
		{
			name:        "Happy path - Conflict (interface)",
			input:       mockErrConflict{"port already allocated"},
			expectedErr: types.ErrConflict,
		},
		{
			name:        "Happy path - Invalid argument (errdefs)",
			input:       mockErrdefs{"invalid argument"},
			expectedErr: types.ErrInvalidInput,
		},
		{
			name:        "Happy path - Permission denied (errdefs)",
			input:       mockErrdefs{"permission denied"},
			expectedErr: types.ErrPermissionsDenied,
		},
		{
			name:        "Happy path - Unauthorized (errdefs)",
			input:       mockErrdefs{"unauth"},
			expectedErr: types.ErrUnauthorized,
		},
		{
			name:        "Happy path - Not found (errdefs)",
			input:       mockErrdefs{"not found"},
			expectedErr: types.ErrNotFound,
		},
		{
			name:        "Happy path - Not found override (errdefs)",
			input:       mockErrdefs{"not found"},
			override:    []error{types.ErrContainerNotFound},
			expectedErr: types.ErrContainerNotFound,
		},
		{
			name:        "Happy path - Connection failed (moby)",
			input:       mockConnectionFailed{},
			expectedErr: types.ErrConnectionFailed,
		},
		{
			name:        "Happy path - Fallback",
			input:       fmt.Errorf("random system crash"),
			expectedErr: types.ErrInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			var res error
			if tc.override != nil {
				res = mapFromMobyError(tc.input, tc.override...)
			} else {
				res = mapFromMobyError(tc.input)
			}

			if tc.expectedErr == nil {
				if res != nil {
					t.Fatalf("expected no error, got %v", res)
				}
				return
			}
			if !errors.Is(res, tc.expectedErr) {
				if !errors.Is(res, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, res)
				}
			}
		})
	}
}
