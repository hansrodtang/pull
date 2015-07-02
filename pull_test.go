package pull

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/github"
)

// Test handling of ping events.
func TestPing(t *testing.T) {

	ts := httptest.NewServer(pingHandler(http.NotFoundHandler()))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("X-GitHub-Event", "ping")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("HTTP Request failed: %s", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusNotImplemented, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	expected := "Received ping"
	actual := string(body)

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}

// Test rejection of unsupported events.
func TestImplements(t *testing.T) {

	ts := httptest.NewServer(implementsHandler(http.NotFoundHandler()))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("X-GitHub-Event", "unsupported")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("HTTP Request failed: %s", err)
	}

	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusNotImplemented, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	expected := "Unsupported event"
	actual := string(body)

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}

// Test handling of HMAC hashes with correct input.
func TestSecret(t *testing.T) {

	secret := "test"
	reqbody := []byte(`{"test":"test"}`)
	hash := "sha1=7c7429c26f63ae28d74597e825c20e1796c167e3"

	expected := "Success"
	ts := httptest.NewServer(secretHandler(secret, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			fmt.Fprint(w, expected)
		})))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBuffer(reqbody))
	req.Header.Set("X-Hub-Signature", hash)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("HTTP Request failed: %s", err)
	}

	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusTeapot, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	actual := string(body)

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}

// Test rejection of no HMAC hash when one is expected.
func TestSecretNoHeader(t *testing.T) {

	ts := httptest.NewServer(secretHandler("secret", http.NotFoundHandler()))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("HTTP Request failed: %s", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusForbidden, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	actual := string(body)
	expected := "X-Hub-Signature required for HMAC verification"

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}

// Test rejection of incorrect HMAC hash.
func TestSecretFailed(t *testing.T) {

	secret := "test"
	reqbody := []byte(`{"test":"test"}`)
	hash := "sha1=mismatchgibberish"

	ts := httptest.NewServer(secretHandler(secret, http.NotFoundHandler()))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBuffer(reqbody))
	req.Header.Set("X-Hub-Signature", hash)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("HTTP Request failed: %s", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusForbidden, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	actual := string(body)
	expected := "HMAC verification failed"

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}

// Test handler with correct input.
func TestMain(t *testing.T) {
	pullrequest := github.PullRequestEvent{
		Action: github.String("opened"),
		Number: github.Int(0),
		Repo: &github.Repository{
			FullName: github.String("test"),
		},
	}
	reqbody, _ := json.Marshal(pullrequest)

	handler := New(Configuration{})
	runner := func(handler *Handler, event github.PullRequestEvent) {}

	ts := httptest.NewServer(mainHandler(handler, runner))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBuffer(reqbody))
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("HTTP Request failed: %s", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusAccepted, resp.StatusCode)
	}
}

// Test rejection of malformed JSON
func TestMainBadJSON(t *testing.T) {
	handler := New(Configuration{})

	runner := func(handler *Handler, event github.PullRequestEvent) {}

	ts := httptest.NewServer(mainHandler(handler, runner))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBuffer([]byte{0}))
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("HTTP Request failed: %s", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected HTTP %d, got HTTP %d", http.StatusBadRequest, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	actual := string(body)
	expected := "Bad JSON"

	if actual != expected {
		t.Errorf("Expected %s got %s", expected, actual)
	}
}
