# Gooo Release Lineage Guard dossier

- Decision schema: `gooo/release-lineage-guard/decision/v1`
- Case ID: `refuted-published-asset-replace`
- Plan ID: `plan-refuted-published-asset-replace`
- State: `REFUTED`
- Terminal: `FIXED_POINT`
- Action: `PRESERVE_AND_ADVANCE`
- Description: A replacement of a published asset is rejected even when the new digest is known.

## Decision

proposed release mutation is REFUTED; retain public history and advance by patch version.

## Lineage and artifact identity

- Repository: `kimjooyoon/gooo-release-lineage-guard`
- Stream: `public-release`
- Lineage ID: `lineage-main-public-v1`
- Artifact `gooo-release-lineage-guard-v0.1.0.tar.gz`: role `main`, sha256 `sha256:5555555555555555555555555555555555555555555555555555555555555555`, bytes `1317`, main `true`

## Proposed outcome

The public outcome is preserved as `OPERATIONAL_REFUTED`; deletion, retagging, replacement, or same-version recreation is not a repair path.

- `op-refuted-asset-02` `published_asset_replace`: published assets are immutable
- `VERSION_NOT_NEXT_PATCH`: requested version is not the next unused patch version
- `SAME_VERSION_REUSE`: same version cannot be recreated after a public artifact exists

## Version advance

- From: `v0.1.0`
- To: `v0.1.1`
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
