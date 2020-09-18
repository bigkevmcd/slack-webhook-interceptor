package interception

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	mimeJSON = "application/json"
	mimeForm = "application/x-www-form-urlencoded"

	prefixHeader         = "Slack-Decodeprefix"
	noFlattenHeader      = "Slack-Decodenoflatten"
	extractPayloadHeader = "Slack-Payload"
	defaultPrefix        = "intercepted"

	// the payload field as defined in interaction messages
	// as per here: https://api.slack.com/messaging/interactivity#understanding_payloads
	slackPayloadField = "payload"
)

// TODO: add some logging.

// Handler processes interception requests.
// The secret is used to authenticate incoming Slack requests.
func MakeHandler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("failed to parse form data: %s", err))
			return
		}
		w.Header().Set("Content-Type", mimeJSON)

		var data interface{}

		switch {
		case payloadExtract(r):
			// mark this string as already json encoded so that it doesn't get encoded again (e.g. quotes escaped out)
			// when it goes through json.Marshal below
			data = json.RawMessage(r.PostForm.Get(slackPayloadField))
		case noFlatten(r):
			data = r.PostForm
		default:
			data = flattenMap(r.PostForm)
		}

		response := map[string]interface{}{prefixFromRequest(r): data}
		payload, err := json.Marshal(response)

		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("failed to marshal form data: %s", err.Error()))
			return
		}

		fmt.Println("Returning response as:")
		fmt.Println(string(payload))

		w.Write(payload)
	}
}

func prefixFromRequest(r *http.Request) string {
	v := r.Header.Get(prefixHeader)
	if v == "" {
		return defaultPrefix
	}
	return v
}

func noFlatten(r *http.Request) bool {
	return strings.ToLower(r.Header.Get(noFlattenHeader)) == "true"
}

func payloadExtract(r *http.Request) bool {
	return strings.ToLower(r.Header.Get(extractPayloadHeader)) == "true"
}

func flattenMap(m url.Values) map[string]string {
	flattened := make(map[string]string)
	for k, v := range m {
		flattened[k] = v[0]
	}
	return flattened
}

type errorResponse struct {
	Message string `json:"message"`
}

func writeErrorResponse(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", mimeJSON)
	w.WriteHeader(http.StatusInternalServerError)
	enc := json.NewEncoder(w)
	err := enc.Encode(errorResponse{Message: msg})
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
