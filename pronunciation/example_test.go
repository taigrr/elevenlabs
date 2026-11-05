package pronunciation_test

import (
	"context"
	"fmt"
	"log"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/pronunciation"
)

func ExampleService_CreateFromRules() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	dict, err := client.PronunciationDictionaries.CreateFromRules(ctx, &pronunciation.CreateFromRulesRequest{
		Name: "My Dictionary",
		Rules: []pronunciation.Rule{
			pronunciation.AliasRule("Thailand", "tie-land"),
			pronunciation.PhonemeRule("tomato", "təˈmeɪtoʊ", "ipa"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dict.ID, dict.VersionID)
}
