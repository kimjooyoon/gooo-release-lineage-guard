package lineage

import "time"

const (
	SourceSchema  = "gooo/release-lineage-guard/source/v1"
	PlanSchema    = "gooo/release-lineage-guard/plan/v1"
	SnapshotSchema = "gooo/release-lineage-guard/remote-snapshot/v1"
	FixtureSchema = "gooo/release-lineage-guard/fixture/v1"
	DecisionSchema = "gooo/release-lineage-guard/decision/v1"
	DossierSchema = "gooo/release-lineage-guard/human-dossier/v1"
	CIDescriptorSchema = "gooo/release-lineage-guard/ci-command/v1"

	StateClosed  State = "CLOSED"
	StateUnknown State = "UNKNOWN"
	StateRefuted State = "REFUTED"

	ActionProceed         Action = "PROCEED"
	ActionHold            Action = "HOLD"
	ActionPreserveAdvance Action = "PRESERVE_AND_ADVANCE"

	TerminalFixedPoint Terminal = "FIXED_POINT"

	ProofFoundation ProofFamily = "FOUNDATION"
	ProofCoherence  ProofFamily = "COHERENCE"
	ProofRegression ProofFamily = "REGRESSION"

	IndicatorDriver    IndicatorClass = "DRIVER"
	IndicatorOutcome   IndicatorClass = "OUTCOME"
	IndicatorGuardrail IndicatorClass = "GUARDRAIL"

	CaseClosed  CaseClass = "CLOSED"
	CaseUnknown CaseClass = "UNKNOWN"
	CaseRefuted CaseClass = "REFUTED"
)

type State string
type Action string
type Terminal string
type ProofFamily string
type IndicatorClass string
type CaseClass string

