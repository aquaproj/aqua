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
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	opts := &jsonschema.Options{
		ModFile: "go.mod",
	}
	if err := jsonschema.Write(&aqua.Config{}, "json-schema/aqua-yaml.json", opts); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	if err := jsonschema.Write(&registry.Config{}, "json-schema/registry.json", opts); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	if err := jsonschema.Write(&policy.ConfigYAML{}, "json-schema/policy.json", opts); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	if err := jsonschema.Write(&genrgst.RawConfig{}, "json-schema/aqua-generate-registry.json", opts); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	return nil
}
