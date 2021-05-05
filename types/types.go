package types

import "github.com/package-url/packageurl-go"

type ProjectList struct {
	Projects   []packageurl.PackageURL
	FileLookup map[string]string
}
