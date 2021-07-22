package server

import (
	"io"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"time"
	"testing"
)

func TestMethodValidation(t *testing.T) {
	srv := httptest.NewServer(New())
	c := http.Client{}
	
	resp, _ := c.Get(srv.URL + "/set")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected StatusMethodNotAllowed, got %d StatusCode\n", resp.StatusCode)
	}
}

func TestSetParamsValidation(t *testing.T) {
	cases := []string{
		`{"Key": "K", "Ttl": "0"}`,
		`{"Kei": "K", "Value": "V", "Ttl": "100"}`,
		`{"Key": "K", "Value": "V", "Ttl": "-10"}`,
		`{"Key": "Ktl": "0"}`,  // wrong json syntax
	}

	srv := httptest.NewServer(New())
	c := http.Client{}

	for i, cs := range cases {
		resp, _ := c.Post(srv.URL + "/set", "application/json", strings.NewReader(cs))
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("TestCase %d: expected StatusBadRequest, got %d StatusCode\n", i, resp.StatusCode)
		}
	}
}

func TestCorrectScenario(t *testing.T) {
	srv := httptest.NewServer(New())
	c := http.Client{}

	var (
		setUrl = srv.URL + "/set"
		getUrl = srv.URL + "/get"
		delUrl = srv.URL + "/del"
		h = "application/json"
	)
	var (
		resString string
		resInt int
		resSlice []string
		resAny interface{}
	)
	
	// set("K", "V", 1s)
	resp, _ := c.Post(setUrl, h, strings.NewReader(`{"Key": "K", "Value": "V", "Ttl": 1000000000}`))
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	err := json.Unmarshal(body, &resString)
	if err != nil {
		t.Fatalf("Expected string in JSON, got:\n%s", string(body))
	}
	if resString != "OK" {
		t.Fatalf("Expected response: \"OK\", got %s\n", resString)
	}

	// get("K")
	resp, _ = c.Post(getUrl, h, strings.NewReader(`{"Key": "K"}`))
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(body, &resString)
	if err != nil {
		t.Fatalf("Expected string in JSON, got:\n%s", string(body))
	}
	if resString != "V" {
		t.Fatalf("Expected response: \"V\", got %s\n", resString)
	}

	// set("KK", ["a", "b"], 0)
	resp, _ = c.Post(setUrl, h, strings.NewReader(`{"Key": "KK", "Value": ["a", "b"]}`))
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(body, &resString)
	if err != nil {
		t.Fatalf("Expected string in JSON, got:\n%s", string(body))
	}
	if resString != "OK" {
		t.Fatalf("Expected response: \"OK\", got %s\n", resString)
	}

	// get("KK")
	resp, _ = c.Post(getUrl, h, strings.NewReader(`{"Key": "KK"}`))
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(body, &resSlice)
	if err != nil {
		t.Fatalf("Expected string in JSON, got:\n%s", string(body))
	}
	if !reflect.DeepEqual(resSlice, []string{"a", "b"}) {
		t.Fatalf("Expected response: [\"a\", \"b\"], got %s\n", resString)
	}

	// wait expiration of key "K" (first op)
	time.Sleep(time.Second)
	
	// get("K")
	resp, _ = c.Post(getUrl, h, strings.NewReader(`{"Key": "K"}`))
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(body, &resAny)
	if err != nil {
		t.Fatalf("Expected string in JSON, got:\n%s", string(body))
	}
	if resAny != nil {
		t.Fatalf("Expected response: null, got %s\n", resString)
	}

	// del("K", "KK", "KKK")
	resp, _ = c.Post(delUrl, h, strings.NewReader(`{"Keys": ["K", "KK", "KK"]}`))
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	err = json.Unmarshal(body, &resInt)
	if err != nil {
		t.Fatalf("Expected int in JSON, got:\n%s", string(body))
	}
	if resInt != 1 {
		t.Fatalf("Expected response: 1, got %d\n", resInt)
	}

}