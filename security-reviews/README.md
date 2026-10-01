# Security reviews

Kensho vendors this provider in, and it holds Jenkins **admin** credentials while talking to `jenkins.infra.kensho.com`. Each review of that risk gets its own dated file here, so the security posture has a history rather than a single snapshot that silently goes stale.

## Log

| Review | Scope | Baseline commit | Verdict |
| --- | --- | --- | --- |
| [2026-08-28](2026-08-28.md) | Upstream provider, pre-vendor diligence. Kensho's `*vault_*` additions out of scope. | `8a024df` | Reasonable to vendor, with fixes. No HIGH. 2 MEDIUM (upstream, deferred), 2 process obligations. |

## Adding a review

Name the file `YYYY-MM-DD.md` for the date the review was performed. Open it with a header block giving the review date, the commit reviewed, who performed it, the scope (state plainly what was *not* covered), the context, and the method. Then: verdict, findings by severity, and an explicit list of areas audited and found clean. Recording the clean areas is as useful as recording the findings, since it tells the next reviewer what has already been covered and at which commit.

Carry forward any finding from an earlier review that is still open, and say what changed. A finding that was deferred should not quietly disappear between reviews.

## When a new review is warranted

- Bumping the vendored upstream to a new commit or tag, especially one that touches `config.go`, the provider schema, the credential resources, or `go.mod`.
- Adding or materially changing Kensho's own resources.
- Changing how the provider authenticates to Jenkins, or what it is trusted to reach.
- Remediating a previously deferred finding, which should be reflected as an update rather than left implicit.

## Open items

Tracked in the most recent review. As of 2026-08-28: M1 (`ca_cert` is a no-op in the pinned client) and M2 (no HTTPS enforcement on `server_url`) are open and deferred, both in upstream code. P1 (re-sign releases with a Kensho-controlled key for the private registry) and P2 (mirror and manually track `bndr/gojenkins`) are process obligations that are not yet met.
