package walker

var (
	// DiscardDirs - removes directories from output
	DiscardDirs = func(info Info) bool {
		return !info.DirEntry.IsDir()
	}

	// DiscardRegular removes regular files from output
	DiscardRegular = func(info Info) bool {
		return !info.DirEntry.Type().IsRegular()
	}
)

// FilterFunc is intended to use with Filter 
// If FilterFunc returns true then the Info is accepted as
// valid and forwarded to the output channel of Filter.
// If FilterFunc returns false then the Info is silently 
// discarded.
type FilterFunc func(Info) bool

// filter accepts chan Info and zero or moreFilterFunc and returns chan Info
// which only emiths those elements for which all funcs returned true.
func filter(c chan Info, funcs ...FilterFunc) chan Info {
	out := make(chan Info)
	go func() {
		for {
			info, ok := <- c
			if !ok {
				// source chan is closed
				break
			}

			if oneOfChecksFails(info, funcs...) {
				continue
			}

			out <- info
		}
		close(out)
	}()
	return out
}

// oneOfChecksFails actually runs all tests. It returns true if 
// at least one of passed FilteFunc failed.
func oneOfChecksFails(info Info, funcs ...FilterFunc) (failed bool) {
	for _, f :=range funcs {
		if !f(info) {
			failed = true
			break
		}
	}
	return 
}
