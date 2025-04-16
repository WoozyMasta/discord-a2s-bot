// Package vars is an internal technical variable store used at build time,
// populated with values ​​based on the state of the git repository.
package vars

var (
	// Version of application (git tag)
	Version string
	// Commit current in git
	Commit string
	// BuildTime of start build app
	BuildTime string
	// URL to repository
	URL string
)
