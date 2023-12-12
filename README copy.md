# Terraform IPM Service Provider

This is a IPM terraform provider which provides CRUD operations for all support IPM Services.
Each service shall have a collection of TF modules for configurations of its resources.

| Service                                                   |  Description                                   | State  |
|-----------------------------------------------------------|------------------------------------------------|--------|
| [Host Management Service](https://github.com/infinera/terraform-ipm_modules/tree/master/module-management-service)                             |                                                | In Progress |
| [Module Management Service](https://github.com/infinera/terraform-ipm_modules/tree/master/module-management-service)                         |                                                | In Progress  |
| [Network Service](https://github.com/infinera/terraform-ipm_modules/tree/master/network-service)                       |                                                | Ready  |
| [Transport Capacity Service](https://github.com/infinera/terraform-ipm_modules/tree/master/transport-capacity-service) |                                                | In Progress  |
| [Network Connection Service](https://github.com/infinera/terraform-ipm_modules/tree/master/network-service) |                                                | In Progress |
| [Module Software Manager](https://github.com/infinera/terraform-ipm_modules/tree/master/module-software-manager)       |                                                |  In Progress  |
| [NDU Service](https://github.com/infinera/terraform-ipm_modules/tree/master/ndu-service)                               |                                                |   In Progress  |
| [Device Aggregator](https://github.com/infinera/terraform-ipm_modules/tree/master/device-aggregator)   |                                                |        |
| [Aggregator Fault Management Service](https://github.com/infinera/terraform-ipm_modules/tree/master/aggregator-fault-management-service)     |                                                |        |
| [Domain Fault Management Service](https://github.com/infinera/terraform-ipm_modules/tree/master/domain-fault-management-service)             |                                                |        |
| [Event Gateway](https://github.com/infinera/terraform-ipm_modules/tree/master/event-gateway)           |                                                |        |
| [Onboarding Tool](https://github.com/infinera/terraform-ipm_modules/tree/master/onboard-tool)          |                                                |        |
| [Author Server](https://github.com/infinera/terraform-ipm_modules/tree/master/author-server) |                  |   |


## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.17
- [make](https://linuxhint.com/install-make-ubuntu/)
- [docker](https://docs.docker.com/engine/install/ubuntu/) Only need to build the IPM modules and provider docker image

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Setup Go for Visual code
    . Add Go for Visual Studio Code extension
    . enter "go mod init terraform-provider-ipm"
    . enter "go mod tidy"
3. Build the IPM provider:

```shell
make install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

# Manage the XR Network via Configuration Intent and IPM Service Docker Image
Please see [Manage the XR Network via Configuration Intent](https://bitbucket.infinera.com/projects/MAR/repos/terraform-provider-ipm/browse/Manage-XR-Network-Using%20-IPM-Services.md) for more details.