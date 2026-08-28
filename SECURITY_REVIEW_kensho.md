# Security review: terraform-provider-jenkins (pre-vendor diligence)

**Reviewed commit:** `8a024df` (upstream `taiidani/terraform-provider-jenkins` `main`, before any Kensho changes)
**Context:** Kensho is vendoring this third-party provider in to manage Jenkins credentials as code and use Vault as the single source of truth for secrets. The provider holds Jenkins **admin** credentials and talks to `jenkins.infra.kensho.com`. Per Dave's caution ("be very careful about using a 3rd party tf provider for jenkins"), this is a supply-chain + code audit of what we're pulling in.
**Method:** read-only audit across three slices — credential/secret handling, transport/injection/XML/egress, and build/release supply chain. Load-bearing transport findings verified against the pinned gojenkins source in the module cache.

## Verdict

**Reasonable to vendor, with fixes.** No HIGH findings, no code-execution / injection / XXE / egress issues. The credential-handling layer is clean by design. There are **two transport weaknesses worth fixing before this holds admin credentials in production**, and **two process obligations Kensho must own** for the private-registry republish. All are addressable in our fork.

Both MEDIUM findings (M1, M2) are in **upstream** code, not in Kensho's additions (the four `*vault_*` resources/data sources). They are **recorded here, deferred — not fixed in this PR**; remediation tracked as follow-up.

## Findings

### MEDIUM (upstream; deferred to follow-up, not fixed in this PR)

**M1 — `ca_cert` provider option is a silent no-op in the pinned client.**
`jenkins/config.go:34-38` assigns the CA bytes to `client.Requester.CACert`, but the pinned `bndr/gojenkins@…20210407143218` **never reads** `CACert` or `SslVerify` — verified: those fields are only declared (`request.go:61-62`), set in the constructor (`jenkins.go:605`), and documented; there is no `tls.Config`/`TLSClientConfig`/`CertPool`/`AppendCertsFromPEM` anywhere in the library. Because the provider calls `CreateJenkins(nil, …)`, the client falls back to `http.DefaultClient` (`jenkins.go:604-607`). Net effect: TLS uses the **system trust pool with full verification** (verification is *not* disabled), but a supplied `ca_cert` does nothing. Practical impact: (a) false assurance of CA pinning; (b) an internal-CA-only Jenkins cert fails to connect rather than silently downgrading — which can tempt an insecure workaround. Also the read error is swallowed (`_`).
**Fix:** in `config.go`, build an explicit `*http.Client` with `tls.Config{RootCAs: pool}` where `pool` is a clone of `x509.SystemCertPool()` plus the supplied CA (augment, never replace), pass it into `CreateJenkins(client, …)`, and surface the read/append errors. Add an integration test against an internal-CA-only cert.

**M2 — Plain `http://` `server_url` is permitted (no scheme enforcement).**
`server_url` (`jenkins/provider.go` schema / `provider_framework.go`) is a free-form string; nothing rejects `http://`. gojenkins sends Basic-Auth admin credentials on **every** request (`request.go` `SetBasicAuth`). A cleartext URL (typo, or a `JENKINS_URL` env default) leaks the admin password to on-path observers.
**Fix:** validate `server_url` requires `https://` (allow `http` only for explicit localhost test targets); fail fast in Configure otherwise. For the Kensho fork, enforce HTTPS unconditionally.

### Process obligations (own before publishing to the private registry)

**P1 — Re-establish the trust chain for the private registry.** Upstream release hygiene is good (goreleaser emits sha256 checksums, GPG-detach-signs the checksum file, release CI SHA-pins goreleaser + import-gpg, secrets confined to a tag-triggered job). But that signature is taiidani's and is meaningless once Kensho republishes. Kensho must pin an immutable upstream tag, verify `*_SHA256SUMS` + `.sig` against taiidani's known key, verify each zip, **then re-sign with a Kensho-controlled key** registered in the private registry — or (preferred) build from source at the audited commit. Nothing in the repo enforces this.

**P2 — `bndr/gojenkins` is the one real provenance wart.** Single-maintainer personal module, pinned to an untagged 2021 commit (`v1.1.1-0.20210407143218-9e2483ff7ebd`), sitting **directly on the Jenkins-admin-credential path**. Not a typosquat, and `go.sum` pins it immutably (no live-swap risk), but dependabot won't meaningfully update a commit-pin. Mirror it through Kensho's GOPROXY, track it for CVEs manually, and note that any fix for M1 that relies on the library honoring `ca_cert` must first move to a version that actually wires TLS — hence M1's fix builds our own `http.Client` instead.

### LOW / informational

- **L1** — AWS `access_key` marked `Sensitive: true` (`resource_jenkins_credential_aws.go:70`, datasource `:47`). Over-classification of a public key ID; harmless.
- **L2** — Pinned gojenkins has a `log.Printf("DEBUG …")` **response**-dump path (`request.go:233-238`) gated on `ctx.Value("debug")`. Dead code here — the provider never sets a `"debug"` context key. Landmine only if Kensho later adds request tracing with that key name.
- **N1 (non-security, upstream functional bug)** — `resource_jenkins_credential_azure_service_principal.go:302-304`: `Update` assigns `certificate_id` into `cred.Data.ClientId` instead of `CertificateId`. Worth upstreaming a fix; not a secret leak.

## Areas audited and confirmed clean

- **Secrets in logs / state:** no secrets to logs/tflog/prints/errors; every `Read()` deliberately does *not* read secrets back into state ("GetSingle is garbage" pattern); all secret schema attributes correctly `Sensitive: true`; data sources omit secret attributes entirely; admin password flows only into Basic-Auth, never logged; no hardcoded creds in the binary.
- **Injection / RCE:** no `os/exec`/shelling out anywhere; command injection surface is nil.
- **XXE:** XML parsed only via Go stdlib `encoding/xml` (`folder.go:42`), which does not resolve external entities; no custom decoder/CharsetReader/entity map/third-party XML parser.
- **SSRF:** user-controlled names/folders/ids are validated (reject `/`) and only ever land in URL **path** positions of a fixed host+scheme; cannot control host or protocol (path-only is out of scope).
- **Egress:** no telemetry, update-check, or phone-home; all traffic goes to the configured `server_url`.
- **Dependency tree (besides gojenkins):** HashiCorp-official + stdlib-adjacent, full `go.sum` hashes, doc-gen deps isolated in `tools/go.mod`.
- **CI:** no `pull_request_target` misuse, no `github.event.*` interpolation into `run:` steps, no build-time remote code fetch/execute.
