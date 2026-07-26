// Insurance policy lifecycle — what a policy administration system's audit
// trail looks like when it is built on kivi instead of bolted onto one.
//
// The point this example exists to make: a policy's insured, policyholder,
// asset, coverages and premium are NOT crammed into kivi's built-in
// Table/Graph views — each meaningful business fact (issuance, endorsement,
// a claim, a reserve change) is its OWN richly-shaped event, exactly as it
// would be named in the business ("policy.issued", not a generic "changed").
// kivi is the event backbone underneath a policy admin system, not a
// replacement for its relational read model — see the docs section this
// example is paired with for the full reasoning.
//
// TYPED events, no raw JSON in sight: each event is a plain struct. Append
// takes it directly (kivi serializes and re-canonicalizes server-side, so
// your field order never matters), and BodyAs decodes a record's body back
// INTO a struct — entity in, entity out, JPA-style, with ZERO extra
// dependency beyond the standard library's encoding/json.
//
// The scene: a claim happens BEFORE a second vehicle is endorsed onto the
// policy. QuerySubject answers "what do we know about this policy RIGHT
// NOW" (both vehicles); the as-of query answers the question a claims
// dispute actually turns on — "was the second vehicle covered on the day
// of THIS loss?" — and the honest answer is no, because the record proves
// the endorsement came later.
//
//	go run ./examples/insurance <addr> [token]
package main

import (
	"context"
	"fmt"
	"os"

	"kivi/client"
	"kivi/ledger"
)

// -- the domain, as plain typed entities (no JSON in the business code) ------

type Asset struct {
	Type  string `json:"type"`
	Plate string `json:"plate"`
}

type Coverage struct {
	Code            string `json:"code"`
	LimitCents      int64  `json:"limit_cents"`
	DeductibleCents int64  `json:"deductible_cents"`
}

type PolicyIssued struct {
	PolicyID      string     `json:"policy_id"`
	Insured       string     `json:"insured"`
	Policyholder  string     `json:"policyholder"`
	Asset         Asset      `json:"asset"`
	PremiumCents  int64      `json:"premium_cents"`
	EffectiveDate string     `json:"effective_date"`
	Coverages     []Coverage `json:"coverages"`
}

type PolicyEndorsed struct {
	PolicyID      string `json:"policy_id"`
	EndorsementNo int    `json:"endorsement_no"`
	Kind          string `json:"kind"`
	Asset         Asset  `json:"asset"`
	EffectiveDate string `json:"effective_date"`
}

type ClaimReported struct {
	ClaimID     string `json:"claim_id"`
	PolicyID    string `json:"policy_id"`
	LossDate    string `json:"loss_date"`
	Description string `json:"description"`
}

type ClaimReserveChanged struct {
	ClaimID      string `json:"claim_id"`
	ReserveCents int64  `json:"reserve_cents"`
	Reason       string `json:"reason"`
}

// Property is the thin, deliberate SHADOW written on the policy_id subject —
// not a duplicate of the rich event, just the handful of fields worth an
// instant, traced whole-row lookup later (QuerySubject).
type Property struct {
	Subject   string `json:"subject"`
	Attribute string `json:"attribute"`
	Value     any    `json:"value"`
}

