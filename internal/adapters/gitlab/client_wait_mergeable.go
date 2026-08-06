package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// mergePollTimeout and mergePollInterval govern how long WaitUntilMergeable
// waits for GitLab to finish computing merge-ability before giving up.
//
// They are package-level variables (not constants) so tests can shrink them to
// keep the suite fast. Production code uses the defaults below.
var (
	mergePollTimeout  = 30 * time.Second
	mergePollInterval = 2 * time.Second
)

// GetMergeRequest fetches a single merge request, including the asynchronously
// computed merge_status / detailed_merge_status fields.
func (c *Client) GetMergeRequest(ctx context.Context, projectID int, mergeRequestIID int) (MergeRequest, error) {
	if projectID <= 0 {
		return MergeRequest{}, errors.New("gitlab project ID must be greater than zero")
	}
	if mergeRequestIID <= 0 {
		return MergeRequest{}, errors.New("gitlab merge request IID must be greater than zero")
	}

	path := fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d", projectID, mergeRequestIID)

	var mergeRequest MergeRequest
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &mergeRequest); err != nil {
		return MergeRequest{}, err
	}
	return mergeRequest, nil
}

// WaitUntilMergeable polls the merge request until GitLab reports it as
// mergeable, then returns it.
//
// GitLab computes merge-ability asynchronously. Immediately after a merge
// request is opened its detailed_merge_status is "checking"/"unchecked" and a
// PUT /merge issued in that window returns HTTP 405 Method Not Allowed. Callers
// must therefore wait for a mergeable state before attempting the merge.
//
// The function is fail-closed: it returns an error on a clearly non-mergeable
// state (for example "conflict") and on timeout, instead of proceeding to a
// merge that GitLab would reject.
func (c *Client) WaitUntilMergeable(ctx context.Context, projectID int, mergeRequestIID int) (MergeRequest, error) {
	if projectID <= 0 {
		return MergeRequest{}, errors.New("gitlab project ID must be greater than zero")
	}
	if mergeRequestIID <= 0 {
		return MergeRequest{}, errors.New("gitlab merge request IID must be greater than zero")
	}

	deadline := time.Now().Add(mergePollTimeout)

	for {
		mergeRequest, err := c.GetMergeRequest(ctx, projectID, mergeRequestIID)
		if err != nil {
			return MergeRequest{}, err
		}

		if isMergeable(mergeRequest) {
			return mergeRequest, nil
		}

		if !isTransientMergeStatus(mergeRequest) {
			return MergeRequest{}, fmt.Errorf(
				"gitlab merge request %d is not mergeable: detailed_merge_status=%q merge_status=%q",
				mergeRequestIID, mergeRequest.DetailedMergeStatus, mergeRequest.MergeStatus,
			)
		}

		if time.Now().After(deadline) {
			return MergeRequest{}, fmt.Errorf(
				"timed out after %s waiting for gitlab merge request %d to become mergeable: detailed_merge_status=%q merge_status=%q",
				mergePollTimeout, mergeRequestIID, mergeRequest.DetailedMergeStatus, mergeRequest.MergeStatus,
			)
		}

		select {
		case <-ctx.Done():
			return MergeRequest{}, ctx.Err()
		case <-time.After(mergePollInterval):
		}
	}
}

// isMergeable reports whether GitLab considers the merge request ready to merge.
// It prefers the modern detailed_merge_status field and falls back to the
// legacy merge_status for older GitLab versions that do not populate it.
func isMergeable(mergeRequest MergeRequest) bool {
	detailed := strings.ToLower(strings.TrimSpace(mergeRequest.DetailedMergeStatus))
	if detailed == "mergeable" {
		return true
	}
	if detailed == "" && strings.ToLower(strings.TrimSpace(mergeRequest.MergeStatus)) == "can_be_merged" {
		return true
	}
	return false
}

// isTransientMergeStatus reports whether the current status is one GitLab is
// expected to resolve on its own, so the caller should keep polling.
func isTransientMergeStatus(mergeRequest MergeRequest) bool {
	switch strings.ToLower(strings.TrimSpace(mergeRequest.DetailedMergeStatus)) {
	case "", "checking", "unchecked", "preparing", "ci_still_running":
		return true
	}
	switch strings.ToLower(strings.TrimSpace(mergeRequest.MergeStatus)) {
	case "checking", "unchecked":
		return true
	}
	return false
}
