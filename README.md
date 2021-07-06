# k3sup 🚀 (said 'ketchup')

<img src="docs/assets/k3sup.png" width="20%" alt="k3sup logo">

k3sup is a light-weight utility to get from zero to KUBECONFIG with [k3s](https://k3s.io/) on any local or remote VM. All you need is `ssh` access and the `k3sup` binary to get `kubectl` access immediately.


The tool is written in Go and is cross-compiled for Linux, Windows, MacOS and even on Raspberry Pi.

How do you say it? Ketchup, as in tomato.

[![Build
Status](https://travis-ci.com/alexellis/k3sup.svg?branch=master)](https://travis-ci.com/alexellis/k3sup)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexellis/k3sup)](https://goreportcard.com/report/github.com/alexellis/k3sup)
[![GoDoc](https://godoc.org/github.com/alexellis/k3sup?status.svg)](https://godoc.org/github.com/alexellis/k3sup) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub All Releases](https://img.shields.io/github/downloads/alexellis/k3sup/total)

## Contents:
- [k3sup 🚀 (said 'ketchup')](#k3sup--said-ketchup)
  - [What's this for? 💻](#whats-this-for-)
    - [Uses](#uses)
    - [Bootstrapping Kubernetes](#bootstrapping-kubernetes)
  - [Download `k3sup` (tl;dr)](#download-k3sup-tldr)
    - [A note for Windows users](#a-note-for-windows-users)
  - [Demo 📼](#demo-)
  - [Who is the author? 👏](#who-is-the-author-)
  - [Usage ✅](#usage-)
    - [👑 Setup a Kubernetes *server* with `k3sup`](#-setup-a-kubernetes-server-with-k3sup)
    - [Advanced KUBECONFIG options](#merging-clusters-into-your-kubeconfig)
    - [😸 Join some agents to your Kubernetes server](#-join-some-agents-to-your-kubernetes-server)
    - [Create a multi-master (HA) setup with external SQL](#create-a-multi-master-ha-setup-with-external-sql)
    - [Create a multi-master (HA) setup with embedded etcd](#create-a-multi-master-ha-setup-with-embedded-etcd)
    - [👨‍💻 Micro-tutorial for Raspberry Pi (2, 3, or 4) 🥧](#-micro-tutorial-for-raspberry-pi-2-3-or-4-)
  - [Caveats on security](#caveats-on-security)
  - [If your ssh-key is password-protected](#if-your-ssh-key-is-password-protected)
  - [Contributing](#contributing)
    - [Insiders Subscription ☕️ 👏](#insiders-subscription-️-)
    - [Blog posts & tweets](#blog-posts--tweets)
    - [Contributing via GitHub](#contributing-via-github)
    - [License](#license)
  - [📢 What are people saying about `k3sup`?](#-what-are-people-saying-about-k3sup)
  - [Similar tools & glossary](#similar-tools--glossary)
  - [Troubleshooting](#troubleshooting)
    - [Getting access to your KUBECONFIG](#getting-access-to-your-kubeconfig)
    - [Smart cards and 2FA](#smart-cards-and-2fa)
    - [Misc note on iptables](#misc-note-on-iptables)

## What's this for? 💻

This tool uses `ssh` to install `k3s` to a remote Linux host. You can also use it to join existing Linux hosts into a k3s cluster as `agents`. First, `k3s` is installed using the utility script from Rancher, along with a flag for your host's public IP so that TLS works properly. The `kubeconfig` file on the server is then fetched and updated so that you can connect from your laptop using `kubectl`.

You may wonder why a tool like this needs to exist when you can do this sort of thing with bash.

k3sup was developed to automate what can be a very manual and confusing process for many developers, who are already short on time. Once you've provisioned a VM with your favourite tooling, `k3sup` means you are only 60 seconds away from running `kubectl get pods` on your own computer. If you are a local computer, you can bypass SSH with `k3sup install --local`

### Uses

* Bootstrap Kubernetes with k3s onto any VM with `k3sup install` - either manually, during CI or through `cloud-init`
* Get from zero to `kubectl` with `k3s` on Raspberry Pi (RPi), VMs, AWS EC2, Packet bare-metal, DigitalOcean, Civo, Scaleway, and others
* Build a HA, multi-master (server) cluster
* Fetch the KUBECONFIG from an existing `k3s` cluster
* Join nodes into an existing `k3s` cluster with `k3sup join`

### Bootstrapping Kubernetes

![Conceptual architecture](./docs/k3sup-cloud.png)
*Conceptual architecture, showing `k3sup` running locally against any VM such as AWS EC2 or a VPS such as DigitalOcean.*

## Do you love `k3sup`?

k3sup is free and open source, but requires time and effort to support users and build and test new features. Support this project via [GitHub Sponsors](https://github.com/users/alexellis/sponsorship).

## Download `k3sup` (tl;dr)

`k3sup` is distributed as a static Go binary. You can use the installer on MacOS and Linux, or visit the [Releases page](https://github.com/alexellis/k3sup/releases) to download the executable for Windows.

```sh
curl -sLS https://get.k3sup.dev | sh
sudo install k3sup /usr/local/bin/

k3sup --help
```

### A note for Windows users

Windows users can use `k3sup install` and `k3sup join` with a normal "Windows command prompt".

## Demo 📼

In the demo I install Kubernetes (`k3s`) onto two separate machines and get my `kubeconfig` downloaded to my laptop each time in around one minute.

1) Ubuntu 18.04 VM created on DigitalOcean with ssh key copied automatically
2) Raspberry Pi 4 with my ssh key copied over via `ssh-copy-id`

Watch the demo:

[![asciicast](https://asciinema.org/a/262630.svg)](https://asciinema.org/a/262630)

## Who is the author? 👏

`k3sup` is Open Source Software (OSS) and was created by [Alex Ellis](https://www.alexellis.io/) - the founder of [OpenFaaS &reg;](https://www.openfaas.com/) & [inlets](https://inlets.dev/). Alex is also an active part of the Docker & Kubernetes community as a [CNCF Ambassador](https://www.cncf.io/people/ambassadors/).

If you've benefitted from his open source projects or blog posts in some way, then and join dozens of other developers today by buying an Insiders Subscription 🏆 via GitHub Sponsors.

* [Sponsor on GitHub](https://github.com/users/alexellis/sponsorship)

## Usage ✅

The `k3sup` tool is a client application which you can run on your own computer. It uses SSH to connect to remote servers and creates a local KUBECONFIG file on your disk. Binaries are provided for MacOS, Windows, and Linux (including ARM).

### 👑 Setup a Kubernetes *server* with `k3sup`

You can setup a server and stop here, or go on to use the `join` command to add some "agents" aka `nodes` or `workers` into the cluster to expand its compute capacity.

Provision a new VM running a compatible operating system such as Ubuntu, Debian, Raspbian, or something else. Make sure that you opt-in to copy your registered SSH keys over to the new VM or host automatically.

> Note: You can copy ssh keys to a remote VM with `ssh-copy-id user@IP`.

Imagine the IP was `192.168.0.1` and the username was `ubuntu`, then you would run this:

* Run `k3sup`:

```sh
export IP=192.168.0.1
k3sup install --ip $IP --user ubuntu

# Or use a hostname and SSH key for EC2
export HOST="ec2-3-250-131-77.eu-west-1.compute.amazonaws.com"
k3sup install --host $HOST --user ubuntu \
  --ssh-key $HOME/ec2-key.pem
```

Other options for `install`:

* `--cluster` - start this server in clustering mode using embedded etcd (embedded HA)
* `--skip-install` - if you already have k3s installed, you can just run this command to get the `kubeconfig`
* `--ssh-key` - specify a specific path for the SSH key for remote login
* `--local-path` - default is `./kubeconfig` - set the file where you want to save your cluster's `kubeconfig`.  By default this file will be overwritten.
* `--merge` - Merge config into existing file instead of overwriting (e.g. to add config to the default kubectl config, use `--local-path ~/.kube/config --merge`).
* `--context` - default is `default` - set the name of the kubeconfig context.
* `--ssh-port` - default is `22`, but you can specify an alternative port i.e. `2222`
* `--k3s-extra-args` - Optional extra arguments to pass to k3s installer, wrapped in quotes, i.e. `--k3s-extra-args '--no-deploy traefik'` or `--k3s-extra-args '--docker'`. For multiple args combine then within single quotes `--k3s-extra-args '--no-deploy traefik --docker'`.
* `--k3s-version` - set the specific version of k3s, i.e. `v0.9.1`
- `--ipsec` - Enforces the optional extra argument for k3s: `--flannel-backend` option: `ipsec`
* `--print-command` - Prints out the command, sent over SSH to the remote computer
* `--datastore` - used to pass a SQL connection-string to the `--datastore-endpoint` flag of k3s. You must use [the format required by k3s in the Rancher docs](https://rancher.com/docs/k3s/latest/en/installation/ha/).

See even more install options by running `k3sup install --help`.

* Now try the access:

```bash
export KUBECONFIG=`pwd`/kubeconfig
kubectl get node
```

Note that you should always use `pwd/` so that a full path is set, and you can change directory if you wish.

### Merging clusters into your KUBECONFIG

You can also merge the remote config into your main KUBECONFIG file `$HOME/.kube/config`, then use `kubectl config get-contexts` or `kubectx` to manage it.

The default "context" name for the remote k3s cluster is `default`, however you can override this as below.

For example:

```bash
k3sup install \
  --ip $IP \
  --user $USER \
  --merge \
  --local-path $HOME/.kube/config \
  --context my-k3s
```

Here we set a context of `my-k3s` and also merge into our main local `KUBECONFIG` file, so we could run `kubectl config set-context my-k3s` or `kubectx my-k3s`.

### 😸 Join some agents to your Kubernetes server

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

### Create a multi-master (HA) setup with external SQL

The easiest way to test out k3s' multi-master (HA) mode with external storage, is to set up a Mysql server using DigitalOcean's managed service.

* Get the connection string from your DigitalOcean dashboard, and adapt it

Before:
```
mysql://doadmin:80624d3936dfc8d2e80593@db-mysql-lon1-90578-do-user-6456202-0.a.db.ondigitalocean.com:25060/defaultdb?ssl-mode=REQUIRED

```

After:
```
mysql://doadmin:80624d3936dfc8d2e80593@tcp(db-mysql-lon1-90578-do-user-6456202-0.a.db.ondigitalocean.com:25060)/defaultdb
```

Note that we've removed `?ssl-mode=REQUIRED` and wrapped the host/port in `tcp()`.

```bash
export DATASTORE="mysql://doadmin:80624d3936dfc8d2e80593@tcp(db-mysql-lon1-90578-do-user-6456202-0.a.db.ondigitalocean.com:25060)/defaultdb
```

You can prefix this command with `  ` two spaces, to prevent it being cached in your bash history.

* Create three VMs

Imagine we have the following three VMs, two will be servers, and one will be an agent.

```bash
export SERVER1=104.248.135.109
export SERVER2=104.248.25.221
export AGENT1=104.248.137.25
```

* Install the first server

```bash
k3sup install --user root --ip $SERVER1 --datastore="${DATASTORE}"
```

* Install the second server

```bash
k3sup install --user root --ip $SERVER2 --datastore="${DATASTORE}"
```

* Join the first agent

You can join the agent to either server, the datastore is not required for this step.

```bash
k3sup join --user root --server-ip $SERVER1 --ip $AGENT1
```

* Additional steps

If you run `kubectl get node`, you'll now see two masters/servers and one agent, however, we joined the agent to the first server. If the first server goes down, the agent will effectively also go offline.

```bash
kubectl get node

NAME              STATUS                        ROLES    AGE     VERSION
k3sup-1           Ready                         master   73s     v1.19.6+k3s1
k3sup-2           Ready                         master   2m31s   v1.19.6+k3s1
k3sup-3           Ready                         <none>   14s     v1.19.6+k3s1
```

There are two ways to prevent a dependency on the IP address of any one host. The first is to create a TCP load-balancer in the cloud of your choice, the second is for you to create a DNS round-robbin record, which contains all of the IPs of your servers.

In your DigitalOcean dashboard, go to the Networking menu and click "Load Balancer", create one in the same region as your Droplets and SQL server. Select your two Droplets, i.e. `104.248.34.61` and `142.93.175.203`, and use `TCP` with port `6443`.

If you want to run `k3sup join` against the IP of the LB, then you should also add `TCP` port `22`

Make sure that the health-check setting is also set to `TCP` and port `6443`. Wait to get your IP, mine was: `174.138.101.83`

Save the LB into an environment variable:

```bash
export LB=174.138.101.83
```

Now use `ssh` to log into both of your servers, and edit their config files at `/etc/systemd/system/k3s.service`, update the lines `--tls-san` and the following address, to that of your LB:

```
ExecStart=/usr/local/bin/k3s \
    server \
        '--tls-san' \
        '104.248.135.109' \
```

Becomes:

```
ExecStart=/usr/local/bin/k3s \
    server \
        '--tls-san' \
        '174.138.101.83' \
```

Now run:

```bash
sudo systemctl daemon-reload && \
  sudo systemctl restart k3s-agent
```

And repeat these steps on the other server.

You can update the agent manually, via ssh and edit `/etc/systemd/system/k3s-agent.service.env` on the host, or use `k3sup join` again, but only if you added port `22` to your LB:

```bash
k3sup join --user root --server-ip $LB --ip $AGENT1
```

Finally, regenerate your KUBECONFIG file with the LB's IP, instead of one of the servers:

```bash
k3sup install --skip-install --ip $LB
```

Log into the first server, and stop k3s `sudo systemctl stop k3s`, then check that kubectl still functions as expected:

```bash
export KUBECONFIG=`pwd`/kubeconfig
kubectl get node -o wide

NAME              STATUS                        ROLES    AGE   VERSION
k3sup-1           NotReady                      master   23m   v1.19.6+k3s1
k3sup-2           Ready                         master   25m   v1.19.6+k3s1
k3sup-3           Ready                         <none>   22m   v1.19.6+k3s1
```

You've just simulated a failure of one of your masters/servers, and you can still access kubectl. Congratulations on building a resilient k3s cluster.

### Create a multi-master (HA) setup with embedded etcd

In k3s `v1.19.5+k3s1` a HA multi-master (multi-server in k3s terminology) configuration is available called "embedded etcd". A quorum of servers will be required, which means having an odd number of nodes and least three. [See more](https://rancher.com/docs/k3s/latest/en/installation/ha-embedded/)

* Initialize the cluster with the first server

Note the `--cluster` flag

```sh
export SERVER_IP=192.168.0.100
export USER=root

k3sup install \
  --ip $SERVER_IP \
  --user $USER \
  --cluster \
  --k3s-version v1.19.1+k3s1
```

* Join each additional server

> Note the new `--server` flag

```sh
export USER=root
export SERVER_IP=192.168.0.100
export NEXT_SERVER_IP=192.168.0.101

k3sup join \
  --ip $NEXT_SERVER_IP \
  --user $USER \
  --server-user $USER \
  --server-ip $SERVER_IP \
  --server \
  --k3s-version v1.19.1+k3s1
```

Now check `kubectl get node`:

```sh
kubectl get node
NAME              STATUS   ROLES    AGE     VERSION
paprika-gregory   Ready    master   8m27s   v1.19.2-k3s
cave-sensor       Ready    master   27m     v1.19.2-k3s
```

### 👨‍💻 Micro-tutorial for Raspberry Pi (2, 3, or 4) 🥧

In a few moments you will have Kubernetes up and running on your Raspberry Pi 2, 3 or 4. Stand by for the fastest possible install. At the end you will have a KUBECONFIG file on your local computer that you can use to access your cluster remotely.

![](./docs/k3sup-rpi.png)
*Conceptual architecture, showing `k3sup` running locally against bare-metal ARM devices.*

* [Download etcher.io](https://www.balena.io/etcher/) for your OS

* Flash an SD card using [Raspbian Lite](https://www.raspberrypi.org/downloads/raspbian/)

* Enable SSH by creating an empty file named `ssh` in the boot partition

* Generate an ssh-key if you don't already have one with `ssh-keygen` (hit enter to all questions)

* Find the RPi IP with `ping -c raspberrypi.local`, then set `export SERVER_IP=""` with the IP

* Enable container features in the kernel, by editing `/boot/cmdline.txt`

* Add the following to the end of the line: ` cgroup_enable=cpuset cgroup_memory=1 cgroup_enable=memory`

* Copy over your ssh key with: `ssh-copy-id pi@raspberrypi.local`

* Run `k3sup install --ip $SERVER_IP --user pi`

* Point at the config file and get the status of the node:

```sh
export KUBECONFIG=`pwd`/kubeconfig
kubectl get node -o wide
```

You now have `kubectl` access from your laptop to your Raspberry Pi running k3s.

If you want to join some nodes, run `export IP=""` for each additional RPi, followed by:

* `k3sup join --ip $IP --server-ip $SERVER_IP --user pi`

> Remember all these commands are run from your computer, not the RPi.

Now where next? I would recommend my detailed tutorial where I spend time looking at how to flash the SD card, deploy k3s, deploy OpenFaaS (for some useful microservices), and then get incoming HTTP traffic.

Try it now: [Will it cluster? K3s on Raspbian](https://blog.alexellis.io/test-drive-k3s-on-raspberry-pi/)

## Caveats on security

If you are using public cloud, then make sure you see the notes from the Rancher team on setting up a Firewall or Security Group.

k3s docs: [k3s configuration / open ports](https://rancher.com/docs/k3s/latest/en/installation/installation-requirements/#networking)

## If your ssh-key is password-protected

If the ssh-key is encrypted the first step is to try to connect to the ssh-agent. If this works, it will be used to connect to the server.
If the ssh-agent is not running, the user will be prompted for the password of the ssh-key.

On most Linux systems and MacOS, ssh-agent is automatically configured and executed at login. No additional actions are required to use it.

To start the ssh-agent manually and add your key run the following commands:

```bash
eval `ssh-agent`
ssh-add ~/.ssh/id_rsa
```

You can now just run k3sup as usual. No special parameters are necessary.

```bash
k3sup --ip $IP --user user
```

## Contributing

### Sponsor on GitHub ☕️ 👏

k3sup is free and open source, but requires time and effort to support users and build and test new features. Support this project via [GitHub Sponsors](https://github.com/users/alexellis/sponsorship).

### Blog posts & tweets

Blogs posts, tutorials, and Tweets about k3sup (`#k3sup`) are appreciated. Please send a PR to the README.md file to add yours.

### Contributing via GitHub

Before contributing code, please see the [CONTRIBUTING guide](https://github.com/alexellis/inlets/blob/master/CONTRIBUTING.md). Note that k3sup uses the same guide as [inlets.dev](https://inlets.dev/).

Both Issues and PRs have their own templates. Please fill out the whole template.

All commits must be signed-off as part of the [Developer Certificate of Origin (DCO)](https://developercertificate.org)

### License

MIT

## 📢 What are people saying about `k3sup`?

* [Five years of Raspberry Pi clusters](https://www.raspberrypi.org/blog/five-years-of-raspberry-pi-clusters/) - raspberrypi.org

* [Multi-master HA Kubernetes in < 5 minutes](https://blog.alexellis.io/multi-master-ha-kubernetes-in-5-minutes/) by Alex Ellis

* [Kubernetes Homelab with Raspberry Pi and k3sup](https://blog.alexellis.io/raspberry-pi-homelab-with-k3sup/)

* [Building a Kubernetes cluster on Raspberry Pi running Ubuntu server](https://medium.com/icetek/building-a-kubernetes-cluster-on-raspberry-pi-running-ubuntu-server-8fc4edb30963) by Jakub Czapliński

* [Multi-node Kubernetes on Civo in 5 minutes flat with k3sup!](https://www.civo.com/learn/kubernetes-on-civo-in-5-minutes-flat) - Civo Learn guide

* [Zero to k3s Kubeconfig in seconds on AWS EC2 with k3sup](https://rancher.com/blog/2019/k3s-kubeconfig-in-seconds) by Saiyam Pathak

* [Create a 3-node k3s cluster with k3sup & DigitalOcean](https://blog.alexellis.io/create-a-3-node-k3s-cluster-with-k3sup-digitalocean/)

* [Cheap k3s cluster using Amazon Lightsail](https://eamonbauman.com/2020/05/09/cheap-k3s-cluster-using-amazon-lightsail/)

* [k3sup mentioned on Kubernetes Podcast episode 67](https://kubernetespodcast.com/episode/067-orka/) by Craig Box & Adam Glick

* [Scheduling Kubernetes workloads to Raspberry Pi using Inlets and Crossplane](https://github.com/crossplaneio/tbs/blob/master/episodes/9/assets/README.md) by [Daniel Mangum](https://github.com/hasheddan)
  * Also checkout the live [demo](https://youtu.be/RVAFEAnirZA)

* Blog post by Ruan Bekker:

    > Provision k3s to all the places with a awesome utility called "k3sup" by @alexellisuk. Definitely worth checking it out, its epic!

    [Provision k3s on the fly with k3sup](https://sysadmins.co.za/provision-k3s-on-the-fly-with-k3sup/)

* [Dave Cadwallader (@geek_dave)](https://twitter.com/geek_dave/status/1162386683200851969?s=20):

    > Alex - Thanks so much for all the effort you put into your tools and tutorials.  My rpi homelab has been a valuable learning playground for CNCF tech thanks to you!

* [k3sup in KubeWeekly #181](https://kubeweekly.io/2019/08/22/kubeweekly-181/)

* [Will it cluster? K3s on Raspbian](https://blog.alexellis.io/test-drive-k3s-on-raspberry-pi/)

* [Kubernetes the Easy Way with k3sup – Cisco DevOps Series](https://blogs.cisco.com/developer/kubernetes-the-easy-way-devops-14)

* [Trying tiny k3s on Google Cloud with k3sup](https://starkandwayne.com/blog/trying-tiny-k3s-on-google-cloud-with-k3sup/) by Stark & Wayne

* [Setting up a Raspberry Pi Kubernetes Cluster with Blinkt! Strips that Show Number of Pods per Node](https://pleasereleaseme.net/setting-up-a-raspberry-pi-kubernetes-cluster-with-blinkt-strips-that-show-number-of-pods-per-node-using-k3sup/)

* [From Zero to Kubernetes Dashboard within Minutes with k3sup and Kontena Lens](https://medium.com/@laurinevala/from-zero-to-kubernetes-dashboard-within-minutes-with-k3sup-and-kontena-lens-84f881400b10) - by Lauri Nevala

* [BYOK - Build your Own Kubernetes Cluster with Raspberry Pis, k3s, and k3sup](https://speakerdeck.com/mikesir87/byok-build-your-own-kubernetes-cluster-with-raspberry-pis-k3s-and-k3sup) by Michael Irwin

* [Run Kubernetes On Your Machine with k3sup](https://itnext.io/run-kubernetes-on-your-machine-7ee463af21a2) by Luc Juggery

* [Trying out k3sup](https://blog.baeke.info/2019/10/25/trying-out-k3sup/) by Geert Baeke

* [Creating your first Kubernetes cluster with k3sup](https://dev.to/kalaspuffar/creating-your-first-kubernetes-cluster-3kp2) by Daniel Persson

* [My 2019 In Review - Hello Open Source](https://blog.heyal.co.uk/My-2019/) by Alistair Hey

* [Kubernetes 104: Create a 2-node k3s cluster with k3sup](https://ahmed-abdelsamad.blogspot.com/2019/09/kubernetes-104-create-2-node-k3s.html) by Ahmed Abelsamad

* [My home Kubernetes cluster driven by GitOps and k3sup](https://github.com/onedr0p/k3s-gitops) - by Devin Buhl

* [Raspberry Pi: From 0 to k3s cluster in 5 min with k3sup and Ansible](https://blog.cloudgsx.es/topic/10/raspberry-pi-from-0-to-k3s-cluster-in-5-min-with-k3sup-and-ansible) - by Pablo Caderno

* [Kubernetes: from Zero to Hero with Kompose, Minikube, k3sup and Helm](https://blog.mi.hdm-stuttgart.de/index.php/2020/02/29/image-editor-on-kubernetes-with-kompose-minikube-k3s-k3sup-and-helm-part-2/) by Leon Klingele, Alexander Merker & Florian Wintel

* [Deploying a highly-available K3s with k3sup](https://ma.ttias.be/deploying-highly-available-k3s-k3sup/) by Dmitriy Akulov

* [Multi-master HA Kubernetes using K3Sup on Windows10/Server 2019](https://github.com/TechGuyTN/K3Sup-Windows10/blob/d7ad4f642ae6ebf441b8137bd71111c9c2890add/README.md) by Aaron Holt

* [Ansible Role: k3sup](https://github.com/vandot/ansible-role-k3sup) by Ivan Vandot

* [HashiCorp Vault on K3s on RPi4](https://github.com/colin-mccarthy/k3s-pi-vault/tree/d5e616de9048da8b990dc7b99a6d2d96bd9e9cc5) by Colin McCarthy

* [Raspberry Pi Cluster Part 2: ToDo API running on Kubernetes with k3s](https://www.dinofizzotti.com/blog/2020-05-09-raspberry-pi-cluster-part-2-todo-api-running-on-kubernetes-with-k3s/) by Dino Fizzotti

* [Cloud Native Tools for Developers Webinar Recap](https://www.openfaas.com/blog/cloud-native-tools-webinar/) by Alex Ellis & Alistair Hey

* [Unobtrusive local development with kubernetes, k3s, traefik2](https://www.codementor.io/@slavko/unobtrusive-local-development-with-kubernetes-k3s-traefik2-15uq596oja) by Vyacheslav

* [Cómo desplegar un clúster de Kubernetes en 60 segundos con k3sup](https://www.cduser.com/como-desplegar-un-cluster-de-kubernetes-en-60-segundos-con-k3sup/) by Ignacio Van Droogenbroeck

* [Raspberry Kubernetes cluster for my homlab with k3sup](https://hybridhacker.com/homelab-raspberry-kubernetes-cluster-with-k3sup.html) by Nicola Ballotta

* [k3sup On MacOS Catalina](https://gizmo.codes/k3sup-on-macos-catalina/) by John Doyle

* [Provision k3s on Google Cloud with Terraform and k3sup — Nimble](https://medium.com/nimble/provision-k3s-on-google-cloud-with-terraform-and-k3sup-nimble-38fa3167b4c3) by Carlos Herrera

* [K3s: Edge Kubernetes](https://redmonk.com/jgovernor/2020/01/31/k3s-edge-kubernetes/) by James Governor

* [OpenStack sur LXD avec Juju et k3sup dans phoenixNAP](https://deep75.medium.com/openstack-sur-lxd-avec-juju-et-k3sup-dans-phoenixnap-e5867a487497)

* [Creating a k3s Cluster with k3sup & Multipass 💻☸️](https://dev.to/tomowatt/creating-a-k3s-cluster-with-k3sup-multipass-h26) by Tom Watt

* [How I’ve set up my highly-available Kubernetes cluster](https://jmac.ph/2021/01/25/how-ive-set-up-my-highly-available-kubernetes-cluster/) by JJ Macalinao

* [Creating a K3S Raspberry PI Cluster with K3Sup to fire up nightscout with MongoDB](https://h3rmanns.medium.com/creating-a-k3s-raspberry-pi-cluster-with-k3sup-to-fire-up-a-nightscout-backend-service-based-on-a-27c1f5727e5b)

Checkout the [Announcement tweet](https://twitter.com/alexellisuk/status/1162272786250735618?s=20)

## Similar tools & glossary

Glossary:

* Kubernetes: master/slave
* k3s: server/agent

Related tools:

* [k3s](https://github.com/rancher/k3s) - Kubernetes as installed by `k3sup`. k3s is a compliant, light-weight, multi-architecture distribution of Kubernetes. It can be used to run Kubernetes locally or remotely for development, or in edge locations.
* [k3d](https://github.com/rancher/k3d) - this tool runs a Docker container on your local laptop with k3s inside
* [kind](https://github.com/kubernetes-sigs/kind) - kind can run a Kubernetes cluster within a Docker container for local development. k3s is also suitable for this purpose through `k3d`. KinD is not suitable for running a remote cluster for development.
* [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/) - a tool to create fully-loaded, production-ready Kubernetes clusters with or without high-availability (HA). Tends to be heavier-weight and slower than k3s. It is aimed at cloud VMs or bare-metal computers which means it doesn't always work well with low-powered ARM devices.
* [k3v](https://github.com/ibuildthecloud/k3v) - "virtual kubernetes" - a very early PoC from the author of k3s aiming to slice up a single cluster for multiple tenants
* [k3sup-multipass](https://github.com/matti/k3sup-multipass) - a helper to launch single node k3s cluster with one command using a [multipass](https://multipass.run) VM and optionally proxy the ingress to localhost for easier development.

## Troubleshooting

If you're having issues, it's likely that this is a problem with K3s, and not with K3sup. How do we know that? Mostly from past issues.

Rancher provides support for K3s [on their Slack](https://slack.rancher.io/) in the `#k3s` channel. This should be your first port of call. Your second port of call is to raise an issue with the K3s maintainers in the [K3s repo](https://github.com/k3s-io/k3s/issues)

Common issues:

* Raspberry Pi - you haven't updated cmdline.txt to enable cgroups for CPU and memory
* K3s server didn't start. Log in and run `sudo systemctl -u k3s`
* The K3s agent didn't start. Log in and run `sudo systemctl -u k3s-agent`
* You tried to remove and re-add a server in an etcd cluster and it failed. This is a known issue, see the [K3s issue tracker](https://github.com/k3s-io/k3s/issues).
* You tried to use an unsupported version of a database for HA. See [this list from Rancher](https://rancher.com/docs/k3s/latest/en/installation/datastore/)

Finally, if everything points to an issue that you can clearly reproduce with k3sup, feel free to open an issue here. To make sure you get a response, fill out the whole template and answer all the questions.

### Getting access to your KUBECONFIG

You may have run into an issue where `sudo` access is required for `kubectl` access.

You should not run kubectl on your server or agent nodes. k3sup is designed to rewrite and/or merge your cluster's config to your local KUBECONFIG file. You should run `kubectl` on your laptop / client machine.

If you've lost your kubeconfig, you can use `k3sup install --skip-install`. See also the various flags for merging and setting a context name.

### Smart cards and 2FA

> Warning: issues requesting support for smart cards / 2FA will be closed immediately. The feature has been proven to work, and is provided as-is. We do not have time to debug your system.

You can use a smart card or 2FA security key such as a Yubikey. You must have your ssh-agent configured correctly, at that point k3sup will defer to the agent to make connections on MacOS and Linux. [Find out more](https://github.com/alexellis/k3sup/pull/312)

### Misc note on `iptables`

> Note added by Eduardo Minguez Perez

Currently there is an issue in k3s involving `iptables >= 1.8` that can affect the network communication. See the [k3s issue](https://github.com/rancher/k3s/issues/703) and the corresponding [kubernetes one](https://github.com/kubernetes/kubernetes/issues/71305) for more information and workarounds. The issue has been observed in Debian Buster but it can affect other distributions as well.
