package mqttserverservice

import (
	"context"
	"encoding/json"
	"strings"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	//"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &MQTTResource{}
	_ resource.ResourceWithConfigure   = &MQTTResource{}
	_ resource.ResourceWithImportState = &MQTTResource{}
)

// NewMQTTResource is a helper function to simplify the provider implementation.
func NewMQTTResource() resource.Resource {
	return &MQTTResource{}
}

type MQTTResource struct {
	client *ipm_pf.Client
}


type MQTTResourceData struct {
	DeviceId         types.String                  `tfsdk:"device_id"`
	ServerId         types.String                  `tfsdk:"server_id"`
	Href             types.String                  `tfsdk:"href"`
	Id               types.String                  `tfsdk:"id"`
	Aid              types.String                  `tfsdk:"aid"`
	Server           types.String                  `tfsdk:"server"`
	Port             types.Int64                   `tfsdk:"port"`
	Kai              types.Int64                   `tfsdk:"kai"`
	CRCode           types.Int64                   `tfsdk:"crcode"`
	Type             types.String                  `tfsdk:"type"`
	SubType          types.Int64                   `tfsdk:"sub_type"`
	Log              types.String                  `tfsdk:"log"`
	LogLevel         types.Int64                   `tfsdk:"log_level"`
	Enabled          types.Bool                    `tfsdk:"enabled"`
	Region           types.String                  `tfsdk:"region"`
}

// Metadata returns the data source type name.
func (r *MQTTResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mqtt_server"
}

// Schema defines the schema for the data source.
func (r *MQTTResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type MQTTResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an MQTT",
		Attributes:  MQTTSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *MQTTResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r MQTTResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MQTTResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "MQTTResource: Create - ", map[string]interface{}{"MQTTResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	//r.create(&data, ctx, &resp.Diagnostics)
	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.Set(ctx, &data)

}

func (r MQTTResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MQTTResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "MQTTResource: Read - ", map[string]interface{}{"MQTTResourceData": data})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r MQTTResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MQTTResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "MQTTResource: Update", map[string]interface{}{"MQTTResourceData": data})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r MQTTResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MQTTResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "MQTTResource: Delete", map[string]interface{}{"MQTTResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *MQTTResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MQTTResource) update(plan *MQTTResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "MQTTResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.DeviceId.IsNull() || plan.ServerId.IsNull() {
		diags.AddError(
			"Error Update MQTTResource",
			"Update: Could not update MQTT Resource, device id or Server id is not specified",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get MQTT config settings
	if !plan.Server.IsNull() {
		updateRequest["server"] = plan.Server.ValueString()
	}
	if !plan.Port.IsNull() {
		updateRequest["port"] = plan.Port.ValueInt64()
	}
	if !plan.Kai.IsNull() {
		updateRequest["kai"] = plan.Kai.ValueInt64()
	}
	if !plan.SubType.IsNull() {
		updateRequest["subType"] = plan.SubType.ValueInt64()
	}
	if !plan.Type.IsNull() {
		updateRequest["type"] = plan.Type.ValueString()
	}
	if !plan.LogLevel.IsNull() {
		updateRequest["logLevel"] = plan.LogLevel.ValueInt64()
	}
	if !plan.Region.IsNull() {
		updateRequest["region"] = plan.Region.ValueString()
	}
	if !plan.Enabled.IsNull() {
		updateRequest["enabled"] = plan.Enabled.ValueBool()
	}
	tflog.Debug(ctx, "MQTTResource: update ## ", map[string]interface{}{"updateRequest": updateRequest})

	// send update request to server
	rb, err := json.Marshal(updateRequest)
	if err != nil {
		diags.AddError(
			"MQTTResource: update ##: Error Update",
			"Update: Could not Marshal MQTTResource, unexpected error: "+err.Error(),
		)
		return
	}

	//command: = devices/e07c21d2-24f7-4e53-4d76-782081410cfa/resources/mqttServers/2'
	body, err := r.client.ExecuteIPMHttpCommand("PUT", "/devices/" + plan.DeviceId.ValueString() + "/resources/mqttServers/" + plan.ServerId.ValueString(), rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"MQTTResource: update ##: Error update MQTTResource",
				"Update:Could not update MQTTResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "MQTTResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"MQTTResource: Update ##: Error Unmarshal response",
			"Update:Could not Update MQTTResource, unexpected error: "+err.Error(),
		)
		return
	}
	plan.Populate(data, ctx, diags)
	//r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "MQTTResource: update failed. Can't find the updated network")
		return
	}

	tflog.Debug(ctx, "MQTTResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *MQTTResource) read(state *MQTTResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "MQTTResource: read ##", map[string]interface{}{"state": state})
	if state.DeviceId.IsNull() || state.ServerId.IsNull() {
		diags.AddError(
			"Error Get MQTTResource",
			"Get: Could not Get MQTT Resource, device id or Server id is not specified",
		)
		return
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", "/devices/" + state.DeviceId.ValueString() + "/resources/mqttServers/" + state.ServerId.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"MQTTResource: read ##: Error Get MQTTResource",
			"Read:Could not get MQTT, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "MQTTResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"MQTTResource: read ##: Error Get MQTT",
			"Read:Could not get MQTT, unexpected error: "+err.Error(),
		)
		return
	}
	// populate network state
	state.Populate(data, ctx, diags)
	tflog.Debug(ctx, "MQTTResource: read SUCCESS ")
}

