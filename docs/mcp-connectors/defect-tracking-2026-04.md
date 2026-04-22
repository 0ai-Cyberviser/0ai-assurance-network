# MCP Connector Defect Tracking - April 2026 Smoke Testing

## Overview

This document tracks MCP connector defects and inconsistencies reproduced during live smoke testing on April 16-17, 2026 against GitHub and Canva connectors.

**Testing Context:**
- Live connector testing (not mocked responses)
- Real GitHub API and Canva API interactions
- Disposable test resources used for validation

**Related Issues:**
- Umbrella: 0ai-Cyberviser/0ai-assurance-network#30
- Child Issues: #31, #32, #33, #34, #35, #36

---

## GitHub MCP Connector Defects

### Defect #31: Invalid GraphQL Field in `mark_pull_request_ready_for_review`

**Issue:** 0ai-Cyberviser/0ai-assurance-network#31

**Status:** Documented

**Severity:** High - Blocks ready-for-review workflow

**Description:**
The `mark_pull_request_ready_for_review` tool uses an invalid GraphQL query that requests `PullRequest.htmlUrl`, which does not exist in GitHub's GraphQL schema.

**Current Behavior:**
```graphql
mutation {
  markPullRequestReadyForReview(input: {...}) {
    pullRequest {
      htmlUrl  # INVALID - this field does not exist
    }
  }
}
```

**Expected Behavior:**
```graphql
mutation {
  markPullRequestReadyForReview(input: {...}) {
    pullRequest {
      url  # Use 'url' instead
    }
  }
}
```

**Root Cause:**
GraphQL schema mismatch - `PullRequest` type exposes `url` field, not `htmlUrl`.

**Required Fix:**
1. Update GraphQL mutation response shape to use `url` instead of `htmlUrl`
2. If normalizer expects `html_url`, map GraphQL `url` to that field explicitly
3. Update any response parsing logic to handle the correct field name

**Validation Plan:**
1. Create a disposable draft PR
2. Execute `mark_pull_request_ready_for_review` tool
3. Verify PR transitions to ready-for-review state
4. Confirm no GraphQL errors in response

**Priority:** 1 (First fix - unblocks review workflows)

---

### Defect #32: Review ID Format Mismatch Between Create and Dismiss

**Issue:** 0ai-Cyberviser/0ai-assurance-network#32

**Status:** Fixed

**Severity:** High - Breaks review lifecycle management

**Description:**
`add_review_to_pr` returns a numeric database ID, but `dismiss_pull_request_review` expects a GraphQL node ID format, causing dismissal operations to fail.

**Current Behavior:**
- `add_review_to_pr` returns: `{ "review_id": 4125974959 }` (numeric)
- `dismiss_pull_request_review` expects: GraphQL node ID like `"PRR_kwDOA..."`

**Expected Behavior:**
Both tools should use compatible ID formats that can be used interchangeably.

**Root Cause:**
REST API returns numeric database IDs, GraphQL mutations require node IDs.

**Required Fix:**
Option A (Preferred):
```json
{
  "review_id": 4125974959,
  "review_node_id": "PRR_kwDOA..."
}
```

Option B:
Internal translation layer that converts numeric ID to node ID before dismissing.

**Validation Plan:**
1. Create a disposable PR review using `add_review_to_pr`
2. Capture returned review ID
3. Attempt to dismiss the review using `dismiss_pull_request_review` with the returned ID
4. Verify dismissal succeeds without ID format errors

**Priority:** 2 (Second fix - enables complete review lifecycle)

---

### Defect #33: `list_pull_request_reviews` Omits COMMENTED Reviews

**Issue:** 0ai-Cyberviser/0ai-assurance-network#33

**Status:** Fixed

**Severity:** Medium - Read path trust failure

**Description:**
`list_pull_request_reviews` returns empty results even when a submitted `COMMENTED` review exists on the PR.

**Current Behavior:**
- Submit a review with state `COMMENTED`
- Call `list_pull_request_reviews`
- Result: Empty list (review is missing)

**Expected Behavior:**
All submitted reviews should be included in the listing, including:
- `COMMENTED`
- `APPROVED`
- `CHANGES_REQUESTED`
- `DISMISSED`
- `PENDING`

