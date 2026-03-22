// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/distribution/reference"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"
	"github.com/samtaborsky/containerlib/types"
)

func (rt *runtime) ImagePull(ctx context.Context, name string, opts *types.ImagePullOptions) error {
	imageRef, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return fmt.Errorf("invalid image name: %w", err)
	}

	imageRef = reference.TagNameOnly(imageRef)
	domain := reference.Domain(imageRef)

	rt.mu.RLock()
	auth, ok := rt.authStore[domain]
	rt.mu.RUnlock()

	options := toMobyImagePullOpts(opts)
	if ok {
		encodedAuth, err := encodeRegistryAuth(auth)
		if err != nil {
			return fmt.Errorf("failed to encode registry auth: %w", err)
		}
		options.RegistryAuth = encodedAuth
	}

	resp, err := rt.api.ImagePull(ctx, imageRef.String(), options)
	if err != nil {
		return mapFromMobyError(err)
	}
	defer resp.Close()

	if err := resp.Wait(ctx); err != nil {
		return fmt.Errorf("image pull failed: %w", err)
	}

	return nil
}

func (rt *runtime) ImageInspect(ctx context.Context, id string, opts *types.ImageInspectOptions) (types.ImageInspectResult, error) {
	resp, err := rt.api.ImageInspect(ctx, id)
	if err != nil {
		return types.ImageInspectResult{}, mapFromMobyError(err)
	}

	return fromMobyImageInspectResult(resp), nil
}

func (rt *runtime) ImageList(ctx context.Context, opts *types.ImageListOptions) (types.ImageListResult, error) {
	resp, err := rt.api.ImageList(ctx, toMobyImageListOpts(opts))
	if err != nil {
		return types.ImageListResult{}, mapFromMobyError(err)
	}

	return fromMobyImageListResult(resp), nil
}

func (rt *runtime) ImageRemove(ctx context.Context, id string, opts *types.ImageRemoveOptions) (types.ImageRemoveResult, error) {
	resp, err := rt.api.ImageRemove(ctx, id, toMobyImageRemoveOpts(opts))
	if err != nil {
		return types.ImageRemoveResult{}, mapFromMobyError(err)
	}

	return fromMobyImageRemoveResult(resp), nil
}

func (rt *runtime) ImagePrune(ctx context.Context, opts *types.ImagePruneOptions) (types.ImagePruneResult, error) {
	resp, err := rt.api.ImagePrune(ctx, toMobyImagePruneOpts(opts))
	if err != nil {
		return types.ImagePruneResult{}, mapFromMobyError(err)
	}

	return fromMobyImagePruneReport(resp), nil
}

// ---------------------------------------------------------------------------------------------------------------------
// --- Helper functions
// ---------------------------------------------------------------------------------------------------------------------

// toMobyImagePullOpts transforms types.ImagePullOptions into a generic type required by the Docker SDK.
func toMobyImagePullOpts(opts *types.ImagePullOptions) client.ImagePullOptions {
	if opts == nil {
		return client.ImagePullOptions{}
	}

	return client.ImagePullOptions{
		All: opts.All,
	}
}

// toMobyImageListOpts transforms types.ImageListOptions into a generic type required by the Docker SDK.
func toMobyImageListOpts(opts *types.ImageListOptions) client.ImageListOptions {
	if opts == nil {
		return client.ImageListOptions{}
	}

	return client.ImageListOptions{
		All: opts.All,
	}
}

// toMobyImageRemoveOpts transforms types.ImageRemoveOptions into a generic type required by the Docker SDK.
func toMobyImageRemoveOpts(opts *types.ImageRemoveOptions) client.ImageRemoveOptions {
	if opts == nil {
		return client.ImageRemoveOptions{}
	}

	return client.ImageRemoveOptions{
		Force:         opts.Force,
		PruneChildren: opts.PruneChildren,
	}
}

// toMobyImagePruneOpts transforms types.ImagePruneOptions into a generic type required by the Docker SDK.
func toMobyImagePruneOpts(opts *types.ImagePruneOptions) client.ImagePruneOptions {
	if opts == nil {
		return client.ImagePruneOptions{}
	}

	return client.ImagePruneOptions{
		Filters: mapToMobyFilters(opts.Filters),
	}
}

// fromMobyImageInspectResult transforms client.ImageInspectResult into types.ImageInspectResult.
func fromMobyImageInspectResult(resp client.ImageInspectResult) types.ImageInspectResult {
	t, err := time.Parse(time.RFC3339Nano, resp.Created)
	if err != nil {
		t = time.Time{}
	}
	return types.ImageInspectResult{
		ID:              resp.ID,
		Created:         t,
		Size:            resp.Size,
		Tags:            resp.RepoTags,
		Architecture:    resp.Architecture,
		OperatingSystem: resp.Os,
	}
}

// fromMobyImageListResult transforms client.ImageListResult into types.ImageListResult.
func fromMobyImageListResult(resp client.ImageListResult) types.ImageListResult {
	var res []types.ImageSummary
	for _, i := range resp.Items {
		img := types.ImageSummary{
			ID:      i.ID,
			Created: time.Unix(i.Created, 0),
			Size:    i.Size,
			Tags:    i.RepoTags,
			Labels:  i.Labels,
		}
		res = append(res, img)
	}
	return types.ImageListResult{
		Images: res,
	}
}

// fromMobyImageRemoveResult transforms client.ImageRemoveResult into types.ImageRemoveResult.
func fromMobyImageRemoveResult(resp client.ImageRemoveResult) types.ImageRemoveResult {
	var res []types.ImageRemoveSummary
	for _, i := range resp.Items {
		img := types.ImageRemoveSummary{
			Untagged: i.Untagged,
			Deleted:  i.Deleted,
		}
		res = append(res, img)
	}
	return types.ImageRemoveResult{
		ImagesRemoved: res,
	}
}

// fromMobyImagePruneReport transforms client.ImagePruneResult into types.ImagePruneResult.
func fromMobyImagePruneReport(resp client.ImagePruneResult) types.ImagePruneResult {
	res := types.ImagePruneResult{
		SpaceReclaimed: resp.Report.SpaceReclaimed,
	}
	for _, i := range resp.Report.ImagesDeleted {
		img := types.ImageRemoveSummary{
			Untagged: i.Untagged,
			Deleted:  i.Deleted,
		}
		res.ImagesRemoved = append(res.ImagesRemoved, img)
	}
	return res
}

// encodeRegistryAuth converts types.AuthConfig into a Base64-encoded JSON string used
// by the Docker daemon for Pull, Push and Build operations.
func encodeRegistryAuth(auth types.AuthConfig) (string, error) {
	mobyAuth := registry.AuthConfig{
		Username:      auth.Username,
		Password:      auth.Password,
		ServerAddress: auth.ServerAddress,
		IdentityToken: auth.IdentityToken,
		RegistryToken: auth.RegistryToken,
	}

	encodedAuth, err := json.Marshal(mobyAuth)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(encodedAuth), nil
}
