package gitfs

// This is an implementation of the Kustomize fs.FileSystem implementation,
// which uses go-git to fetch the files.
//
// It is not a complete implementation of fs.FileSystem, only the necessary
// methods to allow Kustomize to build manifests.
//
// This is a lesson in the Interface Seggregation principal...
