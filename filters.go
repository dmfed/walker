package walker

// FilterFunc is intended to use with Walker.
// If FilterFunc returns true then the Info is accepted as
// valid and forwarded to the output channel of Walker.Walk method.
// If FilterFunc returns false then the Info is silently
// discarded.
type FilterFunc func(Info) bool

// DiscardDirs - removes directories from output
// (Info.DirEntry.IsDir())
func DiscardDirs() FilterFunc {
	return func(info Info) bool {
		return !info.DirEntry.IsDir()
	}
}

// DiscardRegular return FulterFunc which removes
// regular files files from output
// (Info.DirEntry.Type().IsRegular())
func DiscardRegular() FilterFunc {
	return func(info Info) bool {
		return !info.DirEntry.Type().IsRegular()
	}
}
