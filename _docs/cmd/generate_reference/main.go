package main

import (
	"fmt"
	"log"

	"github.com/bitrise-io/bitrise/models"
	// "github.com/google/jsonschema-go/jsonschema"
	"github.com/invopop/jsonschema"
)

func main() {
	// schema, err := jsonschema.For[models.BitriseDataModel](nil)
	// if err != nil {
	// 	log.Fatalf("generate schema: %v", err)
	// }

	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", "  ")
	// if err := enc.Encode(schema); err != nil {
	// 	log.Fatalf("encode schema: %v", err)
	// }

	schema := jsonschema.Reflect(&models.BitriseDataModel{})
	j, err := schema.MarshalJSON()
	if err != nil {
		log.Fatalf("generate schema: %v", err)
	}
	fmt.Println(string(j))
}
