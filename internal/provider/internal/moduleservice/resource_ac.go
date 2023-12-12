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
	_ resource.Resource                = &ACResource{}
	_ resource.ResourceWithConfigure   = &ACResource{}
	_ resource.ResourceWithImportState = &ACResource{}
)

// NewACResource is a helper function to simplify the provider implementation.
func NewACResource() resource.Resource {
	return &ACResource{}
}

type ACResource struct {
	client *ipm_pf.Client
}

type ACResourceData struct {
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	ColId      types.Int64               `tfsdk:"col_id"`
	State      types.Object              `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *ACResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ac"
}

// Schema defines the schema for the data source.
func (r *ACResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type ACResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an Module AC",
		Attributes:  ACResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *ACResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r ACResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ACResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "ACResource: Create - ", map[string]interface{}{"ACResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r ACResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ACResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ACResource: Create - ", map[string]interface{}{"ACResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ACResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ACResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "ACResource: Update", map[string]interface{}{"ACResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ACResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ACResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ACResource: Update", map[string]interface{}{"ACResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ACResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ACResource) read(state *ACResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && ((state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read ACResource",
			"ACResource: Could not read. AC. Href or (ModuleId, EthernetClient ColId and AC ColId) is not specified.",
		)
		return
	}

	tflog.Debug(ctx, "ACResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/ethernetClients/" + state.Identifier.ParentColId.ValueString() + "/acs" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/ethernetClients/" + state.Identifier.ParentColId.ValueString() + "/acs" + queryStr + "&q={\"state.attachmentCircuitAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/ethernetClients/" + state.Identifier.ParentColId.ValueString() + "/acs" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/ethernetClients/" + state.Identifier.ParentColId.ValueString() + "/acs/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/modules" + state.Identifier.DeviceId.ValueString() + "/ethernetClients" + state.Identifier.ParentColId.ValueString() + "/acs" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"ACResource: read ##: Error Read ACResource",
			"Read:Could not get ACResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "ACResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"ACResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal ACResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"ACResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "ACResource: read ## ", map[string]interface{}{"plan": state})
}

func (ac *ACResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "ACResourceData: populate ## ", map[string]interface{}{"plan": data})

	ac.Id = types.StringValue(data["id"].(string))
	ac.Href = types.StringValue(data["href"].(string))
	ac.ParentId = types.StringValue(data["parentId"].(string))
	ac.ColId = types.Int64Value(int64(data["colId"].(float64)))

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		ac.State = types.ObjectValueMust(ACStateAttributeType(), ACStateAttributeValue(state))
	}

	tflog.Debug(ctx, "ACResourceData: read ## ", map[string]interface{}{"plan": state})
}

func ACResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Module.",
			Computed:    true,
		},
		"module_id": schema.StringAttribute{
			Description: "module id",
			Optional:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"col_id": schema.Int64Attribute{
			Description: "col id",
			Optional:    true,
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: ACStateAttributeType(),
		},
	}
}

func ACObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: ACAttributeType(),
	}
}

func ACObjectsValue(data []interface{}) []attr.Value {
	acs := []attr.Value{}
	for _, v := range data {
		ac := v.(map[string]interface{})
		if ac != nil {
			acs = append(acs, types.ObjectValueMust(
				ACAttributeType(),
				ACAttributeValue(ac)))
		}
	}
	return acs
}

func ACAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"state":     types.ObjectType{AttrTypes: ACStateAttributeType()},
	}
}

func ACAttributeValue(ac map[string]interface{}) map[string]attr.Value {
	col_id := types.Int64Null()
	if ac["colId"] != nil {
		col_id = types.Int64Value(int64(ac["colId"].(float64)))
	}
	href := types.StringNull()
	if ac["href"] != nil {
		href = types.StringValue(ac["href"].(string))
	}
	id := types.StringNull()
	if ac["id"] != nil {
		id = types.StringValue(ac["id"].(string))
	}
	parentId := types.StringNull()
	if ac["parentId"] != nil {
		parentId = types.StringValue(ac["parentId"].(string))
	}
	state := types.ObjectNull(ACStateAttributeType())
	if (ac["state"]) != nil {
		state = types.ObjectValueMust(ACStateAttributeType(), ACStateAttributeValue(ac["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"parent_id": parentId,
		"id":        id,
		"href":      href,
		"col_id":    col_id,
		"state":     state,
	}
}

func ACStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"ac_aid":        types.StringType,
		"parent_aid":    types.StringType,
		"ac_ctrl":       types.Int64Type,
		"capacity":      types.Int64Type,
		"imc":           types.StringType,
		"imc_outer_vid": types.StringType,
		"emc":           types.StringType,
		"emc_outer_vid": types.StringType,
	}
}

func ACStateAttributeValue(acState map[string]interface{}) map[string]attr.Value {
	attachmentCircuitAid := types.StringNull()
	if acState["attachmentCircuitAid"] != nil {
		attachmentCircuitAid = types.StringValue(acState["attachmentCircuitAid"].(string))
	}
	parentAid := types.StringNull()
	if acState["parentAid"] != nil {
		parentAids := acState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	acCtrl := types.Int64Null()
	if acState["acCtrl"] != nil {
		acCtrl = types.Int64Value(int64(acState["acCtrl"].(float64)))
	}
	capacity := types.Int64Null()
	if acState["capacity"] != nil {
		capacity = types.Int64Value(int64(acState["capacity"].(float64)))
	}
	imc := types.StringNull()
	if acState["imc"] != nil {
		imc = types.StringValue(acState["imc"].(string))
	}
	imcOuterVID := types.StringNull()
	if acState["imcOuterVID"] != nil {
		imcOuterVID = types.StringValue(acState["imcOuterVID"].(string))
	}
	emc := types.StringNull()
	if acState["emc"] != nil {
		emc = types.StringValue(acState["emc"].(string))
	}
	emcOuterVID := types.StringNull()
	if acState["emcOuterVID"] != nil {
		emcOuterVID = types.StringValue(acState["emcOuterVID"].(string))
	}

	return map[string]attr.Value{
		"ac_aid":        attachmentCircuitAid,
		"parent_aid":    parentAid,
		"ac_ctrl":       acCtrl,
		"capacity":      capacity,
		"imc":           imc,
		"imc_outer_vid": imcOuterVID,
		"emc":           emc,
		"emc_outer_vid": emcOuterVID,
	}
}