**Root Cause:**
Likely filtering or normalization logic that excludes `COMMENTED` review state.

**Required Fix:**
1. Compare `list_pull_request_reviews` implementation with `fetch_pr_comments` (proven working)
2. Remove any filters that exclude `COMMENTED` state
3. Ensure normalization preserves all review states
4. Verify pagination doesn't skip `COMMENTED` reviews

**Validation Plan:**
1. Create a disposable PR
2. Submit a review with state `COMMENTED`
3. Call `list_pull_request_reviews`
4. Verify the `COMMENTED` review appears in results
5. Test with other review states to ensure no regression

**Priority:** 3 (Third fix - restores read-path trust)

**Note:** GitHub product rules prevent dismissing plain `COMMENTED` reviews - this is expected behavior, not a connector defect.

---

### Defect #34: Issue Comment Reactions Not Reflected After Creation

**Issue:** 0ai-Cyberviser/0ai-assurance-network#34

**Status:** Fixed

**Severity:** Medium - Read path trust failure

**Description:**
`get_issue_comment_reactions` returns empty results immediately after successfully creating an issue comment reaction.

**Current Behavior:**
1. Create reaction on issue comment → Success
2. Read reactions via `get_issue_comment_reactions` → Empty list

**Expected Behavior:**
Newly created reactions should appear in the readback immediately.

**Root Cause:**
``get_issue_comment_reactions`` was routing to the issue-level endpoint
``/repos/{owner}/{repo}/issues/{number}/reactions`` instead of the
comment-level endpoint
``/repos/{owner}/{repo}/issues/comments/{comment_id}/reactions``.  GitHub
silently scopes the query and returns an empty list when there are no
issue-level reactions, even when the comment itself has reactions.

**Fix Applied:**
File: ``src/assurancectl/github_mcp.py``

```python
# WRONG (returns empty list for comment reactions):
path = f"/repos/{owner}/{repo}/issues/{issue_number}/reactions"

# CORRECT:
path = f"/repos/{owner}/{repo}/issues/comments/{comment_id}/reactions"
```

The module also shares a single ``_normalize_reaction`` function between the
issue-comment and PR reaction paths, guaranteeing identical output shape and
eliminating any future normalisation divergence.

**Tests:** ``tests/test_github_mcp.py``
- ``TestGetIssueCommentReactions.test_calls_issue_comment_reactions_endpoint``
- ``TestGetIssueCommentReactions.test_does_not_call_issue_level_endpoint``
- ``TestGetIssueCommentReactions.test_immediately_readable_after_creation``
- ``TestGetPrReactions.test_issue_comment_and_pr_reactions_share_normalizer``

**Validation Plan:**
1. Create a disposable issue comment
2. Add a reaction (e.g., 👍) using the connector
3. Immediately read reactions via `get_issue_comment_reactions`
4. Verify created reaction appears in the list
5. Test with multiple reaction types

**Priority:** 4 (Fourth fix - completes read-path trust)

---

### Defect #36: Same-Repo PR Update Fails with `maintainer_can_modify`

**Issue:** 0ai-Cyberviser/0ai-assurance-network#36

**Status:** Fixed

**Severity:** Low - Wrapper validation hardening

**Description:**
`update_pull_request` includes `maintainer_can_modify: true` for same-repository PRs, causing GitHub to return HTTP 422 error because this field is only valid for cross-repo (fork) PRs.

**Current Behavior:**
```json
{
  "title": "Updated title",
  "maintainer_can_modify": true
}
```
Result: 422 error when PR is from same repo.

**Expected Behavior:**
Only include `maintainer_can_modify` when the PR is cross-repo (from a fork).

**Root Cause:**
Missing repository comparison logic - connector sends field unconditionally.

**Required Fix:**
```python
# Pseudo-code
if pr.head_repo != pr.base_repo:  # Cross-repo PR
    payload["maintainer_can_modify"] = value
# else: omit field for same-repo PRs
```

**Validation Plan:**
1. Create same-repo PR
2. Update PR metadata via `update_pull_request`
3. Verify update succeeds without 422 error
4. Create cross-repo (fork) PR
5. Update with `maintainer_can_modify`
6. Verify behavior remains intact for forks

