package interception

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestHandlerProcessingBodySuccessfully(t *testing.T) {

	bodyTests := []struct {
		name    string
		payload io.ReadCloser
		headers map[string]string
		want    []byte
	}{
		{
			name:    "unflattened body no prefix header",
			headers: map[string]string{"Content-Type": mimeForm},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1=value1&field2=value2`)),
			want:    []byte(`{"intercepted":{"field1":["value1"],"field2":["value2"]}}`),
		},
		{
			name:    "flattened body no prefix header",
			headers: map[string]string{"Content-Type": mimeForm, flattenHeader: "true"},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1=value1&field2=value2`)),
			want:    []byte(`{"intercepted":{"field1":"value1","field2":"value2"}}`),
		},
		{
			name:    "flattened body with prefix header",
			headers: map[string]string{"Content-Type": mimeForm, prefixHeader: "slack", flattenHeader: "true"},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1=value1&field2=value2`)),
			want:    []byte(`{"slack":{"field1":"value1","field2":"value2"}}`),
		},
	}

	for _, tt := range bodyTests {
		r, _ := http.NewRequest("POST", "/", tt.payload)
		for k, v := range tt.headers {
			r.Header.Add(k, v)
		}
		w := httptest.NewRecorder()
		interceptor := makeInterceptor()

		interceptor.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected status code, got %d, wanted %d", resp.StatusCode, http.StatusOK)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type incorrect, got %s, wanted %s", ct, "application/json")
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(respBody, tt.want) {
			t.Errorf("decoded response: got %s, wanted %s\n", respBody, tt.want)
		}
	}
}

// func TestHandlerProcessingBodyWithErrors(t *testing.T) {
// 	bodyTests := []struct {
// 		name    string
// 		payload io.ReadCloser
// 		headers map[string]string
// 		wantErr string
// 	}{
// 		{
// 			name:    "bad form data",
// 			headers: map[string]string{"Content-Type": mimeForm},
// 			payload: ioutil.NopCloser(bytes.NewBufferString(`field1%%%====`)),
// 			wantErr: "failed to parse form data",
// 		},
// 	}

// 	for _, tt := range bodyTests {
// 		r, _ := http.NewRequest("POST", "/", tt.payload)
// 		for k, v := range tt.headers {
// 			r.Header.Add(k, v)
// 		}
// 		w := httptest.NewRecorder()

// 		Handler(w, r)

// 		resp := w.Result()
// 		if resp.StatusCode != http.StatusInternalServerError {
// 			t.Errorf("unexpected status code, got %d, wanted %d", resp.StatusCode, http.StatusInternalServerError)
// 		}
// 		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
// 			t.Errorf("Content-Type incorrect, got %s, wanted %s", ct, "application/json")
// 		}
// 		respBody, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		assertErrorResponse(t, respBody, tt.wantErr)
// 	}
// }

func assertErrorResponse(t *testing.T, body []byte, msg string) {
	var r errorResponse
	err := json.Unmarshal(body, &r)
	if err != nil {
		t.Fatalf("error decoding JSON from response: %s", err)
	}
	match, err := regexp.MatchString(msg, r.Message)
	if err != nil {
		t.Fatalf("error matching with regex %#v against %#v: %s", msg, r.Message, err)
	}
	if !match {
		t.Fatalf("errorResponse did not match %#v against %#v", msg, r.Message)
	}

}

func matchError(t *testing.T, s string, e error) bool {
	t.Helper()
	if s == "" && e == nil {
		return true
	}
	if s != "" && e == nil {
		return false
	}
	match, err := regexp.MatchString(s, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	return match
}

func makeInterceptor(objs ...runtime.Object) *SlackDecoder {
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion)
	cl := fake.NewFakeClient(objs...)
	return &SlackDecoder{
		client: cl,
	}
}
