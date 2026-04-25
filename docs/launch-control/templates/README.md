# CSV Template Field Guide

This guide defines columns for the Launch Control Center CSV templates.

## `allowlist.csv`

- `wallet_address`: checksummed recipient wallet.
- `max_mint`: wallet-specific mint cap for phase.
- `tier`: segment label used for pricing or eligibility logic.
- `notes`: operator context (partner, raffle, migration, etc.).

## `vesting_beneficiaries.csv`

- `beneficiary_address`: checksummed recipient wallet.
- `allocation_tokens`: integer token amount in whole units.
- `start_utc`: vesting start timestamp (`YYYY-MM-DDTHH:MM:SSZ`).
- `cliff_days`: lock period before first unlock.
- `duration_days`: full vesting duration.
- `slice_days`: release interval granularity.
- `revocable`: `true` or `false`.
- `label`: internal beneficiary identifier.

## `support_macros.csv`

- `macro_id`: stable short identifier.
- `category`: support queue category.
- `title`: macro name shown to operators.
- `message_template`: reusable response text.
- `escalation_path`: owner chain for escalation.

## `incident_log.csv`

- `incident_id`: unique incident handle.
- `opened_utc`: open timestamp in UTC.
- `severity`: `critical`, `high`, `medium`, or `low`.
- `status`: `open`, `monitoring`, or `resolved`.
- `detected_by`: source of detection.
- `component`: affected system area.
- `symptom`: short issue signature.
- `impact`: user or protocol impact summary.
- `mitigation`: immediate action taken.
- `owner`: current incident owner.
- `next_update_utc`: next comms checkpoint.
- `closed_utc`: resolution timestamp.
- `rca_summary`: brief root-cause summary.
- `followup_ticket`: internal issue tracker ID.
