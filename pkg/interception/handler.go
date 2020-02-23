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

	prefixHeader    = "Slack-Decodeprefix"
	noFlattenHeader = "Slack-Decodenoflatten"
	defaultPrefix   = "intercepted"
)

// TODO validate the shared secret
// TODO: add some logging.

// Handler processes interception requests.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Dirty Hack until this lands https://github.com/tektoncd/triggers/pull/438
	// Tekton Triggers is stripping the method, which means the data isn't being
	// treated as a POST, and so ParseForm() isn't seeing the body.
	r.Method = http.MethodPost

	err := r.ParseForm()
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("failed to parse form data: %s", err))
		return
	}
	w.Header().Set("Content-Type", mimeJSON)
	var data interface{} = flattenMap(r.PostForm)
	if noFlatten(r) {
		data = r.PostForm
	}
	response := map[string]interface{}{prefixFromRequest(r): data}
	payload, err := json.Marshal(response)

	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("failed to marshal form data: %s", err.Error()))
		return
	}
	w.Write(payload)
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
