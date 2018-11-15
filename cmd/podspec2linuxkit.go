package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tjfontaine/podspec2linuxkit/pkg/linuxkit"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"strings"
)

// SEE https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities
var DEFAULT_CAPABILITIES = map[string]bool{
	"CAP_SETPCAP":          true,
	"CAP_MKNOD":            true,
	"CAP_AUDIT_WRITE":      true,
	"CAP_CHOWN":            true,
	"CAP_NET_RAW":          true,
	"CAP_DAC_OVERRIDE":     true,
	"CAP_FOWNER":           true,
	"CAP_FSETID":           true,
	"CAP_KILL":             true,
	"CAP_SETGID":           true,
	"CAP_SETUID":           true,
	"CAP_NET_BIND_SERVICE": true,
	"CAP_SYS_CHROOT":       true,
	"CAP_SETFCAP":          true,
}

func containerToLinuxKitImage(spec *corev1.PodSpec, container corev1.Container, volumeMap map[string]string) (*linuxkit.Image, error) {
	image := &linuxkit.Image{
		Name:  container.Name,
		Image: container.Image,
		ImageConfig: linuxkit.ImageConfig{
			Cwd:      container.WorkingDir,
			Hostname: spec.Hostname,
		},
	}

	if len(container.Command) > 0 {
		image.ImageConfig.Command = &container.Command
	}

	envArr := []string{}
	if len(container.Env) > 0 {
		for _, env := range container.Env {
			if env.ValueFrom != nil {
				log.Warnf("valueFrom for environment variables not implemented: %s unset", env.Name)
				continue
			}
			envArr = append(envArr, fmt.Sprintf("%s=%s", env.Name, env.Value))
		}
	}

	if len(envArr) > 0 {
		image.ImageConfig.Env = &envArr
	}

	if len(container.EnvFrom) > 0 {
		log.Warnf("envFrom not implemented")
	}

	mounts := []string{
		"/etc/resolv.conf:/etc/resolv.conf",
	}

	for _, volume := range container.VolumeMounts {
		mount, propagateMounts, err := volumeMountToLinuxKitMount(&volume, volumeMap)
		if err != nil {
			return nil, err
		}

		if propagateMounts != "" {
			if image.RootfsPropagation != nil {
				log.Warnf("Overwriting RootfsPropagation value -- Old %s New %s", *image.RootfsPropagation, propagateMounts)
			}
			image.RootfsPropagation = &propagateMounts
		}

		mounts = append(mounts, mount)
	}

	if len(mounts) > 0 {
		image.ImageConfig.Binds = &mounts
	}

	capabMap := DEFAULT_CAPABILITIES

	if container.SecurityContext != nil {
		sc := container.SecurityContext

		if sc.Privileged != nil && *sc.Privileged {
			capabMap = map[string]bool{"all": true}
		} else {
			if sc.Capabilities != nil {
				for _, capab := range sc.Capabilities.Add {
					capabMap[fmt.Sprintf("CAP_%s", string(capab))] = true
				}

				for _, capab := range sc.Capabilities.Drop {
					capabMap[fmt.Sprintf("CAP_%s", string(capab))] = false
				}
			}
		}

		if sc.RunAsUser != nil {
			var v interface{} = *sc.RunAsUser
			image.UID = &v
		}

		if sc.RunAsGroup != nil {
			var v interface{} = *sc.RunAsGroup
			image.GID = &v
		}

		image.Readonly = sc.ReadOnlyRootFilesystem
		image.NoNewPrivileges = sc.AllowPrivilegeEscalation

		// TODO sysctls?
	}

	capabArr := []string{}

	for capabKey, capabValue := range capabMap {
		if capabValue {
			capabArr = append(capabArr, capabKey)
		}
	}

	if len(capabArr) > 0 {
		image.Capabilities = &capabArr
	}

	// By default LinuxKit will put containers in their own pid namespace, so we only need to take care of those that
	// want the host PID namespace, or to share a pid namespace among themselves.
	if spec.HostPID {
		image.ImageConfig.Pid = "host"
	} else if spec.ShareProcessNamespace != nil && *spec.ShareProcessNamespace {
		image.ImageConfig.Pid = fmt.Sprintf("/run/pidns/shared-namespace")
	}

	// by default, LinuxKit already runs all containers in the same host, ipc, and utc namespaces -- so
	// spec.HostNetwork, spec.HostIPC have no particular meaning here

	resources := linuxkit.LinuxResources{}
	resourcesSeen := false
	for name, limit := range container.Resources.Limits {
		switch name {
		case "cpu":
			val, ok := limit.AsDec().Unscaled()
			if !ok {
				log.Warnf("couldn't convert limit to int64")
				continue
			}
			uval := uint64(val)
			resources.CPU = &linuxkit.LinuxCPU{
				Shares: &uval,
			}
			resourcesSeen = true
		case "memory":
			resourcesSeen = true
			val := limit.Value()
			resources.Memory = &linuxkit.LinuxMemory{
				Limit: &val,
			}
		default:
			log.Warnf("unknown limit name: %s", name)
		}
	}

	if resourcesSeen {
		image.ImageConfig.Resources = &resources
	}

	// TODO make a pattern for managing the firewall and exposing ports for containers
	for _, port := range container.Ports {
		log.Warnf("TODO Firewall By Default -- Port %s:%d is already open (as are all ports)", port.Name, port.ContainerPort)
	}

	return image, nil
}

