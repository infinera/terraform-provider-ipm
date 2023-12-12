package common

import (
	"context"
	"encoding/json"
	"errors"
	//"go/types"
	"strconv"
	"strings"

	"github.com/fujiwara/tfstate-lookup/tfstate"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getContent(body []byte) (c map[string]interface{}, err error) {
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	var content map[string]interface{}
	content, contentOk := data["content"].(map[string]interface{})
	if !contentOk {
		return nil, errors.New("getResourceIdNContent: No Content data:")
	} else {
		return content, nil
	}
}

func getResourceIdNContent(body []byte) (r map[string]interface{}, c map[string]interface{}, err error) {
	var data1 = make(map[string]interface{})
	err = json.Unmarshal(body, &data1)
	if err != nil {
		return nil, nil, err
	}

	data, ok := data1["data"].(map[string]interface{})
	if !ok {
		return nil, nil, err
	}
	var resourceId map[string]interface{}
	resourceId, resourceIdOk := data["resourceId"].(map[string]interface{})

	var content map[string]interface{}
	content, contentOk := data["content"].(map[string]interface{})

	if !contentOk && !resourceIdOk {
		return nil, nil, errors.New("getResourceIdNContent: No ResourceID and Content data:")
	} else if !contentOk {
		return resourceId, nil, errors.New("getResourceIdNContent: No Content data")
	} else if !resourceIdOk {
		return nil, content, errors.New("getResourceIdNContent: No ResourceID data")
	} else {
		return resourceId, content, nil
	}
}

// func GetData(plan *interface{}, body []byte) (c map[string]interface{}, err error) {
func SetResourceId(deviceName string, Id *types.String, body []byte) (c map[string]interface{}, err error) {

	var resourceId map[string]interface{}
	var content map[string]interface{}
	resourceId, content, _ = getResourceIdNContent(body)

	if resourceId == nil && content == nil {
		return nil, errors.New("SetResourceId: No ResourceID and/or Content data:" + string(body))
	}

	if Id != nil && len(Id.ValueString()) <= 0 {
		if content["href"] == nil {
			*Id = types.StringValue(deviceName + resourceId["href"].(string))
		} else {
			*Id = types.StringValue(deviceName + content["href"].(string))
		}
	}
	return content, nil
}

func StringAfter(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	return value[pos+1:]
}

var mytfstate *tfstate.TFState = nil
var tfstatefile string = ""

func GetTFState(ctx context.Context, file string) (*tfstate.TFState, error) {
	if mytfstate == nil || tfstatefile != file {
		tfstatefile = file
		mytfstate, _ = tfstate.ReadFile(ctx, file)
	}
	return mytfstate, nil
}

func LookupTFState(key string) (interface{}, error) {
	if mytfstate != nil {
		value, _ := mytfstate.Lookup(key)
		if value != nil {
			return value.Value, nil
		}
	}
	return nil, nil
}

func GetAndLookupTFState(ctx context.Context, file string, key string) (interface{}, error) {
	state, _ := GetTFState(ctx, file)
	if state != nil {
		value, _ := state.Lookup(key)
		if value != nil {
			return value.Value, nil
		}
	}
	return nil, nil
}

func Find(what string, where []types.String) (idx int) {
	for i, v := range where {
		if v.ValueString() == what {
			return i
		}
	}
	return -1
}

// Sets the bit at pos in the integer n.
func setBit(n int, pos uint) int {
	n |= (1 << pos)
	return n
}

// Clears the bit at pos in n.
func clearBit(n int, pos uint) int {
	mask := ^(1 << pos)
	n &= mask
	return n
}

func hasBit(n int, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func setBits(positions []int) int {
	n := 0
	for _, v := range positions {
		n = setBit(n, uint(v))
	}
	return n
}

func getBits(n int) []attr.Value {
	var bits []attr.Value
	for i := 0; i < 16; i++ {
		if hasBit(n, uint(i)) {
			bits = append(bits, types.Int64Value(int64(i)))
		}
	}
	return bits
}

func ToAttributeListValues(values []interface{}) []attr.Value {
	var attributeValues []attr.Value
	for _, v := range values {
		attributeValues = append(attributeValues, types.Int64Value(int64(v.(float64))))
	}
	return attributeValues
}

func ToAttributeMapValues(data interface{}) (map[string]attr.Value) {
	capMap := make(map[string]attr.Value)
	for k, v2 := range data.(map[string]interface{}) {
			capMap[k] = types.StringValue(v2.(string))
	}
	return capMap
}

func ToStringList(data []interface{}) []string {
	vs := make([]string, 0, len(data))
	for _, v := range data {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, val)
		}
	}
	return vs
}

func ToInt64List(data []interface{}) []types.Int64 {
	var vi []types.Int64
	for _, v := range data {
		val := types.Int64Value(int64(v.(float64)))
		vi = append(vi, val)
	}
	return vi
}

func ListAttributeInt64Value(data []interface{}) ([]attr.Value) {
	values := []attr.Value{}
	for _, v := range data {
		values = append(values, types.Int64Value(int64(v.(float64))))
	}
	return values
}

func ListAttributeStringValue(data []interface{}) ([]attr.Value) {
	values := []attr.Value{}
	for _, v := range data {
		values = append(values, types.StringValue(v.(string)))
	}
	return values
}


func ListStringValue(data []interface{}) ([]types.String) {
	values := []types.String{}
	for _, v := range data {
		values = append(values, types.StringValue(v.(string)))
	}
	return values
}

func ListValue(data []types.String) ([]string) {
	values := []string{}
	for _, v := range data {
		values = append(values, v.ValueString())
	}
	return values
}

func MapAttributeValue(data map[string]interface{}) (map[string]attr.Value) {
	values := map[string]attr.Value{}
	for k, v := range data {
		values[k] = types.StringValue(v.(string))
	}
	return values
}

// contains checks if a string is present in a slice
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func GetColId(s string, pattern string) int64 {
	colIdStr := GetId(s, pattern)
	if( len(colIdStr) == 0 ) {
		return -1
	}
	colId, err := strconv.ParseInt(colIdStr, 10, 64)
	if err != nil{
		return -1
	}
	return colId
}

func GetId(s string, pattern string) string {
	items := strings.Split(s, "/")
	for i := 0; i < len(items); i++ {
		if items[i] == pattern  {
			return items[i+1]
		}
	}
	return ""
}
