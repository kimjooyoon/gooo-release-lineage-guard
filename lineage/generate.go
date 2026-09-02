package lineage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GeneratedCase struct {
	CaseID      string `json:"case_id"`
	Class       CaseClass `json:"class"`
	DecisionFile string `json:"decision_file"`
	DossierFile  string `json:"dossier_file"`
}

type GeneratedManifest struct {
	Schema string `json:"schema"`
	Source string `json:"source"`
	Cases  []GeneratedCase `json:"cases"`
	CI     string `json:"ci_command_descriptor"`
}

func GenerateArtifacts(model SemanticModel, sourcePath, fixturesDir, outputDir string) (GeneratedManifest, error) {
	entries, err := os.ReadDir(fixturesDir)
	if err != nil { return GeneratedManifest{}, fmt.Errorf("read fixture directory: %w", err) }
	casePaths := make([]string, 0)
	for _, entry := range entries { if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") { casePaths = append(casePaths, filepath.Join(fixturesDir, entry.Name())) } }
	sort.Strings(casePaths)
	if len(casePaths) != len(model.Cases) { return GeneratedManifest{}, fmt.Errorf("fixture directory has %d JSON cases; semantic source declares %d", len(casePaths), len(model.Cases)) }
	manifest := GeneratedManifest{Schema: "gooo/release-lineage-guard/generated-manifest/v1", Source: sourcePath, CI: filepath.Join(outputDir, "ci", "plan-gate-command.json")}
	seen := map[string]bool{}
	for _, fixturePath := range casePaths {
		fixture, err := ReadFixture(fixturePath); if err != nil { return GeneratedManifest{}, err }
		if seen[fixture.CaseID] { return GeneratedManifest{}, fmt.Errorf("duplicate fixture case %q", fixture.CaseID) }; seen[fixture.CaseID] = true
		decision, err := Evaluate(model, fixture, sourcePath, fixturePath); if err != nil { return GeneratedManifest{}, err }
		decisionPath := filepath.Join(outputDir, "decisions", fixture.CaseID+".json")
		dossierPath := filepath.Join(outputDir, "dossiers", fixture.CaseID+".md")
		if err := WriteJSON(decisionPath, decision); err != nil { return GeneratedManifest{}, err }
		if err := os.MkdirAll(filepath.Dir(dossierPath), 0o755); err != nil { return GeneratedManifest{}, fmt.Errorf("create dossier directory: %w", err) }
		if err := os.WriteFile(dossierPath, []byte(Dossier(decision, fixture.Description)), 0o644); err != nil { return GeneratedManifest{}, fmt.Errorf("write dossier %s: %w", dossierPath, err) }
		manifest.Cases = append(manifest.Cases, GeneratedCase{CaseID: fixture.CaseID, Class: caseClass(model, fixture.CaseID), DecisionFile: decisionPath, DossierFile: dossierPath})
	}
	if len(seen) != len(model.Cases) { return GeneratedManifest{}, fmt.Errorf("generated cases do not cover all semantic cases") }
	descriptor := CIDescriptor{Schema: CIDescriptorSchema, Command: "go run ./cmd/gooo-release-lineage-guard conformance --source semantic/release-lineage.gooo --fixtures fixtures/cases", Source: sourcePath, FixtureDirectory: fixturesDir, RequiresState: StateClosed, RequiresTerminal: TerminalFixedPoint, ReadOnly: true, Forbidden: SortedOperationKinds(model), Outputs: []string{"machine decision JSON", "human dossier Markdown", "generated/ci/plan-gate-command.json"}}
	ciPath := filepath.Join(outputDir, "ci", "plan-gate-command.json")
	if err := WriteJSON(ciPath, descriptor); err != nil { return GeneratedManifest{}, err }
	manifestPath := filepath.Join(outputDir, "manifest.json")
	if err := WriteJSON(manifestPath, manifest); err != nil { return GeneratedManifest{}, err }
	return manifest, nil
}

func caseClass(model SemanticModel, caseID string) CaseClass { item, _ := CaseSpecFor(model, caseID); return item.Class }

func Dossier(decision Decision, description string) string {
	var builder strings.Builder
	builder.WriteString("# Gooo Release Lineage Guard dossier\n\n")
	builder.WriteString("- Decision schema: `" + decision.Schema + "`\n")
	builder.WriteString("- Case ID: `" + decision.CaseID + "`\n")
	builder.WriteString("- Plan ID: `" + decision.PlanID + "`\n")
	builder.WriteString("- State: `" + string(decision.State) + "`\n")
	builder.WriteString("- Terminal: `" + string(decision.Terminal) + "`\n")
	builder.WriteString("- Action: `" + string(decision.Action) + "`\n")
	builder.WriteString("- Description: " + description + "\n\n")
	builder.WriteString("## Decision\n\n" + decision.Reason + ".\n\n")
	builder.WriteString("## Lineage and artifact identity\n\n")
	builder.WriteString("- Repository: `" + decision.Lineage.Repository + "`\n")
	builder.WriteString("- Stream: `" + decision.Lineage.Stream + "`\n")
	builder.WriteString("- Lineage ID: `" + decision.Lineage.LineageID + "`\n")
	if len(decision.Artifacts) == 0 { builder.WriteString("- Main artifact: missing\n") }
	for _, artifact := range decision.Artifacts { builder.WriteString(fmt.Sprintf("- Artifact `%s`: role `%s`, sha256 `%s`, bytes `%d`, main `%t`\n", artifact.Name, artifact.Role, artifact.SHA256, artifact.Size, artifact.Main)) }
	builder.WriteString("\n## Proposed outcome\n\n")
	if len(decision.Violations) == 0 { builder.WriteString("No forbidden mutation or fixed-point violation was observed.\n\n") } else { builder.WriteString("The public outcome is preserved as `OPERATIONAL_REFUTED`; deletion, retagging, replacement, or same-version recreation is not a repair path.\n\n") }
	for _, violation := range decision.Violations { if violation.OperationID == "" { builder.WriteString(fmt.Sprintf("- `%s`: %s\n", violation.Kind, violation.Reason)) } else { builder.WriteString(fmt.Sprintf("- `%s` `%s`: %s\n", violation.OperationID, violation.Kind, violation.Reason)) } }
	for _, unknown := range decision.Unknown { builder.WriteString(fmt.Sprintf("- UNKNOWN: stage `%s`, step `%s`, reason `%s`, unknown_class `%s`, next_operation `%s`, blocked_by `%s`\n", unknown.Stage, unknown.Step, unknown.Reason, unknown.UnknownClass, unknown.NextOperation, strings.Join(unknown.BlockedBy, ","))) }
	if len(decision.Violations) > 0 || len(decision.Unknown) > 0 { builder.WriteString("\n") }
	builder.WriteString("## Version advance\n\n")
	builder.WriteString(fmt.Sprintf("- From: `%s`\n- To: `%s`\n- Required: `%t`\n- Reason: %s\n\n", decision.VersionAdvance.FromVersion, decision.VersionAdvance.ToVersion, decision.VersionAdvance.Required, decision.VersionAdvance.Reason))
	builder.WriteString("## Runtime boundary\n\n")
	builder.WriteString("The evaluator reads a caller-provided snapshot and proposed operations only. Repository writes, local test executions, cross-project required gates, and GitHub mutations are all `0`; wall and RSS are `null` because no local performance claim is made.\n\n")
	builder.WriteString("## Meta activities\n\n")
	for _, activity := range decision.Activities { builder.WriteString(fmt.Sprintf("- `%s` / `%s` / `%s` / `%s`: `%s` — %s\n", activity.ID, activity.MetaName, activity.Proof, activity.Indicator, activity.State, activity.Reason)) }
	return builder.String()
}

func DecodeJSON(path string, value any) error { contents, err := os.ReadFile(path); if err != nil { return err }; return json.Unmarshal(contents, value) }
