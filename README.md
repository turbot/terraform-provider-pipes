# Terraform Turbot Pipes provider

- Terraform: https://www.terraform.io
- Steampipe: https://steampipe.io
- Turbot Pipes: https://pipes.turbot.com
- Community: [Steampipe Slack](https://turbot.com/community/join)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.10.x
- [Go](https://golang.org/doc/install) 1.14 (to build the provider plugin)

## Building The Provider

Clone repository to: `$GOPATH/src/github.com/turbot/terraform-provider-pipes`

```sh
$ export GOPATH=$(go env GOPATH)
$ mkdir -p $GOPATH/src/github.com/turbot; cd $GOPATH/src/github.com/turbot
$ git clone git@github.com:turbot/terraform-provider-pipes
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/turbot/terraform-provider-pipes
$ make build
```

## Using the provider

If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory, run `terraform init` to initialize it.

Further [usage documentation is available on the Terraform website](https://registry.terraform.io/providers/turbot/pipes/latest/docs).

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is _required_). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
go build -o bin/terraform-provider-pipes_0.0.1 -ldflags="-X github.com/turbot/terraform-provider-pipes/version.ProviderVersion=0.0.1"
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
