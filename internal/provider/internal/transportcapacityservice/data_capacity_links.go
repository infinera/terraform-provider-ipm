package transportcapacity

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
	_ datasource.DataSource              = &CapacityLinksDataSource{}
	_ datasource.DataSourceWithConfigure = &CapacityLinksDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewCapacityLinksDataSource() datasource.DataSource {
	return &CapacityLinksDataSource{}
}

// coffeesDataSource is the data source implementation.
type CapacityLinksDataSource struct {
	client *ipm_pf.Client
}

type CapacityLinksDataSourceData struct {
	Id             types.String      `tfsdk:"id"`
	CapacityLinks  types.List        `tfsdk:"capacity_links"`
}

// Metadata returns the data source type name.
func (r *CapacityLinksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tc_capacity_links"
}

func (d *CapacityLinksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of transport capacities",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Capacity Link ID",
				Required: true,
			},
			"capacity_links":schema.ListAttribute{
				Computed: true,
				ElementType: TCCapacityLinkObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CapacityLinksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *CapacityLinksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := CapacityLinksDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "CapacityLinksDataSource: get CapacityLinks", map[string]interface{}{"query": query})
	var body []byte
	var err error

	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/capacity-links?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/capacity-links/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"CapacityLinksDataSource: read ##: Error read ACResource",
			"get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Debug(ctx, "CapacityLinksDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"CapacityLinksDataSource: read ##: Error Get NetworkResource",
			"Get:Could not get CapacityLinksDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "CapacityLinksDataSource: get ", map[string]interface{}{"ACs": data})
	switch data.(type) {
		case []interface{}:
			if len(data.([]interface{})) > 0 {
				query.CapacityLinks = types.ListValueMust(TCCapacityLinkObjectType(), TCCapacityLinkObjectsValue(data.([]interface{})))
		}
		case map[string]interface{}:
			capacityLink := []interface{}{data}
			query.CapacityLinks = types.ListValueMust(TCCapacityLinkObjectType(), TCCapacityLinkObjectsValue(capacityLink))
		default:
			// it's something else
	}
	tflog.Debug(ctx, "CapacityLinksDataSource: CapacityLinks ", map[string]interface{}{"CapacityLinks": query.CapacityLinks})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "CapacityLinksDataSource: get ", map[string]interface{}{"CapacityLinks": query})
}


