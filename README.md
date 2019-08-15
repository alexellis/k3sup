# k3sup

k3sup is a light-weight utility to get from zero to KUBECONFIG with [k3s](https://k3s.io/) on any local or remote VM. All you need is `ssh` access and the `k3sup` binary to get `kubectl` access immediately.

The tool is written in Go and is cross-compiled for Linux, Windows, MacOS and even on Raspberry Pi.

## Why does this tool need to exist?

You may wonder why a tool like this needs to exist when you can do this sort of thing with bash.

This tool was developed to automate a manual and off-putting process for developers who are already short of time. Once you've provisioned a VM with your favourite tooling, k3sup makes it a doddle to get access to `kubectl` locally.

Uses:

* Bootstrap Kubernetes with k3s onto any VM - either manually, during CI or through cloudinit
* Get from zero to `kubectl` with `k3s` on Raspberry Pi (RPi), VMs, DigitalOcean, Civo, Scaleway and more
* Fetch a KUBECONFIG from an existing `k3s` cluster

## Demo

I install k3s onto two separate machines and get access to `kubeconfig` within a minute.

* Ubuntu 18.04 VM created on DigitalOcean with ssh key copied automatically
* Raspberry Pi 4 with my ssh key copied over via `ssh-copy-id`

Watch the demo:

[![asciicast](https://asciinema.org/a/262630.svg)](https://asciinema.org/a/262630)

## Usage

```sh
curl -sLS https://raw.githubusercontent.com/alexellis/k3sup/master/get.sh | sh
sudo install k3sup /usr/local/bin/
```

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.1` and the usenrame was `ubuntu`, then you would run this:

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

### Micro-tutorial for Raspberry Pi (2, 3, or 4)

In a few moments you will have Kubernetes up and running on your Raspberry Pi 2, 3 or 4. Stand by for the fastest possible install. At the end you will have a KUBECONFIG file on your local computer that you can use to access your cluster remotely.

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

## License

MIT

## Contributing

As per [OpenFaaS](https://github.com/openfaas/faas/blob/master/CONTRIBUTING.md)

All commits must be signed-off as part of the [Developer Certificate of Origin (DCO)](https://developercertificate.org)