**Priority:** 5 (Fifth fix - wrapper hardening)

---

## Canva MCP Connector Defects

### Defect #35: Asset Upload Not Reflected in Uploads Folder Listing

**Issue:** 0ai-Cyberviser/0ai-assurance-network#35

**Status:** Fixed

**Severity:** Medium - Read path trust failure

**Description:**
Successful asset upload operation not reflected in `Uploads` folder listing immediately after upload.

**Current Behavior:**
1. Upload asset → Success response
2. List Uploads folder → Asset not visible

**Expected Behavior:**
Uploaded assets should appear in folder listing immediately or with documented eventual consistency window.

**Root Cause:**
Possible causes:
- Eventual consistency in Canva's backend
- Wrong folder scope for listing
- Pagination issues
- Upload success doesn't guarantee folder placement

**Required Fix:**
1. Verify upload target folder matches listing scope
2. Check if Canva API has documented eventual consistency
3. Add retry logic with exponential backoff if eventual consistency is expected
4. Validate folder ID used in both upload and listing
5. Check pagination parameters in listing

**Root Cause (confirmed):**
Canva's backend has an eventual-consistency window for newly uploaded assets.
An upload job that returns success and a valid asset ID may not be immediately
visible in the Uploads folder listing.  The fix adds:

- `verify_upload` – retries `list_folder_items` with exponential backoff
  (up to 3 attempts; delays of 2 s, 4 s) until the asset appears.
- `list_all_folder_items` – follows pagination continuation tokens to
  guarantee the full listing is returned (partial pagination was a secondary
  suspected cause).

**Implementation:** `src/assurancectl/canva_connector.py`
**Tests:** `tests/test_canva_connector.py`

**Validation Plan:**
1. Upload a test asset to Canva
2. Capture upload success response and asset ID
3. List Uploads folder contents
4. Verify uploaded asset appears in listing
5. If eventual consistency: wait and retry, document timing
6. Test with multiple uploads

**Priority:** 6 (Sixth fix - after GitHub defects)

**Note:** Canva brand-template generation failures were excluded from tracking as they are plan-gated by Canva Enterprise tier, not connector defects.

---

## Fix Order and Dependencies

### Recommended Triage Order

1. **#31** - `mark_pull_request_ready_for_review` GraphQL fix
   - Unblocks PR state transitions
   - Quick GraphQL schema fix

2. **#32** - Review ID compatibility
   - Enables complete review lifecycle
   - Unblocks dismiss operations

3. **#33** - Review listing completeness
   - Restores read-path trust
   - No dependency on other fixes

4. **#34** - Issue comment reaction readback
   - Restores read-path trust
   - No dependency on other fixes

5. **#36** - Same-repo PR update guardrail
   - Wrapper hardening
   - Improves robustness

6. **#35** - Canva asset upload reflection
   - Separate connector
   - Independent fix path

### Dependency Graph

```
#31 ──┐
      ├─→ Review Workflow Complete
#32 ──┘

#33 ──→ Read Path Trust (Reviews)
#34 ──→ Read Path Trust (Reactions)

#36 ──→ Update Robustness

#35 ──→ Canva Asset Tracking
```

---

## End-to-End Validation Sequence

After all GitHub fixes are implemented, run this complete validation flow:

### Prerequisites
- Disposable GitHub repository with issues enabled
- Test GitHub account with PR/issue permissions

### Validation Steps

1. **Create Draft PR**
   ```
   - Create draft PR from feature branch
   - Note PR number
   ```

