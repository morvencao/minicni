package version

var (
	Version = "0.1.0"
)

// nolint
type VersionInfo struct {
	CniVersion        string   `json:"cniVersion"`
	SupportedVersions []string `json:"supportedVersions"`
}
