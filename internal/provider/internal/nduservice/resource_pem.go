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
	_ resource.Resource                = &PEMResource{}
	_ resource.ResourceWithConfigure   = &PEMResource{}
	_ resource.ResourceWithImportState = &PEMResource{}
)

// NewPEMResource is a helper function to simplify the provider implementation.
func NewPEMResource() resource.Resource {
	return &PEMResource{}
}

type PEMResource struct {
	client *ipm_pf.Client
}

type PEMResourceData struct {
	Id    types.String `tfsdk:"id"`
	ParentId   types.String  `tfsdk:"parent_id"`
	Href  types.String `tfsdk:"href"`
	Identifier common.ResourceIdentifier `tfsdk:"identifier"`
	State types.Object `tfsdk:"state"`
}

// Metadata returns the data source type name.
func (r *PEMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu_pem"
}

// Schema defines the schema for the data source.
func (r *PEMResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type PEMResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages an NDU pem",
		Attributes:  PEMResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *PEMResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r PEMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PEMResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "PEMResource: Create - ", map[string]interface{}{"PEMResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r PEMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PEMResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "PEMResource: Create - ", map[string]interface{}{"PEMResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r PEMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PEMResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "PEMResource: Update", map[string]interface{}{"PEMResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r PEMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PEMResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "PEMResource: Update", map[string]interface{}{"PEMResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *PEMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PEMResource) read(state *PEMResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Href.IsNull() && state.Id.IsNull() && state.Identifier.Href.IsNull() && state.Identifier.DeviceId.IsNull() {
		diags.AddError(
			"Error Read PEMResource",
			"PEMResource: Could not read. Fan. Href or (NDUId and Fan ColId) is not specified.",
		)
		return
	}

	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Identifier.Href.IsNull() {
		queryStr = state.Identifier.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Id.ValueString() + "/pem"  + queryStr 
	} else if !state.Identifier.DeviceId.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + "/pem"  + queryStr 
	} else if !state.Identifier.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.Id.ValueString() + "/pem"  + queryStr 
	}else {
		queryStr = "/ndus" + state.Identifier.DeviceId.ValueString() + "/pem" + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"PEMResource: read ##: Error Read PEMResource",
			"Read:Could not get PEMResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "PEMResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})

	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"PEMResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal PEMResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"PEMResource: read ##: Can not get Module",
				"Read:Could not get EClient for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "PEMResource: read ## ", map[string]interface{}{"plan": state})
}

func (pem *PEMResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "PEMResourceData: populate ## ", map[string]interface{}{"plan": data})

	pem.Id = types.StringValue(data["id"].(string))
	pem.ParentId = types.StringValue(data["parentId"].(string))
	pem.Href = types.StringValue(data["href"].(string))

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		pem.State = types.ObjectValueMust(PEMStateAttributeType(), PEMStateAttributeValue(state))
	}

	tflog.Debug(ctx, "PEMResourceData: read ## ", map[string]interface{}{"plan": state})
}

func PEMResourceSchemaAttributes() map[string]schema.Attribute {
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
		"identifier": common.ResourceIdentifierAttribute(),
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: PEMStateAttributeType(),
		},
	}
}

func PEMObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: PEMAttributeType(),
	}
}

func PEMObjectsValue(data []interface{}) []attr.Value {
	pems := []attr.Value{}
	for _, v := range data {
		pem := v.(map[string]interface{})
		if pem != nil {
			pems = append(pems, types.ObjectValueMust(
				PEMAttributeType(),
				PEMAttributeValue(pem)))
		}
	}
	return pems
}

func PEMAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_id": types.StringType,
		"id":     types.StringType,
		"href":   types.StringType,
		"state":  types.ObjectType{AttrTypes: PEMStateAttributeType()},
	}
}

func PEMAttributeValue(pem map[string]interface{}) map[string]attr.Value {
	href := types.StringNull()
	if pem["href"] != nil {
		href = types.StringValue(pem["href"].(string))
	}
	id := types.StringNull()
	if pem["id"] != nil {
		id = types.StringValue(pem["id"].(string))
	}
	parentId := types.StringNull()
	if pem["parentId"] != nil {
		parentId = types.StringValue(pem["parentId"].(string))
	}
	state := types.ObjectNull(PEMStateAttributeType())
	if (pem["state"]) != nil {
		state = types.ObjectValueMust(PEMStateAttributeType(), PEMStateAttributeValue(pem["state"].(map[string]interface{})))
	}

	return map[string]attr.Value{
		"parent_id": parentId,
		"id":     id,
		"href":   href,
		"state":  state,
	}
}

func PEMStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"parent_aid":       types.StringType,
		"pem_aid":     types.StringType,
		"feed_status": types.StringType,
	}
}

func PEMStateAttributeValue(pemState map[string]interface{}) map[string]attr.Value {
	pemAid := types.StringNull()
	if pemState["pemAid"] != nil {
		pemAid = types.StringValue(pemState["pemAid"].(string))
	}
	parentAid := types.StringNull()
	if pemState["parentAid"] != nil {
		parentAids := pemState["parentAid"].([]interface{})
		parentAid = types.StringValue(parentAids[0].(string))
	}
	feedStatus := types.StringNull()
	if pemState["feedStatus"] != nil {
		feedStatuses := pemState["feedStatus"].([]interface{})
		feedStatus = types.StringValue(feedStatuses[0].(string))
	}

	return map[string]attr.Value{
		"pem_aid":     pemAid,
		"parent_aid":       parentAid,
		"feed_status": feedStatus,
	}
}
