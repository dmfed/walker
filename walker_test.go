package walker

import (
	"context"
	"testing"
)

var testPath = "./testdata"

func TestWalk(t *testing.T) {
	w := New()

	files := w.Walk(context.Background(), testPath)

	for f := range files {
		// t.Log(f)
		if f.Path == "" {
			t.Error("got empty Path from channel")
		} else if f.DirEntry == nil {
			t.Error("got nil DirEntry from channel")
		}
	}
	
	// range should succeed and resulting error should be nil
	if err := w.Err(); err != nil {
		t.Error("walker returned error for testdata", err)
	}
}

func TestWalkWithContextCancelation(t *testing.T) {
	w := New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	files := w.Walk(ctx, testPath)

	if err := w.Err(); err != nil {
		t.Errorf("Err() returns non-nil error before chan closed")
	}


	if err := w.Err(); err != nil {
		t.Error(err) // err should be nil until files chan is closed
	}

	count := 0
	for f := range files {
		// t.Log(f)
		if f.Path == "" {
			t.Error("got empty Path from channel")
		} else if f.DirEntry == nil {
			t.Error("got nil DirEntry from channel")
		}
		if count == 2 {
			// simulating context cancellation here
			cancel()
		}
		count++
	}

	// err should be non-nil because the context has been cancelled
	if err := w.Err(); err == nil {
		t.Error(err) // context cancellation should raise error
	}
}

func TestFilterDirs(t *testing.T) {
	w := New()

	files := w.Walk(context.Background(), testPath, DiscardDirs)

	for f := range files {
		if f.DirEntry.IsDir() {
			t.Error("DiscardDirs  filter fails")
		}
	}
}

func TestFilterRegular(t *testing.T) {
	w := New()

	files := w.Walk(context.Background(), testPath, DiscardRegular)

	for f := range files {
		if f.DirEntry.Type().IsRegular() {
			t.Error("DiscardRegular filter fails")
		}
	}
}

func TestTwoFilters(t *testing.T) {
	w := New()

	files := w.Walk(context.Background(), testPath, DiscardRegular, DiscardDirs)

	for f := range files {
		t.Log(f)
		if f.DirEntry.Type().IsRegular() || f.DirEntry.IsDir() {
			t.Error("Two filters fail")
		}
	}
}
