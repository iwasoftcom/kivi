// Reinsurance without a shared database — federation (G10) made concrete.
//
// Two insurers, Acme (the ceding insurer) and Re-Global (the reinsurer),
// share a treaty but NOT a database, and do not trust each other's systems.
// When Acme cedes part of a large claim to Re-Global, both sides need to
// agree on the FACTS — the claim was reported, reserved, ceded, accepted —
// yet neither can be asked to trust the other's word or the other's ledger.
//
// kivi's answer is mutual WITNESSING, the inter-organization promise of a
// blockchain WITHOUT paying consensus on every write: each side fetches a
// FRESH, signed attestation from the other, verifies it OFFLINE (canonical
// hash + Ed25519 — no trust by assertion), and FILES it in its OWN ledger as
// a kivi.witness record. From that moment the other party's chain head is
// pinned inside your history — and if they ever rewrite their past to deny
// what they committed to, their OWN old signed head, sitting in your records,
// testifies against them.
//
// This demo is SELF-CONTAINED: it launches two independent in-process kivid
// servers (two companies, two ledgers, two signing identities, cross-wired
// as each other's witness peer) and runs the whole scene in one command.
//
//	go run ./examples/federation
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"

	"kivi/api"
	"kivi/client"
	"kivi/server"
)

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "federation:", err)
		os.Exit(1)
	}
}

// must unwraps any (T, error) result, exiting on error. Used as a statement
// (return value discarded) when only the side effect matters.
func must[T any](v T, err error) T { fail(err); return v }

// company is one independent insurer's whole deployment: its own server,
// its own gRPC listener, and an SDK client pointed at itself.
type company struct {
	name string
	addr string
	srv  *server.Server
	gs   *grpc.Server
	cl   *client.Client
}

// start boots an independent kivid in-process on addr, configured to witness
// exactly the peers in `witness` (their addresses). Signing is ON (the
// single-node default) — federation needs seals to attest with.
func start(dir, name, addr string, witness []string) *company {
	s := server.New(
		server.Config{Data: filepath.Join(dir, name+".kivi"), WitnessPeers: witness},
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	)
	lis, err := net.Listen("tcp", addr)
	fail(err)
	unary, stream := s.Interceptors()
	gs := grpc.NewServer(grpc.ChainUnaryInterceptor(unary), grpc.ChainStreamInterceptor(stream))
	api.RegisterKiviServer(gs, s)
	go gs.Serve(lis)
	cl, err := client.Dial(addr, client.Options{})
	fail(err)
	return &company{name: name, addr: addr, srv: s, gs: gs, cl: cl}
}

func (c *company) stop() {
	c.cl.Close()
	c.gs.Stop()
	c.srv.Shutdown()
}

// freeAddrs grabs n OS-assigned free ports and returns their addresses; the
// listeners are closed immediately so the servers can take them (the tiny
// race is fine for a local demo — the exact pattern the federation tests use).
func freeAddrs(n int) []string {
	out := make([]string, n)
	for i := range out {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		fail(err)
		out[i] = l.Addr().String()
		l.Close()
	}
	return out
}

func main() {
	ctx := context.Background()
	dir, err := os.MkdirTemp("", "kivi-federation-demo")
	fail(err)
	defer os.RemoveAll(dir)

	// two free ports, decided up front so each server can be told to witness
	// the other from the start (the witness list is deliberately explicit —
	// you only pin peers you have chosen to)
	addrs := freeAddrs(2)
	acmeAddr, reAddr := addrs[0], addrs[1]

	acme := start(dir, "acme", acmeAddr, []string{reAddr})         // Acme witnesses Re-Global
	reglobal := start(dir, "reglobal", reAddr, []string{acmeAddr}) // and vice-versa (mutual)
	defer acme.stop()
	defer reglobal.stop()
	fmt.Printf("== two independent insurers, two ledgers, no shared database\n")
	fmt.Printf("   Acme (ceding)   @ %s\n   Re-Global (reinsurer) @ %s\n\n", acmeAddr, reAddr)

	// --- each side records the SAME cession, in ITS OWN words, on ITS OWN ledger
	fmt.Println("== Acme records that it is ceding 60% of a large claim to Re-Global")
	must(acme.cl.Append(ctx, "claim.ceded", map[string]any{
		"claim_id": "CLM-5001", "treaty": "TR-2026-QS", "cession_pct": 60,
		"ceded_reserve_cents": 2_700_000, "reinsurer": "Re-Global",
	}))

	fmt.Println("== Re-Global records that it ACCEPTS the cession (its own signed commitment)")
	must(reglobal.cl.Append(ctx, "cession.accepted", map[string]any{
		"claim_id": "CLM-5001", "treaty": "TR-2026-QS", "accepted_pct": 60,
		"accepted_reserve_cents": 2_700_000, "ceding_insurer": "Acme",
	}))

	// --- the mutual witnessing: each pins the other's signed head into its own
	fmt.Println("\n== Acme witnesses Re-Global — fetch a fresh signed attestation, verify OFFLINE, FILE it")
	wAcme := must(acme.cl.Witness(ctx, reAddr))
	fmt.Printf("   Re-Global's chain head is now pinned inside Acme's ledger as record %d (kivi.witness)\n", wAcme.No)

	fmt.Println("== Re-Global witnesses Acme — the same, the other direction (mutual)")
	wRe := must(reglobal.cl.Witness(ctx, acmeAddr))
	fmt.Printf("   Acme's chain head is now pinned inside Re-Global's ledger as record %d (kivi.witness)\n", wRe.No)

	// --- show the testimony: the kivi.witness record inside Acme, pinning Re-Global
	fmt.Println("\n== the testimony itself — Acme's kivi.witness record about Re-Global:")
	rec := must(acme.cl.GetRecord(ctx, wAcme.No, false))
	body, _ := rec["body"].(map[string]any)
	fmt.Printf("   peer       : %v\n", body["peer"])
	fmt.Printf("   covered_no : %v  (Re-Global's history up to this record is pinned)\n", body["covered_no"])
	fmt.Printf("   head_hash  : %v…\n", short(body["head_hash"]))
	fmt.Printf("   pk         : %v…  (Re-Global's OWN signing key — it signed this head itself)\n", short(body["pk"]))

	// --- the honest door: an attestation that fails verification is refused
	//     BEFORE anything is filed, and an unlisted peer is refused by name
	fmt.Println("\n== the door is honest: a peer you never listed cannot be witnessed")
	if _, err := acme.cl.Witness(ctx, "127.0.0.1:1"); err != nil {
		fmt.Printf("   refused, as it must be: %s\n", firstLine(err.Error()))
	}

	fmt.Println("\n== why this matters")
	fmt.Println("   Re-Global signed its acceptance of the 60% cession. That signed head now")
	fmt.Println("   lives inside ACME's ledger. If Re-Global ever rewrites its past to deny the")
	fmt.Println("   cession, its own old signed head — in Acme's records — proves the denial is a")
	fmt.Println("   lie. No shared database, no blockchain, no consensus on every write. Receipts.")

	fmt.Println("\nFEDERATION DEMO OK")
}

func short(v any) string {
	s := fmt.Sprint(v)
	if len(s) > 16 {
		return s[:16]
	}
	return s
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
