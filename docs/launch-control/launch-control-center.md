# Launch Control Center — Token + NFT Drop

## Command Header

- Project:
- Network / Chain ID:
- Launch date/time (UTC):
- Release manager:
- War room channel:
- Status page:
- Multisig:
- Deployer:
- Explorer links:

---

## 1) Go / No-Go Summary

| Domain | Owner | Status (Green/Yellow/Red) | Notes |
|---|---|---|---|
| Security audit remediations |  |  |  |
| Testnet dress rehearsal |  |  |  |
| Contract verification workflow |  |  |  |
| Legal/terms/disclosures |  |  |  |
| Treasury + vesting readiness |  |  |  |
| Monitoring + alerts |  |  |  |
| Support + comms readiness |  |  |  |

**Decision**
- [ ] GO
- [ ] NO-GO
- Timestamp (UTC):
- Approver(s):

---

## 2) Contract Registry

| Component | Address | Verified? | Owner/Admin | Notes |
|---|---|---|---|---|
| ERC-20 Token |  | [ ] |  |  |
| ERC-721/1155 NFT |  | [ ] |  |  |
| Sale Manager |  | [ ] |  |  |
| Vesting |  | [ ] |  |  |
| Treasury Wallet |  | n/a |  |  |
| Royalty Recipient |  | n/a |  |  |

---

## 3) Parameter Snapshot

| Parameter | Value | Source of Truth |
|---|---|---|
| AL start/end UTC |  | Params Sheet |
| Public start UTC |  | Params Sheet |
| Late mint start/end UTC |  | Params Sheet |
| AL price |  | Params Sheet |
| Public price |  | Params Sheet |
| AL wallet cap |  | Params Sheet |
| Public wallet cap |  | Params Sheet |
| AL phase cap |  | Params Sheet |
| Public phase cap |  | Params Sheet |
| Merkle root |  | Build artifact |
| Provenance hash |  | Metadata artifact |
| Royalty bps/recipient |  | Params Sheet |

---

## 4) Transaction Execution Table (Multisig)

| Seq | Contract | Function | Args (short) | Multisig Link | On-chain Tx | Signers | Status |
|---|---|---|---|---|---|---|---|
| 1 | Token | transferOwnership | SAFE |  |  |  |  |
| 2 | NFT | transferOwnership | SAFE |  |  |  |  |
| 3 | SaleManager | transferOwnership | SAFE |  |  |  |  |
| 4 | SaleManager/NFT | grantRole | MINTER, SaleManager |  |  |  |  |
| 5 | All | revokeRole | deployer roles |  |  |  |  |
| 6 | SaleManager | setMerkleRoot | root |  |  |  |  |
| 7 | SaleManager | setPrices | AL, Public, Late |  |  |  |  |
| 8 | SaleManager | setWalletCaps | AL, Public |  |  |  |  |
| 9 | SaleManager | setPhaseWindows | timestamps |  |  |  |  |
|10 | NFT | setRoyaltyInfo | recipient,bps |  |  |  |  |
|11 | NFT | setProvenanceHash | hash |  |  |  |  |
|12 | NFT | setBaseURI | pre-reveal URI |  |  |  |  |
|13 | SaleManager | setSaleState | ALLOWLIST_OPEN |  |  |  |  |
|14 | SaleManager | setSaleState | PUBLIC_OPEN |  |  |  |  |
|15 | SaleManager | setSaleState | CLOSED |  |  |  |  |
|16 | NFT | setBaseURI/freeze | reveal URI + freeze |  |  |  |  |

---

## 5) Timeline (UTC)

| Time | Action | Owner | Success Metric | Fallback |
|---|---|---|---|---|
| T-120 | War room open, checks |  | Team online | Delay launch |
| T-60 | Final simulation + nonce check |  | Sim pass | Hold tx |
| T-30 | Final parameter readback |  | Exact match | Correct params |
| T-10 | Frontend freeze |  | Correct env | Rollback FE |
| T+0 | Allowlist open |  | Mint success starts | Pause + announce |
| T+75 | Public pre-check |  | Stable metrics | Delay public |
| T+135 | Public open |  | Success rate acceptable | Pause/delay |
| T+240 | Mid-run review |  | Error rates normal | Contingency |
| End day | Publish summary |  | Summary posted | Draft fallback |

---

## 6) Live KPI Panel

- Mint success rate (%):
- Revert rate (%):
- Median confirmation time:
- RPC error rate (%):
- Unique minters:
- Total minted:
- Gross proceeds:
- Support queue volume:
- Major incident open? (Y/N)

**Thresholds**
- Revert > 25% for 10 min => consider pause.
- RPC errors > 5% for 5 min => delay next phase.
- Incident comms SLA: first update in <= 10 min.

---

## 7) Incident Quick Actions

1. Pause affected contract(s).
2. Post official incident notice.
3. Triage root cause (RPC/contract/frontend/bot).
4. Execute remediation transaction set.
5. Resume after Eng + Sec sign-off.
6. Publish closure report + corrective actions.

---

## 8) Post-Launch (T+1 to T+7)

- [ ] Supply and treasury reconciliation complete.
- [ ] Incident report published (if needed).
- [ ] Postmortem held with owned action items.
- [ ] Governance / next milestone update published.
