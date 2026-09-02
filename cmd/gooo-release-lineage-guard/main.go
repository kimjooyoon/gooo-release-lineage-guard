package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-release-lineage-guard/lineage"
)

func main() {
	if len(os.Args) < 2 { fail("command is required: plan-gate, conformance, generate, inventory, or parse") }
	switch os.Args[1] {
	case "plan-gate": planGate(os.Args[2:])
	case "conformance": conformance(os.Args[2:])
	case "generate": generate(os.Args[2:])
	case "inventory": inventory(os.Args[2:])
	case "parse": parse(os.Args[2:])
	default: fail("unknown command %q", os.Args[1])
	}
}

func planGate(args []string) {
	flags := flag.NewFlagSet("plan-gate", flag.ExitOnError); source := flags.String("source", "semantic/release-lineage.gooo", "semantic source"); fixturePath := flags.String("fixture", "", "fixture JSON"); output := flags.String("out", "", "optional decision JSON output"); flags.Parse(args)
	if *fixturePath == "" { fail("--fixture is required") }
	model := mustModel(*source); fixture, err := lineage.ReadFixture(*fixturePath); if err != nil { fail("%v", err) }
	decision, err := lineage.Evaluate(model, fixture, *source, *fixturePath); if err != nil { fail("%v", err) }
	if *output != "" { if err := lineage.WriteJSON(*output, decision); err != nil { fail("%v", err) } } else { printJSON(decision) }
	if decision.State != lineage.StateClosed || decision.Terminal != lineage.TerminalFixedPoint { os.Exit(3) }
}

func conformance(args []string) {
	flags := flag.NewFlagSet("conformance", flag.ExitOnError); source := flags.String("source", "semantic/release-lineage.gooo", "semantic source"); fixtures := flags.String("fixtures", "fixtures/cases", "fixture directory"); output := flags.String("out", "generated", "generated output directory"); flags.Parse(args)
	model := mustModel(*source); manifest, err := lineage.GenerateArtifacts(model, *source, *fixtures, *output); if err != nil { fail("%v", err) }
	printJSON(manifest)
}

func generate(args []string) {
	flags := flag.NewFlagSet("generate", flag.ExitOnError); source := flags.String("source", "semantic/release-lineage.gooo", "semantic source"); fixtures := flags.String("fixtures", "fixtures/cases", "fixture directory"); output := flags.String("out", "generated", "generated output directory"); flags.Parse(args)
	model := mustModel(*source); manifest, err := lineage.GenerateArtifacts(model, *source, *fixtures, *output); if err != nil { fail("%v", err) }; printJSON(manifest)
}

func inventory(args []string) {
	flags := flag.NewFlagSet("inventory", flag.ExitOnError); root := flags.String("root", ".", "repository root"); output := flags.String("out", "", "optional JSON output"); flags.Parse(args)
	value, err := lineage.BuildInventory(*root); if err != nil { fail("%v", err) }
	if *output != "" { if err := lineage.WriteJSON(*output, value); err != nil { fail("%v", err) } } else { printJSON(value) }
}

func parse(args []string) {
	flags := flag.NewFlagSet("parse", flag.ExitOnError); source := flags.String("source", "semantic/release-lineage.gooo", "semantic source"); flags.Parse(args); printJSON(mustModel(*source))
}

func mustModel(path string) lineage.SemanticModel { model, err := lineage.ReadSource(path); if err != nil { fail("%v", err) }; return model }
func printJSON(value any) { contents, err := lineage.MarshalJSON(value); if err != nil { fail("%v", err) }; fmt.Println(string(contents)) }
func fail(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(2) }
