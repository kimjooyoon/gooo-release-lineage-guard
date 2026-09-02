package lineage

import (
	"fmt"
	"sort"
)

var requiredEntities = []string{"ReleaseLineage", "PublicArtifact", "MutationPlan", "PreservationRule", "VersionAdvance", "Decision", "MetaActivity"}
var requiredAllowed = []string{"new_annotated_tag", "draft_first_release", "immutable_policy_precheck", "new_patch_version", "exact_main_artifact_digest_upload", "publish_draft_release", "preserve_failed_artifact", "advance_patch_version"}
var requiredForbidden = []string{"release_delete", "tag_delete", "tag_retag", "same_version_recreate", "published_asset_replace", "failed_run_delete"}
var requiredUnknownFields = []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}
var requiredActivities = []string{"01-lineage-identity", "02-policy-precheck", "03-artifact-identity", "04-plan-identity", "05-operation-allowlist", "06-version-advance", "07-draft-first", "08-digest-upload", "09-failure-preservation", "10-refuted-precedence", "11-unknown-causality", "12-authority-zero"}
var requiredCases = []string{"closed-annotated-draft", "closed-patch-advance", "closed-digest-exact", "closed-policy-prechecked", "unknown-policy-missing", "unknown-remote-snapshot-missing", "unknown-lineage-identity-missing", "unknown-external-utility", "refuted-release-delete", "refuted-tag-retag", "refuted-published-asset-replace", "refuted-configured-after-publish"}

func ValidateSemanticModel(model SemanticModel) error {
	if model.Schema != SourceSchema || model.Name != "gooo-release-lineage-guard" || model.Version != "v1" { return fmt.Errorf("invalid semantic source header") }
	if !sameStrings(entityNames(model.Entities), requiredEntities) { return fmt.Errorf("semantic entities are not the fixed seven") }
	if len(model.Activities) != 12 || !sameStrings(activityIDs(model.Activities), requiredActivities) { return fmt.Errorf("semantic activity set is not the fixed twelve") }
	proofCount := map[ProofFamily]int{}; indicatorCount := map[IndicatorClass]int{}
	for _, activity := range model.Activities {
		if activity.Proof != ProofFoundation && activity.Proof != ProofCoherence && activity.Proof != ProofRegression { return fmt.Errorf("activity %q has invalid proof family", activity.ID) }
		if activity.Indicator != IndicatorDriver && activity.Indicator != IndicatorOutcome && activity.Indicator != IndicatorGuardrail { return fmt.Errorf("activity %q has invalid indicator class", activity.ID) }
		proofCount[activity.Proof]++; indicatorCount[activity.Indicator]++
	}
	for _, proof := range []ProofFamily{ProofFoundation, ProofCoherence, ProofRegression} { if proofCount[proof] != 4 { return fmt.Errorf("proof family %s must contain exactly four activities", proof) } }
	for _, indicator := range []IndicatorClass{IndicatorDriver, IndicatorOutcome, IndicatorGuardrail} { if indicatorCount[indicator] != 4 { return fmt.Errorf("indicator class %s must contain exactly four activities", indicator) } }
	if !sameStrings(operationKinds(model.AllowedOperations), requiredAllowed) { return fmt.Errorf("allowed operation set is not fixed") }
	if !sameStrings(operationKinds(model.ForbiddenOperations), requiredForbidden) { return fmt.Errorf("forbidden operation set is not fixed") }
	if len(model.StatePrecedence) != 3 || model.StatePrecedence[0] != StateRefuted || model.StatePrecedence[1] != StateUnknown || model.StatePrecedence[2] != StateClosed { return fmt.Errorf("state precedence must be REFUTED > UNKNOWN > CLOSED") }
	if model.Terminal != TerminalFixedPoint { return fmt.Errorf("only explicit FIXED_POINT is permitted") }
	if !sameStrings(model.UnknownFields, requiredUnknownFields) { return fmt.Errorf("UNKNOWN fields are not the contracted six") }
	if model.Authority.RepositoryWrites != 0 || model.Authority.LocalTestExecutions != 0 || model.Authority.CrossProjectRequiredGates != 0 { return fmt.Errorf("source authority boundary must be zero") }
	if model.ExternalEvidence.Utility != StateUnknown || model.ExternalEvidence.Improvement != StateUnknown { return fmt.Errorf("external utility and improvement must default to UNKNOWN") }
	if len(model.Cases) != 12 || !sameStrings(caseIDs(model.Cases), requiredCases) { return fmt.Errorf("case set is not the fixed twelve") }
	caseCount := map[CaseClass]int{}; for _, item := range model.Cases { if item.Class != CaseClosed && item.Class != CaseUnknown && item.Class != CaseRefuted { return fmt.Errorf("case %q has invalid class", item.ID) }; caseCount[item.Class]++ }
	if caseCount[CaseClosed] != 4 || caseCount[CaseUnknown] != 4 || caseCount[CaseRefuted] != 4 { return fmt.Errorf("case classes must be 4 CLOSED, 4 UNKNOWN, 4 REFUTED") }
	for _, entity := range model.Entities { if len(entity.Fields) == 0 { return fmt.Errorf("entity %q has no fields", entity.Name) } }
	return nil
}

func sameStrings(left, right []string) bool { if len(left) != len(right) { return false }; for index := range left { if left[index] != right[index] { return false } }; return true }
func entityNames(items []EntitySpec) []string { result := make([]string, 0, len(items)); for _, item := range items { result = append(result, item.Name) }; return result }
func activityIDs(items []ActivitySpec) []string { result := make([]string, 0, len(items)); for _, item := range items { result = append(result, item.ID) }; return result }
func operationKinds(items []OperationRule) []string { result := make([]string, 0, len(items)); for _, item := range items { result = append(result, item.Kind) }; return result }
func caseIDs(items []CaseSpec) []string { result := make([]string, 0, len(items)); for _, item := range items { result = append(result, item.ID) }; return result }

func CaseSpecFor(model SemanticModel, id string) (CaseSpec, bool) { for _, item := range model.Cases { if item.ID == id { return item, true } }; return CaseSpec{}, false }
func SortedOperationKinds(model SemanticModel) []string { result := operationKinds(model.ForbiddenOperations); sort.Strings(result); return result }
