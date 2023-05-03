package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Andrew-Klaas/vault-go-demo/users"
)

func TestIndex(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(users.Index))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("status code = %v; want %v", res.StatusCode, http.StatusOK)
	}

	fmt.Printf("here")
	if content, err := ioutil.ReadAll(res.Body); string(content) == "" {
		t.Errorf("content = %v; want %v", content, "not empty")
	} else if err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(string(content))
	}
}

func TestMain(t *testing.T) {
	t.Run("index", TestIndex)
}
