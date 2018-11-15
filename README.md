# PodSpec2LinuxKit

`podspec2linuxkit` will translate a `Deployment`, `ReplicaSet`, `DaemonSet`,
`Pod` or similar manifest for Kubernetes into a
[LinuxKit](https://github.com/linuxkit/linuxkit) image manifest. The resulting
image manifest can be built with `linuxkit build` and run on the infrastructure
provider of your choosing (either supported by `linuxkit run` or by importing
the image or using the iso as you see fit). Raising the question, "do you really
need all of Kubernetes?" or perhaps "do you need the kubelet?"

## Usage

```bash
$ git clone https://github.com/tjfontaine/podspec2linuxkit
$ cd podspec2linuxkit && make
$ ./podspec2linuxkit < my-deployment.yaml > my-linuxkit.yaml
$ linuxkit build -format iso-efi -name my-image -dir ./out templates/base_image.yaml ./my-linuxkit.yaml
$ linuxkit run hyperkit -publish 18080:80 -iso -uefi ./out/my-image
```

## Base Image

The tool itself does not produce a standalone manifest that can be used for
building a LinuxKit image. However, by leveraging how LinuxKit can merge
multiple yaml files to produce the build, you can use the files in `templates`
with the resulting manifest, or use them as a starting point.

## Caveats

Nearly everything you can represent in a `PodSpec` has a direct translation for
the LinuxKit image manifest. You can have multiple containers, init containers,
resource limits, environment variables, mount propagation, security context, and
volumes (see supported volume types).

That being said, there are some things that are currently not possible with this
tool as is.

### External References

Manifest that refer to values stored in other manifests won't work. So things
like `valueFrom` or `envFrom` can't work because they're not defined in the same
manifest.

There are at least two ways we can solve the missing values, we could evolve the
tool to interpret multiple manifests at once such that it could attempt to
dereference all the appropriate values. Alternatively the tool could connect to
a running Kubernetes cluster and try and dereference the values. It does neither
for now.

### Ports

`ports` definitions are currently ignored. Having a pattern to firewall the host
network namespace by default and allow communication in would be ideal, but
doesn't exist yet -- so make sure you're ready for traffic to all your
containers.

### Supported Volume Types

Currently only `hostPath` and `emptyDir` are implemented, though things like
`nfs` and `iscsi` should be relatively straight forward to add.

`persistentVolumeClaim` is *not* currently supported for the "External
References" reason above, but is likely the most desirable option to add.

All the vendor related and external volumes are feasible, they just require some
amount of effort by those who wish to support them.

### Anything Relating to Scheduling

A lot of what's in the `PodSpec` is really only germane to the Kubernetes
scheduler responsible for the type of pod in question. They don't have an analog
in LinuxKit, so they're not included at all.

## Why does this tool exist?

If you are already packaging your software in Docker/OCI compatible containers,
and you're defining Kubernetes manifests as well, why bother with this tool at
all?

That's a tough, but fair, question.

First, I have a bit of a geek crush on LinuxKit, and I wanted an interesting way
to demonstrate its value. It expands on what made `docker` the tool powerful for
creating container images through a declarative model and uses it to create
immutable machine images (by stitching together multiple container images).

Speaking of immutability, I find it slightly hypocritical that many of us
espouse the ethos of immutable infrastructure, but we put the `kubelet` on a
fairly stateful image. So with these images, you get something that behaves a
bit more like a container instance (ACI/Fargate) without needing the
infrastructure platform to support it.

But really, often I see many teams struggling with Kubernetes as it relates to
isolation. This isn't just about isolation for security, but about resource and
fault isolation.

In the beginning, teams start by just deploying pods, replicasets, or
deployments without much care. Then failures start happening, either because of
memory/disk usage, or because the scheduler allowed multiple instances of the
same pod to be scheduled to the same node. As a result, teams then go through
and set their affinity, or use labels and taints to limit what pods can run on
which nodes.

Effectively if you adopted Kubernetes because you thought it would reduce your
cognitive or operational load by not having to think as much about
infrastructure, you're probably pretty disenfranchised at this point.

To be clear, this tool doesn't really solve much of that problem. But if your
problem ends up being infrastructure, why not just rely on an infrastructure
provider to be responsible for scheduling and keeping it running? This tool lets
you keep building your applications as containers, and defining their
relationships with Kubernetes manifesets, and then you can use LinuxKit to build
images you use with your infrastructure provider.

## So what about all that life cycle management then?

Well, if you mean "how can I get Kubernetes' features with Kubernetes' manifests
without using Kubernetes?" I don't really have an answer for you.

If on the other hand, your question is, "how can I do deployments with these
images, or health checking, or ..." then I would say, you should probably check
out [Spinnaker](https://spinnaker.io), which probably has many of the features
you're looking for and works with a variety of cloud providers.

One _could_ imagine however, a world where there was a frontend that "felt" like
Kuberentes, that when given a manifest built an image and then used Spinnaker to
manage the deployment ... but so far as I know it doesn't exist yet.

### PS

I'm not a golang expert, so go easy on my usage :)
