# rollfuse in GitHub Actions: a deploy kill switch

Gate a deploy step on a [rollfuse](https://rollfuse.com) flag using
[`@rollfuse/go-sdk`](https://github.com/rollfuse/go-sdk) — flip the flag
off and the next deploy is blocked, with no new commit, no workflow edit,
and no redeploy of anything. See it run:
**[Actions → deploy-with-flag-gate](actions/workflows/deploy-with-flag-gate.yml)**,
then hit "Run workflow" and pick `simulate: off` to watch it block in
real time.

## The pattern

```
checkout → setup-go → flag-gate → deploy
                          │
                    exits 1 if the flag
                    doesn't match what
                    you asked for — the
                    "deploy" step never
                    runs (GitHub Actions
                    stops the job on a
                    failed step)
```

`cmd/flag-gate` is a small, dependency-light CLI: it evaluates one flag
for one subject key and exits `0` only if the resulting variation matches
`-want-variation`. Nothing GitHub-Actions-specific about it — the same
binary gates a step in GitLab CI, CircleCI, or a plain shell script; it's
a CLI, not a composite Action.

```bash
go run ./cmd/flag-gate -flag deploys-enabled -want-variation on
# or, via env vars (what the workflow actually uses):
FLAG_KEY=deploys-enabled WANT_VARIATION=on go run ./cmd/flag-gate
```

## Use this in your own repo

1. Copy `.github/workflows/deploy-with-flag-gate.yml` into your repo.
2. Point `ROLLFUSE_API_BASE_URL` at your real environment
   (`https://api.rollfuse.com`) and put a real Service Credential in a
   repository secret — reference it as
   `${{ secrets.ROLLFUSE_SERVICE_CREDENTIAL }}`, never a literal like the
   demo below.
3. Delete the "start mock rollfuse API" step — you don't need it against
   a real environment. It exists only so this example runs with no
   rollfuse account, same as every other repo in this org.
4. Replace "deploy (simulated)" with your real deploy command.
5. Create the flag in your rollfuse project (a boolean, two variations —
   `on`/`off` or whatever names you prefer, matching `-want-variation`).

## Why this instead of your CI config's own if-condition

A flag gives you what a hardcoded `if: github.ref == 'refs/heads/main'`
can't: change the gate's outcome **without touching the workflow file or
making a new commit** — flip it from the rollfuse console (or another
automated trigger, like a guardrail auto-rollback), and the very next
workflow run picks it up. Useful as a manual freeze switch during an
incident, or wired to the same guardrail automation that already pauses a
rollout — pausing your deploy pipeline too, from one place.

## Development

```bash
go build ./...
go vet ./...

# exercise both outcomes locally, no CI needed:
CONFIG_PATH=config/deploy-gate-config.json go run ./cmd/mock-rollfuse-api &
FLAG_KEY=deploys-enabled WANT_VARIATION=on  go run ./cmd/flag-gate   # PASSED
FLAG_KEY=deploys-enabled WANT_VARIATION=off go run ./cmd/flag-gate   # BLOCKED, exit 1
```

`cmd/mock-rollfuse-api` is the same self-contained platform-API stand-in
every example in this org uses — see its own doc comment. `flag-gate`
reads `ROLLFUSE_API_BASE_URL` (default `http://localhost:8090`),
`ROLLFUSE_SERVICE_CREDENTIAL`, `FLAG_KEY` (required), `SUBJECT_KEY`
(default `ci`), and `WANT_VARIATION` (default `on`) — or the equivalent
`-base-url`/`-credential`/`-flag`/`-subject`/`-want-variation` flags.

## Related

- [`rollfuse/go-sdk`](https://github.com/rollfuse/go-sdk) — the SDK this
  example runs.
- [`rollfuse/examples-kubernetes`](https://github.com/rollfuse/examples-kubernetes) —
  the same SDK, changing a rollout live on a running deployment instead of
  gating a pipeline step.
