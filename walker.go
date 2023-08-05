package walker

import (
	"context"
	"io/fs"
	"os"
)

// Info represents and entry located by fs.WalkDir
// Path is always relative to base path supplied to Walker
// for example if Walk is called with path == "/tmp"
// the returned entries will not contain "/tmp" in Path field.
type Info struct {
	Path     string
	DirEntry fs.DirEntry
}

// Walker crawls the specified path using fs.WalkDir and sends all encountered
// items as Info to the channel returned by Walk method.
type Walker interface {
	// Walk runs fs.WalkDir under the hood and returns chan of Info.
	// All items found by fs.WalkDir are returned as is. Basically
	// Walk is a custom fs.WalkDirFunc and a wrapper around WalkDir intended
	// to send result to a channel. It also computes lazily waiting for reader
	// to pull results from the channel as opposed to fs.WalkDir (fire and forget,
	// then wait for all recursive calls to finish).
	// Walker should not be used concurrently. Create new instance of Walker instead.
	// However, concurrent reads from output chan are ok.
	Walk(ctx context.Context, path string) <-chan Info

	// WithFilters adds FilterFuncs to an existing Walker and returns instance of Walker.
	// FilerFuncs are run for each entry Walker finds in the same order as passed to this method.
	// When one of funcs returns Action not equal to Pass this action is applied and other
	// checks are not perfomed.
	WithFilters(funcs ...FilterFunc) Walker

	// Err explains why chan returned by Walk was closed. If context was
	// cancelled or context deadline exceeded or whatewer error the
	// WalkDirFunc function encountered it is returned by this method.
	// Err will return nil if chan Info has not yet been closed.
	Err() error
}

func New() Walker {
	return newWalker()
}

// here comes the implementation

type walker struct {
	filters []FilterFunc
	err     error
}

func newWalker() Walker {
	return &walker{}
}

func (w *walker) Err() error {
	return walkerErr(w)
}

func (w *walker) Walk(ctx context.Context, path string) <-chan Info {
	return walkerWalk(ctx, w, path, w.filters...)
}

func (w *walker) WithFilters(funcs ...FilterFunc) Walker {
	for _, f := range funcs {
		if f != nil {
			w.filters = append(w.filters, f)
		}
	}
	return w
}

func walkerWalk(ctx context.Context, w *walker, path string, funcs ...FilterFunc) <-chan Info {
	out := make(chan Info)
	go func(c chan Info) {
		defer close(c)
		err := fs.WalkDir(os.DirFS(path), ".", newWalkDirFunc(ctx, w, c))
		w.err = err
	}(out)
	return out
}

func walkerErr(w *walker) error {
	return w.err
}

func newWalkDirFunc(ctx context.Context, w *walker, infoCh chan Info) fs.WalkDirFunc {
	return func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// if stat/io error occurs then
			// fs.WalkDir calls WalkDirFunc with
			// this stat error
			// nothing to do here
			// but return, since d is nil in this case.
			return err
		}

		info := Info{
			Path:     p,
			DirEntry: d,
		}

		if len(w.filters) > 0 {
			a := applyAllFilters(info, w.filters)
			switch a {
			case Discard:
				return nil
			case SkipDir:
				return fs.SkipDir
			case SkipAll:
				return fs.SkipAll
			}
		}

		select {
		case <-ctx.Done():
			err = ctx.Err()
		case infoCh <- info:
			err = nil
		}
		return err
	}
}

// applyAllFilters actually runs all tests. It returns on encountering
// firts Action not equal to Pass.
func applyAllFilters(info Info, funcs []FilterFunc) Action {
	var a Action
	for _, f := range funcs {
		a = f(info)
		if a != Pass {
			break
		}
	}
	return a
}
