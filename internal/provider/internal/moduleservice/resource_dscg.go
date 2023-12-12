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
	_ resource.Resource                = &DSCGResource{}
	_ resource.ResourceWithConfigure   = &DSCGResource{}
	_ resource.ResourceWithImportState = &DSCGResource{}
)

// NewDSCGResource is a helper function to simplify the provider implementation.
func NewDSCGResource() resource.Resource {
	return &DSCGResource{}
}

type DSCGResource struct {
	client *ipm_pf.Client
}

type DSCGResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	Href       types.String              `tfsdk:"href"`
	ColId      types.Int64               `tfsdk:"col_id"`
	State      types.Object              `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *DSCGResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dscg"
}

// Schema defines the schema for the data source.
func (r *DSCGResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type DSCGResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an ModuleLinePTP",
		Attributes:  DSCGResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *DSCGResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r DSCGResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DSCGResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "DSCGResource: Create - ", map[string]interface{}{"DSCGResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r DSCGResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DSCGResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "DSCGResource: Create - ", map[string]interface{}{"DSCGResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r DSCGResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DSCGResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "DSCGResource: Update", map[string]interface{}{"DSCGResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r DSCGResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DSCGResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "DSCGResource: Update", map[string]interface{}{"DSCGResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *DSCGResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DSCGResource) read(state *DSCGResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && (!state.Identifier.DeviceId.IsNull() && (state.Identifier.ColId.IsNull() || state.Identifier.ParentColId.IsNull() || state.Identifier.GrandParentColId.IsNull()) && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull()) {
		diags.AddError(
			"Error Read DSCGResource",
			"DSCGResource: Could not read. Id, and Href, and identifiers are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "DSCGResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.GrandParentColId.ValueString() + "/carriers/" + state.Identifier.ParentColId.ValueString() + "/dscgs" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.GrandParentColId.ValueString() + "/carriers/" + state.Identifier.ParentColId.ValueString() + "/dscgs" + queryStr + "&q={\"state.dscgAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.GrandParentColId.ValueString() + "/carriers/" + state.Identifier.ParentColId.ValueString() + "/dscgs" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/modules/" + state.Identifier.DeviceId.ValueString() + "/linePtps/" + state.Identifier.GrandParentColId.ValueString() + "/carriers/" + state.Identifier.ParentColId.ValueString() + "/dscgs/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/modules" + state.Identifier.DeviceId.ValueString() + "/linePtps" + state.Identifier.ParentColId.ValueString() + "/carriers" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"DSCGResource: read ##: Error Read DSCGResource",
			"Read:Could not get DSCGResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "DSCGResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"DSCGResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal DSCGResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"DSCGResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "DSCGResource: read ## ", map[string]interface{}{"plan": state})
}

func (dscg *DSCGResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "DSCGResourceData: populate ## ", map[string]interface{}{"plan": data})

	dscg.Id = types.StringValue(data["id"].(string))
	dscg.Href = types.StringValue(data["href"].(string))
	dscg.ParentId = types.StringValue(data["parentId"].(string))
	dscg.ColId = types.Int64Value(int64(data["colId"].(float64)))

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		dscg.State = types.ObjectValueMust(DSCGStateAttributeType(), DSCGStateAttributeValue(state))
	}

	tflog.Debug(ctx, "DSCGResourceData: read ## ", map[string]interface{}{"plan": state})
}

func DSCGResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Module.",
			Computed:    true,
		},
		"parent_id": schema.StringAttribute{
			Description: "module id",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"col_id": schema.Int64Attribute{
			Description: "col id",
			Optional:    true,
		},
		"identifier": common.ResourceIdentifierAttribute(),
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: DSCGStateAttributeType(),
		},
	}
}

func DSCGObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: DSCGAttributeType(),
	}
}

func DSCGObjectsValue(data []interface{}) []attr.Value {
	dscgs := []attr.Value{}
	for _, v := range data {
		dscg := v.(map[string]interface{})
		if dscg != nil {
			dscgs = append(dscgs, types.ObjectValueMust(
				DSCGAttributeType(),
				DSCGAttributeValue(dscg)))
		}
	}
	return dscgs
}

func DSCGAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":        types.StringType,
		"href":      types.StringType,
		"col_id":    types.Int64Type,
		"state":     types.ObjectType{AttrTypes: DSCGStateAttributeType()},
	}
}

func DSCGAttributeValue(dscg map[string]interface{}) map[string]attr.Value {
	col_id := types.Int64Null()
	if dscg["colId"] != nil {
		col_id = types.Int64Value(int64(dscg["colId"].(float64)))
	}
	href := types.StringNull()
	if dscg["href"] != nil {
		href = types.StringValue(dscg["href"].(string))
	}
	parentId := types.StringNull()
	if dscg["parentId"] != nil {
		parentId = types.StringValue(dscg["parentId"].(string))
	}
	id := types.StringNull()
	if dscg["id"] != nil {
		id = types.StringValue(dscg["id"].(string))
	}
	state := types.ObjectNull(DSCGStateAttributeType())
	if (dscg["state"]) != nil {
		state = types.ObjectValueMust(DSCGStateAttributeType(), DSCGStateAttributeValue(dscg["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"col_id":    col_id,
		"parent_id": parentId,
		"id":        id,
		"href":      href,
		"state":     state,
	}
}

func DSCGStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"dscg_aid":   types.StringType,
		"parent_aid": types.StringType,
		"dscg_ctrl":  types.Int64Type,
		"tx_cdscs":   types.ListType{ElemType: types.Int64Type},
		"rx_cdscs":   types.ListType{ElemType: types.Int64Type},
		"idle_cdscs": types.ListType{ElemType: types.Int64Type},
	}
}

func DSCGStateAttributeValue(dscgState map[string]interface{}) map[string]attr.Value {
	dscgAid := types.StringNull()
	if dscgState["dscgAid"] != nil {
		dscgAid = types.StringValue(dscgState["dscgAid"].(string))
	}
	parentAid := types.StringNull()
	if dscgState["parentAid"] != nil {
		parentAids := dscgState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	dscgCtrl := types.Int64Null()
	if dscgState["dscgCtrl"] != nil {
		dscgCtrl = types.Int64Value(int64(dscgState["dscgCtrl"].(float64)))
	}
	txCDSCs := types.ListNull(types.Int64Type)
	if dscgState["txCDSCs"] != nil {
		txCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(dscgState["txCDSCs"].([]interface{})))
	}
	rxCDSCs := types.ListNull(types.Int64Type)
	if dscgState["rxCDSCs"] != nil {
		rxCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(dscgState["rxCDSCs"].([]interface{})))
	}
	idleCDSCs := types.ListNull(types.Int64Type)
	if dscgState["idleCDSCs"] != nil {
		idleCDSCs = types.ListValueMust(types.Int64Type, common.ListAttributeInt64Value(dscgState["idleCDSCs"].([]interface{})))
	}
	return map[string]attr.Value{
		"dscg_aid":   dscgAid,
		"parent_aid": parentAid,
		"dscg_ctrl":  dscgCtrl,
		"tx_cdscs":   txCDSCs,
		"rx_cdscs":   rxCDSCs,
		"idle_cdscs": idleCDSCs,
	}
}