func (mqttData *MQTTResourceData) Populate(mqtt map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	data := mqtt["data"].(map[string]interface{})
	mqttResourceId := data["resourceId"].(map[string]interface{})
	mqttContent := data["content"].(map[string]interface{})

	tflog.Debug(ctx, "MQTTResource: populate ## ", map[string]interface{}{" Module ID ": data["id"]})
	
	mqttData.DeviceId = types.StringValue(mqttResourceId["deviceId"].(string))
	href := mqttResourceId["href"].(string)
	mqttData.Href = types.StringValue(href)
	serverId := common.StringAfter(href, "/")
	mqttData.ServerId = types.StringValue(serverId)
	mqttData.Id = types.StringValue(mqttResourceId["deviceId"].(string) + href)

	for k, v := range mqttContent {
		switch k {
		case "aid":
			mqttData.Aid = types.StringValue(v.(string))
		case "server":
			if !mqttData.Server.IsNull() {
				mqttData.Server = types.StringValue(v.(string))
			}
		case "port":
			if !mqttData.Port.IsNull() {
				mqttData.Port = types.Int64Value(int64(v.(float64)))
			}
		case "kai":
			if !mqttData.Kai.IsNull() {
				mqttData.Kai = types.Int64Value(int64(v.(float64)))
			}
		case "log":
			mqttData.Log = types.StringValue(v.(string))
		case "logLevel":
			if !mqttData.LogLevel.IsNull() {
				mqttData.LogLevel = types.Int64Value(int64(v.(float64)))
			}
		case "crcode":
			mqttData.CRCode = types.Int64Value(int64(v.(float64)))
		case "type":
			if !mqttData.Type.IsNull() {
				mqttData.Type = types.StringValue(v.(string))
			}
		case "subType":
			if !mqttData.SubType.IsNull() {
				mqttData.SubType = types.Int64Value(int64(v.(float64)))
			}
		case "enabled":
			if !mqttData.Enabled.IsNull() {
				mqttData.Enabled = types.BoolValue(v.(bool))
			}
		case "region":
			if !mqttData.Region.IsNull() {
				mqttData.Region = types.StringValue(v.(string))
			}
		}
	}

	tflog.Debug(ctx, "MQTTResource: Populated SUCCESS ", map[string]interface{}{"NW ID ": data["id"]})
}

func MQTTSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Numeric identifier of the MQTT.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
		},
		"device_id": schema.StringAttribute{
			Description: "device_id",
			Optional:    true,
		},
		"server_id": schema.StringAttribute{
			Description: "server_id",
			Optional:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"aid": schema.StringAttribute{
			Description: "aid",
			Computed:    true,
		},
		"log": schema.StringAttribute{
			Description: "log",
			Computed:    true,
		},
		"log_level": schema.Int64Attribute{
			Description: "log_level",
			Optional:    true,
		},
		"crcode": schema.Int64Attribute{
			Description: "crcode",
			Computed:    true,
		},
		"server": schema.StringAttribute{
			Description: "server",
			Optional:    true,
		},
		"port": schema.Int64Attribute{
			Description: "port",
			Optional:    true,
		},
		"kai": schema.Int64Attribute{
			Description: "kai",
			Optional:    true,
		},
		"type": schema.StringAttribute{
			Description: "type",
			Optional:    true,
		},
		"sub_type": schema.Int64Attribute{
			Description: "sub Type",
			Optional:    true,
		},
		"enabled": schema.BoolAttribute{
			Description: "enabled",
			Optional:    true,
		},
		"region": schema.StringAttribute{
			Description: "region",
			Optional:    true,
		},
	}
}

