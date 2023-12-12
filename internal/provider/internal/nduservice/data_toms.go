package nduservice

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
	_ datasource.DataSource              = &TOMsDataSource{}
	_ datasource.DataSourceWithConfigure = &TOMsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewTOMsDataSource() datasource.DataSource {
	return &TOMsDataSource{}
}

// coffeesDataSource is the data source implementation.
type TOMsDataSource struct {
	client *ipm_pf.Client
}

type TOMsDataSourceData struct {
	NDUId     types.String `tfsdk:"ndu_id"`
	PortColId types.String  `tfsdk:"port_col_id"`
	ColId     types.String `tfsdk:"col_id"`
	TOMs      types.List   `tfsdk:"toms"`
}

// Metadata returns the data source type name.
func (r *TOMsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_toms"
}

func (d *TOMsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of ndus' carries information",
		Attributes: map[string]schema.Attribute{
			"ndu_id": schema.StringAttribute{
				Description: "ndu ID",
				Required:    true,
			},
			"port_col_id": schema.StringAttribute{
				Description: "port_col_id",
				Required:    true,
			},
			"col_id": schema.StringAttribute{
				Description: "col_id",
				Required:    true,
			},
			"toms": schema.ListAttribute{
				Computed:    true,
				ElementType: TOMObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TOMsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *TOMsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := TOMsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.NDUId.IsNull() || query.PortColId.IsNull() {
		diags.AddError(
			"Error Get TOMs",
			"TOMsDataSource: Could not get TOM for a ndu. ndu ID or Port Col ID is not specified",
		)
		return
	}
	tflog.Debug(ctx, "TOMsDataSource: get TOMs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.ColId.IsNull() || strings.Compare(strings.ToUpper(query.ColId.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/ports/"+query.PortColId.ValueString()+"/toms?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/ports/"+query.PortColId.ValueString()+"/toms/"+query.ColId.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"TOMsDataSource: read ##: Error read TOMResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "TOMsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TOMsDataSource: read ##: Error Get TOMResource",
			"Get:Could not get TOMsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "TOMsDataSource: get ", map[string]interface{}{"TOMs": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.TOMs = types.ListValueMust(TOMObjectType(), TOMObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.TOMs = types.ListValueMust(TOMObjectType(), TOMObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "TOMsDataSource: get ", map[string]interface{}{"TOMs": query})
}
