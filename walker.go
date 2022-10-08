package walker

import (
	"context"
	"io/fs"
	"os"
	"sync"
)

// Info represents and entry located by fs.WalkDir
// Path is always relative to base path supplied to Walker
// for example if Walk is called with path == "/tmp"
// the returned entries will not contain "/tmp" in Path field.
type Info struct {
	Path string
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
	Walk(ctx context.Context, path string, funcs ...FilterFunc) <-chan Info

	// Err explains why chan returned by Walk was closed. If context was 
	// cancelled or context deadline exceeded or whatewer error the 
	// WalkDirFunc function encountered it is returned by this method.
	// Err will return nil if chan Info has not yet been closed. 
	Err() error
}

func New() Walker {
	return newWalker()
}

func (w *walker) Err() error {
	return walkerErr(w)
}

func (w *walker) Walk(ctx context.Context, path string, funcs ...FilterFunc) (<-chan Info)  {
	return walkerWalk(ctx, w, path, funcs...) 
}

// here comes the implementation

type walker struct {
	err error
	mu sync.Mutex
}

func newWalker() Walker {
	return &walker{}
}


func walkerWalk(ctx context.Context, w *walker, path string, funcs ...FilterFunc) (<-chan Info) {
	filesCh := make(chan Info)

	filtered := filter(filesCh, funcs...)
	// filter func will take care of its output
	// chan itself when input chan closes.
	go func() {
		defer close(filesCh)

		err := fs.WalkDir(os.DirFS(path), ".", newWalkDirFunc(ctx, filesCh))

		w.mu.Lock()
		w.err = err
		w.mu.Unlock()
	}()

	return filtered
}

func walkerErr(w *walker) error {
	var err error

	w.mu.Lock()
	err = w.err
	w.mu.Unlock()

	return err
}

func newWalkDirFunc(ctx context.Context, infoCh chan Info) fs.WalkDirFunc {
	return func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// if stat/io error occurs then
			// fs.WalkDir calls WalkDirFunc with
			// this stat error
			// nothing to do here
			// but return, since d is nil in this case.
			return err
		}

		info := Info {
			Path: p,
			DirEntry: d,
		}

		select {
		case <- ctx.Done():
			err = ctx.Err()
		case infoCh <- info:
			err = nil
		}
		return err
	}
}
