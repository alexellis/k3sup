## go-execute

A wrapper for Go's command execution packages.

`go get github.com/alexellis/go-execute/v2`

## Docs

See docs at pkg.go.dev: [github.com/alexellis/go-execute](https://pkg.go.dev/github.com/alexellis/go-execute)

## go-execute users

[Used by dozens of projects as identified by GitHub](https://github.com/alexellis/go-execute/network/dependents), notably:

* [alexellis/arkade](https://github.com/alexellis/arkade)
* [openfaas/faas-cli](https://github.com/openfaas/faas-cli)
* [inlets/inletsctl](https://github.com/inlets/inletsctl)
* [inlets/cloud-provision](https://github.com/inlets/cloud-provision)
* [alexellis/k3sup](https://github.com/alexellis/k3sup)
* [openfaas/connector-sdk](https://github.com/openfaas/connector-sdk)
* [openfaas-incubator/ofc-bootstrap](https://github.com/openfaas-incubator/ofc-bootstrap)

Community examples:

* [dokku/lambda-builder](https://github.com/dokku/lambda-builder)
* [027xiguapi/pear-rec](https://github.com/027xiguapi/pear-rec)
* [cnrancher/autok3s](https://github.com/cnrancher/autok3s)
* [ainsleydev/hupi](https://github.com/ainsleydev/hupi)
* [andiwork/andictl](https://github.com/andiwork/andictl)
* [tonit/rekind](https://github.com/tonit/rekind)
* [lucasrod16/ec2-k3s](https://github.com/lucasrod16/ec2-k3s)
* [seaweedfs/seaweed-up](https://github.com/seaweedfs/seaweed-up)
* [jsiebens/inlets-on-fly](https://github.com/jsiebens/inlets-on-fly)
* [jsiebens/hashi-up](https://github.com/jsiebens/hashi-up)
* [edgego/ecm](https://github.com/edgego/ecm)
* [ministryofjustice/cloud-platform-terraform-upgrade](https://github.com/ministryofjustice/cloud-platform-terraform-upgrade)
* [mattcanty/go-ffmpeg-transcode](https://github.com/mattcanty/go-ffmpeg-transcode)
* [Popoola-Opeyemi/meeseeks](https://github.com/Popoola-Opeyemi/meeseeks)
* [aidun/minicloud](https://github.com/aidun/minicloud)

Feel free to add a link to your own projects in a PR.

## Main options

* `DisableStdioBuffer` - Discard Stdio, rather than buffering into memory
* `StreamStdio` - Stream stderr and stdout to the console, useful for debugging and testing
* `Shell` - Use bash as a shell to execute the command, rather than exec a binary directly
* `StdOutWriter` - an additional writer for stdout, useful for mutating or filtering the output
* `StdErrWriter` - an additional writer for stderr, useful for mutating or filtering the output
* `PrintCommand` - print the command to stdout before executing it

## Example of exec without streaming to STDIO

This example captures the values from stdout and stderr without relaying to the console. This means the values can be inspected and used for automation.

```golang
package main

import (
	"fmt"

	execute "github.com/alexellis/go-execute/v2"
	"context"
)

func main() {
	cmd := execute.ExecTask{
		Command:     "docker",
		Args:        []string{"version"},
		StreamStdio: false,
	}

	res, err := cmd.Execute(context.Background())
	if err != nil {
		panic(err)
	}

	if res.ExitCode != 0 {
		panic("Non-zero exit code: " + res.Stderr)
	}

	fmt.Printf("stdout: %s, stderr: %s, exit-code: %d\n", res.Stdout, res.Stderr, res.ExitCode)
}
```

## Example with "shell" and exit-code 0

```golang
package main

import (
	"fmt"

	execute "github.com/alexellis/go-execute/v2"
	"context"
)

func main() {
	ls := execute.ExecTask{
		Command: "ls",
		Args:    []string{"-l"},
		Shell:   true,
	}
	res, err := ls.Execute(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Printf("stdout: %q, stderr: %q, exit-code: %d\n", res.Stdout, res.Stderr, res.ExitCode)
}
```

## Example with "shell" and exit-code 1

```golang
package main

import (
	"fmt"

	"context"
	execute "github.com/alexellis/go-execute/v2"
)

func main() {
	ls := execute.ExecTask{
		Command: "exit 1",
		Shell:   true,
	}
	res, err := ls.Execute(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Printf("stdout: %q, stderr: %q, exit-code: %d\n", res.Stdout, res.Stderr, res.ExitCode)
}
```

## Contributing

Commits must be signed off with `git commit -s`

License: MIT