2. **Ready for Review (#31)**
   ```
   - Execute mark_pull_request_ready_for_review
   - Verify PR transitions to ready state
   - Verify no GraphQL errors
   ```

3. **Submit Comment Review (#32, #33)**
   ```
   - Submit review with state COMMENTED
   - Capture review ID from response
   - Verify ID format is usable
   ```

4. **List Reviews (#33)**
   ```
   - Execute list_pull_request_reviews
   - Verify COMMENTED review appears in list
   - Verify review metadata is complete
   ```

5. **Dismiss Review (#32)**
   ```
   - Execute dismiss_pull_request_review with captured ID
   - Verify dismissal succeeds
   - Note: COMMENTED reviews cannot be dismissed per GitHub rules
   - Use APPROVED review for this test
   ```

6. **PR Reaction (#34 control)**
   ```
   - Add reaction to PR
   - Read PR reactions
   - Verify reaction appears (baseline)
   ```

7. **Issue Comment Reaction (#34)**
   ```
   - Create issue comment
   - Add reaction to comment
   - Execute get_issue_comment_reactions
   - Verify reaction appears
   ```

8. **Same-Repo PR Update (#36)**
   ```
   - Update PR metadata (title, body)
   - Verify update succeeds without 422
   - Verify maintainer_can_modify not sent
   ```

9. **Cross-Repo PR Update (#36 regression)**
   ```
   - Create fork PR (if possible)
   - Update with maintainer_can_modify
   - Verify behavior intact for forks
   ```

### Success Criteria

All validation steps complete without:
- GraphQL schema errors
- ID format mismatches
- Missing or filtered results
- HTTP 422 errors on same-repo updates
- Regression in existing functionality

---

## Non-Defect Behaviors

The following are **not** connector defects:

### GitHub Product Constraints

1. **COMMENTED Review Dismissal**
   - GitHub does not allow dismissing plain `COMMENTED` reviews
   - Only `APPROVED` and `CHANGES_REQUESTED` can be dismissed
   - This is GitHub product behavior, not a connector bug

2. **Repository Rules**
   - Force-push protection
   - Branch protection rules
   - Required status checks
   - These may block cleanup but are not connector issues

3. **Rate Limiting**
   - GitHub API rate limits are enforced by GitHub
   - Connectors should handle gracefully but cannot bypass

### Canva Product Constraints

1. **Brand Template Generation**
   - Excluded from bug tracking
   - Gated by Canva Enterprise plan tier
   - Not a connector failure

---

## Testing Artifacts

### Cleanup Artifacts

Some GitHub resources may remain after testing due to product constraints:
- Draft PRs (can be closed but not deleted via API)
- Review comments (permanent unless PR is deleted)
- Reactions (can be removed)
- Issue comments (can be deleted)

### Test Resource Naming

Use clear prefixes for disposable resources:
- Branch: `test/mcp-connector-validation-YYYYMMDD`
- PR title: `[TEST] MCP Connector Validation - DO NOT MERGE`
- Issue title: `[TEST] MCP Connector Validation`

---

## Implementation Checklist

### GitHub MCP Connector

- [ ] Fix #31: Update GraphQL mutation for ready-for-review
- [ ] Fix #32: Return compatible review IDs
- [x] Fix #33: Include COMMENTED reviews in listing
- [x] Fix #32: Return compatible review IDs
- [ ] Fix #33: Include COMMENTED reviews in listing
- [ ] Fix #34: Fix issue-comment reaction readback
- [x] Fix #36: Suppress maintainer_can_modify for same-repo PRs
- [x] Fix #34: Fix issue-comment reaction readback
- [ ] Fix #36: Suppress maintainer_can_modify for same-repo PRs
- [ ] Run end-to-end validation sequence
- [ ] Document any API quirks or edge cases found
- [ ] Update connector documentation

### Canva MCP Connector

- [x] Fix #35: Investigate and fix asset upload reflection
- [x] Document eventual consistency behavior if applicable
- [x] Add retry logic if needed
- [x] Validate folder scoping
- [x] Update connector documentation

### Final Verification

- [ ] All child issues resolved
- [ ] Validation sequence passes
- [ ] No regression in existing functionality
- [ ] Documentation updated
- [ ] Test cleanup completed

---

## Metadata

**Created:** 2026-04-22
**Smoke Test Date:** 2026-04-16 to 2026-04-17
**Connector Versions:** Live production connectors as of April 2026
**Test Environment:** Live GitHub and Canva APIs (not mocked)

**Maintainer:** 0ai-Cyberviser
**Repository:** 0ai-Cyberviser/0ai-assurance-network

**Related Documentation:**
- GitHub GraphQL API: https://docs.github.com/en/graphql
- GitHub REST API: https://docs.github.com/en/rest
- Canva Connect API: (Canva documentation links)

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2026-04-22 | 1.0 | Initial defect tracking document created |