type FieldSpec struct {
	Entity string `json:"entity"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
}

type EntitySpec struct {
	Name   string      `json:"name"`
	Fields []FieldSpec `json:"fields"`
}

type ActivitySpec struct {
	ID        string        `json:"id"`
	MetaName  string        `json:"meta_name"`
	Proof     ProofFamily   `json:"proof"`
	Indicator IndicatorClass `json:"indicator"`
}

type OperationRule struct {
	Kind string `json:"kind"`
}

type AuthorityContract struct {
	RepositoryWrites          int `json:"repository_writes"`
	LocalTestExecutions       int `json:"local_test_executions"`
	CrossProjectRequiredGates int `json:"cross_project_required_gates"`
}

type ExternalEvidenceContract struct {
	Utility     State `json:"utility"`
	Improvement State `json:"improvement"`
}

type CaseSpec struct {
	ID    string    `json:"id"`
	Class CaseClass `json:"class"`
}

type SemanticModel struct {
	Schema            string                    `json:"schema"`
	Name              string                    `json:"name"`
	Version           string                    `json:"version"`
	Entities          []EntitySpec              `json:"entities"`
	Activities        []ActivitySpec            `json:"activities"`
	AllowedOperations []OperationRule           `json:"allowed_operations"`
	ForbiddenOperations []OperationRule         `json:"forbidden_operations"`
	StatePrecedence   []State                    `json:"state_precedence"`
	Terminal          Terminal                   `json:"terminal"`
	UnknownFields     []string                   `json:"unknown_fields"`
	Authority         AuthorityContract          `json:"authority"`
	ExternalEvidence  ExternalEvidenceContract   `json:"external_evidence"`
	Cases             []CaseSpec                 `json:"cases"`
}

type ReleaseLineage struct {
	Repository     string `json:"repository"`
	Stream         string `json:"stream"`
	LineageID      string `json:"lineage_id"`
	CurrentVersion string `json:"current_version"`
	CurrentTag     string `json:"current_tag"`
	CurrentCommit  string `json:"current_commit"`
}

type PublicArtifact struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	SHA256 string `json:"sha256"`
	Main   bool   `json:"main"`
	Size   int64  `json:"size"`
}

type ExternalEvidence struct {
	Source string `json:"source"`
	Digest string `json:"digest"`
}

type MutationPlan struct {
	Schema          string                      `json:"schema"`
	PlanID          string                      `json:"plan_id"`
	Lineage         ReleaseLineage              `json:"lineage"`
	RequestedVersion string                     `json:"requested_version"`
	TargetCommit    string                      `json:"target_commit"`
	Artifacts       []PublicArtifact            `json:"artifacts"`
	ProposedOperations []ProposedOperation      `json:"proposed_operations"`
	ExternalClaims  []string                    `json:"external_claims"`
	ExternalEvidence map[string]ExternalEvidence `json:"external_evidence"`
}

type ProposedOperation struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Tag          string `json:"tag"`
	Version      string `json:"version"`
	ReleaseID    string `json:"release_id"`
	ArtifactName string `json:"artifact_name"`
	Digest       string `json:"digest"`
	TargetCommit string `json:"target_commit"`
	Annotated    bool   `json:"annotated"`
	Draft        bool   `json:"draft"`
}

type RemoteTag struct {
	Name         string `json:"name"`
	ObjectID     string `json:"object_id"`
	Annotated    bool   `json:"annotated"`
}

type RemoteAsset struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

type RemoteRelease struct {
	ID           string        `json:"id"`
	Tag          string        `json:"tag"`
	Version      string        `json:"version"`
	Published    bool          `json:"published"`
	Immutable    bool          `json:"immutable"`
	Assets       []RemoteAsset `json:"assets"`
}

type RemoteSnapshot struct {
	Schema           string          `json:"schema"`
	Repository       string          `json:"repository"`
	DefaultBranch    string          `json:"default_branch"`
	HeadCommit       string          `json:"head_commit"`
	ImmutablePolicy  *bool           `json:"immutable_policy"`
	LineageIdentity  string          `json:"lineage_identity"`
	InventoryComplete *bool          `json:"inventory_complete"`
	Tags             []RemoteTag     `json:"tags"`
	Releases         []RemoteRelease `json:"releases"`
	CapturedAt       string          `json:"captured_at"`
}

type Fixture struct {
	Schema      string         `json:"schema"`
	CaseID      string         `json:"case_id"`
	Description string         `json:"description"`
	Plan        MutationPlan   `json:"plan"`
	Snapshot    RemoteSnapshot `json:"snapshot"`
}

type UnknownDetail struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Violation struct {
	OperationID string `json:"operation_id,omitempty"`
	Kind        string `json:"kind"`
	Reason      string `json:"reason"`
}

type ActivityDecision struct {
	ID        string        `json:"id"`
	MetaName  string        `json:"meta_name"`
	Proof     ProofFamily   `json:"proof"`
	Indicator IndicatorClass `json:"indicator"`
	State     State         `json:"state"`
	Reason    string        `json:"reason"`
}

type PreservationRule struct {
	OnRefuted      string `json:"on_refuted"`
	PublicHistory  string `json:"public_history"`
	FailedArtifact string `json:"failed_artifact"`
	NeverDelete    bool   `json:"never_delete"`
}

type VersionAdvance struct {
	FromVersion string `json:"from_version"`
	ToVersion   string `json:"to_version"`
	Reason      string `json:"reason"`
	Required    bool   `json:"required"`
}

type RuntimeBoundary struct {
	RepositoryWrites          int    `json:"repository_writes"`
	LocalTestExecutions       int    `json:"local_test_executions"`
	CrossProjectRequiredGates int    `json:"cross_project_required_gates"`
	GitHubMutations           int    `json:"github_mutations"`
	ReadOnlySnapshot          bool   `json:"read_only_snapshot"`
	WallMS                    *int64 `json:"wall_ms"`
	PeakRSSKiB                *int64 `json:"peak_rss_kib"`
}

type Decision struct {
	Schema          string             `json:"schema"`
	CaseID          string             `json:"case_id"`
	PlanID          string             `json:"plan_id"`
	State           State              `json:"state"`
	Terminal        Terminal           `json:"terminal"`
	Action          Action             `json:"action"`
	Reason          string             `json:"reason"`
	PreservationRule PreservationRule  `json:"preservation_rule"`
	VersionAdvance  VersionAdvance     `json:"version_advance"`
	Unknown         []UnknownDetail    `json:"unknown,omitempty"`
	Violations      []Violation        `json:"violations,omitempty"`
	Activities      []ActivityDecision `json:"activities"`
	Runtime         RuntimeBoundary    `json:"runtime"`
	Source          string             `json:"source"`
	Snapshot        string             `json:"snapshot"`
	Lineage         ReleaseLineage     `json:"lineage"`
	Artifacts       []PublicArtifact   `json:"artifacts"`
}

type CIDescriptor struct {
	Schema          string   `json:"schema"`
	Command         string   `json:"command"`
	Source          string   `json:"source"`
	FixtureDirectory string  `json:"fixture_directory"`
	RequiresState   State    `json:"requires_state"`
	RequiresTerminal Terminal `json:"requires_terminal"`
	ReadOnly        bool    `json:"read_only"`
	Forbidden       []string `json:"forbidden_operations"`
	Outputs         []string `json:"outputs"`
}

func NowUTC() string { return time.Now().UTC().Format(time.RFC3339) }