func MQTTServerObjectType() (types.ObjectType) {
	return types.ObjectType{	
					AttrTypes: MQTTServerAttributeType(),
				}	
}

func MQTTServerAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"device_id":   types.StringType,
		"server_id":   types.StringType,
		"href":   types.StringType,
		"id":     types.StringType,
		"aid":   types.StringType,
		"server":   types.StringType,
		"port":   types.Int64Type,
		"kai":   types.Int64Type,
		"crcode":   types.Int64Type,
		"type":   types.StringType,
		"sub_type":   types.Int64Type,
		"log":   types.StringType,
		"log_level":   types.Int64Type,
		"enabled":   types.BoolType,
		"region":   types.StringType,
	}
}

func MQTTServerAttributeValue(mqtt map[string]interface{}) map[string]attr.Value {
	data := mqtt["data"].(map[string]interface{})
	mqttResourceId := data["resourceId"].(map[string]interface{})
	mqttContent := data["content"].(map[string]interface{})

	device_id := types.StringNull()
	if mqttResourceId["deviceId"] != nil {
		device_id = types.StringValue(mqttResourceId["deviceId"].(string))
	}
	href := types.StringNull()
	server_id := types.StringNull()
	if mqttResourceId["href"] != nil {
		href = types.StringValue(mqttResourceId["href"].(string))
		server_id = types.StringValue(common.StringAfter(mqttResourceId["href"].(string), "/"))
	}
	id := types.StringValue(device_id.ValueString() + href.ValueString())
	aid := types.StringNull()
	if mqttContent["aid"] != nil {
		aid = types.StringValue(mqttContent["aid"].(string))
	}
	server := types.StringNull()
	if mqttContent["server"] != nil {
		server = types.StringValue(mqttContent["server"].(string))
	}
	port := types.Int64Null()
	if mqttContent["port"] != nil {
		port = types.Int64Value(int64(mqttContent["port"].(float64)))
	}
	kai := types.Int64Null()
	if mqttContent["kai"] != nil {
		kai = types.Int64Value(int64(mqttContent["kai"].(float64)))
	}
	crcode := types.Int64Null()
	if mqttContent["crcode"] != nil {
		crcode = types.Int64Value(int64(mqttContent["crcode"].(float64)))
	}
	mqttType := types.StringNull()
	if mqttContent["type"] != nil {
		mqttType = types.StringValue(mqttContent["type"].(string))
	}
	mqttSubType := types.Int64Null()
	if mqttContent["subType"] != nil {
		mqttSubType = types.Int64Value(int64(mqttContent["subType"].(float64)))
	}
	log := types.StringNull()
	if mqttContent["log"] != nil {
		log = types.StringValue(mqttContent["log"].(string))
	}
	log_level := types.Int64Null()
	if mqttContent["logLevel"] != nil {
		log_level = types.Int64Value(int64(mqttContent["logLevel"].(float64)))
	}
	region := types.StringNull()
	if mqttContent["region"] != nil {
		region = types.StringValue(mqttContent["region"].(string))
	}
	enabled := types.BoolNull()
	if mqttContent["enabled"] != nil {
		enabled = types.BoolValue(mqttContent["enabled"].(bool))
	}
	
	return map[string]attr.Value{
		"id" : id,
		"device_id": device_id,
		"server_id": server_id,
		"href":  href,
		"aid":   aid,
		"server":   server,
		"port":   port,
		"kai":   kai,
		"crcode":   crcode,
		"type":   mqttType,
		"sub_type":   mqttSubType,
		"log":   log,
		"log_level":   log_level,
		"enabled":   enabled,
		"region":   region,
	}
}


