package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/clivm/clivm/pkg/config/aqua"
	"github.com/clivm/clivm/pkg/config/registry"
	"github.com/invopop/jsonschema"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	if err := gen(&aqua.Config{}, "json-schema/aqua-yaml.json"); err != nil {
		return err
	}
	if err := gen(&registry.Config{}, "json-schema/registry.json"); err != nil {
		return err
	}
	return nil
}

func gen(input interface{}, p string) error {
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("create a file %s: %w", p, err)
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	s := jsonschema.Reflect(input)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("mashal schema as JSON: %w", err)
	}
	if err := os.WriteFile(p, []byte(strings.ReplaceAll(string(b), "http://json-schema.org", "https://json-schema.org")+"\n"), 0o644); err != nil { //nolint:gosec,gomnd
		return fmt.Errorf("write JSON Schema to %s: %w", p, err)
	}
	return nil
}
