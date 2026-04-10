// SPDX-License-Identifier: MIT

package types

import "time"

// ImagePullOptions holds optional arguments used for pulling images from registry.
type ImagePullOptions struct {
	// All specifies whether to pull every single tag of the image.
	All bool `json:"all"`

	// Progress receives updates during the image pull.
	// This channel will be closed when the pull is finished.
	Progress chan<- ImagePullProgress
}

// ImagePullProgress contains information about the current progress of an image pull.
type ImagePullProgress struct {
	// ID represents the ID of the current layer.
	ID string `json:"id"`

	// Status represents the description of the current action (e.g. "Downloading")
	//
	// When Status contains "Pull complete" message, it informs about the pull of the current layer being
	// complete. The image pull is complete only when the channel closes or the ImageOps.ImagePull function returns.
	Status string `json:"status"`

	// Current is the number of bytes processed for this layer.
	Current int64 `json:"current"`
	// Total is the total number of bytes of this layer.
	// If 0, the size of the layer is not currently known.
	Total int64 `json:"total"`

	// Digest is the final SHA256 identifier of the image.
	// This is typically present in the final message in the stream.
	Digest string `json:"message"`
}

// ImageInspectOptions may hold future optional arguments used for inspecting images.
type ImageInspectOptions struct{}

// ImageInspectResult contains information about an image.
type ImageInspectResult struct {
	// ID represents the image ID.
	ID string `json:"id"`

	// Created is a timestamp of the image creation.
	Created time.Time `json:"created"`
	// Size is the size of all the layers of the image on disk.
	Size int64 `json:"size"`

	// Tags contain list of tags associated with the image.
	Tags []string `json:"tags,omitempty"`

	// Architecture specifies the architecture the image is built to run on.
	Architecture string `json:"architecture"`
	// OperatingSystem specifies the operating system the image is built to run on.
	OperatingSystem string `json:"os"`
}

// ImageListOptions holds optional arguments used for listing images on the host.
type ImageListOptions struct {
	// All specifies whether to also list intermediate image layers.
	All bool `json:"all"`

	// Filters contain predicates for filtering the request.
	Filters Filters `json:"filters"`
}

// ImageListResult holds a list of images on the host system and basic information about them.
type ImageListResult struct {
	Images []ImageSummary `json:"images"`
}

// ImageSummary contains basic information about an image.
type ImageSummary struct {
	// ID represents the image ID.
	ID string `json:"id"`

	// Created is a timestamp of the image creation.
	Created time.Time `json:"created"`
	// Size is the size of all the layers of the image on disk.
	Size int64 `json:"size"`

	// Tags contain list of tags associated with the image.
	Tags []string `json:"tags,omitempty"`

	// Labels contains metadata for the image.
	Labels map[string]string `json:"labels,omitempty"`
}

// ImageRemoveOptions holds optional arguments used for removing images.
type ImageRemoveOptions struct {
	// Force tells the daemon to remove the image even if there are stopped containers using it.
	Force bool `json:"force"`

	// PruneChildren tells the daemon to clean up any not needed untagged parent layers.
	PruneChildren bool `json:"pruneChildren"`
}

// ImageRemoveResult holds a list of removed images (and tags).
type ImageRemoveResult struct {
	ImagesRemoved []ImageRemoveSummary `json:"imagesRemoved"`
}

// ImageRemoveSummary contains information about the removed images.
type ImageRemoveSummary struct {
	// Untagged is the ID of the untagged image.
	Untagged string `json:"untagged"`
	// Deleted is the ID of the deleted image.
	Deleted string `json:"deleted"`
}

// ImagePruneOptions holds optional arguments used for pruning dangling images.
type ImagePruneOptions struct {
	// Filters contains predicates for filtering the request.
	Filters Filters `json:"filters"`
}

// ImagePruneResult contains a summary of the Prune operation.
type ImagePruneResult struct {
	// ImagesRemoved contains information about all pruned images.
	ImagesRemoved []ImageRemoveSummary `json:"imagesRemoved"`
	// SpaceReclaimed is the amount of space freed by the Prune operation.
	SpaceReclaimed uint64 `json:"spaceReclaimed"`
}
