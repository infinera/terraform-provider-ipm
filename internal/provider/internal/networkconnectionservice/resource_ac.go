package networkconnection

import (
	"context"
	"encoding/json"

	"terraform-provider-ipm/internal/ipm_pf"

	//	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	//"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
	Id       types.String `tfsdk:"id"`
	ParentId types.String `tfsdk:"parent_id"`
	Href     types.String `tfsdk:"href"`
	Config   types.Object `tfsdk:"config"`
	State    types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *ACResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nc_ac"
}

// Schema defines the schema for the data source.
func (r *ACResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type ACResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an AC",
		Attributes:  map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "AC ID",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"href": schema.StringAttribute{
				Description: "href of the network module",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			//Config     types.Object   `tfsdk:"config"`
			"config": schema.ObjectAttribute{
				Computed: true,
				AttributeTypes: ACConfigAttributeType(),
			},
			//State     types.Object   `tfsdk:"state"`
			"state": schema.ObjectAttribute{
				Computed: true,
				AttributeTypes: ACStateAttributeType(),
			},
		},
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
	tflog.Debug(ctx, "NetworkModuleResource: Create - ", map[string]interface{}{"NetworkModuleResourceData": data})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

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
	tflog.Debug(ctx, "CfgResource: Update", map[string]interface{}{"ACResourceData": data})

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

	tflog.Debug(ctx, "CfgResource: Delete", map[string]interface{}{"ACResourceData": data})

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

	if state.Id.IsNull() {
		diags.AddError(
			"Error Update Module",
			"Update: Could not Update Module. Id and Href are not specified.",
		)
		return
	}
	body, err := r.client.ExecuteIPMHttpCommand("GET", "/acs/"+state.Id.ValueString()+"?content=expanded", nil)
	if err != nil {
		diags.AddError(
			"ACResource: read ##: Error Read ACResource",
			"Update:Could not read ACResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "ACResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var data = make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		diags.AddError(
			"ACResource: read ##: Error Unmarshal response",
			"Update:Could not read ACResource, unexpected error: "+err.Error(),
		)
		return
	}

	var content = data["content"].(map[string]interface{})

	state.Populate(content, ctx, diags)

	tflog.Debug(ctx, "ACResource: read ## ", map[string]interface{}{"plan": state})

}

func (acData *ACResourceData) Populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics, computeOnly ...bool) {

	computeFlag := false
	if len(computeOnly) > 0 {
		computeFlag = computeOnly[0]
	}

	tflog.Debug(ctx, "ACResourceData: populate ## ", map[string]interface{}{"plan": data})
	if computeFlag {
		acData.Id = types.StringValue(data["id"].(string))
	}

	acData.Href = types.StringValue(data["href"].(string))
	//populate state
	if data["state"] != nil {
		acData.State = types.ObjectValueMust(
			ACStateAttributeType(), ACStateAttributeValue(data["state"].(map[string]interface{})))
	}
		//populate config
	if data["config"] != nil {
		acData.Config = types.ObjectValueMust(
			ACConfigAttributeType(), ACConfigAttributeValue(data["config"].(map[string]interface{})))
	}

	tflog.Debug(ctx, "ACResourceData: populate ## ", map[string]interface{}{"acDate": acData})
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
		"id":     types.StringType,
		"href":   types.StringType,
		"config": types.ObjectType{AttrTypes: ACConfigAttributeType()},
		"state":  types.ObjectType{AttrTypes: ACStateAttributeType()},
	}
}

