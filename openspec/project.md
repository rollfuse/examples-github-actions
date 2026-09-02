# Project Context

## What This Repo Is

An example repository under the `rollfuse` GitHub organization: gating a
GitHub Actions deploy step on a rollfuse flag using
[`@rollfuse/go-sdk`](https://github.com/rollfuse/go-sdk) — a kill-switch
pattern. `.github/workflows/deploy-with-flag-gate.yml` is the example
itself, runnable with no rollfuse account (a bundled mock stands in for
the platform API). It is not part of the rollfuse platform itself.

## Repository Shape

```text
/
├── cmd/
│   ├── flag-gate/            The CLI: evaluates one flag, exits 0 only
│   │                         if the variation matches -want-variation.
│   └── mock-rollfuse-api/    Self-contained stand-in for the real platform
│                             API (GET /v1/config, POST /v1/exposure-events).
├── config/
│   └── deploy-gate-config.json   Static Configuration the mock API serves.
└── .github/workflows/
    ├── ci.yml                          This repo's own CI.
    └── deploy-with-flag-gate.yml       THE example — copy this into your
                                         own repo and adapt.
```

## Working On This Repo

- `flag-gate` stays a plain CLI, not a GitHub composite Action — the
  point is that the same binary works in any CI system, not just GitHub
  Actions. Don't turn it into an `action.yml`-wrapped Action without
  discussing the scope change first (that's closer to what a future
  `rollfuse/github-action` repo would be).
- `go build ./...`, `go vet ./...`, and `gofmt -l .` (clean) are the only
  static checks; correctness beyond that is verified by CI actually
  running `flag-gate` against the mock API and asserting BOTH outcomes
  (passes when the flag matches, exits non-zero when it doesn't) — this
  is a demo, not a library, so "does the documented flow actually work"
  is the bar, not unit tests.
- The README's claims about the pattern (what blocks, what doesn't, why
  it's better than a hardcoded `if:` condition) should stay grounded in
  what `deploy-with-flag-gate.yml` and `flag-gate` actually do.
- OpenSpec here is for planning nontrivial changes to the demo itself,
  not for specifying rollfuse platform behavior (that lives in
  `rollfuse/rollfuse` and `rollfuse/go-sdk`).
