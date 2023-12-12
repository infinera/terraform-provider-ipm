package moduleservice

import (
	"context"
	"encoding/json"
	"strings"

	"terraform-provider-ipm/internal/ipm_pf"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &ModulesDataSource{}
	_ datasource.DataSourceWithConfigure = &ModulesDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewModulesDataSource() datasource.DataSource {
	return &ModulesDataSource{}
}

// coffeesDataSource is the data source implementation.
type ModulesDataSource struct {
	client *ipm_pf.Client
}

type ModulesDataSourceData struct {
	Id      types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	SerialNumber types.String `tfsdk:"serial_number"`
	MACAddress types.String `tfsdk:"mac_address"`
	Modules types.List   `tfsdk:"modules"`
}

// Metadata returns the data source type name.
func (r *ModulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_modules"
}

func (d *ModulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Network ID",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "name",
				Optional:    true,
			},
			"serial_number": schema.StringAttribute{
				Description: "serial_number",
				Optional:    true,
			},
			"mac_address": schema.StringAttribute{
				Description: "mac_address",
				Optional:    true,
			},
			"modules": schema.ListAttribute{
				Computed:    true,
				ElementType: ModuleObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ModulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *ModulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := ModulesDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "ModulesDataSource: get Modules", map[string]interface{}{"queryNetworks": query})

	queryStr := "?content=expanded"
	if !query.Id.IsNull() {
		if strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
			queryStr = "/modules/" + queryStr
		} else {
			queryStr = "/modules/"+ query.Id.ValueString() + queryStr
		}
	} else if !query.Name.IsNull(){
		queryStr = "/modules" + queryStr + "&q={\"state.moduleName\":\"" + query.Name.ValueString() + "\"}"
	} else if !query.MACAddress.IsNull() {
		queryStr = "/modules" + queryStr + "&q={\"state.hwDescription.macAddress\":\"" + query.MACAddress.ValueString() + "\"}"
	} else if !query.SerialNumber.IsNull(){
		queryStr = "/modules" + queryStr + "&q={\"state.hwDescription.serialNumber\":\"" + query.SerialNumber.ValueString() + "\"}"
	} else {
		queryStr = "/modules/" + queryStr
	}
	body, err := d.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"ModulesDataSource: read ##: Error read ModuleResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "ModulesDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"ModulesDataSource: read ##: Error Get ModuleResource",
			"Get:Could not get ModulesDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "ModulesDataSource: get ", map[string]interface{}{"Modules": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.Modules = types.ListValueMust(ModuleObjectType(), ModuleObjectsValue(data.([]interface{})))
		} else {
			query.Modules = types.ListNull(ModuleObjectType())
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.Modules = types.ListValueMust(ModuleObjectType(), ModuleObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "ModulesDataSource: get ", map[string]interface{}{"Modules": query})
}
