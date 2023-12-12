package eventservice

import (
	"context"
	"encoding/json"
	"strings"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &EventResource{}
	_ resource.ResourceWithConfigure   = &EventResource{}
	_ resource.ResourceWithImportState = &EventResource{}
)

// NewEventResource is a helper function to simplify the provider implementation.
func NewEventResource() resource.Resource {
	return &EventResource{}
}

type EventResource struct {
	client *ipm_pf.Client
}

type RequestedResource struct {
	ResourceType   types.String  `tfsdk:"resource_type"`
	Ids            []types.String    `tfsdk:"ids"`
	ModuleIds      []types.String    `tfsdk:"module_ids"`
	Hrefs          []types.String    `tfsdk:"hrefs"`
}

type SubscriptionFilter struct {
	RequestedNotificationTypes    []types.String `tfsdk:"requested_notification_types"`
	RequestedResources         []RequestedResource  `tfsdk:"requested_resources"`
}


type EventResourceData struct {
	Id        types.String `tfsdk:"id"`
	Href      types.String `tfsdk:"href"`
	Name      types.String `tfsdk:"name"`
	NotificationChannel    types.String   `tfsdk:"notification_channel"`
	SubscriptionFilters     []SubscriptionFilter `tfsdk:"subscription_filters"`
	ConState    types.String `tfsdk:"con_state"`
	LastConnectionTime   types.String `tfsdk:"last_connection_time"`
}

// Metadata returns the data source type name.
func (r *EventResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_Event"
}

// Schema defines the schema for the data source.
func (r *EventResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type EventResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages Event",
		Attributes:  EventResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *EventResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r EventResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EventResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "EventResource: Create - ", map[string]interface{}{"EventResourceData": data})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.create(&data, ctx, &resp.Diagnostics)
	resp.State.Set(ctx, &data)

}

