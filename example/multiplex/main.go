package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/dmfed/walker/v2"
)

// creates two channels and sorts files depending on file extension
func multiplex(c <-chan walker.Info) (pythonFiles chan string, goFiles chan string) {
	pythonFiles, goFiles = make(chan string), make(chan string)
	go func() {
		defer close(goFiles)
		defer close(pythonFiles)
		for entry := range c {
			switch filepath.Ext(entry.Path) {
			case ".py":
				pythonFiles <- entry.Path
			case ".go":
				goFiles <- entry.Path
			}
		}
	}()

	return pythonFiles, goFiles
}

func main() {
	var (
		src string
	)
	flag.StringVar(&src, "src", "", "source dir")
	flag.Parse()

	// First we'll create a Walker instance.
	w := walker.New().WithFilters(walker.DiscardDirs())

	// Then we tell Walker to crawl our source directory.
	infos := w.Walk(context.Background(), src)

	// Then we'll apply our multiplexer to the results channel.
	// Note that there needs to be a consumer for both channels,
	// otherwise we'll have a deadlock. Also not a very good example
	// since the consumer for one of resulting channels can be much slower
	// which will slow down overall performance if consumer's work is done
	// synchronously on read from channel (not in a goroutine).
	pythonFiles, goFiles := multiplex(infos)

	// so let's add them consumers
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		for d := range pythonFiles {
			fmt.Println("PYTHON FILE:", d)
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		for f := range goFiles {
			fmt.Println("GO FILE:", f)
		}
		wg.Done()
	}()

	wg.Wait()

	// let's check whether Walker encountered errors
	// or finished it's work normally
	if err := w.Err(); err != nil {
		fmt.Println("Walker failed with error:", err)
	}
}
