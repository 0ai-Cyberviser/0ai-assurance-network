# Launch Control Center Pack

This folder provides a ready-to-use operations package for a token + NFT launch.

## Contents

- `launch-control-center.md` — single control document (go/no-go, tx table, timeline, KPIs, incident response).
- `templates/allowlist.csv` — allowlist import template.
- `templates/vesting_beneficiaries.csv` — vesting schedule import template.
- `templates/support_macros.csv` — support response macro template.
- `templates/incident_log.csv` — launch incident tracking template.

## Suggested usage

1. Copy this folder into your launch workspace.
2. Fill addresses, timestamps (UTC), and signer details.
3. Lock parameters before final signer review.
4. Use `launch-control-center.md` as the war-room source of truth.

## Notes

- Keep all times in UTC (`YYYY-MM-DDTHH:MM:SSZ`).
- Keep contract addresses checksummed.
- Treat this pack as an operational template; adapt to your legal and security requirements.
