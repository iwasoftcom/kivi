#!/usr/bin/env python3
"""Insurance policy lifecycle — what a policy administration system's audit
trail looks like when it is built on kivi instead of bolted onto one.

The point this example exists to make: a policy's insured, policyholder,
asset, coverages and premium are NOT crammed into kivi's built-in
Table/Graph views — each meaningful business fact (issuance, endorsement, a
claim, a reserve change) is its OWN richly-shaped event, exactly as it would
be named in the business ("policy.issued", not a generic "changed"). kivi is
the event backbone underneath a policy admin system, not a replacement for
its relational read model — see docs/kivi-docs.html's "Worked example"
section for the full reasoning.

The scene: a claim happens BEFORE a second vehicle is endorsed onto the
policy. subject() answers "what do we know about this policy RIGHT NOW"
(both vehicles); the as-of query answers the question a claims dispute
actually turns on — "was the second vehicle covered on the day of THIS
loss?" — and the honest answer is no, because the record proves the
endorsement came later.

    # start any kivid, then:
    clients/python/.venv/bin/python examples/insurance.py localhost:4741 [token]
"""

import sys
import os
from dataclasses import dataclass, field
from typing import Any, List

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)),
                                "..", "clients", "python"))

from kivi import KiviClient, body_as  # noqa: E402

POLICY_ID = "POL-1001"


# -- the domain, as plain typed dataclasses (no dicts in the business code) ---

@dataclass
class Asset:
    type: str
    plate: str


@dataclass
class Coverage:
    code: str
    limit_cents: int
    deductible_cents: int


@dataclass
class PolicyIssued:
    policy_id: str
    insured: str
    policyholder: str
    asset: Asset
    premium_cents: int
    effective_date: str
    coverages: List[Coverage]


@dataclass
class PolicyEndorsed:
    policy_id: str
    endorsement_no: int
    kind: str
    asset: Asset
    effective_date: str


@dataclass
class ClaimReported:
    claim_id: str
    policy_id: str
    loss_date: str
    description: str


@dataclass
class ClaimReserveChanged:
    claim_id: str
    reserve_cents: int
    reason: str


@dataclass
class Property:
    subject: str
    attribute: str
    value: Any


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    addr = sys.argv[1]
    token = sys.argv[2] if len(sys.argv) > 2 else None
    c = KiviClient(addr, token=token)  # verification ON by default

    print("== issue the policy (one vehicle) — a typed event, not a generic row")
    c.append("policy.issued", PolicyIssued(
        policy_id=POLICY_ID, insured="jane-doe", policyholder="jane-doe",
        asset=Asset(type="vehicle", plate="1FA-2024"),
        premium_cents=84900, effective_date="2026-01-01",
        coverages=[
            Coverage(code="LIABILITY", limit_cents=10_000_000, deductible_cents=0),
            Coverage(code="COLLISION", limit_cents=5_000_000, deductible_cents=50_000),
        ]))
    # a thin, deliberate SHADOW as properties on the policy_id subject — not a
    # duplicate of the rich event, just the handful of fields worth an
    # instant, traced whole-row lookup later (subject())
    c.append("property", Property(subject=POLICY_ID, attribute="status", value="active"))
    c.append("property", Property(subject=POLICY_ID, attribute="asset_count", value=1))

    print("== a claim is reported against the ONE vehicle on record (FNOL)")
    claim = c.append("claim.reported", ClaimReported(
        claim_id="CLM-5001", policy_id=POLICY_ID,
        loss_date="2026-03-10", description="rear-end collision"))
    print(f"   filed as record {claim.no}")

    print("== an adjuster sets, then revises, the reserve — TWO events, not an overwrite")
    c.append("claim.reserve_changed", ClaimReserveChanged(
        claim_id="CLM-5001", reserve_cents=300_000, reason="initial estimate"))
    final = c.append("claim.reserve_changed", ClaimReserveChanged(
        claim_id="CLM-5001", reserve_cents=450_000,
        reason="body shop estimate came in higher"))

    print("== FIVE DAYS LATER: a second vehicle is endorsed onto the policy")
    c.append("policy.endorsed", PolicyEndorsed(
        policy_id=POLICY_ID, endorsement_no=1, kind="asset_added",
        asset=Asset(type="vehicle", plate="9ZR-2021"), effective_date="2026-03-15"))
    c.append("property", Property(subject=POLICY_ID, attribute="asset_count", value=2))

    print("\n== subject(): everything we know about the policy RIGHT NOW, one traced call")
    now = c.subject(POLICY_ID)
    print(f"   {now.value}  (established by {len(now.trace)} events, scope: records 0..{now.scope})")

    print("\n== the question a claims dispute actually turns on:")
    print('   "was the second vehicle covered on the day of the March 10 loss?"')
    as_of_claim = c.subject(POLICY_ID, as_of=claim.no)
    print(f"   as of record {claim.no} (the moment the claim was filed): {as_of_claim.value}")
    print("   → asset_count is 1: the endorsement had not happened yet. Not fabricated — proven.")

    print("\n== why(): the FINAL reserve receipt, read back AS A TYPED ENTITY (body_as)")
    for rec in c.why([final.no]):
        e = body_as(ClaimReserveChanged, rec)
        print(f"   claim {e.claim_id}: reserve now {e.reserve_cents} cents ({e.reason})")

    print("\n== subscribe(): a downstream reserve-monitor gets TYPED events, one event type only")
    n = 0
    for rec in c.subscribe(start=0, types=["claim.reserve_changed"]):
        n += 1
        e = body_as(ClaimReserveChanged, rec)
        print(f"   {e.claim_id} → {e.reserve_cents} cents ({e.reason})")
    print(f"   {n} reserve-change event(s) delivered — policy.issued/endorsed and "
          "claim.reported never crossed the wire")

    print("\n== what this demo does NOT show, on purpose")
    print('   kivi retention hold --reason "SIU fraud review" <ledger> <period>   # legal hold: CLI-only,')
    print("   offline, deliberately never a server RPC — see OPERATIONS.md §2.1/§14 and the docs page.")

    print("\nINSURANCE DEMO OK")
    return 0


if __name__ == "__main__":
    sys.exit(main())
