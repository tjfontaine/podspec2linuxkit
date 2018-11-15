package linuxkit

import "github.com/containerd/containerd/reference"

// https://raw.githubusercontent.com/linuxkit/linuxkit/v0.6/src/cmd/linuxkit/moby/config.go
/*
Copyright 2015-2017 Docker, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// removed "github.com/opencontainers/runtime-spec/specs-go" to customize yaml
/*
   Copyright 2015 The Linux Foundation.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Moby is the type of a Moby config file
type Moby struct {
	Kernel     *KernelConfig `yaml:"kernel,omitempty" json:"kernel,omitempty"`
	Init       *[]string     `yaml:"init,omitempty" json:"init,omitempty"`
	Onboot     *[]*Image     `yaml:"onboot,omitempty" json:"onboot,omitempty"`
	Onshutdown *[]*Image     `yaml:"onshutdown,omitempty" json:"onshutdown,omitempty"`
	Services   *[]*Image     `yaml:"services" json:"services"`
	Trust      TrustConfig   `yaml:"trust,omitempty" json:"trust,omitempty"`
	Files      *[]File       `yaml:"files,omitempty" json:"files,omitempty"`

	initRefs []*reference.Spec
}

// KernelConfig is the type of the config for a kernel
type KernelConfig struct {
	Image   string  `yaml:"image" json:"image"`
	Cmdline string  `yaml:"cmdline,omitempty" json:"cmdline,omitempty"`
	Binary  string  `yaml:"binary,omitempty" json:"binary,omitempty"`
	Tar     *string `yaml:"tar,omitempty" json:"tar,omitempty"`
	UCode   *string `yaml:"ucode,omitempty" json:"ucode,omitempty"`

	ref *reference.Spec
}

// TrustConfig is the type of a content trust config
type TrustConfig struct {
	Image []string `yaml:"image,omitempty" json:"image,omitempty"`
	Org   []string `yaml:"org,omitempty" json:"org,omitempty"`
}

// File is the type of a file specification
type File struct {
	Path      string      `yaml:"path" json:"path"`
	Directory bool        `yaml:"directory" json:"directory"`
	Symlink   string      `yaml:"symlink,omitempty" json:"symlink,omitempty"`
	Contents  *string     `yaml:"contents,omitempty" json:"contents,omitempty"`
	Source    string      `yaml:"source,omitempty" json:"source,omitempty"`
	Metadata  string      `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Optional  bool        `yaml:"optional" json:"optional"`
	Mode      string      `yaml:"mode,omitempty" json:"mode,omitempty"`
	UID       interface{} `yaml:"uid,omitempty" json:"uid,omitempty"`
	GID       interface{} `yaml:"gid,omitempty" json:"gid,omitempty"`
}

// Image is the type of an image config
type Image struct {
	Name        string `yaml:"name" json:"name"`
	Image       string `yaml:"image" json:"image"`
	ImageConfig `yaml:",inline"`
}

// ImageConfig is the configuration part of Image, it is the subset
// which is valid in a "org.mobyproject.config" label on an image.
// Everything except Runtime and ref is used to build the OCI spec
type ImageConfig struct {
	Capabilities      *[]string          `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Ambient           *[]string          `yaml:"ambient,omitempty" json:"ambient,omitempty"`
	Mounts            *[]Mount           `yaml:"mounts,omitempty" json:"mounts,omitempty"`
	Binds             *[]string          `yaml:"binds,omitempty" json:"binds,omitempty"`
	Tmpfs             *[]string          `yaml:"tmpfs,omitempty" json:"tmpfs,omitempty"`
	Command           *[]string          `yaml:"command,omitempty" json:"command,omitempty"`
	Env               *[]string          `yaml:"env,omitempty" json:"env,omitempty"`
	Cwd               string             `yaml:"cwd,omitempty" json:"cwd,omitempty"`
	Net               string             `yaml:"net,omitempty" json:"net,omitempty"`
	Pid               string             `yaml:"pid,omitempty" json:"pid,omitempty"`
	Ipc               string             `yaml:"ipc,omitempty" json:"ipc,omitempty"`
	Uts               string             `yaml:"uts,omitempty" json:"uts,omitempty"`
	Userns            string             `yaml:"userns,omitempty" json:"userns,omitempty"`
	Hostname          string             `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	Readonly          *bool              `yaml:"readonly,omitempty" json:"readonly,omitempty"`
	MaskedPaths       *[]string          `yaml:"maskedPaths,omitempty" json:"maskedPaths,omitempty"`
	ReadonlyPaths     *[]string          `yaml:"readonlyPaths,omitempty" json:"readonlyPaths,omitempty"`
	UID               *interface{}       `yaml:"uid,omitempty" json:"uid,omitempty"`
	GID               *interface{}       `yaml:"gid,omitempty" json:"gid,omitempty"`
	AdditionalGids    *[]interface{}     `yaml:"additionalGids,omitempty" json:"additionalGids,omitempty"`
	NoNewPrivileges   *bool              `yaml:"noNewPrivileges,omitempty" json:"noNewPrivileges,omitempty"`
	OOMScoreAdj       *int               `yaml:"oomScoreAdj,omitempty" json:"oomScoreAdj,omitempty"`
	RootfsPropagation *string            `yaml:"rootfsPropagation,omitempty" json:"rootfsPropagation,omitempty"`
	CgroupsPath       *string            `yaml:"cgroupsPath,omitempty" json:"cgroupsPath,omitempty"`
	Resources         *LinuxResources    `yaml:"resources,omitempty" json:"resources,omitempty"`
	Sysctl            *map[string]string `yaml:"sysctl,omitempty" json:"sysctl,omitempty"`
	Rlimits           *[]string          `yaml:"rlimits,omitempty" json:"rlimits,omitempty"`
	UIDMappings       *[]LinuxIDMapping  `yaml:"uidMappings,omitempty" json:"uidMappings,omitempty"`
	GIDMappings       *[]LinuxIDMapping  `yaml:"gidMappings,omitempty" json:"gidMappings,omitempty"`
	Annotations       *map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`

	Runtime *Runtime `yaml:"runtime,omitempty" json:"runtime,omitempty"`

	ref *reference.Spec
}

// Runtime is the type of config processed at runtime, not used to build the OCI spec
type Runtime struct {
	Cgroups    *[]string    `yaml:"cgroups,omitempty" json:"cgroups,omitempty"`
	Mounts     *[]Mount     `yaml:"mounts,omitempty" json:"mounts,omitempty"`
	Mkdir      *[]string    `yaml:"mkdir,omitempty" json:"mkdir,omitempty"`
	Interfaces *[]Interface `yaml:"interfaces,omitempty,omitempty" json:"interfaces,omitempty"`
	BindNS     Namespaces   `yaml:"bindNS,omitempty" json:"bindNS,omitempty"`
	Namespace  *string      `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

// Namespaces is the type for configuring paths to bind namespaces
type Namespaces struct {
	Cgroup *string `yaml:"cgroup,omitempty" json:"cgroup,omitempty"`
	Ipc    *string `yaml:"ipc,omitempty" json:"ipc,omitempty"`
	Mnt    *string `yaml:"mnt,omitempty" json:"mnt,omitempty"`
	Net    *string `yaml:"net,omitempty" json:"net,omitempty"`
	Pid    *string `yaml:"pid,omitempty" json:"pid,omitempty"`
	User   *string `yaml:"user,omitempty" json:"user,omitempty"`
	Uts    *string `yaml:"uts,omitempty" json:"uts,omitempty"`
}

// Interface is the runtime config for network interfaces
type Interface struct {
	Name         string `yaml:"name,omitempty" json:"name,omitempty"`
	Add          string `yaml:"add,omitempty" json:"add,omitempty"`
	Peer         string `yaml:"peer,omitempty" json:"peer,omitempty"`
	CreateInRoot bool   `yaml:"createInRoot" json:"createInRoot"`
}

// LinuxResources has container runtime resource constraints
type LinuxResources struct {
	// Devices configures the device whitelist.
	Devices []LinuxDeviceCgroup `json:"devices,omitempty" yaml:"devices,omitempty"`
	// Memory restriction configuration
	Memory *LinuxMemory `json:"memory,omitempty" yaml:"memory,omitempty"`
	// CPU resource restriction configuration
	CPU *LinuxCPU `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	// Task resource restriction configuration.
	Pids *LinuxPids `json:"pids,omitempty" yaml:"pids,omitempty"`
	// BlockIO restriction configuration
	BlockIO *LinuxBlockIO `json:"blockIO,omitempty" yaml:"blockIO,omitempty"`
	// Hugetlb limit (in bytes)
	HugepageLimits []LinuxHugepageLimit `json:"hugepageLimits,omitempty" yaml:"hugepageLimits,omitempty"`
	// Network restriction configuration
	Network *LinuxNetwork `json:"network,omitempty" yaml:"network,omitempty"`
}

// LinuxHugepageLimit structure corresponds to limiting kernel hugepages
type LinuxHugepageLimit struct {
	// Pagesize is the hugepage size
	Pagesize string `json:"pageSize"`
	// Limit is the limit of "hugepagesize" hugetlb usage
	Limit uint64 `json:"limit"`
}

// LinuxInterfacePriority for network interfaces
type LinuxInterfacePriority struct {
	// Name is the name of the network interface
	Name string `json:"name"`
	// Priority for the interface
	Priority uint32 `json:"priority"`
}

// linuxBlockIODevice holds major:minor format supported in blkio cgroup
type linuxBlockIODevice struct {
	// Major is the device's major number.
	Major int64 `json:"major"`
	// Minor is the device's minor number.
	Minor int64 `json:"minor"`
}

// LinuxWeightDevice struct holds a `major:minor weight` pair for weightDevice
type LinuxWeightDevice struct {
	linuxBlockIODevice
	// Weight is the bandwidth rate for the device.
	Weight *uint16 `json:"weight,omitempty"`
	// LeafWeight is the bandwidth rate for the device while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
}

// LinuxThrottleDevice struct holds a `major:minor rate_per_second` pair
type LinuxThrottleDevice struct {
	linuxBlockIODevice
	// Rate is the IO rate limit per cgroup per device
	Rate uint64 `json:"rate"`
}

// LinuxBlockIO for Linux cgroup 'blkio' resource management
type LinuxBlockIO struct {
	// Specifies per cgroup weight
	Weight *uint16 `json:"weight,omitempty"`
	// Specifies tasks' weight in the given cgroup while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
	// Weight per cgroup per device, can override BlkioWeight
	WeightDevice []LinuxWeightDevice `json:"weightDevice,omitempty"`
	// IO read rate limit per cgroup per device, bytes per second
	ThrottleReadBpsDevice []LinuxThrottleDevice `json:"throttleReadBpsDevice,omitempty"`
	// IO write rate limit per cgroup per device, bytes per second
	ThrottleWriteBpsDevice []LinuxThrottleDevice `json:"throttleWriteBpsDevice,omitempty"`
	// IO read rate limit per cgroup per device, IO per second
	ThrottleReadIOPSDevice []LinuxThrottleDevice `json:"throttleReadIOPSDevice,omitempty"`
	// IO write rate limit per cgroup per device, IO per second
	ThrottleWriteIOPSDevice []LinuxThrottleDevice `json:"throttleWriteIOPSDevice,omitempty"`
}

// LinuxMemory for Linux cgroup 'memory' resource management
type LinuxMemory struct {
	// Memory limit (in bytes).
	Limit *int64 `json:"limit,omitempty" yaml:"limit,omitempty"`
	// Memory reservation or soft_limit (in bytes).
	Reservation *int64 `json:"reservation,omitempty" yaml:"reservation,omitempty"`
	// Total memory limit (memory + swap).
	Swap *int64 `json:"swap,omitempty" yaml:"swap,omitempty"`
	// Kernel memory limit (in bytes).
	Kernel *int64 `json:"kernel,omitempty" yaml:"kernel,omitempty"`
	// Kernel memory limit for tcp (in bytes)
	KernelTCP *int64 `json:"kernelTCP,omitempty" yaml:"kernelTCP,omitempty"`
	// How aggressive the kernel will swap memory pages.
	Swappiness *uint64 `json:"swappiness,omitempty" yaml:"swappiness,omitempty"`
	// DisableOOMKiller disables the OOM killer for out of memory conditions
	DisableOOMKiller *bool `json:"disableOOMKiller,omitempty" yaml:"disableOOMKiller,omitempty"`
}

// LinuxCPU for Linux cgroup 'cpu' resource management
type LinuxCPU struct {
	// CPU shares (relative weight (ratio) vs. other cgroups with cpu shares).
	Shares *uint64 `json:"shares,omitempty" yaml:"shares,omitempty"`
	// CPU hardcap limit (in usecs). Allowed cpu time in a given period.
	Quota *int64 `json:"quota,omitempty" yaml:"quota,omitempty"`
	// CPU period to be used for hardcapping (in usecs).
	Period *uint64 `json:"period,omitempty" yaml:"period,omitempty"`
	// How much time realtime scheduling may use (in usecs).
	RealtimeRuntime *int64 `json:"realtimeRuntime,omitempty" yaml:"realtimeRuntime,omitempty"`
	// CPU period to be used for realtime scheduling (in usecs).
	RealtimePeriod *uint64 `json:"realtimePeriod,omitempty" yaml:"realtimePeriod,omitempty"`
	// CPUs to use within the cpuset. Default is to use any CPU available.
	Cpus string `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	// List of memory nodes in the cpuset. Default is to use any available memory node.
	Mems string `json:"mems,omitempty" yaml:"mems,omitempty"`
}

// LinuxPids for Linux cgroup 'pids' resource management (Linux 4.3)
type LinuxPids struct {
	// Maximum number of PIDs. Default is "no limit".
	Limit int64 `json:"limit"`
}

// LinuxNetwork identification and priority configuration
type LinuxNetwork struct {
	// Set class identifier for container's network packets
	ClassID *uint32 `json:"classID,omitempty"`
	// Set priority of network traffic for container
	Priorities []LinuxInterfacePriority `json:"priorities,omitempty"`
}

// LinuxDeviceCgroup represents a device rule for the whitelist controller
type LinuxDeviceCgroup struct {
	// Allow or deny
	Allow bool `json:"allow"`
	// Device type, block, char, etc.
	Type string `json:"type,omitempty"`
	// Major is the device's major number.
	Major *int64 `json:"major,omitempty"`
	// Minor is the device's minor number.
	Minor *int64 `json:"minor,omitempty"`
	// Cgroup access permissions format, rwm.
	Access string `json:"access,omitempty"`
}

// LinuxIDMapping specifies UID/GID mappings
type LinuxIDMapping struct {
	// HostID is the starting UID/GID on the host to be mapped to 'ContainerID'
	HostID uint32 `json:"hostID"`
	// ContainerID is the starting UID/GID in the container
	ContainerID uint32 `json:"containerID"`
	// Size is the number of IDs to be mapped
	Size uint32 `json:"size"`
}

// Mount specifies a mount for a container.
type Mount struct {
	// Destination is the absolute path where the mount will be placed in the container.
	Destination string `json:"destination"`
	// Type specifies the mount kind.
	Type string `json:"type,omitempty" platform:"linux,solaris"`
	// Source specifies the source path of the mount.
	Source string `json:"source,omitempty"`
	// Options are fstab style mount options.
	Options []string `json:"options,omitempty"`
}
