# Development Environment setup for Terraform Plugin Development

Note: Terraform is a linux tool

## Step 1: Windows 10

1. Install WLS (windows linux subsystem) https://docs.microsoft.com/en-us/windows/wsl/install
2. Install golang (defualt GOROOT)
3. Install Visual Code
4. source code directory %USERDIR%/go/src (default GOPATH)

## WSL Setup to access your code

1. install golang
2.
3. link directory ln $HOME/go/src /mnt/<user>/go/src
4. setup wsl to connect to external server with VPN : Excellent article follow the steps https://jamespotz.github.io/blog/how-to-fix-wsl2-and-cisco-vpn
   -- Issue: resolv.conf get writen when WSL get restarted, the global config .wslconfig does not work. Need to redo at the start of the development
5. setup git to connect to bitbucket using ssh - TBD

## Visual Code - Setup

1. WSL setup
2. Open the ipm folder in VSC - important, otherwise you will have broken links to

## VPN command to connect to external server from WSL

Get-NetAdapter | Where-Object {$\_.InterfaceDescription -Match "Cisco AnyConnect"} | Set-NetIPInterface -InterfaceMetric 6000
