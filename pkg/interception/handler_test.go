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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHandlerProcessingBodySuccessfully(t *testing.T) {
	bodyTests := []struct {
		name    string
		payload io.ReadCloser
		headers map[string]string
		want    []byte
		secret  *corev1.Secret
	}{
		{
			name:    "flattened body no prefix header",
			headers: map[string]string{"Content-Type": mimeForm},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1=value1&field2=value2`)),
			want:    []byte(`{"intercepted":{"field1":"value1","field2":"value2"}}`),
		},
		{
			name:    "flattened body with prefix header",
			headers: map[string]string{"Content-Type": mimeForm, prefixHeader: "slack"},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1=value1&field2=value2`)),
			want:    []byte(`{"slack":{"field1":"value1","field2":"value2"}}`),
		},
		{
			name:    "unflattened body no prefix header",
			headers: map[string]string{"Content-Type": mimeForm, noFlattenHeader: "true"},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1=value1&field2=value2`)),
			want:    []byte(`{"intercepted":{"field1":["value1"],"field2":["value2"]}}`),
		},
		{
			name:    "special payload parsing",
			headers: map[string]string{"Content-Type": mimeForm, extractPayloadHeader: "true"},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1=value1&payload={"field2":"value2","field3":["value3a","value3b"]}`)),
			want:    []byte(`{"intercepted":{"field2":"value2","field3":["value3a","value3b"]}}`),
		},
		{
			name:    "valid secret",
			headers: map[string]string{"Content-Type": mimeForm},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1=value1&field2=value2`)),
			want:    []byte(`{"intercepted":{"field1":"value1","field2":"value2"}}`),
			secret:  makeSecret("secret-token"),
		},
	}

	for _, tt := range bodyTests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("POST", "/", tt.payload)
			for k, v := range tt.headers {
				r.Header.Add(k, v)
			}
			w := httptest.NewRecorder()

			makeHandler().ServeHTTP(w, r)

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
		})
	}
}

func TestHandlerProcessingBodyWithErrors(t *testing.T) {
	bodyTests := []struct {
		name    string
		payload io.ReadCloser
		headers map[string]string
		wantErr string
	}{
		{
			name:    "bad form data",
			headers: map[string]string{"Content-Type": mimeForm},
			payload: ioutil.NopCloser(bytes.NewBufferString(`field1%%%====`)),
			wantErr: "failed to parse form data",
		},
	}

	for _, tt := range bodyTests {
		r, _ := http.NewRequest("POST", "/", tt.payload)
		for k, v := range tt.headers {
			r.Header.Add(k, v)
		}
		w := httptest.NewRecorder()

		makeHandler().ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("unexpected status code, got %d, wanted %d", resp.StatusCode, http.StatusInternalServerError)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type incorrect, got %s, wanted %s", ct, "application/json")
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		assertErrorResponse(t, respBody, tt.wantErr)
	}
}

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

func makeHandler(o ...runtime.Object) *SlackHandler {
	fakeClient := fake.NewSimpleClientset()
	return &SlackHandler{kubeClient: fakeClient}
}

func makeSecret(token string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tekton-ci-auth",
			Namespace: "testing",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"token": []byte(token),
		},
	}
}
