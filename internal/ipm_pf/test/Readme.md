WSL - commands to connect to plgd server; Cisco VPN connect
PowerShell - run in super user mode
Get-NetAdapter | Where-Object {$_.InterfaceDescription -Match "Cisco AnyConnect"} | Set-NetIPInterface -InterfaceMetric 6000


Plgd server:https://sv-kube-prd.infinera.com:443

1. Access Token
export ACCESS_TOKEN=$(curl -ks 'https://sv-kube-prd.infinera.com:443/oauth/token?client_id=test&audience=test' | jq -r .access_token)

2. Get Devices
curl -ks -XGET 'https://sv-kube-prd.infinera.com:443/api/v1/devices' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $ACCESS_TOKEN"

3. Get Cfg

curl -ks -XGET 'https://sv-kube-prd.infinera.com:443/api/v1/devices/1a05b61c-da5c-52ec-8680-563065b8fc00/reources/cfg' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $ACCESS_TOKEN"

4. Update Cfg


 curl -ks -XPUT 'https://sv-kube-prd.infinera.com:443/api/v1/devices/de6706f6-f380-47ba-4490-abc7692ee0d0/resou
rces/cfg' --header 'Content-Type: application/json' --header "Authorization: Bearer $ACCESS_TOKEN" -d ' {
    "configuredRole": "hub",
    "trafficMode": "L1Mode",
    "fiberConnectionMode": "dual"
}'


Terrafrom debug

export TF_LOG=DEBUG

TODO : VS code Debugger 
https://www.terraform.io/plugin/sdkv2/debugging