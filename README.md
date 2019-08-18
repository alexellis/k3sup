# k3sup 🚀 (said 'ketchup')

k3sup is a light-weight utility to get from zero to KUBECONFIG with [k3s](https://k3s.io/) on any local or remote VM. All you need is `ssh` access and the `k3sup` binary to get `kubectl` access immediately.

The tool is written in Go and is cross-compiled for Linux, Windows, MacOS and even on Raspberry Pi.

How do you say it? Ketchup, as in tomato.

## What's this for? 💻

This tool uses `ssh` to install `k3s` to a remote Linux host. You can also use it to join existing Linux hosts into a k3s cluster as `agents`. First, `k3s` is installed using the utility script from Rancher, along with a flag for your host's public IP so that TLS works properly. The `kubeconfig` file on the server is then fetched and updated so that you can connect from your laptop using `kubectl`.

You may wonder why a tool like this needs to exist when you can do this sort of thing with bash.

k3sup was developed to automate what can be a very manual and confusing process for many developers, who are already short on time. Once you've provisioned a VM with your favourite tooling, `k3sup` means you are only 60 seconds away from running `kubectl get pods` on your own computer. With version 0.2.0, you can even `join` other nodes into any existing k3s cluster.

Uses:

* Bootstrap Kubernetes with k3s onto any VM - either manually, during CI or through `cloudinit`
* Get from zero to `kubectl` with `k3s` on Raspberry Pi (RPi), VMs, AWS EC2, DigitalOcean, Civo, Scaleway and more
* Fetch a working KUBECONFIG from an existing `k3s` cluster
* Join nodes into an existing `k3s` cluster with `k3sup join`

![](./docs/k3sup-cloud.png)
*Conceptual architecture, showing `k3sup` running locally against any VM such as AWS EC2 or a VPS such as DigitalOcean.*

## Demo 📼

In the demo I install Kubernetes (`k3s`) onto two separate machines and get my `kubeconfig` downloaded to my laptop each time in around one minute.

1) Ubuntu 18.04 VM created on DigitalOcean with ssh key copied automatically
2) Raspberry Pi 4 with my ssh key copied over via `ssh-copy-id`

Watch the demo:

[![asciicast](https://asciinema.org/a/262630.svg)](https://asciinema.org/a/262630)

## Usage ✅

The `k3sup` tool is designed to be run on your desktop/laptop computer, but binaries are provided for MacOS, Windows, and Linux (including ARM).

### Setup a Kubernetes server

You can setup a server and stop here, or go on to use the `join` command to add some "agents" aka `nodes` or `workers` into the cluster to expand its compute capacity.

```sh
curl -sLS https://get.k3sup.dev | sh
sudo install k3sup /usr/local/bin/
```

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.1` and the username was `ubuntu`, then you would run this:

* Run `k3sup`:

```sh
export IP=192.168.0.1
k3sup install --ip $IP --user ubuntu
```

Other options for `install`:

* `--skip-install` - if you already have k3s installed, you can just run this command to get the `kubeconfig`
* `--ssh-key` - specify a specific path for the SSH key for remote login
* `--local-path` - default is `./kubeconfig` - set the path into which you want to save your VM's `kubeconfig`
* `--ssh-port` - default is `22`, but you can specify an alternative port i.e. `2222`

* Now try the access:

```sh
export KUBECONFIG=`pwd`/kubeconfig
kubectl get node
```

### Join some agents to your Kubernetes server

Let's say that you have a server, and have already run the following:

```sh
export SERVER_IP=192.168.0.100
export USER=root

k3sup install --ip $SERVER_IP --user $USER
```

Next join one or more `agents` to the cluster:

```sh
export AGENT_IP=192.168.0.101

export SERVER_IP=192.168.0.100
export USER=root

k3sup join --ip $AGENT_IP --server-ip $SERVER_IP --user $USER
```

That's all, so with the above command you can have a two-node cluster up and running, whether that's using VMs on-premises, using Raspberry Pis, 64-bit ARM or even cloud VMs on EC2.

### Micro-tutorial for Raspberry Pi (2, 3, or 4) 🥧

In a few moments you will have Kubernetes up and running on your Raspberry Pi 2, 3 or 4. Stand by for the fastest possible install. At the end you will have a KUBECONFIG file on your local computer that you can use to access your cluster remotely.

![](./docs/k3sup-rpi.png)
*Conceptual architecture, showing `k3sup` running locally against bare-metal ARM devices.*

* [Download etcher.io](https://www.balena.io/etcher/) for your OS

* Flash an SD card using [Raspbian Lite](https://www.raspberrypi.org/downloads/raspbian/)

* Generate an ssh-key if you don't already have one with `ssh-keygen` (hit enter to all questions)

* Find the RPi IP with `ping -c raspberrypi.local`, then set `export IP=""` with the IP

* Copy over your ssh key with: `ssh-copy-id pi@raspberrypi.local`

* Run `k3sup --ip $IP --user pi`

* Point at the config file and get the status of the node:

```sh
export KUBECONFIG=`pwd`/kubeconfig
kubectl get node -o wide
```

You now have `kubectl` access from your laptop to your Raspberry Pi running k3s.

See also: [Blog: Will it cluster? K3s on Raspbian](https://blog.alexellis.io/test-drive-k3s-on-raspberry-pi/)

## Caveats on security

If you are using public cloud, then make sure you see the notes from the Rancher team on setting up a Firewall or Security Group.

k3s docs: [k3s configuration / open ports](https://rancher.com/docs/k3s/latest/en/configuration/#open-ports-network-security)

## What are people saying about `k3sup`?

* Blog post by Ruan Bekker:

    > Provision k3s to all the places with a awesome utility called "k3sup" by @alexellisuk. Definitely worth checking it out, its epic!

    [Provision k3s on the fly with k3sup](https://sysadmins.co.za/provision-k3s-on-the-fly-with-k3sup/)

* [Dave Cadwallader (@geek_dave)](https://twitter.com/geek_dave/status/1162386683200851969?s=20):

    > Alex - Thanks so much for all the effort you put into your tools and tutorials.  My rpi homelab has been a valuable learning playground for CNCF tech thanks to you!

* Checkout the [Announcement tweet](https://twitter.com/alexellisuk/status/1162272786250735618?s=20)

## What other cool tools exist for k3s?

* [k3s](https://github.com/rancher/k3s) - Kubernetes as installed by `k3sup`. k3s is a compliant, light-weight, multi-architecture distribution of Kubernetes
* [k3d](https://github.com/rancher/k3d) - this tool runs a Docker container on your local laptop with k3s started up inside
* [k3v](https://github.com/ibuildthecloud/k3v) - "virtual kubernetes" - a very early PoC from the author of k3s aiming to slice up a single cluster for multiple tenants

## License

MIT

## Contributing

### Blog posts & tweets

Blogs posts, tutorials, and Tweets about k3sup are appreciated. Please send a PR to the README.md file to add yours.

### Say thanks

Buy the author a coffee and show your support for `k3sup` through [GitHub Sponsors](https://github.com/users/alexellis/sponsorship).

### Contributing via GitHub

Before contributing code, please see the [CONTRIBUTING guide](https://github.com/alexellis/inlets/blob/master/CONTRIBUTING.md). Note that k3sup uses the same guide as [inlets.dev](https://inlets.dev/).

Both Issues and PRs have their own templates. Please fill out the whole template.

All commits must be signed-off as part of the [Developer Certificate of Origin (DCO)](https://developercertificate.org)
