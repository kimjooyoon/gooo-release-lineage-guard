# Gooo Release Lineage Guard dossier

- Decision schema: `gooo/release-lineage-guard/decision/v1`
- Case ID: `refuted-release-delete`
- Plan ID: `plan-refuted-release-delete`
- State: `REFUTED`
- Terminal: `FIXED_POINT`
- Action: `PRESERVE_AND_ADVANCE`
- Description: A proposed release deletion is rejected and must remain as public history.

## Decision

proposed release mutation is REFUTED; retain public history and advance by patch version.

## Lineage and artifact identity

- Repository: `kimjooyoon/gooo-release-lineage-guard`
- Stream: `public-release`
- Lineage ID: `lineage-main-public-v1`
- Artifact `gooo-release-lineage-guard-v0.1.1.tar.gz`: role `main`, sha256 `sha256:3333333333333333333333333333333333333333333333333333333333333333`, bytes `1293`, main `true`

## Proposed outcome

The public outcome is preserved as `OPERATIONAL_REFUTED`; deletion, retagging, replacement, or same-version recreation is not a repair path.

- `op-refuted-delete-02` `release_delete`: release deletion would erase public history

## Version advance

- From: `v0.1.1`
- To: `v0.1.2`
- Required: `true`
- Reason: failed public output is preserved; a new patch version is the only forward path

## Runtime boundary

The evaluator reads a caller-provided snapshot and proposed operations only. Repository writes, local test executions, cross-project required gates, and GitHub mutations are all `0`; wall and RSS are `null` because no local performance claim is made.

## Meta activities

- `01-lineage-identity` / `BindLineageIdentity` / `FOUNDATION` / `DRIVER`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `02-policy-precheck` / `PrecheckImmutablePolicy` / `FOUNDATION` / `DRIVER`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `03-artifact-identity` / `BindPublicArtifact` / `FOUNDATION` / `DRIVER`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `04-plan-identity` / `BindMutationPlan` / `FOUNDATION` / `DRIVER`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `05-operation-allowlist` / `EvaluateProposedOperations` / `COHERENCE` / `OUTCOME`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `06-version-advance` / `RequireNextPatchVersion` / `COHERENCE` / `OUTCOME`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `07-draft-first` / `RequireDraftFirstRelease` / `COHERENCE` / `OUTCOME`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `08-digest-upload` / `RequireExactMainDigest` / `COHERENCE` / `OUTCOME`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `09-failure-preservation` / `PreserveOperationalRefuted` / `REGRESSION` / `GUARDRAIL`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `10-refuted-precedence` / `EnforceRefutedPrecedence` / `REGRESSION` / `GUARDRAIL`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `11-unknown-causality` / `PreserveUnknownFields` / `REGRESSION` / `GUARDRAIL`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
- `12-authority-zero` / `EnforceReadOnlyRuntime` / `REGRESSION` / `GUARDRAIL`: `REFUTED` — proposed release mutation is REFUTED; retain public history and advance by patch version
