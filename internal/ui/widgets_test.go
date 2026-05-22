package ui

import (
	"reflect"
	"testing"
)

func TestSplitMessageLinesPreservesVisibleBlankLines(t *testing.T) {
	got := splitMessageLines("First line\n\nThird line")
	want := []string{"First line", " ", "Third line"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("message lines mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestSplitMessageLinesHandlesSingleLineMessage(t *testing.T) {
	got := splitMessageLines("Help text")
	want := []string{"Help text"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("message lines mismatch\n got: %#v\nwant: %#v", got, want)
	}
}
