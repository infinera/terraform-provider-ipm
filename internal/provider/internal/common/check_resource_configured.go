package common

import (
	"context"
	"encoding/json"
	"errors"
	"terraform-provider-ipm/internal/ipm_pf"
	"time"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)


func CheckResourceState(ctx context.Context, client *ipm_pf.Client, queryString string, tryCount ...int) ( found bool, error error ) {

	numRetry := 1
	if len(tryCount) > 0 {
		numRetry = tryCount[0]
	}

	tflog.Debug(ctx, "CheckResourceState:  ", map[string]interface{}{"query 1": queryString })

	for i := 1; i <= numRetry; i++ {
		data, err := FindResource(client, queryString)
		if err == nil && data["state"] != nil {
				state := data["state"].(map[string]interface{})
				if state["lifecycleState"] != nil  && state["lifecycleState"].(string) == "configured" {
					return true, nil
				}
		} else {
			return false, err
		}
		if i < numRetry {
			time.Sleep(15 * time.Second)
		}
	} 
	return false, errors.New("The resource is not in \"Configured\" state")
}

func CheckNetworkState(ctx context.Context, client *ipm_pf.Client, queryStrings []string, endpointIds []string, tryCount ...int) ( found bool, error error ) {
	numRetry := 1
	if len(tryCount) > 0 {
		numRetry = tryCount[0]
	}

	tflog.Debug(ctx, "CheckNetworkState:  ", map[string]interface{}{"query 1": queryStrings[0], "query 2": queryStrings[1] })

	query := queryStrings[0]
	useQuery2 := false
	for i := 1; i <= numRetry; i++ {
		data, err := FindResource(client, query)
		if err != nil && !useQuery2 {
			useQuery2 = true
			query = queryStrings[1]
			data, err = FindResource(client, query)
		}
		if err == nil && data["state"] != nil {
			networkState := data["state"].(map[string]interface{})
			networkLCState := networkState["lifecycleState"]
			if networkLCState != nil && (networkLCState.(string) == "configured" ||  networkLCState.(string) == "pendingConfiguration") {
				leafModules := data["leafModules"].([]interface{})
				if len(leafModules) > 0  {
					for _, v := range leafModules {
						leaf := v.(map[string]interface{})
						state := leaf["state"].(map[string]interface{})
						if state["lifecycleState"] == "configured" {
							leafConfig := leaf["config"].(map[string]interface{})
							if leafConfig["managedBy"].(string) != "host" {
								leafConfigSelector := leafConfig["selector"].(map[string]interface{})
								for k2, v2 := range leafConfigSelector {
									leafSelector := v2.(map[string]interface{})
									switch k2 {
									case "moduleSelectorByModuleId":
										id := leafSelector["moduleId"].(string)
										if id == endpointIds[0] || id == endpointIds[1] {
											return true, nil
										}
									case "moduleSelectorByModuleName":
										id := leafSelector["moduleName"].(string)
										if id == endpointIds[0] || id == endpointIds[1] {
											return true, nil
										}
									case "moduleSelectorByModuleMAC":
										id := leafSelector["moduleMAC"].(string)
										if id == endpointIds[0] || id == endpointIds[1] {
											return true, nil
										}
									case "moduleSelectorByModuleSerialNumber":
										id := leafSelector["moduleSerialNumber"].(string)
										if id == endpointIds[0] || id == endpointIds[1] {
											return true, nil
										}
									case "hostPortSelectorByName":
										id := leafSelector["hostName"].(string) + ":" + leafSelector["hostPortName"].(string)
										if id == endpointIds[0] || id == endpointIds[1] {
											return true, nil
										}
									case "hostPortSelectorByPortId":
										id := leafSelector["chassisId"].(string) + ":" + leafSelector["portId"].(string)
										if id == endpointIds[0] || id == endpointIds[1] {
											return true, nil
										}
									case "hostPortSelectorBySysName":
										id := leafSelector["sysName"].(string) + ":" + leafSelector["portId"].(string)
										if id == endpointIds[0] || id == endpointIds[1] {
											return true, nil
										}
									case "hostPortSelectorByPortSourceMAC":
										id := leafSelector["portSourceMAC"].(string) 
										if id == endpointIds[0] || id == endpointIds[1] {
											return true, nil
										}
									}
								}
							} else {
								leafModule := state["module"].(map[string]interface{})
								if leafModule["moduleId"].(string) == endpointIds[0] || leafModule["moduleId"].(string) == endpointIds[1] || leafModule["moduleName"].(string) == endpointIds[0] || leafModule["moduleName"].(string) == endpointIds[1] || leafModule["macAddress"].(string) == endpointIds[0] || leafModule["macAddress"].(string) == endpointIds[1] || leafModule["serialNumber"].(string) == endpointIds[0] || leafModule["serialNumber"].(string) == endpointIds[1] {
									return true, nil
								}
							}
						}
					}
				} else {
					return false, errors.New("Constellation has no leaf: " + networkState["name"].(string))
				}
			} 
		} else {
			return false, err
		}
		if i < numRetry {
			time.Sleep(15 * time.Second)
		}
	} 
	return false, errors.New("The resource is not in \"Configured\" state")
}


func FindResource(client *ipm_pf.Client, queryString string) ( data map[string]interface{}, error error ) {
	body, err := client.ExecuteIPMHttpCommand("GET", queryString, nil)
	if err != nil {
		return nil, errors.New("Can't get the resource: " + queryString)
	}
	var resps []interface{}
	err = json.Unmarshal(body, &resps)
	if err != nil {
		return nil, errors.New("Can't unmarshall the resource's data: " + queryString)
	}
	if len(resps) > 0 {
		return resps[0].(map[string]interface{}), nil
	}
	return nil, errors.New("Can't find resource for query string: " + queryString)
}