package nduservice

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

type LinePTPConfig struct {
	Name          types.String `tfsdk:"name"`
	TxLaserEnable types.Bool   `tfsdk:"tx_laser_enable"`
}

type LinePTPResourceData struct {
	Id         types.String  `tfsdk:"id"`
	ParentId   types.String  `tfsdk:"parent_id"`
	Href       types.String  `tfsdk:"href"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	ColId      types.Int64   `tfsdk:"colid"`
	Config    *LinePTPConfig `tfsdk:"config"`
	State     types.Object   `tfsdk:"state"`
	Carriers  types.List     `tfsdk:"carriers"`
}

// Metadata returns the data source type name.
func (r *LinePTPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_linePtp"
}

// Schema defines the schema for the data source.
func (r *LinePTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type LinePTPResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU LinePTP",
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

	r.update(&data, ctx, &resp.Diagnostics)

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

	r.update(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LinePTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LinePTPResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LinePTPResource: Update", map[string]interface{}{"LinePTPResourceData": data})

	resp.Diagnostics.Append(diags...)

	resp.State.RemoveResource(ctx)
}

func (r *LinePTPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *LinePTPResource) create(plan *LinePTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "LinePTPResource: create ##", map[string]interface{}{"plan": plan})
}

func (r *LinePTPResource) update(plan *LinePTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "LinePTPResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && (plan.ColId.IsNull() || plan.ParentId.IsNull()) && (plan.Identifier.DeviceId.IsNull() || plan.Identifier.ParentColId.IsNull() || plan.Identifier.ColId.IsNull()) {
		diags.AddError(
			"LinePTPResource: Error update LinePTP",
			"LinePTPResource: Could not update LinePTP. Href, NDUId, PortColId, LinePTPColId, LinePTP ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})
	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.TxLaserEnable.IsNull() {
		updateRequest["txLaserEnable"] = plan.Config.TxLaserEnable.ValueBool()
	}

	tflog.Debug(ctx, "LinePTPResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"LinePTPResource: update ##: Error Create AC",
				"Create: Could not Marshal LinePTPResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Identifier.DeviceId.IsNull() && !plan.Identifier.ParentColId.IsNull() && !plan.Identifier.ColId.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString() + "/ports/" +  plan.Identifier.ParentColId.ValueString()  + "/linePtps/" +  plan.Identifier.ColId.ValueString(), rb)
		} else {
			diags.AddError(
				"VOAResource: update ##: Error update VOAResource",
				"Update: Could not update VOAResource, Identfier (DeviceID or ColId) is not specified: ",
			)
			return
		}

		tflog.Debug(ctx, "LinePTPResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"LinePTPResource: Create ##: Error Unmarshal response",
				"Update:Could not Create LinePTPResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "LinePTPResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "LinePTPResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *LinePTPResource) read(state *LinePTPResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read LinePTPResource",
			"LinePTPResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "LinePTPResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/linePtps" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/linePtps" + queryStr + "&q={\"state.edfaAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/linePtps" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/linePtps/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() +"/ports/"+ state.Identifier.ParentColId.String() + "/linePtps" + state.Identifier.ParentColId.ValueString() + "/linePtps" + queryStr
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
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "LinePTPResource: read ## ", map[string]interface{}{"plan": state})
}

func (linePtpData *LinePTPResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "LinePTPResourceData: populate ## ", map[string]interface{}{"plan": data})

	linePtpData.Id = types.StringValue(data["id"].(string))
	linePtpData.ColId = types.Int64Value(int64(data["colid"].(float64)))
	linePtpData.Href = types.StringValue(data["href"].(string))
	linePtpData.ParentId = types.StringValue(data["parentId"].(string))


	// populate config
	var config = data["config"].(map[string]interface{})
	if config != nil {
		if linePtpData.Config == nil {
			linePtpData.Config = &LinePTPConfig{}
		}
		for k, v := range config {
			switch k {
			case "name":
				if !linePtpData.Config.Name.IsNull() {
					linePtpData.Config.Name = types.StringValue(v.(string))
				}
			case "txLaserEnable":
				if !linePtpData.Config.TxLaserEnable.IsNull() {
					linePtpData.Config.TxLaserEnable = types.BoolValue(v.(bool))
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		linePtpData.State = types.ObjectValueMust(LinePTPStateAttributeType(), LinePTPStateAttributeValue(state))
	}

	var carriers = data["carriers"].([]interface{})
	if carriers != nil {
		linePtpData.Carriers = types.ListValueMust(CarrierObjectType(), CarrierObjectsValue(carriers))
	}

	tflog.Debug(ctx, "LinePTPResourceData: read ## ", map[string]interface{}{"linePtpData": linePtpData})
}

func LinePTPResourceSchemaAttributes() map[string]schema.Attribute {
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
		"config": schema.SingleNestedAttribute{
			Description: "Network Connection LC Config",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "term_lb",
					Optional:    true,
				},
				"tx_laser_enable": schema.BoolAttribute{
					Description: "tx_laser_enable",
					Optional:    true,
				},
			},
		},
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
	linePtps := []attr.Value{}
	for _, v := range data {
		linePtp := v.(map[string]interface{})
		if linePtp != nil {
			linePtps = append(linePtps, types.ObjectValueMust(
				LinePTPAttributeType(),
				LinePTPAttributeValue(linePtp)))
		}
	}
	return linePtps
}

func LinePTPAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":          types.StringType,
		"href":        types.StringType,
		"col_id":      types.Int64Type,
		"config":      types.ObjectType{AttrTypes: LinePTPConfigAttributeType()},
		"state":       types.ObjectType{AttrTypes: LinePTPStateAttributeType()},
		"carriers":    types.ListType{ElemType: CarrierObjectType()},
	}
}

func LinePTPAttributeValue(linePtp map[string]interface{}) map[string]attr.Value {
	colId := types.Int64Null()
	if linePtp["colId"] != nil {
		colId = types.Int64Value(int64(linePtp["colId"].(float64)))
	}
	href := types.StringNull()
	if linePtp["href"] != nil {
		href = types.StringValue(linePtp["href"].(string))
	}
	id := types.StringNull()
	if linePtp["id"] != nil {
		id = types.StringValue(linePtp["id"].(string))
	}
	parentId := types.StringNull()
	if linePtp["parentId"] != nil {
		parentId = types.StringValue(linePtp["parentId"].(string))
	}
	config := types.ObjectNull(LinePTPConfigAttributeType())
	if (linePtp["config"]) != nil {
		config = types.ObjectValueMust(LinePTPConfigAttributeType(), LinePTPConfigAttributeValue(linePtp["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(LinePTPStateAttributeType())
	if (linePtp["state"]) != nil {
		state = types.ObjectValueMust(LinePTPStateAttributeType(), LinePTPStateAttributeValue(linePtp["state"].(map[string]interface{})))
	}
	carriers := types.ListNull(CarrierObjectType())
	if (linePtp["carriers"]) != nil {
		carriers = types.ListValueMust(CarrierObjectType(), CarrierObjectsValue(linePtp["carriers"].([]interface{})))
	}

	return map[string]attr.Value{
		"col_id":      colId,
		"parent_id": parentId,
		"id":          id,
		"href":        href,
		"config":      config,
		"state":       state,
		"carriers":    carriers,
	}
}
func LinePTPConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":            types.StringType,
		"tx_laser_enable": types.BoolType,
	}
}

func LinePTPConfigAttributeValue(linePtpConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	txLaserEnable := types.BoolNull()

	for k, v := range linePtpConfig {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		case "txLaserEnable":
			txLaserEnable = types.BoolValue(v.(bool))
		}
	}

	return map[string]attr.Value{
		"name":            name,
		"tx_laser_enable": txLaserEnable,
	}
}

func LinePTPStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":            types.StringType,
		"parent_aid":      types.StringType,
		"line_ptp_aid":    types.StringType,
		"tx_laser_enable": types.BoolType,
		"current_role":    types.StringType,
		"lifecycle_state": types.StringType,
	}
}

func LinePTPStateAttributeValue(linePtpState map[string]interface{}) map[string]attr.Value {
	linePtpAid := types.StringNull()
	parentAid := types.StringNull()
	name := types.StringNull()
	currentRole := types.StringNull()
	txLaserEnable := types.BoolNull()
	lifecycleState := types.StringNull()

	for k, v := range linePtpState {
		switch k {
		case "linePtpAid":
			linePtpAid = types.StringValue(v.(string))
		case "parentAid":
			parentAids := v.([]interface{})
			parentAid = types.StringValue(parentAids[0].(string))
		case "name":
			name = types.StringValue(v.(string))
		case "currentRole":
			currentRole = types.StringValue(v.(string))
		case "txLaserEnable":
			txLaserEnable = types.BoolValue(v.(bool))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		}
	}

	return map[string]attr.Value{
		"line_ptp_aid":    linePtpAid,
		"parent_aid":          parentAid,
		"name":            name,
		"tx_laser_enable": txLaserEnable,
		"current_role":    currentRole,
		"lifecycle_state": lifecycleState,
	}
}