func must[T any](v T, err error) T {
	if err != nil {
		fmt.Fprintln(os.Stderr, "insurance:", err)
		os.Exit(1)
	}
	return v
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: insurance <addr> [token]")
		os.Exit(2)
	}
	token := ""
	if len(os.Args) > 2 {
		token = os.Args[2]
	}
	ctx := context.Background()
	c := must(client.Dial(os.Args[1], client.Options{Token: token})) // verify ON by default
	defer c.Close()

	const policyID = "POL-1001"

	fmt.Println("== issue the policy (one vehicle) — a typed event, not a generic row")
	must(c.Append(ctx, "policy.issued", PolicyIssued{
		PolicyID: policyID, Insured: "jane-doe", Policyholder: "jane-doe",
		Asset:        Asset{Type: "vehicle", Plate: "1FA-2024"},
		PremiumCents: 84900, EffectiveDate: "2026-01-01",
		Coverages: []Coverage{
			{Code: "LIABILITY", LimitCents: 10_000_000, DeductibleCents: 0},
			{Code: "COLLISION", LimitCents: 5_000_000, DeductibleCents: 50_000},
		},
	}))
	must(c.Append(ctx, "property", Property{Subject: policyID, Attribute: "status", Value: "active"}))
	must(c.Append(ctx, "property", Property{Subject: policyID, Attribute: "asset_count", Value: 1}))

	fmt.Println("== a claim is reported against the ONE vehicle on record (FNOL)")
	claim := must(c.Append(ctx, "claim.reported", ClaimReported{
		ClaimID: "CLM-5001", PolicyID: policyID,
		LossDate: "2026-03-10", Description: "rear-end collision",
	}))
	fmt.Printf("   filed as record %d\n", claim.No)

	fmt.Println("== an adjuster sets, then revises, the reserve — TWO events, not an overwrite")
	must(c.Append(ctx, "claim.reserve_changed", ClaimReserveChanged{
		ClaimID: "CLM-5001", ReserveCents: 300_000, Reason: "initial estimate"}))
	reserveFinal := must(c.Append(ctx, "claim.reserve_changed", ClaimReserveChanged{
		ClaimID: "CLM-5001", ReserveCents: 450_000, Reason: "body shop estimate came in higher"}))

	fmt.Println("== FIVE DAYS LATER: a second vehicle is endorsed onto the policy")
	must(c.Append(ctx, "policy.endorsed", PolicyEndorsed{
		PolicyID: policyID, EndorsementNo: 1, Kind: "asset_added",
		Asset: Asset{Type: "vehicle", Plate: "9ZR-2021"}, EffectiveDate: "2026-03-15",
	}))
	must(c.Append(ctx, "property", Property{Subject: policyID, Attribute: "asset_count", Value: 2}))

	fmt.Println("\n== QuerySubject: everything we know about the policy RIGHT NOW, one traced call")
	now := must(c.Subject(ctx, policyID))
	fmt.Printf("   %v  (established by %d events, scope: records 0..%d)\n", now.Value, len(now.Trace), now.Scope)

	fmt.Println("\n== the question a claims dispute actually turns on:")
	fmt.Println("   \"was the second vehicle covered on the day of the March 10 loss?\"")
	asOfClaim := must(c.SubjectAt(ctx, policyID, claim.No))
	fmt.Printf("   as of record %d (the moment the claim was filed): %v\n", claim.No, asOfClaim.Value)
	fmt.Println("   → asset_count is 1: the endorsement had not happened yet. Not fabricated — proven.")

	fmt.Println("\n== Why: the FINAL reserve receipt, read back AS A TYPED ENTITY (BodyAs)")
	for _, rec := range must(c.Why(ctx, []int64{reserveFinal.No})) {
		e := must(client.BodyAs[ClaimReserveChanged](rec))
		fmt.Printf("   claim %s: reserve now %d cents (%s)\n", e.ClaimID, e.ReserveCents, e.Reason)
	}

	fmt.Println("\n== Subscribe: a downstream reserve-monitor gets TYPED events, one event type only")
	n := 0
	must(0, c.Subscribe(ctx, 0, false, []string{"claim.reserve_changed"}, func(rec ledger.Record) error {
		n++
		e := must(client.BodyAs[ClaimReserveChanged](rec)) // entity out, in the callback
		fmt.Printf("   %s → %d cents (%s)\n", e.ClaimID, e.ReserveCents, e.Reason)
		return nil
	}))
	fmt.Printf("   %d reserve-change event(s) delivered — policy.issued/endorsed and claim.reported never crossed the wire\n", n)

	fmt.Println("\n== what this demo does NOT show, on purpose")
	fmt.Println("   kivi retention hold --reason \"SIU fraud review\" <ledger> <period>   # legal hold: CLI-only,")
	fmt.Println("   offline, deliberately never a server RPC — see OPERATIONS.md §2.1/§14 and the docs page.")

	fmt.Println("\nINSURANCE DEMO OK")
}
