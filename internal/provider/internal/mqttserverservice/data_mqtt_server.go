package mqttserverservice

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &MQTTServerDataSource{}
	_ datasource.DataSourceWithConfigure = &MQTTServerDataSource{}
)

// NewCoffeesDataSource is a helper function toMQTTServermqtt_server simplify the provider implementation.
func NewMQTTServerDataSource() datasource.DataSource {
	return &MQTTServerDataSource{}
}

// coffeesDataSource is the data source implementation.
type MQTTServerDataSource struct {
	client *ipm_pf.Client
}

type MQTTServerDataSourceData struct {
	DeviceId     types.String       `tfsdk:"device_id"`
	ServerId types.String           `tfsdk:"server_id"`
	MQTTServer   types.Object       `tfsdk:"mqtt_server"`
}

// Metadata returns the data source type name.
func (r *MQTTServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mqtt_server"
}

func (d *MQTTServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the MQTT Server information",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Description: "Device ID",
				Optional:    true,
			},
			"server_id": schema.StringAttribute{
				Description: "MQTTServer server id",
				Optional:    true,
			},
			"mqtt_server":schema.ObjectAttribute{
				Computed:       true,
				AttributeTypes: MQTTServerAttributeType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *MQTTServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *MQTTServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := MQTTServerDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.DeviceId.IsNull() ||query.ServerId.IsNull() {
		diags.AddError(
			"MQTTServerDataSource: read ##: Error read MQTTServerDataSource",
			"Get: Device ID or MQTT server ID is not specified",
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Debug(ctx, "MQTTServerDataSource: get MQTTServer", map[string]interface{}{"Device ID": query.DeviceId.ValueString(), "MQTT Server ID": query.ServerId.ValueString()})
	
	body, err := d.client.ExecuteIPMHttpCommand("GET", "/devices/" + query.DeviceId.ValueString() + "/resources/mqttServers/" + query.ServerId.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"MQTTServerDataSource: read ##: Error read MQTTServerDataSource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "MQTTServerDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"MQTTServerDataSource: read ##: Error Get MQTTServerDataSource",
			"Get:Could not get MQTTServerDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "MQTTServerDataSource: get ", map[string]interface{}{"MQTTServer": data})
	query.MQTTServer = types.ObjectValueMust(MQTTServerAttributeType(), MQTTServerAttributeValue(data))
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "MQTTServerDataSource: get ", map[string]interface{}{"MQTTServer": query})
}
