package lineage

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseSource(source string) (SemanticModel, error) {
	model := SemanticModel{Schema: SourceSchema}
	seenEntity := map[string]int{}
	seenField := map[string]bool{}
	seenActivity := map[string]bool{}
	seenAllowed := map[string]bool{}
	seenForbidden := map[string]bool{}
	seenCase := map[string]bool{}
	programSeen := false

	for lineNo, raw := range strings.Split(source, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") { continue }
		fields := strings.Fields(line)
		switch fields[0] {
		case "program":
			if programSeen || len(fields) != 3 || fields[1] != "gooo-release-lineage-guard" || fields[2] != "v1" {
				return SemanticModel{}, fmt.Errorf("line %d: invalid program declaration", lineNo+1)
			}
			programSeen = true
			model.Name, model.Version = fields[1], fields[2]
		case "entity":
			if len(fields) != 2 || fields[1] == "" || seenEntity[fields[1]] > 0 {
				return SemanticModel{}, fmt.Errorf("line %d: invalid or duplicate entity", lineNo+1)
			}
			seenEntity[fields[1]] = len(model.Entities) + 1
			model.Entities = append(model.Entities, EntitySpec{Name: fields[1]})
		case "field":
			if len(fields) != 4 || fields[1] == "" || fields[2] == "" || fields[3] == "" {
				return SemanticModel{}, fmt.Errorf("line %d: field requires entity, name, and kind", lineNo+1)
			}
			if seenEntity[fields[1]] == 0 { return SemanticModel{}, fmt.Errorf("line %d: field refers to undeclared entity %q", lineNo+1, fields[1]) }
			key := fields[1] + "." + fields[2]
			if seenField[key] { return SemanticModel{}, fmt.Errorf("line %d: duplicate field %q", lineNo+1, key) }
			seenField[key] = true
			model.Entities[seenEntity[fields[1]]-1].Fields = append(model.Entities[seenEntity[fields[1]]-1].Fields, FieldSpec{Entity: fields[1], Name: fields[2], Kind: fields[3]})
		case "activity":
			if len(fields) != 5 || seenActivity[fields[1]] { return SemanticModel{}, fmt.Errorf("line %d: invalid or duplicate activity", lineNo+1) }
			attrs, err := attributes(fields[2:])
			if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			activity := ActivitySpec{ID: fields[1], MetaName: attrs["meta"], Proof: ProofFamily(attrs["proof"]), Indicator: IndicatorClass(attrs["indicator"])}
			if activity.MetaName == "" || activity.Proof == "" || activity.Indicator == "" { return SemanticModel{}, fmt.Errorf("line %d: activity requires meta, proof, and indicator", lineNo+1) }
			seenActivity[activity.ID] = true
			model.Activities = append(model.Activities, activity)
		case "allow":
			if len(fields) != 2 { return SemanticModel{}, fmt.Errorf("line %d: invalid allow declaration", lineNo+1) }
			attrs, err := attributes(fields[1:]); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			kind := attrs["operation"]; if kind == "" || seenAllowed[kind] { return SemanticModel{}, fmt.Errorf("line %d: invalid or duplicate allowed operation", lineNo+1) }
			seenAllowed[kind] = true; model.AllowedOperations = append(model.AllowedOperations, OperationRule{Kind: kind})
		case "forbid":
			if len(fields) != 2 { return SemanticModel{}, fmt.Errorf("line %d: invalid forbid declaration", lineNo+1) }
			attrs, err := attributes(fields[1:]); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			kind := attrs["operation"]; if kind == "" || seenForbidden[kind] { return SemanticModel{}, fmt.Errorf("line %d: invalid or duplicate forbidden operation", lineNo+1) }
			seenForbidden[kind] = true; model.ForbiddenOperations = append(model.ForbiddenOperations, OperationRule{Kind: kind})
		case "precedence":
			if len(fields) != 2 { return SemanticModel{}, fmt.Errorf("line %d: invalid precedence declaration", lineNo+1) }
			attrs, err := attributes(fields[1:]); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			if model.StatePrecedence != nil { return SemanticModel{}, fmt.Errorf("line %d: duplicate precedence declaration", lineNo+1) }
			for _, state := range strings.Split(attrs["states"], ">") { model.StatePrecedence = append(model.StatePrecedence, State(state)) }
		case "terminal":
			if len(fields) != 2 { return SemanticModel{}, fmt.Errorf("line %d: invalid terminal declaration", lineNo+1) }
			attrs, err := attributes(fields[1:]); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			model.Terminal = Terminal(attrs["value"])
		case "unknown":
			if len(fields) != 2 { return SemanticModel{}, fmt.Errorf("line %d: invalid unknown declaration", lineNo+1) }
			attrs, err := attributes(fields[1:]); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			model.UnknownFields = strings.Split(attrs["fields"], ",")
		case "authority":
			attrs, err := attributes(fields[1:]); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			model.Authority.RepositoryWrites, err = integerAttribute(attrs, "repository_writes"); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			model.Authority.LocalTestExecutions, err = integerAttribute(attrs, "local_test_executions"); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			model.Authority.CrossProjectRequiredGates, err = integerAttribute(attrs, "cross_project_required_gates"); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
		case "external_evidence":
			attrs, err := attributes(fields[1:]); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			model.ExternalEvidence = ExternalEvidenceContract{Utility: State(attrs["utility"]), Improvement: State(attrs["improvement"])}
		case "case":
			if len(fields) != 3 { return SemanticModel{}, fmt.Errorf("line %d: invalid case", lineNo+1) }
			attrs, err := attributes(fields[1:]); if err != nil { return SemanticModel{}, fmt.Errorf("line %d: %w", lineNo+1, err) }
			caseSpec := CaseSpec{ID: attrs["id"], Class: CaseClass(attrs["class"])}
			if caseSpec.ID == "" || caseSpec.Class == "" { return SemanticModel{}, fmt.Errorf("line %d: case requires id and class", lineNo+1) }
			if seenCase[caseSpec.ID] { return SemanticModel{}, fmt.Errorf("line %d: duplicate case %q", lineNo+1, caseSpec.ID) }
			seenCase[caseSpec.ID] = true; model.Cases = append(model.Cases, caseSpec)
		default:
			return SemanticModel{}, fmt.Errorf("line %d: unknown declaration %q", lineNo+1, fields[0])
		}
	}
	if !programSeen { return SemanticModel{}, fmt.Errorf("missing program declaration") }
	if err := ValidateSemanticModel(model); err != nil { return SemanticModel{}, err }
	return model, nil
}

func attributes(fields []string) (map[string]string, error) {
	result := make(map[string]string, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" { return nil, fmt.Errorf("malformed attribute %q", field) }
		if _, exists := result[parts[0]]; exists { return nil, fmt.Errorf("duplicate attribute %q", parts[0]) }
		result[parts[0]] = parts[1]
	}
	return result, nil
}

func integerAttribute(attrs map[string]string, key string) (int, error) {
	value, exists := attrs[key]
	if !exists { return 0, fmt.Errorf("missing attribute %q", key) }
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 { return 0, fmt.Errorf("attribute %q must be a non-negative integer", key) }
	return parsed, nil
}
