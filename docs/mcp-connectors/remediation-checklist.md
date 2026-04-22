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

**File:** `src/assurancectl/canva_connector.py`

**Root Cause:** Canva backend eventual consistency — uploaded assets may not
appear in the Uploads folder listing immediately even when the upload job
has returned success.

**Investigation outcome:**
- [x] Upload folder (`uploads`) matches the listing scope
- [x] Eventual consistency window confirmed as root cause
- [x] Pagination in folder listing resolved via `list_all_folder_items`
- [x] Upload response and list response compared — asset IDs are compatible

**Change:**
```python
# verify_upload: retry a fully paginated folder listing with exponential backoff
def verify_upload(asset_id, *, client, folder_id=UPLOADS_FOLDER_ID,
                  item_types=None, max_retries=3, base_delay=2.0,
                  _sleep=time.sleep):
    for attempt in range(max_retries):
        items = list_all_folder_items(
            folder_id, client=client, item_types=item_types
        )
        if any(item.get("asset_id") == asset_id for item in items):
            return True
        if attempt < max_retries - 1:
            _sleep(base_delay * (2 ** attempt))
    return False

# list_all_folder_items: follow pagination tokens to return the full listing
def list_all_folder_items(folder_id, *, client, item_types=None):
    items, continuation_token = [], None
    while True:
        result = client.list_folder_items(
            folder_id, item_types=item_types,
            continuation_token=continuation_token
        )
        items.extend(result.get("items", []))
        continuation_token = result.get("continuation_token")
        if not continuation_token:
            break
    return items
```

**Tests:** `tests/test_canva_connector.py` (13 unit tests)

**Validate:**
```bash
python -m unittest tests/test_canva_connector.py -v
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
| #36 | GitHub MCP | Low | ⏳ Pending |
| #35 | Canva MCP | Medium | ✅ Fixed |

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
