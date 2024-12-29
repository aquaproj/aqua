package main

import (
	"fmt"
	"log"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/suzuki-shunsuke/gen-go-jsonschema/jsonschema"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	if err := jsonschema.Write(&aqua.Config{}, "json-schema/aqua-yaml.json"); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	if err := jsonschema.Write(&registry.Config{}, "json-schema/registry.json"); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	if err := jsonschema.Write(&policy.ConfigYAML{}, "json-schema/policy.json"); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	return nil
}
