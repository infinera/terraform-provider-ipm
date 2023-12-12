package moduleservice

import (
	"context"
	"encoding/json"

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
	_ resource.Resource                = &LinePTPResource{}
	_ resource.ResourceWithConfigure   = &LinePTPResource{}
	_ resource.ResourceWithImportState = &LinePTPResource{}
)

// NewLinePTPResource is a helper function to simplify the provider implementation.
func NewLinePTPResource() resource.Resource {
	return &LinePTPResource{}
}

type LinePTPResource struct {
	client *ipm_pf.Client
}

type LinePTPResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	ColId      types.Int64               `tfsdk:"colid"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	State      types.Object              `tfsdk:"state"`
	Carriers   types.List                `tfsdk:"carriers"`
}

// Metadata returns the data source type name.
func (r *LinePTPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_line_ptp"
}

// Schema defines the schema for the data source.
func (r *LinePTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type LinePTPResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an ModuleLinePTP",
		Attributes:  LinePTPResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *LinePTPResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r LinePTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LinePTPResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "LinePTPResource: Create - ", map[string]interface{}{"LinePTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r LinePTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LinePTPResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LinePTPResource: Create - ", map[string]interface{}{"LinePTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LinePTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LinePTPResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "LinePTPResource: Update", map[string]interface{}{"LinePTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LinePTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LinePTPResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LinePTPResource: Update", map[string]interface{}{"LinePTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	r.delete(&data, ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *LinePTPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *LinePTPResource) read(state *LinePTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && state.Identifier.ColId.IsNull() && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull() {
		diags.AddError(
			"Error Read LinePTPResource",
			"LinePTPResource: Could not read. Id, and Href, and identifier are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "LinePTPResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps" + queryStr + "&q={\"state.linePtpAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/modules" + state.Identifier.DeviceId.ValueString() + "/linePtps" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"LinePTPResource: read ##: Error Read LinePTPResource",
			"Read:Could not get LinePTPResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "LinePTPResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"LinePTPResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal LinePTPResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"LinePTPResource: read ##: Can not get Module",
				"Read:Could not get EClient for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "LinePTPResource: read ## ", map[string]interface{}{"plan": state})
}

func (r *LinePTPResource) delete(plan *LinePTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "LinePTPResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (lineptpData *LinePTPResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "LinePTPResourceData: populate ## ", map[string]interface{}{"plan": data})

	if !lineptpData.Id.IsNull() {
		lineptpData.Id = types.StringValue(data["id"].(string))
	}

	lineptpData.Id = types.StringValue(data["id"].(string))
	lineptpData.ParentId = types.StringValue(data["parentId"].(string))
	lineptpData.Href = types.StringValue(data["href"].(string))
	lineptpData.ColId = types.Int64Value(int64(data["colid"].(float64)))

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		lineptpData.State = types.ObjectValueMust(LinePTPStateAttributeType(), LinePTPStateAttributeValue(state))
	}
	// populate Carriers
	lineptpData.Carriers = types.ListNull(CarrierObjectType())
	if data["carriers"] != nil {
		lineptpData.Carriers = types.ListValueMust(CarrierObjectType(), CarrierObjectsValue(data["carriers"].([]interface{})))
	}

	tflog.Debug(ctx, "LinePTPResourceData: read ## ", map[string]interface{}{"plan": state})
}

func LinePTPResourceSchemaAttributes(computeEntity_optional ...bool) map[string]schema.Attribute {
	computeFlag := false
	optionalFlag := true
	if len(computeEntity_optional) > 0 {
		computeFlag = computeEntity_optional[0]
		optionalFlag = !computeFlag
	}
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Module.",
			Computed:    computeFlag,
			Optional:    optionalFlag,
		},
		"parent_id": schema.StringAttribute{
			Description: "parent id",
			Computed:    computeFlag,
			Optional:    optionalFlag,
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
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: LinePTPStateAttributeType(),
		},
		"carriers": schema.ListAttribute{
			Computed:    true,
			ElementType: CarrierObjectType(),
		},
	}
}

func LinePTPObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: LinePTPAttributeType(),
	}
}

func LinePTPObjectsValue(data []interface{}) []attr.Value {
	lineptps := []attr.Value{}
	for _, v := range data {
		lineptp := v.(map[string]interface{})
		if lineptp != nil {
			lineptps = append(lineptps, types.ObjectValueMust(
				LinePTPAttributeType(),
				LinePTPAttributeValue(lineptp)))
		}
	}
	return lineptps
}

func LinePTPAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"state":     types.ObjectType{AttrTypes: LinePTPStateAttributeType()},
		"carriers":  types.ListType{ElemType: CarrierObjectType()},
	}
}

func LinePTPAttributeValue(lineptp map[string]interface{}) map[string]attr.Value {
	col_id := types.Int64Null()
	if lineptp["colId"] != nil {
		col_id = types.Int64Value(int64(lineptp["colId"].(float64)))
	}
	href := types.StringNull()
	if lineptp["href"] != nil {
		href = types.StringValue(lineptp["href"].(string))
	}
	parentId := types.StringNull()
	if lineptp["parentId"] != nil {
		parentId = types.StringValue(lineptp["parentId"].(string))
	}
	id := types.StringNull()
	if lineptp["id"] != nil {
		id = types.StringValue(lineptp["id"].(string))
	}
	state := types.ObjectNull(LinePTPStateAttributeType())
	if (lineptp["state"]) != nil {
		state = types.ObjectValueMust(LinePTPStateAttributeType(), LinePTPStateAttributeValue(lineptp["state"].(map[string]interface{})))
	}
	carriers := types.ListNull(CarrierObjectType())
	if (lineptp["carriers"]) != nil {
		carriers = types.ListValueMust(CarrierObjectType(), CarrierObjectsValue(lineptp["carriers"].([]interface{})))
	}

	return map[string]attr.Value{
		"col_id":    col_id,
		"parent_id": parentId,
		"id":        id,
		"href":      href,
		"state":     state,
		"carriers":  carriers,
	}
}

func LinePTPStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"line_ptp_aid":         types.StringType,
		"parent_aid":           types.StringType,
		"cp_counter":           types.Int64Type,
		"discovered_neighbors": types.ListType{ElemType: LinePTPDiscoveredNeighborObjectType()},
		"ctrl_plane_neighbors": types.ListType{ElemType: LinePTPCtrlPlaneNeighborObjectType()},
	}
}

func LinePTPStateAttributeValue(lldp map[string]interface{}) map[string]attr.Value {
	linePtpAid := types.StringNull()
	if lldp["linePtpAid"] != nil {
		linePtpAid = types.StringValue(lldp["linePtpAid"].(string))
	}
	parentAid := types.StringNull()
	if lldp["parentAid"] != nil {
		parentAids := lldp["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	cpCounter := types.Int64Null()
	if lldp["cpCounter"] != nil {
		cpCounter = types.Int64Value(int64(lldp["cpCounter"].(float64)))
	}
	discoveredNeighbors := types.ListNull(LinePTPDiscoveredNeighborObjectType())
	if (lldp["discoveredNeighbors"]) != nil {
		discoveredNeighbors = types.ListValueMust(LinePTPDiscoveredNeighborObjectType(), LinePTPDiscoveredNeighborObjectsValue(lldp["discoveredNeighbors"].([]interface{})))
	}
	ctrlPlaneNeighbors := types.ListNull(LinePTPCtrlPlaneNeighborObjectType())
	if (lldp["ctrlPlaneNeighbors"]) != nil {
		ctrlPlaneNeighbors = types.ListValueMust(LinePTPCtrlPlaneNeighborObjectType(), LinePTPCtrlPlaneNeighborObjectsValue(lldp["ctrlPlaneNeighbors"].([]interface{})))
	}

	return map[string]attr.Value{
		"line_ptp_aid":         linePtpAid,
		"parent_aid":           parentAid,
		"cp_counter":           cpCounter,
		"discovered_neighbors": discoveredNeighbors,
		"ctrl_plane_neighbors": ctrlPlaneNeighbors,
	}
}

func LinePTPCtrlPlaneNeighborObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: LinePTPCtrlPlaneNeighborAttributeType(),
	}
}

func LinePTPCtrlPlaneNeighborObjectsValue(data []interface{}) []attr.Value {
	lineptps := []attr.Value{}
	for _, v := range data {
		lineptp := v.(map[string]interface{})
		if lineptp != nil {
			lineptps = append(lineptps, types.ObjectValueMust(
				LinePTPCtrlPlaneNeighborAttributeType(),
				LinePTPCtrlPlaneNeighborAttributeValue(lineptp)))
		}
	}
	return lineptps
}

func LinePTPCtrlPlaneNeighborAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"remote_cp_counter":       types.Int64Type,
		"mac_address":             types.StringType,
		"current_role":            types.StringType,
		"constellation_frequency": types.Int64Type,
		"con_state":               types.StringType,
		"last_con_state_change":   types.StringType,
	}
}

func LinePTPCtrlPlaneNeighborAttributeValue(neighbors map[string]interface{}) map[string]attr.Value {
	remoteCpCounter := types.Int64Null()
	macAddress := types.StringNull()
	currentRole := types.StringNull()
	constellationFrequency := types.Int64Null()
	conState := types.StringNull()
	lastConStateChange := types.StringNull()

	for k, v := range neighbors {
		switch k {
		case "remoteCpCounter":
			remoteCpCounter = types.Int64Value(int64(v.(float64)))
		case "macAddress":
			macAddress = types.StringValue(v.(string))
		case "currentRole":
			currentRole = types.StringValue(v.(string))
		case "constellationFrequency":
			constellationFrequency = types.Int64Value(int64(v.(float64)))
		case "conState":
			conState = types.StringValue(v.(string))
		case "lastConStateChange":
			lastConStateChange = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"remote_cp_counter":       remoteCpCounter,
		"mac_address":             macAddress,
		"current_role":            currentRole,
		"constellation_frequency": constellationFrequency,
		"con_state":               conState,
		"last_con_state_change":   lastConStateChange,
	}
}

func LinePTPDiscoveredNeighborObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: LinePTPDiscoveredNeighborAttributeType(),
	}
}

func LinePTPDiscoveredNeighborObjectsValue(data []interface{}) []attr.Value {
	lineptps := []attr.Value{}
	for _, v := range data {
		lineptp := v.(map[string]interface{})
		if lineptp != nil {
			lineptps = append(lineptps, types.ObjectValueMust(
				LinePTPDiscoveredNeighborAttributeType(),
				LinePTPDiscoveredNeighborAttributeValue(lineptp)))
		}
	}
	return lineptps
}

func LinePTPDiscoveredNeighborAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"remote_cp_counter":       types.Int64Type,
		"mac_address":             types.StringType,
		"current_role":            types.StringType,
		"constellation_frequency": types.Int64Type,
	}
}

func LinePTPDiscoveredNeighborAttributeValue(neighbors map[string]interface{}) map[string]attr.Value {
	remoteCpCounter := types.Int64Null()
	macAddress := types.StringNull()
	currentRole := types.StringNull()
	constellationFrequency := types.Int64Null()

	for k, v := range neighbors {
		switch k {
		case "remoteCpCounter":
			remoteCpCounter = types.Int64Value(int64(v.(float64)))
		case "macAddress":
			macAddress = types.StringValue(v.(string))
		case "currentRole":
			currentRole = types.StringValue(v.(string))
		case "constellationFrequency":
			constellationFrequency = types.Int64Value(int64(v.(float64)))
		}
	}

	return map[string]attr.Value{
		"remote_cp_counter":       remoteCpCounter,
		"mac_address":             macAddress,
		"current_role":            currentRole,
		"constellation_frequency": constellationFrequency,
	}
}
