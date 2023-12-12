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
	_ datasource.DataSource              = &EventsDataSource{}
	_ datasource.DataSourceWithConfigure = &EventsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewEventsDataSource() datasource.DataSource {
	return &EventsDataSource{}
}

// coffeesDataSource is the data source implementation.
type EventsDataSource struct {
	client *ipm_pf.Client
}

type EventsDataSourceData struct {
	Id        types.String  `tfsdk:"id"`
	Events  types.List    `tfsdk:"events"`
}

// Metadata returns the data source type name.
func (r *EventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_events"
}

func (d *EventsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Modules' carries information",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Event ID",
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
func (d *EventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ipm_pf.Client)
}

func (d *EventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	query := EventsDataSourceData{}

	diags := req.Config.Get(ctx, &query)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "EventsDataSource: get Events", map[string]interface{}{"event id": query.Id.ValueString()})

	var body []byte
	var err error
	if query.Id.IsNull() || strings.Compare(strings.ToUpper(query.Id.ValueString()), "ALL") == 0 {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/subscriptions/events?content=expanded", nil)
	} else {
		body, err = d.client.ExecuteIPMHttpCommand("GET", "/subscriptions/events/"+query.Id.ValueString()+"?content=expanded", nil)
	}
	if err != nil {
		diags.AddError(
			"EventsDataSource: read ##: Error Get Events",
			"Update:Could not read, unexpected error: "+err.Error(),
		)
		resp.Diagnostics.Append(diags...)
		return
	}
	tflog.Debug(ctx, "EventsDataSource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"EventsDataSource: read ##: Error Get EventResource",
			"Get:Could not get EventsDataSource, unexpected error: "+err.Error(),
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
	tflog.Debug(ctx, "EventsDataSource: Events ", map[string]interface{}{"Events": query.Events})
	diags = resp.State.Set(ctx, query)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "EventsDataSource: get ", map[string]interface{}{"Events": query})
}

