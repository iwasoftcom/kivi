#!/usr/bin/env python3
"""The mind journal (PLAN3 G5, kivi side) — what it looks like when an honest
mind keeps its memory on kivi.

This is the shape an honest mind's memory takes on kivi: perceptions, feelings and values
become EVENTS; the mind's current state is a COMPILED VIEW; and the question a
mere log can never answer — "why do I hold this value? SHOW me" — gets an
answer made of receipts, verified CLIENT-SIDE. Identity = replayable history.

    # start any kivid, then:
    clients/python/.venv/bin/python examples/mind_journal.py localhost:4741 [token]
"""

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)),
                                "..", "clients", "python"))

from kivi import KiviClient  # noqa: E402


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    addr = sys.argv[1]
    token = sys.argv[2] if len(sys.argv) > 2 else None
    c = KiviClient(addr, token=token)  # verification ON by default

    # a day in a small mind's life — every experience is an event, nothing more
    print("== living (append-only, hash-chained, receipted)")
    c.append("perception", {"sense": "hearing", "event": "noise", "intensity": 8})
    c.append("his", {"ad": "tedirginlik", "source": "noise", "yön": -1})
    c.append("perception", {"sense": "hearing", "event": "silence", "intensity": 1})
    c.append("his", {"ad": "huzur", "source": "silence", "yön": 1})
    r = c.append("property", {"subject": "mind", "attribute": "value",
                              "value": "I value silence"})
    print(f"   value formed at record {r.no} (hash {r.hash[:12]}…)")

    # the mind's current state: a compiled view, disposable, reborn on demand
    a = c.table("mind", "value")
    print(f"== current value: {a.value!r}  (scope: records 0..{a.scope})")

    # THE question — "why?" — answered with receipts, not narrative
    print("== why do I hold this value? (the receipts themselves)")
    for rec in c.why(a.trace):
        print(f"   record {rec['no']}: {rec['body']}")

    # and nothing above was taken on faith: replay re-verifies EVERY record
    n = sum(1 for _ in c.replay(verify=True))
    print(f"== identity = replayable history: {n} records re-verified client-side")
    print("MIND JOURNAL OK")
    return 0


if __name__ == "__main__":
    sys.exit(main())
