package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/invopop/jsonschema"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	if err := gen(&config.Config{}, "json-schema/aqua-yaml.json"); err != nil {
		return err
	}
	if err := gen(&config.RegistryContent{}, "json-schema/registry.json"); err != nil {
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
	if err := encoder.Encode(s); err != nil {
		return fmt.Errorf("encode config as JSON: %w", err)
	}
	return nil
}