var mountPropMap = map[string]string{
	"None":            "",
	"HostToContainer": "rslave",
	"Bidirectional":   "rshared",
}

func volumeMountToLinuxKitMount(volume *corev1.VolumeMount, volumeMap map[string]string) (string, string, error) {
	propagateMounts := ""

	path, ok := volumeMap[volume.Name]

	if !ok {
		return "", propagateMounts, fmt.Errorf("failed to find volume in pod spec: %s", volume.Name)
	}

	mount := fmt.Sprintf("%s:%s", path, volume.MountPath)

	opts := []string{}

	if volume.ReadOnly {
		opts = append(opts, "ro")
	}

	if volume.MountPropagation != nil {
		propagateMounts, ok = mountPropMap[string(*volume.MountPropagation)]
		if !ok {
			log.Warnf("Unknown mount propagation value: %s", string(*volume.MountPropagation))
		} else if propagateMounts != "" {
			opts = append(opts, propagateMounts)
		}
	}

	if len(opts) > 0 {
		mount = fmt.Sprintf("%s:%s", mount, strings.Join(opts, ","))
	}

	return mount, propagateMounts, nil
}

func volumeToLinuxKitMount(volume *corev1.Volume, volumeMap map[string]string) (*linuxkit.Image, error) {
	var image *linuxkit.Image = nil

	if volume.HostPath != nil {
		var command []string
		volumeMap[volume.Name] = volume.HostPath.Path
		if volume.HostPath.Type != nil {
			switch *volume.HostPath.Type {
			case "DirectoryOrCreate":
				command = []string{"mkdir", "-p", volume.HostPath.Path}
			case "FileOrCreate":
				command = []string{"touch", volume.HostPath.Path}
			case "Directory":
				fallthrough
			case "File":
				fallthrough
			case "Socket":
				fallthrough
			case "CharDevice":
				fallthrough
			case "BlockDevice":
				break
			}
		}
		if len(command) > 0 {
			image = &linuxkit.Image{
				Name:  fmt.Sprintf("create-volume-%s", volume.Name),
				Image: "busybox:latest",
			}
			image.ImageConfig.Command = &command
		}
	} else if volume.EmptyDir != nil {
		path := fmt.Sprintf("/var/lib/volumes/%s", volume.Name)
		volumeMap[volume.Name] = path
		image = &linuxkit.Image{
			Name:  fmt.Sprintf("create-volume-%s", volume.Name),
			Image: "busybox:latest",
			ImageConfig: linuxkit.ImageConfig{
				Command: &[]string{"mkdir", "-p", path},
			},
		}
	} else {
		return nil, fmt.Errorf("Unhandled volume type: %#v", volume)
	}

	return image, nil
}

