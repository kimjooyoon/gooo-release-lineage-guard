package lineage

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var versionPattern = regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.([0-9]+)$`)

func Evaluate(model SemanticModel, fixture Fixture, sourcePath, snapshotPath string) (Decision, error) {
	if err := ValidateSemanticModel(model); err != nil { return Decision{}, err }
	if fixture.Schema != FixtureSchema || fixture.CaseID == "" || fixture.Description == "" { return Decision{}, fmt.Errorf("invalid fixture header") }
	caseSpec, exists := CaseSpecFor(model, fixture.CaseID); if !exists { return Decision{}, fmt.Errorf("fixture case %q is not declared by .gooo", fixture.CaseID) }
	if fixture.Plan.Schema != PlanSchema { return Decision{}, fmt.Errorf("fixture plan has invalid schema") }
	if fixture.Snapshot.Schema != SnapshotSchema { return Decision{}, fmt.Errorf("fixture snapshot has invalid schema") }

	decision := Decision{
		Schema: DecisionSchema, CaseID: fixture.CaseID, PlanID: fixture.Plan.PlanID, Terminal: model.Terminal,
		Action: ActionProceed, Reason: "read-only plan satisfies the explicit FIXED_POINT release lineage policy",
		PreservationRule: PreservationRule{OnRefuted: "retain_failed_public_artifact_as_OPERATIONAL_REFUTED", PublicHistory: "preserve", FailedArtifact: "retain", NeverDelete: true},
		VersionAdvance: VersionAdvance{FromVersion: fixture.Plan.Lineage.CurrentVersion, ToVersion: fixture.Plan.RequestedVersion, Reason: "new patch version only", Required: fixture.Plan.Lineage.CurrentVersion != fixture.Plan.RequestedVersion},
		Runtime: RuntimeBoundary{RepositoryWrites: 0, LocalTestExecutions: 0, CrossProjectRequiredGates: 0, GitHubMutations: 0, ReadOnlySnapshot: true, WallMS: nil, PeakRSSKiB: nil},
		Source: sourcePath, Snapshot: snapshotPath, Lineage: fixture.Plan.Lineage, Artifacts: fixture.Plan.Artifacts,
	}

	unknowns := make([]UnknownDetail, 0)
	violations := make([]Violation, 0)
	addUnknown := func(stage, step, reason, class, next string, blocked []string) { unknowns = append(unknowns, UnknownDetail{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: blocked}) }
	addViolation := func(operationID, kind, reason string) { violations = append(violations, Violation{OperationID: operationID, Kind: kind, Reason: reason}) }

	if fixture.Plan.Lineage.Repository == "" || fixture.Plan.Lineage.Stream == "" || fixture.Plan.Lineage.LineageID == "" || fixture.Plan.TargetCommit == "" {
		addUnknown("LINEAGE", "BIND_LINEAGE_IDENTITY", "release lineage identity is incomplete", "LINEAGE_IDENTITY_MISSING", "supply repository, stream, lineage_id, and target_commit", []string{"MutationPlan.lineage", "MutationPlan.target_commit"})
	}
	if fixture.Snapshot.Repository == "" || fixture.Snapshot.DefaultBranch == "" || fixture.Snapshot.HeadCommit == "" || fixture.Snapshot.LineageIdentity == "" {
		addUnknown("SNAPSHOT", "READ_REMOTE_IDENTITY", "remote identity is incomplete", "REMOTE_SNAPSHOT_IDENTITY_MISSING", "capture repository, default_branch, head_commit, and lineage_identity", []string{"RemoteSnapshot.identity"})
	}
	if fixture.Snapshot.ImmutablePolicy == nil {
		addUnknown("POLICY", "READ_IMMUTABLE_POLICY", "immutable release policy was not present in the read-only snapshot", "POLICY_MISSING", "capture repository immutable release policy before proposing publication", []string{"RemoteSnapshot.immutable_policy"})
	}
	if fixture.Snapshot.InventoryComplete == nil || !valueOrFalse(fixture.Snapshot.InventoryComplete) {
		addUnknown("SNAPSHOT", "READ_RELEASE_AND_TAG_INVENTORY", "remote release and tag inventory is not complete", "REMOTE_SNAPSHOT_MISSING", "capture all releases, tags, and assets before evaluating version reuse", []string{"RemoteSnapshot.inventory_complete", "RemoteSnapshot.tags", "RemoteSnapshot.releases"})
	}
	if fixture.Plan.Lineage.Repository != "" && fixture.Snapshot.Repository != "" && fixture.Plan.Lineage.Repository != fixture.Snapshot.Repository {
		addViolation("", "LINEAGE_REPOSITORY_MISMATCH", "plan lineage repository does not match the remote snapshot")
	}
	if fixture.Plan.Lineage.LineageID != "" && fixture.Snapshot.LineageIdentity != "" && fixture.Plan.Lineage.LineageID != fixture.Snapshot.LineageIdentity {
		addViolation("", "LINEAGE_IDENTITY_MISMATCH", "plan lineage identity does not match the remote snapshot")
	}
	if fixture.Snapshot.ImmutablePolicy != nil && !*fixture.Snapshot.ImmutablePolicy {
		addViolation("", "IMMUTABLE_POLICY_DISABLED", "publication cannot reach FIXED_POINT before immutable releases are enabled")
	}

	mainArtifact, artifactOK := mainArtifact(fixture.Plan.Artifacts)
	if !artifactOK { addUnknown("ARTIFACT", "BIND_MAIN_ARTIFACT", "exactly one main public artifact is required", "MAIN_ARTIFACT_IDENTITY_MISSING", "supply one main artifact name, size, and SHA-256 digest", []string{"MutationPlan.artifacts"}) }
	if fixture.Plan.RequestedVersion == "" { addUnknown("VERSION", "BIND_REQUESTED_VERSION", "requested patch version is missing", "VERSION_IDENTITY_MISSING", "supply a new patch version", []string{"MutationPlan.requested_version"}) }

	seenKinds := map[string]int{}
	hasPrecheck, hasDraft, hasDigestUpload, hasPublish := false, false, false, false
	for _, operation := range fixture.Plan.ProposedOperations {
		seenKinds[operation.Kind]++
		if containsOperation(model.ForbiddenOperations, operation.Kind) {
			addViolation(operation.ID, operation.Kind, forbiddenReason(operation.Kind))
			continue
		}
		switch operation.Kind {
		case "immutable_policy_precheck": hasPrecheck = true
		case "new_annotated_tag":
			if !operation.Annotated { addViolation(operation.ID, operation.Kind, "new release tags must be annotated") }
			if operation.TargetCommit != fixture.Plan.TargetCommit { addViolation(operation.ID, operation.Kind, "annotated tag target commit does not match the plan target commit") }
			if existingTag(fixture.Snapshot.Tags, operation.Tag) { addViolation(operation.ID, "same_version_recreate", "tag name already exists; tag reuse is not a new lineage") }
		case "draft_first_release":
			if !operation.Draft { addViolation(operation.ID, operation.Kind, "release must be created as a draft before assets are attached") }
			if existingRelease(fixture.Snapshot.Releases, operation.Tag) { addViolation(operation.ID, "same_version_recreate", "release version already exists") }
			hasDraft = operation.Draft
		case "exact_main_artifact_digest_upload":
			if !artifactOK || operation.ArtifactName != mainArtifact.Name || normalizeDigest(operation.Digest) != normalizeDigest(mainArtifact.SHA256) { addViolation(operation.ID, operation.Kind, "uploaded main artifact digest is not the exact declared digest") }
			hasDigestUpload = true
		case "publish_draft_release": hasPublish = true
		case "new_patch_version", "preserve_failed_artifact", "advance_patch_version":
		default:
			addUnknown("PLAN", "CLASSIFY_OPERATION", fmt.Sprintf("operation kind %q is not declared by the semantic source", operation.Kind), "OPERATION_KIND_UNKNOWN", "classify the operation against the .gooo allowlist", []string{operation.ID})
		}
	}
	if len(fixture.Plan.ProposedOperations) == 0 { addUnknown("PLAN", "READ_PROPOSED_OPERATIONS", "the mutation plan has no proposed operations", "PROPOSED_OPERATIONS_MISSING", "provide an immutable policy precheck and draft-first operation sequence", []string{"MutationPlan.proposed_operations"}) }
	if !hasPrecheck { addUnknown("POLICY", "REQUIRE_POLICY_PRECHECK", "the plan has no explicit immutable policy precheck", "POLICY_PRECHECK_MISSING", "add immutable_policy_precheck before any public mutation", []string{"MutationPlan.proposed_operations"}) }
	if hasPublish && (!hasDraft || !hasDigestUpload) { addViolation("", "PUBLISH_NOT_FIXED_POINT", "published release must follow draft creation and exact main artifact upload") }
	if _, exists := seenKinds["publish_release"]; exists { addViolation("", "PUBLISH_NOT_DRAFT_FIRST", "direct publish operation is not permitted; publish only a prepared draft") }

	if fixture.Plan.ExternalClaims != nil {
		for _, claim := range fixture.Plan.ExternalClaims {
			if _, evidence := fixture.Plan.ExternalEvidence[claim]; !evidence { addUnknown("EVIDENCE", "READ_EXTERNAL_EVIDENCE", fmt.Sprintf("external %s evidence is absent", claim), "EXTERNAL_EVIDENCE_MISSING", "capture source-backed external evidence before making the claim", []string{"MutationPlan.external_claims", "MutationPlan.external_evidence"}) }
		}
	}
	if fixture.Snapshot.ImmutablePolicy != nil && *fixture.Snapshot.ImmutablePolicy && fixture.Snapshot.Repository != "" && fixture.Snapshot.LineageIdentity != "" && artifactOK && fixture.Plan.RequestedVersion != "" {
		if !hasExpectedPatchAdvance(fixture.Plan.RequestedVersion, fixture.Plan.Lineage.CurrentVersion, fixture.Snapshot) { addViolation("", "VERSION_NOT_NEXT_PATCH", "requested version is not the next unused patch version") }
	}
	if fixture.Plan.Lineage.CurrentVersion == fixture.Plan.RequestedVersion && fixture.Plan.RequestedVersion != "" && len(fixture.Snapshot.Releases) > 0 { addViolation("", "SAME_VERSION_REUSE", "same version cannot be recreated after a public artifact exists") }

	decision.Unknown, decision.Violations = unknowns, violations
	decision.State, decision.Action, decision.Reason = StateClosed, ActionProceed, "read-only plan satisfies the explicit FIXED_POINT release lineage policy"
	if len(unknowns) > 0 { decision.State, decision.Action, decision.Reason = StateUnknown, ActionHold, "required policy, snapshot, lineage, or external evidence is missing; no mutation is authorized" }
	if len(violations) > 0 { decision.State, decision.Action, decision.Reason = StateRefuted, ActionPreserveAdvance, "proposed release mutation is REFUTED; retain public history and advance by patch version" }
	decision.Activities = activityDecisions(model.Activities, decision.State, decision.Reason)
	if decision.State == StateRefuted {
		decision.PreservationRule = PreservationRule{OnRefuted: "retain_failed_public_artifact_as_OPERATIONAL_REFUTED", PublicHistory: "preserve", FailedArtifact: "retain", NeverDelete: true}
		decision.VersionAdvance = VersionAdvance{FromVersion: fixture.Plan.RequestedVersion, ToVersion: nextPatch(fixture.Plan.RequestedVersion), Reason: "failed public output is preserved; a new patch version is the only forward path", Required: true}
	}
	if caseSpec.Class == CaseClosed && decision.State != StateClosed { return Decision{}, fmt.Errorf("fixture %q is declared CLOSED but evaluates to %s", fixture.CaseID, decision.State) }
	if caseSpec.Class == CaseUnknown && decision.State != StateUnknown { return Decision{}, fmt.Errorf("fixture %q is declared UNKNOWN but evaluates to %s", fixture.CaseID, decision.State) }
	if caseSpec.Class == CaseRefuted && decision.State != StateRefuted { return Decision{}, fmt.Errorf("fixture %q is declared REFUTED but evaluates to %s", fixture.CaseID, decision.State) }
	return decision, nil
}

func activityDecisions(activities []ActivitySpec, state State, reason string) []ActivityDecision { result := make([]ActivityDecision, 0, len(activities)); for _, item := range activities { result = append(result, ActivityDecision{ID: item.ID, MetaName: item.MetaName, Proof: item.Proof, Indicator: item.Indicator, State: state, Reason: reason}) }; return result }
func valueOrFalse(value *bool) bool { return value != nil && *value }
func mainArtifact(items []PublicArtifact) (PublicArtifact, bool) { var result PublicArtifact; count := 0; for _, item := range items { if item.Main { result = item; count++ } }; return result, count == 1 && result.Name != "" && normalizeDigest(result.SHA256) != "" }
func containsOperation(items []OperationRule, kind string) bool { for _, item := range items { if item.Kind == kind { return true } }; return false }
func existingTag(items []RemoteTag, name string) bool { for _, item := range items { if item.Name == name { return true } }; return false }
func existingRelease(items []RemoteRelease, tag string) bool { for _, item := range items { if item.Tag == tag || item.Version == tag { return true } }; return false }
func normalizeDigest(value string) string { return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "sha256:") }
func forbiddenReason(kind string) string { switch kind { case "release_delete": return "release deletion would erase public history"; case "tag_delete": return "tag deletion would erase the public ref"; case "tag_retag": return "retagging would move an established ref"; case "same_version_recreate": return "same-version recreation is lineage reuse"; case "published_asset_replace": return "published assets are immutable"; case "failed_run_delete": return "failed runs are evidence and must not be deleted"; default: return "operation is forbidden by the semantic source" } }

type semanticVersion struct { major, minor, patch int }
func parseVersion(value string) (semanticVersion, bool) { matches := versionPattern.FindStringSubmatch(value); if len(matches) != 4 { return semanticVersion{}, false }; var result semanticVersion; for index, target := range []*int{&result.major, &result.minor, &result.patch} { parsed, ok := atoi(matches[index+1]); if !ok { return semanticVersion{}, false }; *target = parsed }; return result, true }
func atoi(value string) (int, bool) { parsed := 0; for _, character := range value { if character < '0' || character > '9' { return 0, false }; parsed = parsed*10 + int(character-'0') }; return parsed, true }
func nextPatch(value string) string { parsed, ok := parseVersion(value); if !ok { return "v0.1.0" }; return fmt.Sprintf("v%d.%d.%d", parsed.major, parsed.minor, parsed.patch+1) }
func hasExpectedPatchAdvance(requested, current string, snapshot RemoteSnapshot) bool {
	requestedVersion, requestedOK := parseVersion(requested); if !requestedOK { return false }
	versions := make([]semanticVersion, 0, len(snapshot.Releases)+len(snapshot.Tags)+1)
	if current != "" { if value, ok := parseVersion(current); ok { versions = append(versions, value) } }
	for _, release := range snapshot.Releases { if value, ok := parseVersion(release.Version); ok { versions = append(versions, value) } }
	for _, tag := range snapshot.Tags { if value, ok := parseVersion(tag.Name); ok { versions = append(versions, value) } }
	if len(versions) == 0 { return requestedVersion.major == 0 && requestedVersion.minor == 1 && requestedVersion.patch == 0 }
	sort.Slice(versions, func(left, right int) bool { if versions[left].major != versions[right].major { return versions[left].major < versions[right].major }; if versions[left].minor != versions[right].minor { return versions[left].minor < versions[right].minor }; return versions[left].patch < versions[right].patch })
	latest := versions[len(versions)-1]
	return requestedVersion.major == latest.major && requestedVersion.minor == latest.minor && requestedVersion.patch == latest.patch+1
}
