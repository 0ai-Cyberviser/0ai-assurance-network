# MCP Connector Remediation Checklist

Quick reference for fixing GitHub and Canva MCP connector defects from April 2026 smoke testing.

**Full Documentation:** [defect-tracking-2026-04.md](./defect-tracking-2026-04.md)

---

## GitHub MCP Fixes

### ✅ Fix #31: Ready-for-Review GraphQL Field

**File:** (GitHub MCP connector GraphQL mutations)

**Change:**
```diff
mutation {
  markPullRequestReadyForReview(input: {...}) {
    pullRequest {
-     htmlUrl
+     url
    }
  }
}
```

**Validate:**
```bash
# Create draft PR and transition to ready
# Verify no GraphQL errors
```

---

### ✅ Fix #32: Review ID Compatibility

**File:** (GitHub MCP connector review handlers)

**Change:**
```diff
# add_review_to_pr response
{
  "review_id": 4125974959,
+ "review_node_id": "PRR_kwDOA..."
}
```

**Validate:**
```bash
# Create review, capture ID
# Dismiss review with returned ID
# Verify success
```

---

### ✅ Fix #33: Include COMMENTED Reviews

**File:** (GitHub MCP connector review listing)

**Change:**
```diff
# Ensure all review states are included
ALLOWED_STATES = [
+ "COMMENTED",
  "APPROVED",
  "CHANGES_REQUESTED",
  "DISMISSED",
  "PENDING"
]
```

**Validate:**
```bash
# Submit COMMENTED review
# List reviews
# Verify COMMENTED review appears
```

---

### ✅ Fix #34: Issue Comment Reaction Readback

**File:** `src/assurancectl/github_mcp.py`

**Root Cause:** The read path was using the issue-level endpoint
`/issues/{number}/reactions` instead of the comment-level endpoint
`/issues/comments/{comment_id}/reactions`.  GitHub silently scopes
the query and returns an empty list, causing the connector to miss
reactions that were just created.

**Change:**
```diff
-# WRONG – returns reactions on the issue, not the comment
-path = f"/repos/{owner}/{repo}/issues/{issue_number}/reactions"

+# CORRECT – returns reactions on the specific comment
+path = f"/repos/{owner}/{repo}/issues/comments/{comment_id}/reactions"
```

A shared `_normalize_reaction` function is used by both the issue-comment
and PR reaction paths so the output shape is guaranteed to be identical.

**Tests:** `tests/test_github_mcp.py`

**Validate:**
```bash
python -m unittest tests/test_github_mcp.py -v
# Expected: 15 tests pass, including:
#   test_calls_issue_comment_reactions_endpoint
#   test_does_not_call_issue_level_endpoint
#   test_immediately_readable_after_creation
#   test_issue_comment_and_pr_reactions_share_normalizer
```

---

### ✅ Fix #36: Same-Repo PR Update Guard

**File:** (GitHub MCP connector PR update)

**Change:**
```diff
def update_pull_request(pr, updates):
  payload = {...}

+ # Only include for cross-repo PRs
+ if pr.head.repo.id != pr.base.repo.id:
    payload["maintainer_can_modify"] = updates.get("maintainer_can_modify")

  return github.patch(f"/pulls/{pr.number}", payload)
```

**Validate:**
```bash
# Update same-repo PR metadata
# Verify no 422 error
# Update cross-repo PR with maintainer_can_modify
# Verify fork behavior intact
```

---

## Canva MCP Fixes

### ✅ Fix #35: Asset Upload Reflection

**File:** (Canva MCP connector asset handlers)

**Investigation:**
- [ ] Verify upload folder matches listing folder
- [ ] Check for eventual consistency window
- [ ] Validate pagination in folder listing
- [ ] Compare upload response with list response

**Change:**
```python
# Add retry logic if eventual consistency
def verify_upload(asset_id, max_retries=3):
  for attempt in range(max_retries):
    assets = list_uploads_folder()
    if asset_id in assets:
      return True
    time.sleep(2 ** attempt)  # Exponential backoff
  return False
```

**Validate:**
```bash
# Upload asset
# List Uploads folder
# Verify asset appears (with retry if needed)
```

---

## End-to-End Validation

### Complete Test Sequence

```bash
# 1. Create draft PR
gh pr create --draft --title "[TEST] MCP Validation" --body "Test PR"

# 2. Mark ready for review (#31)
# <use connector tool>

# 3. Submit COMMENTED review (#33)
# <use connector tool>

# 4. List reviews - verify COMMENTED appears (#33)
# <use connector tool>

# 5. Submit APPROVED review (#32)
# <use connector tool>

# 6. Dismiss review with returned ID (#32)
# <use connector tool>

# 7. Add issue comment reaction (#34)
# <use connector tool>

# 8. Read issue comment reactions (#34)
# <use connector tool>

# 9. Update PR metadata (#36)
# <use connector tool>

# 10. Cleanup
gh pr close <pr-number>
```

### Success Criteria

- [ ] No GraphQL schema errors
- [ ] Review IDs are compatible between create/dismiss
- [ ] COMMENTED reviews appear in listings
- [ ] Issue comment reactions are readable
- [ ] Same-repo PR updates succeed
- [ ] No regression in existing features

---

## Quick Reference

| Issue | Component | Severity | Status |
|-------|-----------|----------|--------|
| #31 | GitHub MCP | High | ⏳ Pending |
| #32 | GitHub MCP | High | ⏳ Pending |
| #33 | GitHub MCP | Medium | ⏳ Pending |
| #34 | GitHub MCP | Medium | ✅ Fixed |
| #36 | GitHub MCP | Low | ⏳ Pending |
| #35 | Canva MCP | Medium | ⏳ Pending |

**Legend:**
- ⏳ Pending
- 🔧 In Progress
- ✅ Fixed
- ✓ Validated

---

## Notes

### Non-Bugs
- ❌ Cannot dismiss `COMMENTED` reviews (GitHub product rule)
- ❌ Repository protection rules may block cleanup
- ❌ Canva brand templates require Enterprise plan

### Testing
- Use disposable test resources
- Prefix branches with `test/mcp-validation-`
- Mark PRs with `[TEST] DO NOT MERGE`
- Clean up after validation

---

**Last Updated:** 2026-04-22
**Related Issues:** #30, #31, #32, #33, #34, #35, #36
