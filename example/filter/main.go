
package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/dmfed/walker"
)

// onlyGoFiles implements walker.FilterFunc
func onlyGoFiles(info walker.Info) bool {
	fileExt := filepath.Ext(info.Path)
	isRegular := info.DirEntry.Type().IsRegular()
	return (fileExt == ".go") && isRegular 
}

func main() {
	var (
		src string
	)
	flag.StringVar(&src, "src", "", "source")
	flag.Parse()

	// First we'll create a Walker instance.
	w := walker.New()

	// Then we tell Walker to crawl our source directory with
	// a custom filter (onlyGoFiles func).
	infos := w.Walk(context.Background(), src, onlyGoFiles)

	for f := range infos {
		// this will output only names of Go files found
		fmt.Println(f.DirEntry.Name())
	}

	// let's check whether Walker encountered errors 
	// or finished it's work normally
	if err := w.Err(); err != nil {
		fmt.Println("Walker failed with error:", err)
	}
}
