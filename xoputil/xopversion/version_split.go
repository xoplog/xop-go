package xopversion

import (
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

var namespaceVersionRE = regexp.MustCompile(`^(.+)[- ]v?(\d+\.\d+\.\d+(?:-\S+)?)$`)

func SplitVersionWithError(namespace string) (string, *semver.Version, error) {
	var version string
	if m := namespaceVersionRE.FindStringSubmatch(namespace); m != nil {
		namespace = m[1]
		version = m[2]
	} else {
		version = "0.0.0"
	}
	sver, err := semver.StrictNewVersion(version)
	if err != nil {
		return "", nil, errors.Wrapf(err, "semver '%s' is not valid", version)
	}
	return namespace, sver, nil
}

var ZeroVersion = func() *semver.Version {
	sver, err := semver.StrictNewVersion("0.0.0")
	if err != nil {
		panic(err)
	}
	return sver
}()

func SplitVersion(namespace string) (string, *semver.Version) {
	n, v, err := SplitVersionWithError(namespace)
	if err != nil {
		return namespace, ZeroVersion
	}
	return n, v
}
