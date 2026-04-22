# MCP Connector Documentation

This directory contains documentation and defect tracking for MCP (Model Context Protocol) connector implementations.

## Contents

- **[defect-tracking-2026-04.md](./defect-tracking-2026-04.md)** - Comprehensive tracking document for defects found during April 2026 live smoke testing of GitHub and Canva MCP connectors
- **[remediation-checklist.md](./remediation-checklist.md)** - Quick reference checklist for implementing fixes

## Overview

MCP connectors provide integration between the 0AI Assurance Network and external platforms:

- **GitHub MCP Connector** - Pull requests, reviews, issues, reactions, and repository operations
- **Canva MCP Connector** - Asset uploads, folder management, and design operations

## Current Status

### April 2026 Smoke Testing Results

Live smoke testing was conducted on April 16-17, 2026 against production GitHub and Canva APIs.

**Defects Found:**

| ID | Platform | Component | Severity | Status |
|----|----------|-----------|----------|--------|
| #31 | GitHub | ready-for-review | High | Tracked |
| #32 | GitHub | review lifecycle | High | Tracked |
| #33 | GitHub | review listing | Medium | Tracked |
| #34 | GitHub | issue reactions | Medium | Tracked |
| #36 | GitHub | PR updates | Low | Tracked |
| #35 | Canva | asset upload | Medium | Tracked |

## Quick Links

**For Implementers:**
- Start with [remediation-checklist.md](./remediation-checklist.md)
- Refer to [defect-tracking-2026-04.md](./defect-tracking-2026-04.md) for detailed context

**For QA/Validation:**
- See [End-to-End Validation Sequence](./defect-tracking-2026-04.md#end-to-end-validation-sequence)
- Review [Non-Defect Behaviors](./defect-tracking-2026-04.md#non-defect-behaviors)

## Fix Priority Order

1. **#31** - GraphQL schema fix for ready-for-review (unblocks workflows)
2. **#32** - Review ID compatibility (enables full review lifecycle)
3. **#33** - Review listing completeness (restores read trust)
4. **#34** - Reaction readback (restores read trust)
5. **#36** - PR update guardrail (hardens robustness)
6. **#35** - Canva asset reflection (independent path)

## Related Issues

- **Umbrella Issue:** 0ai-Cyberviser/0ai-assurance-network#30
- **Child Issues:** #31, #32, #33, #34, #35, #36

---

**Maintainer:** 0ai-Cyberviser
**Last Updated:** 2026-04-22