func (r EventResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EventResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "EventResource: Create - ", map[string]interface{}{"EventResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r EventResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EventResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"EventResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r EventResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EventResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"EventResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *EventResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve Event ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}


func (r *EventResource) create(plan *EventResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "EventResource: create ## ", map[string]interface{}{"plan": plan})

	var createRequest = make(map[string]interface{})

	// get Network config settings
	if !plan.Name.IsNull() {
		createRequest["subscriptionName"] = plan.Name.ValueString()
	}

	subscriptionFilters := []interface{}{}
	for _, v := range plan.SubscriptionFilters {
		SubscriptionFilter := make(map[string]interface{})
		requestedNotificationTypes := []string{}
		for _, rnt := range v.RequestedNotificationTypes {
			requestedNotificationType := rnt.ValueString()
			requestedNotificationTypes =append(requestedNotificationTypes, requestedNotificationType)
		}
		SubscriptionFilter["requestedNotificationTypes"] = requestedNotificationTypes
		requestedResources := []interface{}{}
		for _, rr := range v.RequestedResources {
			requestedResource := make(map[string]interface{})
			requestedResource["resourceType"] = rr.ResourceType.ValueString()
			event_ids := []string{}
			for _, value := range rr.Ids {
				event_id :=  value.ValueString()
				event_ids =append(event_ids, event_id)
			}
			requestedResource["ids"] = event_ids
			module_ids := []string{}
			for _, value := range rr.ModuleIds {
				module_id :=  value.ValueString()
				module_ids =append(module_ids, module_id)
			}
			requestedResource["moduleIds"] = module_ids
			hrefs := []string{}
			for _, value := range rr.Hrefs {
				href :=  value.ValueString()
				hrefs =append(hrefs, href)
			}
			requestedResource["hrefs"] = hrefs
			requestedResources = append(requestedResources, requestedResource)
		}
		SubscriptionFilter["requestedResources"] = requestedResources
		subscriptionFilters = append(subscriptionFilters, SubscriptionFilter)
	}
	createRequest["subscriptionFilters"] = subscriptionFilters
	tflog.Debug(ctx, "EventResource: create ## ", map[string]interface{}{"createRequest": createRequest})

	// send create request to server
	rb, err := json.Marshal(createRequest)
	if err != nil {
		diags.AddError(
			"EventResource: create ##: Error Create EventResource",
			"Create: Could not Marshal EventResource, unexpected error: "+err.Error(),
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("POST", "/subscriptions/events", rb)
	if err != nil {
		if !strings.Contains(err.Error(), "status: 202") {
			diags.AddError(
				"EventResource: create ##: Error create EventResource",
				"Create:Could not create EventResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "EventResource: create ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data []interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"EventResource: Create ##: Error Unmarshal response",
			"Update:Could not Create EventResource, unexpected error: "+err.Error(),
		)
		return
	}

	result := data[0].(map[string]interface{})

	href := result["href"].(string)
	splits := strings.Split(href, "/")
	id := splits[len(splits)-1]
	plan.Href = types.StringValue(href)
	plan.Id = types.StringValue(id)

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "EventResource: create failed. Can't find the created network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "EventResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *EventResource) update(plan *EventResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "EventResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Id.IsNull() {
		diags.AddError(
			"EventResource: Error update Event",
			"EventResource: Could not update Event. Id is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})

	// get Network config settings
	if !plan.Name.IsNull() {
		updateRequest["subscriptionName"] = plan.Name.ValueString()
	}

	subscriptionFilters := []interface{}{}
	if len(plan.SubscriptionFilters) > 0 {
		for _, v := range plan.SubscriptionFilters {
			SubscriptionFilter := make(map[string]interface{})
			requestedNotificationTypes := []string{}
			for _, rnt := range v.RequestedNotificationTypes {
				requestedNotificationType := rnt.ValueString()
				requestedNotificationTypes =append(requestedNotificationTypes, requestedNotificationType)
			}
			SubscriptionFilter["requestedNotificationTypes"] = requestedNotificationTypes
			requestedResources := []interface{}{}
			for _, rr := range v.RequestedResources {
				requestedResource := make(map[string]interface{})
				requestedResource["resourceType"] = rr.ResourceType.ValueString()
				event_ids := []string{}
				for _, value := range rr.Ids {
					event_id :=  value.ValueString()
					event_ids =append(event_ids, event_id)
				}
				requestedResource["ids"] = event_ids
				module_ids := []string{}
				for _, value := range rr.ModuleIds {
					module_id :=  value.ValueString()
					module_ids =append(module_ids, module_id)
				}
				requestedResource["moduleIds"] = module_ids
				hrefs := []string{}
				for _, value := range rr.Hrefs {
					href :=  value.ValueString()
					hrefs =append(hrefs, href)
				}
				requestedResource["hrefs"] = hrefs
				requestedResources = append(requestedResources, requestedResource)
			}
			SubscriptionFilter["requestedResources"] = requestedResources
			subscriptionFilters = append(subscriptionFilters, SubscriptionFilter)
		}
		updateRequest["subscriptionFilters"] = subscriptionFilters
	}
	tflog.Debug(ctx, "EventResource: update ## ", map[string]interface{}{"Update Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"EventResource: update ##: Error Create AC",
				"Create: Could not Marshal EventResource, unexpected error: "+err.Error(),
			)
			return
		}
		body, err := r.client.ExecuteIPMHttpCommand("PUT", "/subscriptions/events/" + plan.Id.ValueString(), rb)
		if err != nil {
			if !strings.Contains(err.Error(), "status: 202") {
				diags.AddError(
					"EventResource: update ##: Error update EventResource",
					"Create:Could not update EventResource, unexpected error: "+err.Error(),
				)
				return
			}
		}

		tflog.Debug(ctx, "EventResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"EventResource: Create ##: Error Unmarshal response",
				"Update:Could not Create EventResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "EventResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "EventResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *EventResource) read(state *EventResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() {
		diags.AddError(
			"Error Read EventResource",
			"EventResource: Could not get Event. Event ID is not specified.",
		)
		return
	}

	tflog.Debug(ctx, "EventResource: read ## ", map[string]interface{}{"plan": state})

	body, err := r.client.ExecuteIPMHttpCommand("GET", "/subscriptions/events/" + state.Id.ValueString(), nil)
	
	if err != nil {
		diags.AddError(
			"EventResource: read ##: Error Read EventResource",
			"Read:Could not get EventResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "EventResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"EventResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal EventResource, unexpected error: "+err.Error(),
		)
		return
	}
	state.populate(data, ctx, diags)

	tflog.Debug(ctx, "EventResource: read ## ", map[string]interface{}{"plan": state})
}

func (e *EventResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "EventResourceData: populate ## ", map[string]interface{}{"plan": data})

	if data["subscriptionId"] != nil {
		e.Id = types.StringValue(data["subscriptionId"].(string))
	}
	if data["href"] != nil {
		e.Href = types.StringValue(data["href"].(string))
	}
	if data["subscriptionName"] != nil {
		e.Name = types.StringValue(data["subscriptionName"].(string))
	}
	if data["notificationChannel"] != nil {
		e.NotificationChannel = types.StringValue(data["notificationChannel"].(string))
	}
	if data["conState"] != nil {
		e.ConState = types.StringValue(data["conState"].(string))
	}
	if data["lastConnectionTime"] != nil {
		e.LastConnectionTime = types.StringValue(data["lastConnectionTime"].(string))
	}
	if data["subscriptionFilters"] != nil {
		sfs := data["subscriptionFilters"].([]interface{})
		e.SubscriptionFilters = []SubscriptionFilter{}
		for _,v := range sfs {
			subscriptionFilter := SubscriptionFilter{}
			sf := v.(map[string]interface{})
			for k1, v1 := range sf {
				if k1 == "requestedNotificationTypes" {
					requestedNotificationTypes := v1.([]interface{})
					subscriptionFilter.RequestedNotificationTypes = common.ListStringValue(requestedNotificationTypes)
				} else if k1 == "requestedResources" {
					rrs := v1.([]interface{})
					subscriptionFilter.RequestedResources = []RequestedResource{}
					for _, v2 := range rrs {
						rr := v2.(map[string]interface{})
						requestedResource := RequestedResource{}
						for k3, v3 := range rr {
							if k3 == "resourceType" {
								requestedResource.ResourceType = types.StringValue(v3.(string))
							} else if k3 == "ids" {
								requestedResource.Ids = common.ListStringValue(v3.([]interface{}))
							} else if k3 == "moduleIds" {
								requestedResource.ModuleIds = common.ListStringValue(v3.([]interface{}))
							} else if k3 == "hrefs" {
								requestedResource.Hrefs = common.ListStringValue(v3.([]interface{}))
							}
						}
						subscriptionFilter.RequestedResources = append (subscriptionFilter.RequestedResources,requestedResource )
					}
				}
			}
			e.SubscriptionFilters = append(e.SubscriptionFilters, subscriptionFilter)
		}
	}
	tflog.Debug(ctx, "EventResourceData: read ## ", map[string]interface{}{"e": e})
}

func EventResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Event.",
			Optional:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "name",
			Computed:    true,
		},
		"notification_channel": schema.StringAttribute{
			Description: "notification_channel",
			Computed:    true,
		},
		"con_state": schema.StringAttribute{
			Description: "con_state",
			Computed:    true,
		},
		"last_connection_time": schema.StringAttribute{
			Description: "last_connection_time",
			Computed:    true,
		},
		//subscription_filters_config    ]SubscriptionFilter `tfsdk:"subscription_filters_config"`
		"subscription_filters": schema.ListNestedAttribute{
			Description: "List of SubscriptionFilter Config",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: SubscriptionFilterSchemaAttributes(),
			},
		},
	}
}

func SubscriptionFilterSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
						"requested_notification_types": schema.ListAttribute{
							Description: "requested notification types",
							Optional:    true,
							ElementType: types.StringType,
						},
						"requested_resources": schema.ListNestedAttribute{
							Description: "requested_resources",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"resource_type": schema.StringAttribute{
										Description: "resource_type",
										Optional:    true,
									},
									"ids": schema.ListAttribute{
										Description: "ids",
										Optional:    true,
										ElementType: types.StringType,
									},
									"module_ids": schema.ListAttribute{
										Description: "module_ids",
										Optional:    true,
										ElementType: types.StringType,
									},
									"hrefs": schema.ListAttribute{
										Description: "hrefs",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					}
}

func EventObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: EventAttributeType(),
	}
}

func EventObjectsValue(data []interface{}) []attr.Value {
	events := []attr.Value{}
	for _, v := range data {
		event := v.(map[string]interface{})
		if event != nil {
			events = append(events, types.ObjectValueMust(
				EventAttributeType(),
				EventAttributeValue(event)))
		}
	}
	return events
}

func EventAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.StringType,
		"href":      types.StringType,
		"notification_channel":    types.StringType,
		"con_state":    types.StringType,
		"last_connection_time":    types.StringType,
		"subscription_filters":     types.ListType{ElemType: SubscriptionFilterObjectType()},
	}
}

