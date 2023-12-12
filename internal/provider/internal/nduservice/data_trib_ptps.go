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
	_ datasource.DataSource              = &TribPTPsDataSource{}
	_ datasource.DataSourceWithConfigure = &TribPTPsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewTribPTPsDataSource() datasource.DataSource {
	return &TribPTPsDataSource{}
}

// coffeesDataSource is the data source implementation.
type TribPTPsDataSource struct {
	client *ipm_pf.Client
}

type TribPTPsDataSourceData struct {
	NDUId     types.String `tfsdk:"ndu_id"`
	PortColId types.String  `tfsdk:"port_col_id"`
	ColId     types.String `tfsdk:"col_id"`
	TribPTPs  types.List   `tfsdk:"trib_ptps"`
}

// Metadata returns the data source type name.
func (r *TribPTPsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_trib_ptps"
}

func (d *TribPTPsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"trib_ptps": schema.ListAttribute{
				Computed:    true,
				ElementType: TribPTPObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TribPTPsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *TribPTPsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := TribPTPsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.NDUId.IsNull() || query.PortColId.IsNull() {
		diags.AddError(
			"Error Get TribPTPs",
			"TribPTPsDataSource: Could not get TribPTP for a ndu. ndu ID or Port Col ID is not specified",
		)
		return
	}
	tflog.Debug(ctx, "TribPTPsDataSource: get TribPTPs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.ColId.IsNull() || strings.Compare(strings.ToUpper(query.ColId.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/ports/"+query.PortColId.ValueString()+"/tribptps?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/ports/"+query.PortColId.ValueString()+"/tribptps/"+query.ColId.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"TribPTPsDataSource: read ##: Error read TribPTPResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "TribPTPsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"TribPTPsDataSource: read ##: Error Get TribPTPResource",
			"Get:Could not get TribPTPsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "TribPTPsDataSource: get ", map[string]interface{}{"TribPTPs": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.TribPTPs = types.ListValueMust(TribPTPObjectType(), TribPTPObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.TribPTPs = types.ListValueMust(TribPTPObjectType(), TribPTPObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "TribPTPsDataSource: get ", map[string]interface{}{"TribPTPs": query})
}
