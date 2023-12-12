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
	_ datasource.DataSource              = &XRsDataSource{}
	_ datasource.DataSourceWithConfigure = &XRsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewXRsDataSource() datasource.DataSource {
	return &XRsDataSource{}
}

// coffeesDataSource is the data source implementation.
type XRsDataSource struct {
	client *ipm_pf.Client
}

type XRsDataSourceData struct {
	NDUId     types.String `tfsdk:"ndu_id"`
	PortColId types.Int64  `tfsdk:"port_col_id"`
	ColId     types.String `tfsdk:"col_id"`
	XRs       types.List   `tfsdk:"xrs"`
}

// Metadata returns the data source type name.
func (r *XRsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_xrs"
}

func (d *XRsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"xrs": schema.ListAttribute{
				Computed:    true,
				ElementType: XRObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *XRsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *XRsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := XRsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.NDUId.IsNull() || query.PortColId.IsNull() {
		diags.AddError(
			"Error Get XRs",
			"XRsDataSource: Could not get XR for a ndu. ndu ID or Port Col ID is not specified",
		)
		return
	}
	tflog.Debug(ctx, "XRsDataSource: get XRs", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.ColId.IsNull() || strings.Compare(strings.ToUpper(query.ColId.String()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/ports/"+query.PortColId.String()+"/xrs?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/ports/"+query.PortColId.String()+"/xrs/"+query.ColId.String()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"XRsDataSource: read ##: Error read XRResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "XRsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"XRsDataSource: read ##: Error Get XRResource",
			"Get:Could not get XRsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "XRsDataSource: get ", map[string]interface{}{"XRs": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.XRs = types.ListValueMust(XRObjectType(), XRObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.XRs = types.ListValueMust(XRObjectType(), XRObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "XRsDataSource: get ", map[string]interface{}{"XRs": query})
}
