//coverage:ignore file
package version

const (
	AppName        = "reactenv"
	BinaryFilename = "reactenv"
)

var (
	// The git commit that was compiled. This will be filled in by the compiler.
	GitCommit   string
	GitBranch   string
	GitDescribe string

	// The compilation date. This will be filled in by the compiler.
	BuildDate string

	Version           = "0.1.116"
	VersionPrerelease = ""
	VersionMetadata   = ""
)
