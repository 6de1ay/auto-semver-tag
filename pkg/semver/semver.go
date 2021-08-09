package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	SemVerRegExp = `^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)`

	IncrementTypeMajor = "major"
	IncrementTypeMinor = "minor"
	IncrementTypePatch = "patch"
)

type SemVer struct {
	major uint64
	minor uint64
	patch uint64
}

func newInvalidSemVerError(semVer string) (SemVer, error) {
	return SemVer{}, fmt.Errorf("invalid semver: %s", semVer)
}

func New(semVer string) (SemVer, error) {

	isSemVerValid, err := regexp.MatchString(SemVerRegExp, semVer)
	if err != nil || !isSemVerValid {
		return newInvalidSemVerError(semVer)
	}

	// semVer[1:] because the first symbol is 'v'
	parts := strings.SplitN(semVer[1:], ".", 3)

	major, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return newInvalidSemVerError(semVer)
	}

	minor, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return newInvalidSemVerError(semVer)
	}

	patch, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return newInvalidSemVerError(semVer)
	}

	res := SemVer{
		major,
		minor,
		patch,
	}

	return res, err
}

func (s SemVer) IncrementVersion(incrementType string) SemVer {
	switch incrementType {
	case IncrementTypeMajor:
		s.major += 1
		s.minor = 0
		s.patch = 0
	case IncrementTypeMinor:
		s.minor += 1
		s.patch = 0
	case IncrementTypePatch:
		s.patch += 1
	default:
		panic("invalid increment type")
	}

	return s
}

func (s SemVer) String() string {
	return fmt.Sprintf("v%d.%d.%d", s.major, s.minor, s.patch)
}

func (s SemVer) IsGreaterThan(o SemVer) bool {
	if s.major > o.major {
		return true
	}
	if s.major < o.major {
		return false
	}

	// Major versions are equal

	if s.minor > o.minor {
		return true
	}
	if s.minor < o.minor {
		return false
	}

	// Major and minor versions are equal

	if s.patch > o.patch {
		return true
	}
	return false
}
