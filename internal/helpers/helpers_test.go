package helpers

import "testing"

func TestGetTitleFromUrl(t *testing.T) {
	title, err := GetTitleFromUrl("https://go.dev")
	want := "The Go Programming Language"

	if err != nil {
		t.Fatalf("got an error but didn't want one: %q", err)
	}

	if title != want {
		t.Fatalf("got %s, want %s", title, want)
	}
}

func TestGetTitleFromUrl_Error(t *testing.T) {
	_, err := GetTitleFromUrl("foobar")

	if err == nil {
		t.Fatal("did not got an error but wanted one")
	}
}
