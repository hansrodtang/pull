package pull

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {

	ts := httptest.NewServer(pingHandler(http.NotFoundHandler()))
	defer ts.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("X-GitHub-Event", "ping")
	resp, err := client.Do(req)

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

func TestImplements(t *testing.T) {

	ts := httptest.NewServer(implementsHandler(http.NotFoundHandler()))
	defer ts.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("X-GitHub-Event", "unsupported")
	resp, err := client.Do(req)

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

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBuffer(reqbody))
	req.Header.Set("X-Hub-Signature", hash)
	resp, err := client.Do(req)

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

func TestSecretNoHeader(t *testing.T) {

	ts := httptest.NewServer(secretHandler("secret", http.NotFoundHandler()))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := client.Do(req)

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

func TestSecretFailed(t *testing.T) {

	secret := "test"
	reqbody := []byte(`{"test":"test"}`)
	hash := "sha1=mismatchgibberish"

	ts := httptest.NewServer(secretHandler(secret, http.NotFoundHandler()))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBuffer(reqbody))
	req.Header.Set("X-Hub-Signature", hash)
	resp, err := client.Do(req)

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
