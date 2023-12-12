package host

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
	_ datasource.DataSource              = &HostPortsDataSource{}
	_ datasource.DataSourceWithConfigure = &HostPortsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewHostPortsDataSource() datasource.DataSource {
	return &HostPortsDataSource{}
}

// coffeesDataSource is the data source implementation.
type HostPortsDataSource struct {
	client *ipm_pf.Client
}

type HostPortsDataSourceData struct {
	HostId     types.String          `tfsdk:"host_id"`
	HostPorts  types.List            `tfsdk:"host_ports"`
}

// Metadata returns the data source type name.
func (r *HostPortsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host_ports"
}

func (d *HostPortsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Host Port information",
		Attributes: map[string]schema.Attribute{
			"host_id": schema.StringAttribute{
				Description: "Host ID",
				Optional: true,
			},
			"host_ports":schema.ListAttribute{
				Computed: true,
				ElementType: HostPortObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *HostPortsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *HostPortsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := HostPortsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.HostId.IsNull() {
		diags.AddError(
			"HostPortsDataSource: read ##: Error Read HostPorts Data",
			"Read:Could not read, Host ID is not specified",
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Debug(ctx, "HostPortsDataSource: get Host Ports", map[string]interface{}{"host Port id": query.HostId.ValueString()})

	body, err := d.client.ExecuteIPMHttpCommand("GET", "/hosts/"+query.HostId.ValueString()+"/ports?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"HostPortsDataSource: read ##: Error Read HostPortsDataSource",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "HostPortsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"HostPortsDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get HostPortsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	if len(data.([]interface{})) > 0 {
		query.HostPorts = types.ListValueMust(HostPortObjectType(), HostPortObjectsValue(data.([]interface{})))
	}
	
	tflog.Debug(ctx, "HostPortsDataSource: Hosts ", map[string]interface{}{"HostPorts": query.HostPorts})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "HostPortsDataSource: get ", map[string]interface{}{"HostPorts": query})
}

