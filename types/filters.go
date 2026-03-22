// SPDX-License-Identifier: MIT

package types

// Filters represents a set of criteria for filtering results (e.g. containers or images).
// It acts as an AND filter between different keys, and an OR filter for values of the same key.
//
// Always initialize it using make(Filters) or by explicit definition/
//
// Example:
//
//	f := make(Filters).Add("status", "running").Add("name", "foo", "bar")
type Filters map[string][]string

// Add appends values to the specific filter key.
// It returns the Filters for chaining.
func (f Filters) Add(key string, values ...string) Filters {
	f[key] = append(f[key], values...)
	return f
}
