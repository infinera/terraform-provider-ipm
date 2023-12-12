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
	_ resource.Resource                = &LCResource{}
	_ resource.ResourceWithConfigure   = &LCResource{}
	_ resource.ResourceWithImportState = &LCResource{}
)

// NewLCResource is a helper function to simplify the provider implementation.
func NewLCResource() resource.Resource {
	return &LCResource{}
}

type LCResource struct {
	client *ipm_pf.Client
}

type LCResourceData struct {
	Id         types.String              `tfsdk:"id"`
	ParentId   types.String              `tfsdk:"parent_id"`
	Href       types.String              `tfsdk:"href"`
	ColId      types.Int64               `tfsdk:"col_id"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	State    types.Object      `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *LCResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_lc"
}

// Schema defines the schema for the data source.
func (r *LCResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type LCResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an NDU LC",
		Attributes:  LCResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *LCResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r LCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LCResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "LCResource: Create - ", map[string]interface{}{"LCResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r LCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LCResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LCResource: Create - ", map[string]interface{}{"LCResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LCResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "LCResource: Update", map[string]interface{}{"LCResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r LCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LCResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "LCResource: Update", map[string]interface{}{"LCResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *LCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}


func (r *LCResource) read(state *LCResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() && state.Identifier.ColId.IsNull() && state.Identifier.Aid.IsNull() && state.Identifier.Id.IsNull() {
		diags.AddError(
			"Error Read LCResource",
			"LCResource: Could not read. Id, and Href, and identifier are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "LCResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/lcs" + queryStr + "&q={\"id\":\"" + state.Id.ValueString() + "\"}"
	} else if !state.Identifier.Aid.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/lcs" + queryStr + "&q={\"state.lcAid\":\"" + state.Identifier.Aid.ValueString() + "\"}"
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/lcs" + queryStr + "&q={\"id\":\"" + state.Identifier.Id.ValueString() + "\"}"
	} else if !state.Identifier.ColId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/lcs/" + state.Identifier.ColId.ValueString() + queryStr
	} else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() + "/lcs" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"LCResource: read ##: Error Read LCResource",
			"Read:Could not get LCResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "LCResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"LCResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal LCResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"LCResource: read ##: Can not get Module",
				"Read:Could not get EClient for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "LCResource: read ## ", map[string]interface{}{"plan": state})
}


func (lc *LCResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "LCResourceData: populate ## ", map[string]interface{}{"plan": data})
	lc.Id = types.StringValue(data["id"].(string))
	lc.ParentId = types.StringValue(data["parentId"].(string))
	lc.Href = types.StringValue(data["href"].(string))
	lc.ColId = types.Int64Value(int64(data["colid"].(float64)))

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		lc.State = types.ObjectValueMust(LCStateAttributeType(),LCStateAttributeValue(state))
	}

	tflog.Debug(ctx, "LCResourceData: read ## ", map[string]interface{}{"plan": state})
}

func LCResourceSchemaAttributes() map[string]schema.Attribute {
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
			Computed:    true,
		},
		"identifier": common.ResourceIdentifierAttribute(),
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed: true,
			AttributeTypes: LCStateAttributeType(),
		},
	}
}

func LCObjectType() (types.ObjectType) {
	return types.ObjectType{	
						AttrTypes: LCAttributeType(),
				}
}

func LCObjectsValue(data []interface{}) []attr.Value {
	lcs := []attr.Value{}
	for _, v := range data {
		lc := v.(map[string]interface{})
		if lc != nil {
			lcs = append(lcs, types.ObjectValueMust(
				LCAttributeType(),
				LCAttributeValue(lc)))
		}
	}
	return lcs
}

func LCAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":     types.StringType,
		"href":   types.StringType,
		"col_id": types.Int64Type,
		"state":  types.ObjectType {AttrTypes:LCStateAttributeType()},
	}
}

func LCAttributeValue(lc map[string]interface{}) map[string]attr.Value {
	col_id := types.Int64Null()
	if lc["colId"] != nil {
		col_id = types.Int64Value(int64(lc["colId"].(float64)))
	}
	href := types.StringNull()
	if lc["href"] != nil {
		href = types.StringValue(lc["href"].(string))
	}
	id := types.StringNull()
	if lc["id"] != nil {
		id = types.StringValue(lc["id"].(string))
	}
	parentId := types.StringNull()
	if lc["parentId"] != nil {
		id = types.StringValue(lc["parentId"].(string))
	}
	state := types.ObjectNull(LCStateAttributeType())
	if (lc["state"]) != nil {
		state = types.ObjectValueMust(LCStateAttributeType(), LCStateAttributeValue(lc["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"col_id" : col_id,
		"parent_id": parentId,
		"id":     id,
		"href":   href,
		"state":  state,
	}
}


func LCStateAttributeType() (map[string]attr.Type) {
	return map[string]attr.Type{
		"parent_aid":       types.StringType,
		"trail_type":    types.StringType,
		"client_aid":           types.StringType,
		"line_aid" : types.StringType,
	}
}

func LCStateAttributeValue(lcState map[string]interface{}) (map[string]attr.Value) {
	parentAid := types.StringNull()
	if lcState["parentAid"] != nil {
		parentAids := lcState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	trailType := types.StringNull()
	if lcState["trailType"] != nil {
		trailType = types.StringValue(lcState["trailType"].(string))
	}
	clientAid := types.StringNull()
	if lcState["clientAid"] != nil {
		clientAid = types.StringValue(lcState["clientAid"].(string))
	}
	lineAid := types.StringNull()
	if lcState["lineAid"] != nil {
		lineAid = types.StringValue(lcState["lineAid"].(string))
	}
	
	return map[string]attr.Value {
		"parent_aid":       parentAid,
		"trail_type":    trailType,
		"client_aid":    clientAid,
		"line_aid" :   lineAid,
	}
}
