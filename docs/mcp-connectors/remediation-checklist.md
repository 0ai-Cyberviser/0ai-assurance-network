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

**File:** (GitHub MCP connector reaction handlers)

**Change:**
```python
# Compare with working PR reaction path
# Fix endpoint, pagination, or normalization
# Ensure issue-comment reactions use correct API endpoint
```

**Validate:**
```bash
# Create issue comment
# Add reaction
# Read reactions
# Verify reaction appears
```

---

### ✅ Fix #36: Same-Repo PR Update Guard

**File:** `src/github_mcp/pull_requests.py`

**Change:**
```diff
def build_update_payload(pr, updates):
  payload = {}

+ # Only include maintainer_can_modify for cross-repo (fork) PRs.
+ # GitHub returns HTTP 422 if this field is sent for same-repo PRs.
  for key, value in updates.items():
+   if key == "maintainer_can_modify":
+     if is_cross_repo_pr(pr):
+       payload[key] = value
+   else:
      payload[key] = value

  return payload
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
| #34 | GitHub MCP | Medium | ⏳ Pending |
| #36 | GitHub MCP | Low | ✅ Fixed |
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
