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
	_ datasource.DataSource              = &EClientsDataSource{}
	_ datasource.DataSourceWithConfigure = &EClientsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewEClientsDataSource() datasource.DataSource {
	return &EClientsDataSource{}
}

// coffeesDataSource is the data source implementation.
type EClientsDataSource struct {
	client *ipm_pf.Client
}

type EClientsDataSourceData struct {
	NDUId    types.String `tfsdk:"ndu_id"`
	ColId    types.String `tfsdk:"col_id"`
	EClients types.List   `tfsdk:"eclients"`
}

// Metadata returns the data source type name.
func (r *EClientsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_eclients"
}

func (d *EClientsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of ndus' carries information",
		Attributes: map[string]schema.Attribute{
			"ndu_id": schema.StringAttribute{
				Description: "ndu ID",
				Required:    true,
			},
			"col_id": schema.StringAttribute{
				Description: "col_id",
				Required:    true,
			},
			"eclients": schema.ListAttribute{
				Computed:    true,
				ElementType: EClientObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *EClientsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *EClientsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := EClientsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if query.NDUId.IsNull() {
		diags.AddError(
			"Error Get EClients",
			"EClientsDataSource: Could not get EClient for a ndu. NDU ID  is not specified",
		)
		return
	}
	tflog.Debug(ctx, "EClientsDataSource: get EClients", map[string]interface{}{"queryNetworks": query})

	var body []byte
	var err error
	if query.ColId.IsNull() || strings.Compare(strings.ToUpper(query.ColId.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/ethernets/?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/ndus/"+query.NDUId.ValueString()+"/ethernets/"+query.ColId.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"EClientsDataSource: read ##: Error read EClientResource",
			"Get:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "EClientsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"EClientsDataSource: read ##: Error Get EClientResource",
			"Get:Could not get EClientsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "EClientsDataSource: get ", map[string]interface{}{"EClients": data})
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.EClients = types.ListValueMust(EClientObjectType(), EClientObjectsValue(data.([]interface{})))
		}
	case map[string]interface{}:
		networks := []interface{}{data}
		query.EClients = types.ListValueMust(EClientObjectType(), EClientObjectsValue(networks))
	default:
		// it's something else
	}
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "EClientsDataSource: get ", map[string]interface{}{"EClients": query})
}
