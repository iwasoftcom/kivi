// The mind journal (PLAN3 G5) — what it looks like when an honest mind keeps
// its memory on kivi, in kivi's own language.
//
// Perceptions, feelings and values become EVENTS; the mind's current state is
// a COMPILED VIEW; and the question a mere log can never answer — "why do I
// hold this value? SHOW me" — gets an answer made of receipts, every one of
// them re-verified CLIENT-SIDE. Identity = replayable history. As the finale,
// the Go-only audit: the client folds the verified stream itself and checks
// that the server's compiled answer digest matches — the mind can audit its
// own librarian.
//
//	go run ./examples/mindjournal localhost:4741 [token]
package main

import (
	"context"
	"fmt"
	"os"

	"kivi/client"
	"kivi/ledger"
	"kivi/views"
)

func must[T any](v T, err error) T {
	if err != nil {
		fmt.Fprintln(os.Stderr, "mindjournal:", err)
		os.Exit(1)
	}
	return v
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: mindjournal <addr> [token]")
		os.Exit(2)
	}
	token := ""
	if len(os.Args) > 2 {
		token = os.Args[2]
	}
	ctx := context.Background()
	c := must(client.Dial(os.Args[1], client.Options{Token: token})) // verify ON by default
	defer c.Close()

	fmt.Println("== living (append-only, hash-chained, receipted)")
	must(c.Append(ctx, "perception", map[string]any{"sense": "hearing", "event": "noise", "intensity": 8}))
	must(c.Append(ctx, "his", map[string]any{"ad": "tedirginlik", "source": "noise", "yön": -1}))
	must(c.Append(ctx, "perception", map[string]any{"sense": "hearing", "event": "silence", "intensity": 1}))
	must(c.Append(ctx, "his", map[string]any{"ad": "huzur", "source": "silence", "yön": 1}))
	r := must(c.Append(ctx, "property", map[string]any{
		"subject": "mind", "attribute": "value", "value": "I value silence"}))
	must(c.Append(ctx, "relation", map[string]any{
		"source": "silence", "relation": "doğurdu", "target": "huzur"}))
	fmt.Printf("   value formed at record %d (hash %s…)\n", r.No, r.Hash[:12])

	fmt.Println("== current value (a compiled view — disposable, reborn on demand)")
	a := must(c.Table(ctx, "mind", "value"))
	fmt.Printf("   %v  (scope: records 0..%d)\n", a.Value, a.Scope)

	fmt.Println("== why do I hold this value? (the receipts themselves)")
	for _, rec := range must(c.Why(ctx, a.Trace)) {
		no, _ := ledger.RecordNo(rec)
		fmt.Printf("   record %d: %v\n", no, rec["body"])
	}

	n := 0
	must(0, c.Replay(ctx, 0, false, func(ledger.Record) error { n++; return nil }))
	fmt.Printf("== identity = replayable history: %d records re-verified client-side\n", n)

	match, local, server, err := c.VerifyView(ctx, views.Graph{})
	if err != nil || !match {
		fmt.Fprintf(os.Stderr, "mindjournal: librarian audit failed: %v (%s vs %s)\n", err, local, server)
		os.Exit(1)
	}
	fmt.Printf("== the mind audits its librarian: local fold %s… == server %s… ✓\n",
		local[:12], server[:12])
	fmt.Println("MIND JOURNAL OK")
}
