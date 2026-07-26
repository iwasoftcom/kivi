#!/usr/bin/env bash
# Legal hold vs. routine retention — the compliance question a plain database
# cannot answer honestly, run entirely through the kivi CLI.
#
# The scene continues the insurance story: claims are archived as they age
# (kivi rotate moves a closed period to the COLD tier), and a routine
# retention policy eventually retires old archives to reclaim storage. But
# ONE claim — CLM-5001 — is under SIU fraud investigation and a litigation
# hold: its archive must NOT be deleted, by anyone, automatic or manual,
# until the hold is lifted. And every one of these steps has to leave a
# receipt, because "why was this deleted?" is a question a regulator will
# ask years later.
#
# Everything here is CLI-only and OFFLINE by design (the writer lock): kivi's
# core never deletes on its own, and there is no scheduler inside kivid that
# could fire a retention sweep with no human at the console. An external cron
# calling `kivi retention sweep` is the supported way to automate the daily
# decision — the deletion mechanism itself stays a deliberate, --yes-confirmed
# human action.
#
#   examples/retention.sh
set -euo pipefail

cd "$(dirname "$0")/.."
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

KV="$TMP/kivi"
echo "== building the kivi CLI (one static binary)…"
go build -o "$KV" ./cmd/kivi
L="$TMP/claims.ledger"

echo
echo "== two claims are handled, each closed into its own archived period"
"$KV" append "$L" claim.reported '{"claim_id":"CLM-5001","note":"flagged: possible fraud"}' >/dev/null
"$KV" rotate "$L" >/dev/null   # period 1 archived (holds CLM-5001)
"$KV" append "$L" claim.reported '{"claim_id":"CLM-7002","note":"routine auto glass"}' >/dev/null
"$KV" rotate "$L" >/dev/null   # period 2 archived (holds CLM-7002)
echo "   period 1 = the fraud-flagged claim, period 2 = the routine claim"

echo
echo "== retention status under a 7-year policy — nothing eligible yet (archives are seconds old)"
"$KV" retention status --days 2555 "$L"

echo
echo "== CLM-5001 is placed under a litigation hold — receipted, like every fact in kivi"
"$KV" retention hold --reason "SIU fraud review — CLM-5001" "$L" 1

echo
echo "== status again: period 1 is now HELD (and never eligible, regardless of policy age)"
"$KV" retention status --days 2555 "$L"

echo
echo "== a retention run tries to retire the HELD period → REFUSED (the gate holds for EVERY caller)"
if "$KV" retire --reason "routine 7-year retention" --yes "$L" 1; then
  echo "   !! unexpected: a held period was retired"; exit 1
fi
echo "   ↑ exactly right — the hold blocks a manual retire, and it blocks the automatic sweep too"

echo
echo "== the routine, unheld period 2 CAN be retired — receipted BEFORE the file is deleted"
"$KV" retire --reason "routine 7-year retention (no hold)" --yes "$L" 2

echo
echo "== verify --retention: the deletion is a DOCUMENTED absence, not tampering — ledger still ok"
"$KV" verify --retention "$L"

echo
echo "== the investigation closes with no fraud found → the hold is LIFTED (also receipted)"
"$KV" retention release --reason "SIU investigation closed, no fraud found" "$L" 1

echo
echo "== now period 1 may finally be retired, like any other archive"
"$KV" retire --reason "routine 7-year retention (hold lifted)" --yes "$L" 1

echo
echo "== the automation layer, in one line (not run here — these archives are seconds old):"
echo "   kivi retention sweep --days 2555 --reason \"7-year policy\" --yes $L"
echo "   → retires EVERY eligible, non-held period at once, through the same receipted-"
echo "     before-deleted door. Held periods are skipped; each retirement leaves a receipt."

echo
echo "RETENTION DEMO OK"
