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
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := gen(&config.Config{}, encoder); err != nil {
		return err
	}
	if err := gen(&config.RegistryContent{}, encoder); err != nil {
		return err
	}
	return nil
}

func gen(input interface{}, encoder *json.Encoder) error {
	s := jsonschema.Reflect(input)
	if err := encoder.Encode(s); err != nil {
		return fmt.Errorf("encode config as JSON: %w", err)
	}
	return nil
}
