package audioeffects_test

import (
	"context"
	"log"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/audioeffects"
)

func ExampleService_GetCatalog() {
	client := elevenlabs.NewClient()
	var catalog *audioeffects.Catalog
	catalog, err := client.AudioEffects.GetCatalog(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, preset := range catalog.FilterPresets {
		log.Println(preset.FilterPresetID, preset.Name)
	}
}