func ACAttributeValue(ac map[string]interface{}) map[string]attr.Value {
	id := types.StringNull()
	if ac["id"] != nil {
		id = types.StringValue(ac["id"].(string))
	}
	href := types.StringNull()
	if ac["href"] != nil {
		href = types.StringValue(ac["href"].(string))
	}
	config := types.ObjectNull(ACConfigAttributeType())
	if (ac["config"]) != nil {
		config = types.ObjectValueMust(ACConfigAttributeType(), ACConfigAttributeValue(ac["config"].(map[string]interface{})))
	}
	state := types.ObjectNull(ACStateAttributeType())
	if (ac["state"]) != nil {
		state = types.ObjectValueMust(ACStateAttributeType(), ACStateAttributeValue(ac["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"id":     id,
		"href":   href,
		"config": config,
		"state":  state,
	}
}

func ACStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"col_id":        types.Int64Type,
		"capacity":      types.Int64Type,
		"imc":           types.StringType,
		"imc_outer_vid": types.StringType,
		"emc":           types.StringType,
		"emc_outer_vid": types.StringType,
		"ac_ctrl":       types.Int64Type,
		"lifecycle_state" : types.StringType,
	}
}

func ACStateAttributeValue(state map[string]interface{}) map[string]attr.Value {
	lifecycleState := types.StringNull()
	if state["lifecycleState"] != nil {
		lifecycleState = types.StringValue(state["lifecycleState"].(string))
	}
	colId := types.Int64Null()
	if state["colId"] != nil {
		colId = types.Int64Value(int64(state["colId"].(float64)))
	}
	capacity := types.Int64Null()
	if state["capacity"] != nil {
		capacity = types.Int64Value(int64(state["capacity"].(float64)))
	}
	imc := types.StringNull()
	if state["imc"] != nil {
		imc = types.StringValue(state["imc"].(string))
	}
	imcOuterVID := types.StringNull()
	if state["imcOuterVID"] != nil {
		imcOuterVID = types.StringValue(state["imcOuterVID"].(string))
	}
	emc := types.StringNull()
	if state["emc"] != nil {
		emc = types.StringValue(state["emc"].(string))
	}
	emcOuterVID := types.StringNull()
	if state["emcOuterVID"] != nil {
		emcOuterVID = types.StringValue(state["emcOuterVID"].(string))
	}
	acCtrl := types.Int64Null()
	if state["acCtrl"] != nil {
		acCtrl = types.Int64Value(int64(state["acCtrl"].(float64)))
	}
	
	return map[string]attr.Value{
		"col_id":        colId,
		"capacity":      capacity,
		"imc":           imc,
		"imc_outer_vid": imcOuterVID,
		"emc":           emc,
		"emc_outer_vid": emcOuterVID,
		"ac_ctrl":       acCtrl,
		"lifecycle_state": lifecycleState,
	}
}

func ACConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"capacity":      types.Int64Type,
		"imc":           types.StringType,
		"imc_outer_vid": types.StringType,
		"emc":           types.StringType,
		"emc_outer_vid": types.StringType,
		"ac_ctrl":       types.Int64Type,
	}
}

func ACConfigAttributeValue(config map[string]interface{}) map[string]attr.Value {
	capacity := types.Int64Null()
	if config["capacity"] != nil {
		capacity = types.Int64Value(int64(config["capacity"].(float64)))
	}
	imc := types.StringNull()
	if config["imc"] != nil {
		imc = types.StringValue(config["imc"].(string))
	}
	imcOuterVID := types.StringNull()
	if config["imcOuterVID"] != nil {
		imcOuterVID = types.StringValue(config["imcOuterVID"].(string))
	}
	emc := types.StringNull()
	if config["emc"] != nil {
		emc = types.StringValue(config["emc"].(string))
	}
	emcOuterVID := types.StringNull()
	if config["emcOuterVID"] != nil {
		emcOuterVID = types.StringValue(config["emcOuterVID"].(string))
	}
	acCtrl := types.Int64Null()
	if config["acCtrl"] != nil {
		acCtrl = types.Int64Value(int64(config["acCtrl"].(float64)))
	}
	
	return map[string]attr.Value{
		"capacity":      capacity,
		"imc":           imc,
		"imc_outer_vid": imcOuterVID,
		"emc":           emc,
		"emc_outer_vid": emcOuterVID,
		"ac_ctrl":       acCtrl,
	}
}





