// SPDX-License-Identifier: MIT

package types

// Mount represents a bind between a path on the host FS and a path inside the container FS.
// It is used when there is need to access files or directories inside the container from the host.
type Mount struct {
	// Source specifies the path on the host or a volume name.
	Source string
	// Destination specifies the path inside the container.
	Destination string

	// Type specifies the mount type (e.g. "bind", "volume").
	Type string

	// ReadOnly specifies if the mount is read-only.
	ReadOnly bool
}
