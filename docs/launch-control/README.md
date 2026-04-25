# Launch Control Center Pack

This folder provides an operations template pack for coordinating pre-launch
readiness drills and controlled launch simulations.

## Scope alignment with this repository

The root repository is explicitly pre-launch and does **not** authorize a live
public token sale. This pack is for tabletop and operational preparedness
workflows so teams can rehearse signer operations, incident handling, and
parameter-control hygiene.

## Contents

- `launch-control-center.md` — single control document for go/no-go, multisig
  sequencing, timeline, KPIs, and incident response.
- `manifest.json` — machine-readable index of included artifacts and owners.
- `templates/README.md` — column-level field reference for each CSV template.
- `templates/allowlist.csv` — allowlist import template.
- `templates/vesting_beneficiaries.csv` — vesting schedule import template.
- `templates/support_macros.csv` — support response macro template.
- `templates/incident_log.csv` — launch incident tracking template.

## Suggested usage

1. Copy this folder into your launch workspace.
2. Replace sample rows with organization-specific values.
3. Fill addresses, UTC timestamps, and signer details.
4. Lock parameters before final signer review.
5. Use `launch-control-center.md` as the war-room source of truth.

## Data hygiene checklist

- Keep all times in UTC (`YYYY-MM-DDTHH:MM:SSZ`).
- Keep contract addresses checksummed.
- Keep signer and owner IDs consistent across files.
- Keep an immutable copy of final CSVs referenced in approval records.
