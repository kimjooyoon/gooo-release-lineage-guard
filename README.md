# Gooo Release Lineage Guard

Gooo Release Lineage Guard is a read-only release mutation plan evaluator. It evaluates a caller-provided remote snapshot and proposed operations before any GitHub mutation is attempted. The product runtime never deletes releases or tags, retags, recreates a version, replaces a published asset, or deletes a failed run.

The only semantic authority is [`semantic/release-lineage.gooo`](semantic/release-lineage.gooo). It declares the seven semantic entities, the fixed twelve meta activities, the proof and indicator partitions, the allowlist, the forbidden mutation set, the `REFUTED > UNKNOWN > CLOSED` precedence, the six UNKNOWN fields, the explicit `FIXED_POINT` terminal, and the twelve conformance cases. Go parses and evaluates that source; JSON fixtures are observations and proposed operations, not policy.

The twelve activity cells are partitioned exactly `FOUNDATION 4 / COHERENCE 4 / REGRESSION 4` and `DRIVER 4 / OUTCOME 4 / GUARDRAIL 4`. The fixture corpus is exactly `CLOSED 4 / UNKNOWN 4 / REFUTED 4`.

The generated evidence is indexed by [`generated/manifest.json`](generated/manifest.json), with exact file IDs and SHA-256 digests in [`generated/inventory.json`](generated/inventory.json). The motivating anonymized counterexample is [`12-refuted-configured-after-publish.json`](fixtures/cases/12-refuted-configured-after-publish.json) and its machine decision is [`refuted-configured-after-publish.json`](generated/decisions/refuted-configured-after-publish.json).

## Decision contract

`CLOSED` means the read-only plan reaches the explicit `FIXED_POINT`: immutable policy is known enabled, the remote inventory and lineage identity are bound, a new annotated tag is proposed, the release is created draft-first, the main artifact digest is exact, and publication follows the draft.

`UNKNOWN` holds the plan when policy, remote snapshot completeness, lineage identity, or source-backed external utility/improvement evidence is missing. Every UNKNOWN detail carries exactly `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.

`REFUTED` preserves the public outcome as `OPERATIONAL_REFUTED`. The failed artifact and public history are retained; the only forward path is a new patch version. `REFUTED` dominates `UNKNOWN`, which dominates `CLOSED`.

There are no aggregate scores, percentages, weighted sums, or scalar rankings in the decision output. The runtime boundary is explicit: `repository_writes=0`, `local_test_executions=0`, `cross_project_required_gates=0`, and `github_mutations=0`. Wall and RSS are `null` when no performance observation was supplied.

## Use

```text
go run ./cmd/gooo-release-lineage-guard plan-gate \
  --source semantic/release-lineage.gooo \
  --fixture fixtures/cases/01-closed-annotated-draft.json

go run ./cmd/gooo-release-lineage-guard conformance \
  --source semantic/release-lineage.gooo \
  --fixtures fixtures/cases

go run ./cmd/gooo-release-lineage-guard generate \
  --source semantic/release-lineage.gooo \
  --fixtures fixtures/cases \
  --out generated
```

`plan-gate` prints machine decision JSON and exits successfully only for `CLOSED` + `FIXED_POINT`. `generate` writes machine decisions, human dossiers, a reusable CI command descriptor, and a manifest. The generated CI descriptor is [`generated/ci/plan-gate-command.json`](generated/ci/plan-gate-command.json). The PR workflow is authoritative for Go checks and artifact regeneration; no local test/build/vet/fmt/actionlint/bash/jq/conformance claim is made by this repository.

## Release runbook

1. Open a PR with the semantic source, evaluator, fixtures, and generated artifacts.
2. Require the PR Actions checks to pass, then merge the green PR into `main`.
3. On `main`, read the repository immutable-release policy first. A missing policy is `UNKNOWN`; a known-disabled policy is `REFUTED` for publication.
4. Create only a new annotated tag and a draft release, attach the exact main artifact digest, and publish that draft once.
5. If a pre-policy public release already exists, do not delete or recreate it. Preserve it as `OPERATIONAL_REFUTED` and advance to `v0.1.1`.

GitHub's immutable-release documentation says published release tags cannot be moved or deleted while their release exists, published assets cannot be modified or deleted, and a draft-first workflow attaches assets before publication. The repository-settings procedure says immutability applies to future releases. See [Immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases), [Preventing changes to your releases](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/establish-provenance-and-integrity/prevent-release-changes), and [Managing releases in a repository](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository).

Go 1.27 is the declared toolchain line for this repository. The implementation uses the Go standard library and follows the [Go 1.27 release notes](https://go.dev/doc/go1.27) and [Go language specification](https://go.dev/ref/spec).
