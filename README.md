## Walker
Package **walker** is a wrapper around fs.WalkDir using custom 
fs.WalkDirFunc to pass results of recursive directory walk to
caller through a channel.

It makes easier to use the result concurrenly.

The package is quite small and presumably does not need extensive
documentation. Just browse it on https://pkg.go.dev/github.com/dmfed/walker 
to see what is has to offer. 

Also take a look at the examples in **examples** dir of the repository
for an overview.
