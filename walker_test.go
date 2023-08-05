package walker

import (
	"context"
	"strings"
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
		t.Error("Walker retuned nil error after context cancellation") // context cancellation should raise error
	}
}

func TestFilterDirs(t *testing.T) {
	w := New().WithFilters(DiscardDirs())

	files := w.Walk(context.Background(), testPath)

	for f := range files {
		// t.Log(f.Path)
		if f.DirEntry.IsDir() {
			t.Error("DiscardDirs  filter fails")
		}
	}
	if err := w.Err(); err != nil {
		t.Error(err)
	}
}

func TestFilterRegular(t *testing.T) {
	w := New().WithFilters(DiscardRegular())

	files := w.Walk(context.Background(), testPath)

	for f := range files {
		if f.DirEntry.Type().IsRegular() {
			t.Error("DiscardRegular filter fails")
		}
	}
	if err := w.Err(); err != nil {
		t.Error(err)
	}
}

func TestSkipDirs(t *testing.T) {
	w := New().WithFilters(SkipDirs("b", "c/ert"))
	files := w.Walk(context.Background(), testPath)

	for f := range files {
		if strings.Contains(f.Path, "b") || strings.Contains(f.Path, "c/ert") {
			t.Errorf("SkipDir filter failed for %s", f.Path)
		}
	}
	if err := w.Err(); err != nil {
		t.Error(err)
	}
}

func TestSkipAll(t *testing.T) {
	w := New().WithFilters(
		func(Info) Action {
			return SkipAll
		})
	var count int
	files := w.Walk(context.Background(), testPath)
	for range files {
		count++
	}
	if count > 0 {
		t.Errorf("SkipAll failed, found %d files", count)
	}
}

func TestTwoFilters(t *testing.T) {
	w := New().WithFilters(DiscardRegular(), DiscardDirs())

	files := w.Walk(context.Background(), testPath)

	for f := range files {
		t.Log(f)
		if f.DirEntry.Type().IsRegular() || f.DirEntry.IsDir() {
			t.Error("Two filters fail")
		}
	}
}

func TestSetupWithZeroAndNilValues(t *testing.T) {
	readAll := func(c <-chan Info) {
		for range c {
			continue
		}
	}

	w := New().WithFilters()
	infos := w.Walk(context.Background(), testPath)
	readAll(infos)
	if err := w.Err(); err != nil {
		t.Error(err)
	}

	w = New().WithFilters(nil)
	infos = w.Walk(context.Background(), testPath)
	readAll(infos)
	if err := w.Err(); err != nil {
		t.Error(err)
	}

	w = New().WithFilters([]FilterFunc{}...)
	infos = w.Walk(context.Background(), testPath)
	readAll(infos)
	if err := w.Err(); err != nil {
		t.Error(err)
	}
}
