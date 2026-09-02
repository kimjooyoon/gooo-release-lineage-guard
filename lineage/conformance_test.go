package lineage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok { t.Fatal("runtime caller unavailable") }
	return filepath.Dir(filepath.Dir(file))
}

func TestSemanticAuthorityIsFixed(t *testing.T) {
	root := repositoryRoot(t)
	model, err := ReadSource(filepath.Join(root, "semantic", "release-lineage.gooo"))
	if err != nil { t.Fatal(err) }
	if len(model.Entities) != 7 || len(model.Activities) != 12 || len(model.Cases) != 12 { t.Fatalf("unexpected fixed authority cardinality: entities=%d activities=%d cases=%d", len(model.Entities), len(model.Activities), len(model.Cases)) }
}

func TestFixtureCorpusHasRequiredStates(t *testing.T) {
	root := repositoryRoot(t)
	model, err := ReadSource(filepath.Join(root, "semantic", "release-lineage.gooo"))
	if err != nil { t.Fatal(err) }
	entries, err := os.ReadDir(filepath.Join(root, "fixtures", "cases")); if err != nil { t.Fatal(err) }
	counts := map[State]int{}
	for _, entry := range entries {
		if entry.IsDir() { continue }
		fixture, err := ReadFixture(filepath.Join(root, "fixtures", "cases", entry.Name())); if err != nil { t.Fatal(err) }
		decision, err := Evaluate(model, fixture, "semantic/release-lineage.gooo", filepath.Join("fixtures/cases", entry.Name())); if err != nil { t.Fatal(err) }
		counts[decision.State]++
	}
	if counts[StateClosed] != 4 || counts[StateUnknown] != 4 || counts[StateRefuted] != 4 { t.Fatalf("state corpus must be 4/4/4, got CLOSED=%d UNKNOWN=%d REFUTED=%d", counts[StateClosed], counts[StateUnknown], counts[StateRefuted]) }
}

func TestCounterexampleIsPreservedAndAdvances(t *testing.T) {
	root := repositoryRoot(t)
	model, err := ReadSource(filepath.Join(root, "semantic", "release-lineage.gooo")); if err != nil { t.Fatal(err) }
	fixture, err := ReadFixture(filepath.Join(root, "fixtures", "cases", "12-refuted-configured-after-publish.json")); if err != nil { t.Fatal(err) }
	decision, err := Evaluate(model, fixture, "semantic/release-lineage.gooo", "fixtures/cases/12-refuted-configured-after-publish.json"); if err != nil { t.Fatal(err) }
	if decision.State != StateRefuted || decision.Action != ActionPreserveAdvance { t.Fatalf("counterexample state/action = %s/%s", decision.State, decision.Action) }
	if decision.PreservationRule.OnRefuted != "retain_failed_public_artifact_as_OPERATIONAL_REFUTED" || decision.VersionAdvance.ToVersion != "v0.1.1" { t.Fatalf("counterexample preservation/advance not enforced") }
}
