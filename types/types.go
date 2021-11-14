package types

import "github.com/package-url/packageurl-go"

type ProjectList struct {
	Projects   []packageurl.PackageURL
	FileLookup map[string]string
}

// Wild guess at values for these. need confirmation
const ConanCacheDirName = ".cheque"
const ConanCacheFileName = "conan-cache.json"
