package main

import (
	"fmt"
	"log"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	genrgst "github.com/aquaproj/aqua/v2/pkg/controller/generate-registry"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/suzuki-shunsuke/gen-go-jsonschema/jsonschema"
)

func main() {
	err := core()
	if err != nil {
		log.Fatal(err)
	}
}

func core() error {
	err := jsonschema.Write(&aqua.Config{}, "json-schema/aqua-yaml.json")
	if err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}

	err := jsonschema.Write(&registry.Config{}, "json-schema/registry.json")
	if err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}

	err := jsonschema.Write(&policy.ConfigYAML{}, "json-schema/policy.json")
	if err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}

	err := jsonschema.Write(&genrgst.RawConfig{}, "json-schema/aqua-generate-registry.json")
	if err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}

	return nil
}
