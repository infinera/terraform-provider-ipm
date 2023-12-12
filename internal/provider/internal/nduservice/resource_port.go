package nduservice

import (
	"context"
	"encoding/json"
	"strconv"

	"terraform-provider-ipm/internal/ipm_pf"
	"terraform-provider-ipm/internal/provider/internal/common"

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
	_ resource.Resource                = &PortResource{}
	_ resource.ResourceWithConfigure   = &PortResource{}
	_ resource.ResourceWithImportState = &PortResource{}
)

// NewPortResource is a helper function to simplify the provider implementation.
func NewPortResource() resource.Resource {
	return &PortResource{}
}

type PortResource struct {
	client *ipm_pf.Client
}

type PortConfig struct {
	Name        types.String `tfsdk:"name"`
	ConnectedTo types.String `tfsdk:"connected_to"`
}

type PortResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String   `tfsdk:"parent_id"`
	Href       types.String   `tfsdk:"href"`
	ColId      types.Int64    `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Config   *PortConfig  `tfsdk:"config"`
	State    types.Object `tfsdk:"state"`
	TOMs     types.List   `tfsdk:"toms"`
	XRs      types.List   `tfsdk:"xrs"`
	EDFAs    types.List   `tfsdk:"edfas"`
	VOAs     types.List   `tfsdk:"voas"`
	LinePTPs types.List   `tfsdk:"line_ptps"`
	TribPTPs types.List   `tfsdk:"trib_ptps"`
	PolPTPs  types.List   `tfsdk:"pol_ptps"`
}

// Metadata returns the data source type name.
func (r *PortResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_port"
}

// Schema defines the schema for the data source.
func (r *PortResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type PortResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an Port port",
		Attributes:  PortResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *PortResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r PortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PortResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "PortResource: Create - ", map[string]interface{}{"PortResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r PortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PortResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "PortResource: Create - ", map[string]interface{}{"PortResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r PortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PortResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "PortResource: Update", map[string]interface{}{"PortResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r PortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PortResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "PortResource: Update", map[string]interface{}{"PortResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *PortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PortResource) update(plan *PortResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "PortResource: update ## ", map[string]interface{}{"plan": plan})
	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"PortResource: Error update Port",
			"PortResource: Could not update Port. Href or Port ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})
	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.ConnectedTo.IsNull() {
		updateRequest["connectedTo"] = plan.Config.ConnectedTo.ValueString()
	}

	tflog.Debug(ctx, "PortResource: update ## ", map[string]interface{}{"update Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"PortResource: update ##: Error update Port",
				"update: Could not Marshal PortResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.ColId.IsNull() && !plan.ParentId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.ParentId.ValueString() + "/ports/" +  strconv.FormatInt(plan.ColId.ValueInt64(),10), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"PortResource: update ##: Error update porr}",
				"Update: Could not update PortResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "PortResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"PortResource: update ##: Error Unmarshal response",
				"Update:Could not update PortResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "PortResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "PortResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *PortResource) read(state *PortResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() &&(state.Identifier.ColId.IsNull() ) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read PortResource",
			"PortResource: Could not read. Port Href or (NDUId and Port ColId) is not specified.",
		)
		return
	}

	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/ports" + queryStr + "&q={\"state.edfaAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/ports" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/ports/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() + "/ports" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"PortResource: read ##: Error Read PortResource",
			"Read:Could not get PortResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "PortResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"PortResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal PortResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"PortResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "PortResource: read ## ", map[string]interface{}{"plan": state})
}

func (port *PortResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "PortResourceData: populate ## ", map[string]interface{}{"plan": data})

	port.Id = types.StringValue(data["id"].(string))
	port.ColId = types.Int64Value(int64(data["colid"].(float64)))
	port.Href = types.StringValue(data["href"].(string))
	port.ParentId = types.StringValue(data["parentId"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if port.Config == nil {
		port.Config = &PortConfig{}
	}
	if config != nil {
		if port.Config == nil {
			port.Config = &PortConfig{}
		}
		for k, v := range config {
			switch k {
			case "name":
				if !port.Config.Name.IsNull() {
					port.Config.Name = types.StringValue(v.(string))
				}
			case "connectedTo":
				if !port.Config.ConnectedTo.IsNull() {
					port.Config.ConnectedTo = types.StringValue(v.(string))
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		port.State = types.ObjectValueMust(PortStateAttributeType(), PortStateAttributeValue(state))
	}

	var toms = data["toms"].([]interface{})
	if toms != nil && len(toms) > 0 {
		port.TOMs = types.ListValueMust(TOMObjectType(), TOMObjectsValue(toms))
	}

	var xrs = data["xrs"].([]interface{})
	if xrs != nil {
		port.XRs = types.ListValueMust(XRObjectType(), XRObjectsValue(xrs))
	}

	var edfas = data["edfas"].([]interface{})
	if edfas != nil {
		port.EDFAs = types.ListValueMust(EDFAObjectType(), EDFAObjectsValue(edfas))
	}

	var voas = data["voas"].([]interface{})
	if voas != nil {
		port.VOAs = types.ListValueMust(VOAObjectType(), VOAObjectsValue(voas))
	}
	var linePtps = data["linePtps"].([]interface{})
	if linePtps != nil {
		port.LinePTPs = types.ListValueMust(LinePTPObjectType(), LinePTPObjectsValue(linePtps))
	}

	var polPtps = data["polPtps"].([]interface{})
	if polPtps != nil {
		port.PolPTPs = types.ListValueMust(PolPTPObjectType(), PolPTPObjectsValue(polPtps))
	}

	tflog.Debug(ctx, "PortResourceData: read ## ", map[string]interface{}{"port": port})
}

func PortResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Carrier.",
			Computed:    true,
		},
		"parent_id": schema.StringAttribute{
			Description: "parent id",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"col_id": schema.Int64Attribute{
			Description: "col id",
			Computed:    true,
		},
		"identifier": common.ResourceIdentifierAttribute(),
		//Config    NodeConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "name",
					Optional:    true,
				},
				"connected_to": schema.StringAttribute{
					Description: "connected_to",
					Optional:    true,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: PortStateAttributeType(),
		},
		"toms": schema.ListAttribute{
			Computed:    true,
			ElementType: TOMObjectType(),
		},
		"xrs": schema.ListAttribute{
			Computed:    true,
			ElementType: XRObjectType(),
		},
		"edfas": schema.ListAttribute{
			Computed:    true,
			ElementType: EDFAObjectType(),
		},
		"voas": schema.ListAttribute{
			Computed:    true,
			ElementType: VOAObjectType(),
		},
		"line_ptps": schema.ListAttribute{
			Computed:    true,
			ElementType: LinePTPObjectType(),
		},
		"trib_ptps": schema.ListAttribute{
			Computed:    true,
			ElementType: TribPTPObjectType(),
		},
		"pol_ptps": schema.ListAttribute{
			Computed:    true,
			ElementType: PolPTPObjectType(),
		},
	}
}

func PortObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: PortAttributeType(),
	}
}

func PortObjectsValue(data []interface{}) []attr.Value {
	ports := []attr.Value{}
	for _, v := range data {
		port := v.(map[string]interface{})
		if port != nil {
			ports = append(ports, types.ObjectValueMust(
				PortAttributeType(),
				PortAttributeValue(port)))
		}
	}
	return ports
}

func PortAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"config":    types.ObjectType{AttrTypes: PortConfigAttributeType()},
		"state":     types.ObjectType{AttrTypes: PortStateAttributeType()},
		"toms":      types.ListType{ElemType: TOMObjectType()},
		"xrs":       types.ListType{ElemType: XRObjectType()},
		"edfas":     types.ListType{ElemType: EDFAObjectType()},
		"voas":      types.ListType{ElemType: VOAObjectType()},
		"line_ptps": types.ListType{ElemType: LinePTPObjectType()},
		"trib_ptps": types.ListType{ElemType: TribPTPObjectType()},
		"pol_ptps":  types.ListType{ElemType: PolPTPObjectType()},
	}
}

func PortAttributeValue(port map[string]interface{}) map[string]attr.Value {
	href := types.StringNull()
	parentId := types.StringNull()
	id := types.StringNull()
	config := types.ObjectNull(PortConfigAttributeType())
	state := types.ObjectNull(PortStateAttributeType())
	toms := types.ListNull(TOMObjectType())
	xrs := types.ListNull(XRObjectType())
	edfas := types.ListNull(EDFAObjectType())
	voas := types.ListNull(VOAObjectType())
	linePtps := types.ListNull(LinePTPObjectType())
	tribPtps := types.ListNull(TribPTPObjectType())
	polPtps := types.ListNull(PolPTPObjectType())

	for k, v := range port {
		switch k {
		case "parentId":
			parentId = types.StringValue(v.(string))
		case "href":
			href = types.StringValue(v.(string))
		case "id":
			id = types.StringValue(v.(string))
		case "config":
			config = types.ObjectValueMust(PortConfigAttributeType(), PortConfigAttributeValue(v.(map[string]interface{})))
		case "state":
			state = types.ObjectValueMust(PortStateAttributeType(), PortStateAttributeValue(v.(map[string]interface{})))
		case "toms":
			toms = types.ListValueMust(TOMObjectType(), TOMObjectsValue(v.([]interface{})))
		case "xrs":
			xrs = types.ListValueMust(XRObjectType(), XRObjectsValue(v.([]interface{})))
		case "edfas":
			edfas = types.ListValueMust(EDFAObjectType(), EDFAObjectsValue(v.([]interface{})))
		case "voas":
			voas = types.ListValueMust(VOAObjectType(), VOAObjectsValue(v.([]interface{})))
		case "linePtps":
			linePtps = types.ListValueMust(LinePTPObjectType(), LinePTPObjectsValue(v.([]interface{})))
		case "tribPtps":
			tribPtps = types.ListValueMust(TribPTPObjectType(), TribPTPObjectsValue(v.([]interface{})))
		case "polPtps":
			polPtps = types.ListValueMust(PolPTPObjectType(), PolPTPObjectsValue(v.([]interface{})))
		}
	}

	return map[string]attr.Value{
		"parent_id": parentId,
		"id":        id,
		"href":      href,
		"config":    config,
		"state":     state,
		"toms":      toms,
		"xrs":       xrs,
		"edfas":     edfas,
		"voas":      voas,
		"line_ptps": linePtps,
		"trib_ptps": tribPtps,
		"pol_ptps":  polPtps,
	}
}

func PortConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"label":        types.StringType,
		"connected_to": types.StringType,
	}
}

func PortConfigAttributeValue(portConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	connectedTo := types.StringNull()

	for k, v := range portConfig {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		case "connectedTo":
			connectedTo = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"name":         name,
		"connected_to": connectedTo,
	}
}

func PortStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":         types.StringType,
		"port_aid":        types.StringType,
		"name":            types.StringType,
		"port_type":       types.StringType,
		"support_types":   types.ListType{ElemType: types.StringType},
		"connected_to":    types.StringType,
		"lifecycle_state": types.StringType,
	}
}

func PortStateAttributeValue(portState map[string]interface{}) map[string]attr.Value {
	parentAid := types.StringNull()
	portAid := types.StringNull()
	name := types.StringNull()
	portType := types.StringNull()
	connectedTo := types.StringNull()
	lifecycleState := types.StringNull()
	supportTypes := types.ListNull(types.StringType)

	for k, v := range portState {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		case "portType":
			portType = types.StringValue(v.(string))
		case "portAid":
			portAid = types.StringValue(v.(string))
		case "connectedTo":
			connectedTo = types.StringValue(v.(string))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		case "supportTypes":
			supportTypes = types.ListValueMust(types.StringType, common.ListAttributeStringValue(v.([]interface{})))
		}
	}

	return map[string]attr.Value{
		"parent_aid":          parentAid,
		"port_aid":        portAid,
		"name":            name,
		"port_type":        portType,
		"connected_to":    connectedTo,
		"lifecycle_state": lifecycleState,
		"support_types":   supportTypes,
	}
}
