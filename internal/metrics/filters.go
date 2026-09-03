package metrics

import "strings"

// Filters holds prefix-exclude lists, applied Go-side only.
type Filters struct {
	MountPoints []string
	FSTypes     []string
	NetDevices  []string
	DiskDevices []string
}

func NewFilters(mountPoints, fsTypes, netDevices, diskDevices []string) Filters {
	return Filters{
		MountPoints: mountPoints,
		FSTypes:     fsTypes,
		NetDevices:  netDevices,
		DiskDevices: diskDevices,
	}
}

func excluded(prefixes []string, v string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(v, p) {
			return true
		}
	}
	return false
}

func (f Filters) MountExcluded(mountpoint string) bool { return excluded(f.MountPoints, mountpoint) }
func (f Filters) FSTypeExcluded(fs string) bool        { return excluded(f.FSTypes, fs) }
func (f Filters) NetDeviceExcluded(dev string) bool    { return excluded(f.NetDevices, dev) }
func (f Filters) DiskExcluded(dev string) bool         { return excluded(f.DiskDevices, dev) }
