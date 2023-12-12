package eventservice

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
	_ datasource.DataSource              = &FoundEventsDataSource{}
	_ datasource.DataSourceWithConfigure = &FoundEventsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewFoundEventsDataSource() datasource.DataSource {
	return &FoundEventsDataSource{}
}

// coffeesDataSource is the data source implementation.
type FoundEventsDataSource struct {
	client *ipm_pf.Client
}

type FoundEventsDataSourceData struct {
	Id        types.String  `tfsdk:"id"`
	Href        types.String  `tfsdk:"href"`
	Name        types.String  `tfsdk:"name"`
	Events  types.List    `tfsdk:"events"`
}

// Metadata returns the data source type name.
func (r *FoundEventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_found_events"
}

func (d *FoundEventsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Event ID",
				Optional:    true,
			},
			"href": schema.StringAttribute{
				Description: "href",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "name",
				Optional:    true,
			},
			"events":schema.ListAttribute{
				Computed: true,
				ElementType: EventObjectType(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *FoundEventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *FoundEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := FoundEventsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "FoundEventsDataSource: get FoundEvents", map[string]interface{}{"event id": query.Id.ValueString()})

	queryString := "?content=expanded"
	if !query.Id.IsNull() {
		if strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") != 0 {
			queryString = "/" + query.Id.ValueString() + queryString
		}
	} else if !query.Href.IsNull() {
		queryString = queryString + "&q={\"href\":\"" + query.Href.ValueString() + "\"}"
	} else {
		queryString = queryString + "&q={\"name\":\"" + query.Name.ValueString() + "\"}"
	}
	tflog.Debug(ctx, "FoundNetworksDataSource: get Event", map[string]interface{}{"queryString": "/xr-networks" + queryString})
	body, err := d.client.ExecuteIPMHttpCommand("GET", "/subscriptions/events"+queryString, nil)
	if err != nil {
		diags.AddError(
			"FoundNetworksDataSource: read ##: Error Get Network",
			"Find Network:Could not find, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "FoundNetworksDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"FoundEventsDataSource: read ##: Error Get FoundEventResource",
			"Get:Could not get FoundEventsDataSource, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	query.Events = types.ListNull(EventObjectType())
	switch data.(type) {
	case []interface{}:
		if len(data.([]interface{})) > 0 {
			query.Events = types.ListValueMust(EventObjectType(), EventObjectsValue(data.([]interface{})))
	}
	case map[string]interface{}:
		events := []interface{}{data}
		query.Events = types.ListValueMust(EventObjectType(), EventObjectsValue(events))
	default:
		// it's something else
	}
	tflog.Debug(ctx, "FoundEventsDataSource: FoundEvents ", map[string]interface{}{"FoundEvents": query.Events})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "FoundEventsDataSource: get ", map[string]interface{}{"FoundEvents": query})
}

