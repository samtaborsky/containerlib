// SPDX-License-Identifier: MIT

package types

import "time"

// ImagePullOptions may hold future optional arguments used for pulling images from registry.
type ImagePullOptions struct {
	// All specifies whether to pull every single tag of the image.
	All bool
}

// ImageInspectOptions may hold future optional arguments used for inspecting images.
type ImageInspectOptions struct{}

// ImageInspectResult contains information about an image.
type ImageInspectResult struct {
	// ID represents the image ID.
	ID string

	// Created is a timestamp of the image creation.
	Created time.Time
	// Size is the size of all the layers of the image on disk.
	Size int64

	// Tags contain list of tags associated with the image.
	Tags []string

	// Architecture specifies the architecture the image is built to run on.
	Architecture string
	// OperatingSystem specifies the operating system the image is built to run on.
	OperatingSystem string
}

// ImageListOptions holds optional arguments used for listing images on the host.
type ImageListOptions struct {
	// All specifies whether to also list intermediate image layers.
	All bool

	// Filters contain predicates for filtering the request.
	Filters Filters
}

// ImageListResult holds a list of images on the host system and basic information about them.
type ImageListResult struct {
	Images []ImageSummary
}

// ImageSummary contains basic information about an image.
type ImageSummary struct {
	// ID represents the image ID.
	ID string

	// Created is a timestamp of the image creation.
	Created time.Time
	// Size is the size of all the layers of the image on disk.
	Size int64

	// Tags contain list of tags associated with the image.
	Tags []string

	// Labels contains metadata for the image.
	Labels map[string]string
}

// ImageRemoveOptions holds optional arguments used for removing images.
type ImageRemoveOptions struct {
	// Force tells the daemon to remove the image even if there are stopped containers using it.
	Force bool

	// PruneChildren tells the daemon to clean up any not needed untagged parent layers.
	PruneChildren bool
}

// ImageRemoveResult holds a list of removed images (and tags).
type ImageRemoveResult struct {
	ImagesRemoved []ImageRemoveSummary
}

// ImageRemoveSummary contains information about the removed images.
type ImageRemoveSummary struct {
	// Untagged is the ID of the untagged image.
	Untagged string
	// Deleted is the ID of the deleted image.
	Deleted string
}

// ImagePruneOptions holds optional arguments used for pruning dangling images.
type ImagePruneOptions struct {
	// Filters contains predicates for filtering the request.
	Filters Filters
}

// ImagePruneResult contains a summary of the Prune operation.
type ImagePruneResult struct {
	// ImagesRemoved contains information about all pruned images.
	ImagesRemoved []ImageRemoveSummary
	// SpaceReclaimed is the amount of space freed by the Prune operation.
	SpaceReclaimed uint64
}
