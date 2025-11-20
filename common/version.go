package common

import "fmt"

type Version uint32

func (v Version) Major() int {
	return int(uint8(v >> 24))
}

func (v Version) Minor() int {
	return int(uint8(v >> 16))
}

func (v Version) Patch() int {
	return int(uint8(v >> 8))
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
}

func (v *Version) Set(major, minor, patch uint8) {
	*v = Version(uint32(major)<<24 | uint32(minor)<<16 | uint32(patch)<<8)
}

func NewVersion(major, minor, patch uint8) Version {
	var version Version
	version.Set(major, minor, patch)
	return version
}
