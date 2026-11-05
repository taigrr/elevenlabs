package webhooks_test

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/taigrr/elevenlabs/v2/webhooks"
)

func ExampleConstructEvent() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		event, err := webhooks.ConstructEvent(payload, r.Header.Get(webhooks.SignatureHeader), os.Getenv("WEBHOOK_SECRET"))
		if err != nil {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		fmt.Println(event.Type)
	})
	mux := http.NewServeMux()
	mux.Handle("/webhooks/elevenlabs", handler)
}