func EventAttributeValue(data map[string]interface{}) map[string]attr.Value {
	href := types.StringNull()
	id := types.StringNull()
	notification_channel := types.StringNull()
	con_state := types.StringNull()
	last_connection_time := types.StringNull()
	subscription_filters := types.ListNull(SubscriptionFilterObjectType())

	for k, v := range data {
		switch k {
		case "href":
			href = types.StringValue(v.(string))
		case "id":
			id = types.StringValue(v.(string))
		case "conState":
			con_state = types.StringValue(v.(string))
		case "notificationChannel":
			notification_channel = types.StringValue(v.(string))
		case "lastConnectionTime":
			last_connection_time = types.StringValue(v.(string))
		case "SubscriptionFilters":
			subscription_filters = types.ListValueMust(SubscriptionFilterObjectType(), SubscriptionFilterObjectsValue(v.([]interface{})))
		}
	}

	return map[string]attr.Value{
		"id":        id,
		"href":      href,
		"con_state":    con_state,
		"notification_channel":     notification_channel,
		"last_connection_time":  last_connection_time,
		"subscription_filters":      subscription_filters,
	}
}

func SubscriptionFilterObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: SubscriptionFilterAttributeType(),
	}
}

func SubscriptionFilterObjectsValue(data []interface{}) []attr.Value {
	subscriptionFilters := []attr.Value{}
	for _, v := range data {
		SubscriptionFilter := v.(map[string]interface{})
		if SubscriptionFilter != nil {
			subscriptionFilters = append(subscriptionFilters, types.ObjectValueMust(
				SubscriptionFilterAttributeType(),
				SubscriptionFilterAttributeValue(SubscriptionFilter)))
		}
	}
	return subscriptionFilters
}

func SubscriptionFilterAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"requested_notification_types":   types.ListType{ElemType: types.StringType},
		"requested_resources":            types.ListType{ElemType: RequestedResourceObjectType()},
	}
}

func SubscriptionFilterAttributeValue(data map[string]interface{}) map[string]attr.Value {
	requested_notification_types := types.ListNull(types.StringType)
	requested_resources := types.ListNull(RequestedResourceObjectType())

	for k, v := range data {
		switch k {
		case "requestedNotificationTypes":
			requested_notification_types = types.ListValueMust(types.StringType, common.ListAttributeStringValue(v.([]interface{})))
		case "requestedResources":
			requested_resources = types.ListValueMust(RequestedResourceObjectType(), RequestedResourceObjectsValue(v.([]interface{})))
		}
	}
	return map[string]attr.Value{
		"requested_notification_types": requested_notification_types,
		"requested_resources":          requested_resources,
	}
}

func RequestedResourceObjectsValue(data []interface{}) []attr.Value {
	requestedResources := []attr.Value{}
	for _, v := range data {
		requestedResource := v.(map[string]interface{})
		if requestedResource != nil {
			requestedResources = append(requestedResources, types.ObjectValueMust(
				RequestedResourceAttributeType(),
				RequestedResourceAttributeValue(requestedResource)))
		}
	}
	return requestedResources
}

func RequestedResourceObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: RequestedResourceAttributeType(),
	}
}

func RequestedResourceAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"resource_type":   types.StringType,
		"ids":   types.ListType{ElemType: types.StringType},
		"module_ids":   types.ListType{ElemType: types.StringType},
		"hrefs":   types.ListType{ElemType: types.StringType},
	}
}

func RequestedResourceAttributeValue(RequestedResource map[string]interface{}) map[string]attr.Value {
	resource_type := types.StringNull()
	ids := types.ListNull(types.StringType)
	module_ids := types.ListNull(types.StringType)
	hrefs := types.ListNull(types.StringType)

	for k, v := range RequestedResource {
		switch k {
			case "resourceType":
				resource_type = types.StringValue(v.(string))
			case "ids":
				ids = types.ListValueMust(types.StringType, common.ListAttributeStringValue(v.([]interface{})))
			case "module_ids":
				module_ids = types.ListValueMust(types.StringType, common.ListAttributeStringValue(v.([]interface{})))
			case "hrefs":
				hrefs = types.ListValueMust(types.StringType, common.ListAttributeStringValue(v.([]interface{})))
		}
	}
	return map[string]attr.Value{
		"resource_type": resource_type,
		"ids":          ids,
		"module_ids": module_ids,
		"hrefs":          hrefs,
	}
}