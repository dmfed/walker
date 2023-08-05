package walker

type Action uint

const (
	Pass    Action = iota // Pass tells that Info should be sent to output chan
	Discard               // Disard tell to silently discard Info and NOT send it to output chan
	SkipDir               // SkipDir tells to skip directory.
	SkipAll               // SkipAll tells Walker to stop completely.
)

// FilterFunc is intended to use with Walker.
// It must return one of Actions which tell Walker
// how to proceed. If FilterFunc returns Pass, the Info
// is sent to output chan, Discard tells Walker to not sent
// Info to output, SkipDir tells to not visit the directory
// (but proceed with walk), SkipAll stops Walker (no further
// entries are read).
type FilterFunc func(Info) Action

// DiscardDirs returns FilterFunc which removes directories from output
// when Info.DirEntry.IsDir().
func DiscardDirs() FilterFunc {
	return func(info Info) Action {
		var a Action
		if info.DirEntry.IsDir() {
			a = Discard
		}
		return a
	}
}

// DiscardRegular return FilterFunc which removes
// regular files files from output
// when Info.DirEntry.Type().IsRegular().
func DiscardRegular() FilterFunc {
	return func(info Info) Action {
		var a Action
		if info.DirEntry.Type().IsRegular() {
			a = Discard
		}
		return a
	}
}

// SkipDirs returns FilterFunc which tells Walker
// to skip directory (and all subdirectories) when
// Info.Path equals any of paths passed as arguments.
func SkipDirs(paths ...string) FilterFunc {
	return func(info Info) Action {
		var a Action
		for i := range paths {
			if paths[i] == info.Path {
				a = SkipDir
				break
			}
		}
		return a
	}
}
