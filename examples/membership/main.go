// Membership organization — a civic initiative's roster on kivi, and the one
// place the Graph view genuinely earns its keep.
//
// The insurance example deliberately did NOT use the Graph view: a policy's
// relationships (which vehicles, which coverages) have a LIFECYCLE — they get
// added and dropped — and the Graph view has no delete semantics, so those
// belong in events. This example is the other half of that lesson. A
// membership organization has a DURABLE structural skeleton — a branch
// hierarchy (Kadıköy is part of İstanbul is part of the national body), the
// roles defined at each level — that essentially never gets rewritten. THAT
// is what the Graph view is for.
//
// So this demo splits the two on purpose:
//
//   - the org's fixed structure  → `relation` events → the Graph view;
//
//   - a member's mutable profile → rich `member.joined` events + `property`
//     shadows → QuerySubject, with time travel for "who was where WHEN".
//
//     go run ./examples/membership <addr> [token]
package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"kivi/client"
)

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "membership:", err)
		os.Exit(1)
	}
}

func must[T any](v T, err error) T { fail(err); return v }

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: membership <addr> [token]")
		os.Exit(2)
	}
	token := ""
	if len(os.Args) > 2 {
		token = os.Args[2]
	}
	ctx := context.Background()
	c := must(client.Dial(os.Args[1], client.Options{Token: token})) // verify ON by default
	defer c.Close()

	// --- the DURABLE structure: relation events → the Graph view. This skeleton
	//     is set up once and essentially never rewritten, which is exactly the
	//     shape the Graph view fits (no edge lifecycle, no per-edge properties).
	fmt.Println("== build the org's structural skeleton (relation events → the Graph view)")
	rel := func(source, relation, target string) {
		must(c.Append(ctx, "relation", map[string]any{
			"source": source, "relation": relation, "target": target}))
	}
	rel("branch:kadikoy", "part_of", "branch:istanbul")
	rel("branch:uskudar", "part_of", "branch:istanbul")
	rel("branch:istanbul", "part_of", "branch:hq")
	rel("role:coordinator", "defined_at", "branch") // a role that exists at branch level
	rel("role:delegate", "defined_at", "branch")

	// --- a member's MUTABLE profile: a rich joined event (the audit-grade
	//     document) PLUS thin property shadows for an instant, traced whole-row
	//     read. NOTE: kivi does not enforce a uniqueness constraint on
	//     tc_kimlik — that is the APP layer's job (check-before-append, or a
	//     periodic audit); kivi guarantees the immutable, traced HISTORY, which
	//     is what a "why was this membership approved?" question actually needs.
	fmt.Println("== a member joins — a rich event, plus property shadows for QuerySubject")
	must(c.Append(ctx, "member.joined", map[string]any{
		"member_id": "M-1001", "name": "Deniz Aydın", "tc_kimlik_sha256": "a1b2c3…",
		"address":     map[string]any{"il": "İstanbul", "ilce": "Kadıköy", "mahalle": "Caferağa"},
		"joined_date": "2026-02-01",
	}))
	setProp := func(subject, attr string, value any) client.Receipt {
		return must(c.Append(ctx, "property", map[string]any{
			"subject": subject, "attribute": attr, "value": value}))
	}
	setProp("M-1001", "status", "active")
	setProp("M-1001", "branch", "branch:kadikoy")
	setProp("M-1001", "role", "role:delegate")

	fmt.Println("\n== QuerySubject: this member's full CURRENT profile, one traced call")
	prof := must(c.Subject(ctx, "M-1001"))
	fmt.Printf("   %v  (established by %d events, scope 0..%d)\n", prof.Value, len(prof.Trace), prof.Scope)

	// --- a branch transfer is NOT an overwrite — it is a new event. The old
	//     affiliation stays in history, which is the whole point when a vote's
	//     validity depends on who was a member of which branch AT THE TIME.
	fmt.Println("\n== the member transfers to another branch — a NEW event, not an overwrite")
	hNo, _ := must2(c.Head(ctx)) // the head record number just BEFORE the transfer
	setProp("M-1001", "branch", "branch:uskudar")

	fmt.Println("== the question that decides a vote's validity: which branch on a past date?")
	nowBranch := must(c.Subject(ctx, "M-1001"))
	fmt.Printf("   now:            branch = %v\n", asMap(nowBranch.Value)["branch"])
	asOf := must(c.SubjectAt(ctx, "M-1001", hNo)) // as of just before the transfer
	fmt.Printf("   as of record %d: branch = %v  (the affiliation that was valid then — proven)\n",
		hNo, asMap(asOf.Value)["branch"])

	// --- a proposal escalates up the branch hierarchy: domain events again,
	//     while the Graph view answers "what IS the hierarchy it climbs?"
	fmt.Println("\n== a local proposal gathers support and ESCALATES up the hierarchy (domain events)")
	must(c.Append(ctx, "proposal.raised", map[string]any{
		"proposal_id": "P-42", "branch": "branch:kadikoy", "title": "car-free Sunday on the waterfront"}))
	must(c.Append(ctx, "proposal.escalated", map[string]any{
		"proposal_id": "P-42", "from": "branch:kadikoy", "to": "branch:istanbul",
		"reason": "passed the local support threshold"}))

	fmt.Println("\n== the Graph view: the org's whole structural skeleton, digest-stamped")
	g := must(c.Graph(ctx))
	edges := make([]string, 0)
	for edge := range asMap(g.State) {
		edges = append(edges, edge)
	}
	sort.Strings(edges)
	for _, e := range edges {
		fmt.Printf("   %s\n", e)
	}
	fmt.Printf("   (scope 0..%d, digest %s…)\n", g.Scope, short(g.Hash))

	fmt.Println("\n== the split, restated")
	fmt.Println("   Graph  → the durable skeleton (branch hierarchy, role definitions)")
	fmt.Println("   events → everything with a lifecycle (a member's branch, a proposal's climb)")
	fmt.Println("   QuerySubject + time travel → \"who was where WHEN\", the membership analog of")
	fmt.Println("   the insurance coverage dispute.")

	fmt.Println("\nMEMBERSHIP DEMO OK")
}

// -- tiny helpers ------------------------------------------------------------

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func short(s string) string {
	if len(s) > 16 {
		return s[:16]
	}
	return s
}

// must2 unwraps Head's (int64, string, error) into (int64, string).
func must2(no int64, hash string, err error) (int64, string) { fail(err); return no, hash }
