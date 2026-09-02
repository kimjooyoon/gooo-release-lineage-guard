package lineage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

func ReadSource(path string) (SemanticModel, error) {
	contents, err := os.ReadFile(path)
	if err != nil { return SemanticModel{}, fmt.Errorf("read source %s: %w", path, err) }
	model, err := ParseSource(string(contents))
	if err != nil { return SemanticModel{}, fmt.Errorf("parse source %s: %w", path, err) }
	return model, nil
}

func ReadFixture(path string) (Fixture, error) {
	contents, err := os.ReadFile(path)
	if err != nil { return Fixture{}, fmt.Errorf("read fixture %s: %w", path, err) }
	var fixture Fixture
	if err := json.Unmarshal(contents, &fixture); err != nil { return Fixture{}, fmt.Errorf("decode fixture %s: %w", path, err) }
	return fixture, nil
}

func MarshalJSON(value any) ([]byte, error) { return json.MarshalIndent(value, "", "  ") }

func WriteJSON(path string, value any) error {
	contents, err := MarshalJSON(value)
	if err != nil { return fmt.Errorf("marshal %s: %w", path, err) }
	contents = append(contents, '\n')
	if err := os.MkdirAll(directory(path), 0o755); err != nil { return fmt.Errorf("create output directory: %w", err) }
	if err := os.WriteFile(path, contents, 0o644); err != nil { return fmt.Errorf("write %s: %w", path, err) }
	return nil
}

func FileDigest(path string) (string, int64, error) {
	contents, err := os.ReadFile(path)
	if err != nil { return "", 0, err }
	digest := sha256.Sum256(contents)
	return "sha256:" + hex.EncodeToString(digest[:]), int64(len(contents)), nil
}

func directory(path string) string { index := len(path) - 1; for index >= 0 && path[index] != '/' { index-- }; if index < 0 { return "." }; if index == 0 { return "/" }; return path[:index] }
