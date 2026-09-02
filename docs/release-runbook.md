# Release runbook

This runbook is intentionally PR-first and mutation-aware.

## Before merge

The pull request must carry the `.gooo` source, Go parser/evaluator/CLI, all twelve anonymized conformance fixtures, and generated decision/dossier artifacts. PR Actions is the only validation authority for this repository. Local test, build, vet, format, actionlint, shell, jq assertion, and conformance executions are outside the evidence boundary.

## Before publication

After the green PR is merged, capture a read-only snapshot from `main`. The snapshot must identify the repository, default branch, head commit, lineage identity, complete tags/releases/assets inventory, and immutable-release policy. Policy is checked before proposing any release operation. A missing field is `UNKNOWN`, not permission to continue.

The safe sequence is:

- `immutable_policy_precheck`
- `new_annotated_tag`
- `draft_first_release`
- `exact_main_artifact_digest_upload`
- `publish_draft_release`

The evaluator does not execute this sequence; it evaluates a proposed sequence and returns a decision. A `CLOSED` decision is the explicit `FIXED_POINT` authorization boundary for an external release operator.

## Incident handling

The fixture `refuted-configured-after-publish` anonymizes the motivating counterexample: a public `v0.1.0` existed before immutable policy was configured, and a cleanup plan proposed release deletion, tag deletion, and same-version recreation. The evaluator returns `REFUTED`, preserves the outcome as `OPERATIONAL_REFUTED`, never authorizes deletion, and advances only to `v0.1.1`.

GitHub documents that immutable releases protect tags and assets after publication and recommends draft-first publication. Immutability applies to future releases, so policy timing is part of the evidence: [immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases) and [preventing release changes](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/establish-provenance-and-integrity/prevent-release-changes).
