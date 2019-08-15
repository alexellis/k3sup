# k3sup

k3sup is a light-weight utility to get from zero to KUBECONFIG with [k3s](https://k3s.io/) on any local or remote VM. All you need is `ssh` access and the `k3sup` binary to get `kubectl` access immediately.

The tool is written in Go and is cross-compiled for Linux, Windows, MacOS and even on Raspberry Pi.

## Why does this tool need to exist?

You may wonder why a tool like this needs to exist when you can do this sort of thing with bash.

This tool was developed to automate a manual and off-putting process for developers who are already short of time. Once you've provisioned a VM with your favourite tooling, k3sup makes it a doddle to get access to `kubectl` locally.

## Usage

```sh
curl -sLS https://raw.githubusercontent.com/alexellis/k3sup/master/get.sh | sh
sudo install k3sup /usr/local/bin/
```

Provision a VM and make sure that your SSH keys are installed.

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

* Now try the access:

```sh
export KUBECONFIG=`pwd`/kubeconfig
kubectl get node
```

## License

MIT