func podSpec2LinuxKit(spec *corev1.PodSpec) (*linuxkit.Moby, error) {
	result := &linuxkit.Moby{}

	onboot := []*linuxkit.Image{}

	volumeMap := map[string]string{}

	for _, volume := range spec.Volumes {
		mount, err := volumeToLinuxKitMount(&volume, volumeMap)
		if err != nil {
			return nil, err
		}
		if mount != nil {
			onboot = append(onboot, mount)
		}
	}

	for idx, initContainer := range spec.InitContainers {
		image, err := containerToLinuxKitImage(spec, initContainer, volumeMap)
		if err != nil {
			return nil, err
		}
		image.Name = fmt.Sprintf("initContainer-%d-%s", idx, initContainer.Name)
		onboot = append(onboot, image)
	}

	if len(onboot) > 0 {
		result.Onboot = &onboot
	}

	services := []*linuxkit.Image{}
	for _, container := range spec.Containers {
		image, err := containerToLinuxKitImage(spec, container, volumeMap)
		if err != nil {
			return nil, err
		}
		image.Name = fmt.Sprintf("container-%s", container.Name)
		services = append(services, image)
	}

	if len(services) > 0 {
		result.Services = &services
	}

	return result, nil
}

type KindToLookup map[string]func(interface{}) corev1.PodSpec
type VersionLookup map[string]KindToLookup
type GroupLookup map[string]VersionLookup

// We could use reflection instead of this lookup table, but I'm not sure it buys us much?
var GroupMap = GroupLookup{
	"apps": VersionLookup{
		"v1": KindToLookup{
			"Deployment": func(reference interface{}) corev1.PodSpec {
				return reference.(*appsv1.Deployment).Spec.Template.Spec
			},
			"ReplicaSet": func(reference interface{}) corev1.PodSpec {
				return reference.(*appsv1.ReplicaSet).Spec.Template.Spec
			},
			"DaemonSet": func(reference interface{}) corev1.PodSpec {
				return reference.(*appsv1.DaemonSet).Spec.Template.Spec
			},
		},
		"v1beta1": KindToLookup{
			"Deployment": func(reference interface{}) corev1.PodSpec {
				return reference.(*appsv1beta1.Deployment).Spec.Template.Spec
			},
		},
		"v1beta2": KindToLookup{
			"Deployment": func(reference interface{}) corev1.PodSpec {
				return reference.(*appsv1beta2.Deployment).Spec.Template.Spec
			},
			"ReplicaSet": func(reference interface{}) corev1.PodSpec {
				return reference.(*appsv1beta2.ReplicaSet).Spec.Template.Spec
			},
			"DaemonSet": func(reference interface{}) corev1.PodSpec {
				return reference.(*appsv1beta2.DaemonSet).Spec.Template.Spec
			},
		},
	},
	"core": VersionLookup{
		"v1": KindToLookup{
			"Pod": func(reference interface{}) corev1.PodSpec {
				return reference.(*corev1.Pod).Spec
			},
		},
	},
	"extensions": VersionLookup{
		"v1beta1": KindToLookup{
			"DaemonSet": func(reference interface{}) corev1.PodSpec {
				return reference.(*extv1beta1.DaemonSet).Spec.Template.Spec
			},
			"ReplicaSet": func(reference interface{}) corev1.PodSpec {
				return reference.(*extv1beta1.ReplicaSet).Spec.Template.Spec
			},
			"Deployment": func(reference interface{}) corev1.PodSpec {
				return reference.(*extv1beta1.Deployment).Spec.Template.Spec
			},
		},
	},
}

func main() {

	rawYaml, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		log.Errorf("Failed to load pod spec: %v", err)
		os.Exit(1)
		return
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, groupVersionKind, err := decode(rawYaml, nil, nil)

	if err != nil {
		log.Errorf("Failed to decode pod spec: %v", err)
		os.Exit(1)
		return
	}

	log.Debugf("%#v", groupVersionKind)

	group, ok := GroupMap[groupVersionKind.Group]

	if !ok {
		log.Errorf("Unknown Group: %s", group)
		os.Exit(1)
	}

	version, ok := group[groupVersionKind.Version]

	if !ok {
		log.Errorf("Unknown Group/Version %s/%s", groupVersionKind.Group, groupVersionKind.Version)
		os.Exit(1)
	}

	kind, ok := version[groupVersionKind.Kind]

	if !ok {
		log.Errorf("Unknown Group/Version/Kind %s/%s/%s", groupVersionKind.Group, groupVersionKind.Version, groupVersionKind.Kind)
		os.Exit(1)
	}

	podSpec := kind(obj)

	foo, err := podSpec2LinuxKit(&podSpec)
	if err != nil {
		log.Errorf("Failed to convert: %v", err)
		os.Exit(1)
	} else {
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.Encode(foo)
	}
}
