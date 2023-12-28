package nduservice

import (
	"context"
	"encoding/json"
	"strings"

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
	_ resource.Resource                = &NDUResource{}
	_ resource.ResourceWithConfigure   = &NDUResource{}
	_ resource.ResourceWithImportState = &NDUResource{}
)

// NewNDUResource is a helper function to simplify the provider implementation.
func NewNDUResource() resource.Resource {
	return &NDUResource{}
}

type NDUResource struct {
	client *ipm_pf.Client
}

type NDULocation struct {
	Description types.String `tfsdk:"description"`
	Clli        types.String `tfsdk:"clli"`
	Latitude    types.Int64  `tfsdk:"latitude"`
	Longitude   types.Int64  `tfsdk:"longitude"`
	Altitude    types.Int64  `tfsdk:"altitude"`
}

type NDUConfig struct {
	Name             types.String `tfsdk:"name"`
	Location         NDULocation  `tfsdk:"location"`
	Contact          types.String `tfsdk:"contact"`
	ManagedBy        types.String `tfsdk:"managed_by"`
	Labels           types.Map    `tfsdk:"labels"`
	PolPowerCtrlMode types.String `tfsdk:"pol_power_ctrl_mode"`
}

type NDUResourceData struct {
	Id        types.String `tfsdk:"id"`
	Href      types.String `tfsdk:"href"`
	Identifier common.DeviceIdentifier `tfsdk:"identifier"`
	Config    *NDUConfig   `tfsdk:"config"`
	State     types.Object `tfsdk:"state"`
	FanUnit   types.Object `tfsdk:"fan_unit"`
	PEM       types.Object `tfsdk:"pem"`
	LEDs      types.List   `tfsdk:"leds"`
	Ports     types.List   `tfsdk:"ports"`
	LCs       types.List   `tfsdk:"lcs"`
	OTUs      types.List   `tfsdk:"otus"`
	Ethernets types.List   `tfsdk:"ethernets"`
}

// Metadata returns the data source type name.
func (r *NDUResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ndu"
}

// Schema defines the schema for the data source.
func (r *NDUResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type NDUResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages NDU",
		Attributes:  NDUResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *NDUResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r NDUResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NDUResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "NDUResource: Create - ", map[string]interface{}{"NDUResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.update(&data, ctx, &resp.Diagnostics)

	resp.Diagnostics.Append(diags...)
}

func (r NDUResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NDUResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "NDUResource: Create - ", map[string]interface{}{"NDUResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.read(&data, ctx, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r NDUResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NDUResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "NDUResource: Update", map[string]interface{}{"NDUResourceData": data})

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

func (r NDUResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NDUResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "NDUResource: Delete", map[string]interface{}{"NDUResourceData": data})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.delete(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *NDUResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve NDU ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NDUResource) update(plan *NDUResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "NDUResource: update ## ", map[string]interface{}{"plan": plan})

	if plan.Href.IsNull() && plan.Id.IsNull() && plan.Identifier.DeviceId.IsNull()  {
		diags.AddError(
			"NDUResource: Error update NDU",
			"NDUResource: Could not update NDU. Href or NDU ColId is not specified.",
		)
		return
	}

	var updateRequest = make(map[string]interface{})
	// get TC config settings
	if !plan.Config.Name.IsNull() {
		updateRequest["name"] = plan.Config.Name.ValueString()
	}
	if !plan.Config.ManagedBy.IsNull() {
		updateRequest["managedBy"] = plan.Config.ManagedBy.ValueString()
	}
	if !plan.Config.Contact.IsNull() {
		updateRequest["contact"] = plan.Config.Contact.ValueString()
	}
	if !plan.Config.ManagedBy.IsNull() {
		updateRequest["managedBy"] = plan.Config.ManagedBy.ValueString()
	}
	location := make(map[string]interface{})
	if !plan.Config.Location.Description.IsNull() {
		location["description"] = plan.Config.Location.Description.ValueString()
	}
	if !plan.Config.Location.Clli.IsNull() {
		location["clli"] = plan.Config.Location.Clli.ValueString()
	}
	if !plan.Config.Location.Latitude.IsNull() {
		location["latitude"] = plan.Config.Location.Latitude.ValueInt64()
	}
	if !plan.Config.Location.Longitude.IsNull() {
		location["longitude"] = plan.Config.Location.Longitude.ValueInt64()
	}
	if !plan.Config.Location.Altitude.IsNull() {
		location["altitude"] = plan.Config.Location.Altitude.ValueInt64()
	}
	if len(location) > 0 {
		updateRequest["location"] = location
	}
	if !plan.Config.Labels.IsNull() {
		labels := map[string]string{}
		diag := plan.Config.Labels.ElementsAs(ctx, &labels, true)
		if !diag.HasError() {
			updateRequest["labels"] = labels
		}
	}

	tflog.Debug(ctx, "NDUResource: update ## ", map[string]interface{}{"Create Request": updateRequest})

	if len(updateRequest) > 0 {
		// send update request to server
		rb, err := json.Marshal(updateRequest)
		if err != nil {
			diags.AddError(
				"NDUResource: update ##: Error Create AC",
				"Create: Could not Marshal NDUResource, unexpected error: "+err.Error(),
			)
			return
		}
		var body []byte
		if !plan.Href.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", plan.Href.ValueString(), rb)
		} else if !plan.Id.IsNull() {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Id.ValueString(), rb)
		} else {
			body, err = r.client.ExecuteIPMHttpCommand("PUT", "/ndus/" + plan.Identifier.DeviceId.ValueString(), rb)
		}
		if err != nil {
			if !strings.Contains(err.Error(), "status: 202") {
				diags.AddError(
					"ModuleResource: update ##: Error update Module",
					"Update: Could not update Module, unexpected error: "+err.Error(),
				)
				return
			}
		}

		tflog.Debug(ctx, "NDUResource: update ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
		var data []interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			diags.AddError(
				"NDUResource: Create ##: Error Unmarshal response",
				"Update:Could not Create NDUResource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	r.read(plan, ctx, diags)
	if diags.HasError() {
		tflog.Debug(ctx, "NDUResource: update failed. Can't find the updated network")
		plan.Id = types.StringNull()
		plan.Href = types.StringNull()
		return
	}

	tflog.Debug(ctx, "NDUResource: update ##", map[string]interface{}{"plan": plan})
}

func (r *NDUResource) delete(plan *NDUResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if plan.Href.IsNull() && plan.Id.IsNull() {
		diags.AddError(
			"Error Delete NDUResource",
			"Read: Could not delete. Both NC ID and Href are not specified",
		)
		return
	}

	_, err := r.client.ExecuteIPMHttpCommand("DELETE", plan.Href.ValueString(), nil)
	if err != nil {
		diags.AddError(
			"NDUResource: delete ##: Error Delete NDUResource",
			"Update:Could not delete NDUResource, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "NDUResource: delete ## ", map[string]interface{}{"plan": plan})
}

func (r *NDUResource) read(state *NDUResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if state.Id.IsNull() && state.Href.IsNull() && state.Identifier.DeviceId.IsNull() {
		diags.AddError(
			"Error Read NDUResource",
			"NDUResource: Could not read. NDU. Href and Identifier are not specified.",
		)
		return
	}

	tflog.Debug(ctx, "NDUResource: read ## ", map[string]interface{}{"plan": state})
	queryStr := "?content=expanded"
	if !state.Href.IsNull() {
		queryStr = state.Href.ValueString() + queryStr
	} else if !state.Id.IsNull() {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString() + queryStr
	} else {
		queryStr = "/ndus/" + state.Identifier.DeviceId.ValueString()  + queryStr
	}

	body, err := r.client.ExecuteIPMHttpCommand("GET", queryStr, nil)
	if err != nil {
		diags.AddError(
			"NDUResource: read ##: Error Read NDUResource",
			"Read:Could not get NDUResource, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "NDUResource: read ## ExecuteIPMHttpCommand ..", map[string]interface{}{"response": string(body)})
	var resp interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		diags.AddError(
			"NDUResource: read ##: Error Unmarshal response",
			"Read:Could not Unmarshal NDUResource, unexpected error: "+err.Error(),
		)
		return
	}
	switch resp.(type) {
	case []interface{}:
		if len(resp.([]interface{})) > 0 {
			state.populate((resp.([]interface{})[0]).(map[string]interface{}), ctx, diags)
		} else {
			diags.AddError(
				"NDUResource: read ##: Can not get Module",
				"Read:Could not get ODU for query: "+queryStr,
			)
			return
		}
	case map[string]interface{}:
		state.populate(resp.(map[string]interface{}), ctx, diags)
	}

	tflog.Debug(ctx, "NDUResource: read ## ", map[string]interface{}{"plan": state})
}

func (ndu *NDUResourceData) populate(data map[string]interface{}, ctx context.Context, diags *diag.Diagnostics) {
	tflog.Debug(ctx, "NDUResourceData: populate ## ", map[string]interface{}{"plan": data})

	ndu.Id = types.StringValue(data["id"].(string))
	ndu.Href = types.StringValue(data["href"].(string))

	// populate config
	var config = data["config"].(map[string]interface{})
	if config != nil {
		if ndu.Config == nil {
			ndu.Config = &NDUConfig{}
		}
		for k, v := range config {
			switch k {
			case "name":
				if !ndu.Config.Name.IsNull() {
					ndu.Config.Name = types.StringValue(v.(string))
				}
			case "contact":
				if !ndu.Config.Contact.IsNull() {
					ndu.Config.Contact = types.StringValue(v.(string))
				}
			case "managedBy":
				if !ndu.Config.ManagedBy.IsNull() {
					ndu.Config.ManagedBy = types.StringValue(v.(string))
				}
			case "polPowerCtrlMode":
				if !ndu.Config.PolPowerCtrlMode.IsNull() {
					ndu.Config.PolPowerCtrlMode = types.StringValue(v.(string))
				}
			case "location":
				location := v.(map[string]interface{})
				for k1, v1 := range location {
					switch k1 {
					case "description":
						if !ndu.Config.Location.Description.IsNull() {
							ndu.Config.Location.Description = types.StringValue(v1.(string))
						}
					case "clli":
						if !ndu.Config.Location.Clli.IsNull() {
							ndu.Config.Location.Clli = types.StringValue(v1.(string))
						}
					case "latitude":
						if !ndu.Config.Location.Latitude.IsNull() {
							ndu.Config.Location.Latitude = types.Int64Value(int64(v1.(float64)))
						}
					case "longitude":
						if !ndu.Config.Location.Longitude.IsNull() {
							ndu.Config.Location.Longitude = types.Int64Value(int64(v1.(float64)))
						}
					case "altitude":
						if !ndu.Config.Location.Altitude.IsNull() {
							ndu.Config.Location.Altitude = types.Int64Value(int64(v1.(float64)))
						}
					}
				}
			case "labels":
				labels := types.MapNull(types.StringType)
				data := make(map[string]attr.Value)
				for k, label := range v.(map[string]interface{}) {
					data[k] = types.StringValue(label.(string))
				}
				labels = types.MapValueMust(types.StringType, data)
				if !ndu.Config.Labels.IsNull() {
					ndu.Config.Labels = labels
				}
			}
		}
	}

	// populate state
	var state = data["state"].(map[string]interface{})
	if state != nil {
		ndu.State = types.ObjectValueMust(NDUStateAttributeType(), NDUStateAttributeValue(state))
	}
	// populate fanunit
	var fanUnit = data["fanUnit"].(map[string]interface{})
	if fanUnit != nil {
		ndu.FanUnit = types.ObjectValueMust(FanUnitStateAttributeType(), FanUnitStateAttributeValue(fanUnit))
	}
	// populate PEMs
	var pem = data["pem"].(map[string]interface{})
	if pem != nil {
		ndu.PEM = types.ObjectValueMust(PEMStateAttributeType(), PEMStateAttributeValue(pem))
	}
	// populate PEM
	var leds = data["leds"].([]interface{})
	if len(leds) > 0 {
		ndu.LEDs = types.ListValueMust(LEDsObjectType(), LEDsObjectsValue(leds))
	}
	// populate ports
	var ports = data["ports"].([]interface{})
	if len(ports) > 0 {
		ndu.Ports = types.ListValueMust(PortObjectType(), PortObjectsValue(ports))
	}
	// populate ports
	var lcs = data["lcs"].([]interface{})
	if len(lcs) > 0 {
		ndu.LCs = types.ListValueMust(LCObjectType(), LCObjectsValue(lcs))
	}
	// populate otus
	var otus = data["otus"].([]interface{})
	if len(otus) > 0 {
		ndu.OTUs = types.ListValueMust(OTUObjectType(), OTUObjectsValue(otus))
	}
	// populate ethernets
	var ethernets = data["ethernet"].([]interface{})
	if ethernets != nil && len(otus) > 0 {
		ndu.Ethernets = types.ListValueMust(EClientObjectType(), EClientObjectsValue(ethernets))
	}

	tflog.Debug(ctx, "NDUResourceData: read ## ", map[string]interface{}{"plan": state})
}

func NDUResourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the Carrier.",
			Computed:    true,
		},
		"href": schema.StringAttribute{
			Description: "href",
			Computed:    true,
		},
		"identifier": common.DeviceIdentifierAttribute(),
		//Config    NodeConfig `tfsdk:"config"`
		"config": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "name",
					Optional:    true,
				},
				"managed_by": schema.StringAttribute{
					Description: "managed_by",
					Optional:    true,
				},
				"contact": schema.StringAttribute{
					Description: "contact",
					Optional:    true,
				},
				"pol_power_ctrl_mode": schema.StringAttribute{
					Description: "polPowerCtrlMode",
					Optional:    true,
				},
				"location": schema.SingleNestedAttribute{
					Description: "location",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Description: "description",
							Optional:    true,
						},
						"clli": schema.StringAttribute{
							Description: "clli",
							Optional:    true,
						},
						"latitude": schema.Int64Attribute{
							Description: "latitude",
							Optional:    true,
						},
						"longitude": schema.Int64Attribute{
							Description: "longitude",
							Optional:    true,
						},
						"altitude": schema.Int64Attribute{
							Description: "altitude",
							Optional:    true,
						},
					},
				},
				"labels": schema.MapAttribute{
					Description: "labels",
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
		//State     types.Object   `tfsdk:"state"`
		"state": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: NDUStateAttributeType(),
		},
		"fan_unit": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: FanUnitStateAttributeType(),
		},
		"pem": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: PEMStateAttributeType(),
		},
		"leds": schema.ObjectAttribute{
			Computed:       true,
			AttributeTypes: LEDsStateAttributeType(),
		},
		"ports": schema.ListAttribute{
			Computed:    true,
			ElementType: PortObjectType(),
		},
		"lcs": schema.ListAttribute{
			Computed:    true,
			ElementType: LCObjectType(),
		},
		"otus": schema.ListAttribute{
			Computed:    true,
			ElementType: OTUObjectType(),
		},
		"ethernet": schema.ListAttribute{
			Computed:    true,
			ElementType: EClientObjectType(),
		},
	}
}

func NDUObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: NDUAttributeType(),
	}
}

func NDUObjectsValue(data []interface{}) []attr.Value {
	ndus := []attr.Value{}
	for _, v := range data {
		ndu := v.(map[string]interface{})
		if ndu != nil {
			ndus = append(ndus, types.ObjectValueMust(
				NDUAttributeType(),
				NDUAttributeValue(ndu)))
		}
	}
	return ndus
}

func NDUAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.StringType,
		"href":      types.StringType,
		"config":    types.ObjectType{AttrTypes: NDUConfigAttributeType()},
		"state":     types.ObjectType{AttrTypes: NDUStateAttributeType()},
		"fan_unit":  types.ObjectType{AttrTypes: FanUnitAttributeType()},
		"pem":       types.ObjectType{AttrTypes: PEMAttributeType()},
		"leds":      types.ObjectType{AttrTypes: LEDsAttributeType()},
		"ports":     types.ListType{ElemType: PortObjectType()},
		"lcs":       types.ListType{ElemType: LCObjectType()},
		"otus":      types.ListType{ElemType: OTUObjectType()},
		"ethernets": types.ListType{ElemType: EClientObjectType()},
	}
}

func NDUAttributeValue(ndu map[string]interface{}) map[string]attr.Value {
	href := types.StringNull()
	id := types.StringNull()
	config := types.ObjectNull(NDUConfigAttributeType())
	state := types.ObjectNull(NDUStateAttributeType())
	fanUnit := types.ObjectNull(FanUnitAttributeType())
	pem := types.ObjectNull(PEMAttributeType())
	leds := types.ObjectNull(LEDsAttributeType())
	ports := types.ListNull(PortObjectType())
	lcs := types.ListNull(LCObjectType())
	otus := types.ListNull(OTUObjectType())
	ethernets := types.ListNull(EClientObjectType())

	for k, v := range ndu {
		switch k {
		case "href":
			href = types.StringValue(v.(string))
		case "id":
			id = types.StringValue(v.(string))
		case "config":
			config = types.ObjectValueMust(NDUConfigAttributeType(), NDUConfigAttributeValue(v.(map[string]interface{})))
		case "state":
			state = types.ObjectValueMust(NDUStateAttributeType(), NDUStateAttributeValue(v.(map[string]interface{})))
		case "fanUnit":
			fanUnit = types.ObjectValueMust(FanUnitAttributeType(), FanUnitAttributeValue(v.(map[string]interface{})))
		case "pem":
			pem = types.ObjectValueMust(PEMAttributeType(), PEMAttributeValue(v.(map[string]interface{})))
		case "leds":
			leds = types.ObjectValueMust(LEDsAttributeType(), LEDsAttributeValue(v.(map[string]interface{})))
		case "ports":
			ports = types.ListValueMust(PortObjectType(), PortObjectsValue(v.([]interface{})))
		case "lcs":
			lcs = types.ListValueMust(LCObjectType(), LCObjectsValue(v.([]interface{})))
		case "otus":
			otus = types.ListValueMust(OTUObjectType(), OTUObjectsValue(v.([]interface{})))
		case "ethernet":
			ethernets = types.ListValueMust(EClientObjectType(), EClientObjectsValue(v.([]interface{})))
		}
	}

	return map[string]attr.Value{
		"id":        id,
		"href":      href,
		"config":    config,
		"state":     state,
		"fan_unit":  fanUnit,
		"pem":       pem,
		"leds":      leds,
		"ports":     ports,
		"lcs":       lcs,
		"otus":      otus,
		"ethernets": ethernets,
	}
}

func NDUConfigAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"location":            types.ObjectType{AttrTypes: LocationAttributeType()},
		"contact":             types.StringType,
		"managed_by":          types.StringType,
		"pol_power_ctrl_mode": types.StringType,
		"labels":              types.MapType{ElemType: types.StringType},
	}
}

func NDUConfigAttributeValue(nduConfig map[string]interface{}) map[string]attr.Value {
	name := types.StringNull()
	location := types.ObjectNull(LocationAttributeType())
	contact := types.StringNull()
	managedBy := types.StringNull()
	polPowerCtrlMode := types.StringNull()
	labels := types.MapNull(types.StringType)

	for k, v := range nduConfig {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		case "location":
			location = types.ObjectValueMust(LocationAttributeType(), LocationAttributeValue(v.(map[string]interface{})))
		case "contact":
			contact = types.StringValue(v.(string))
		case "managedBy":
			managedBy = types.StringValue(v.(string))
		case "polPowerCtrlMode":
			polPowerCtrlMode = types.StringValue(v.(string))
		case "labels":
			data := make(map[string]attr.Value)
			for k1, v1 := range v.(map[string]interface{}) {
				data[k1] = types.StringValue(v1.(string))
			}
			labels, _ = types.MapValue(types.StringType, data)
		}
	}

	return map[string]attr.Value{
		"name":             name,
		"location":         location,
		"contact":          contact,
		"managed_by":       managedBy,
		"polPowerCtrlMode": polPowerCtrlMode,
		"labels":           labels,
	}
}

func LocationAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"description": types.StringType,
		"clli":        types.StringType,
		"longitude":   types.Int64Type,
		"altitude":    types.Int64Type,
	}
}

func LocationAttributeValue(location map[string]interface{}) map[string]attr.Value {
	description := types.StringNull()
	clli := types.StringNull()
	longitude := types.Int64Null()
	altitude := types.Int64Null()

	for k, v := range location {
		switch k {
		case "description":
			description = types.StringValue(v.(string))
		case "clli":
			clli = types.StringValue(v.(string))
		case "longitude":
			longitude = types.Int64Value(int64(v.(float64)))
		case "altitude":
			altitude = types.Int64Value(int64(v.(float64)))
		}
	}

	return map[string]attr.Value{
		"description": description,
		"clli":        clli,
		"longitude":   longitude,
		"altitude":    altitude,
	}
}

func NDUStateAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"ndu_aid":              types.StringType,
		"name":                 types.StringType,
		"location":             types.ObjectType{AttrTypes: LocationAttributeType()},
		"contact":              types.StringType,
		"managed_by":           types.StringType,
		"pol_power_ctrl_mode":  types.StringType,
		"labels":               types.MapType{ElemType: types.StringType},
		"connectivity_state":   types.StringType,
		"lifecycle_state":      types.StringType,
		"restart_action":       types.StringType,
		"factory_reset_action": types.BoolType,
		"hw_description":       types.ObjectType{AttrTypes: NDUHWDescriptionAttributeType()},
	}
}

func NDUStateAttributeValue(nduState map[string]interface{}) map[string]attr.Value {
	nduAid := types.StringNull()
	name := types.StringNull()
	location := types.ObjectNull(LocationAttributeType())
	contact := types.StringNull()
	managedBy := types.StringNull()
	polPowerCtrlMode := types.StringNull()
	labels := types.MapNull(types.StringType)
	connectivityState := types.StringNull()
	lifecycleState := types.StringNull()
	restartAction := types.StringNull()
	factoryResetAction := types.BoolNull()
	hwDescription := types.ObjectNull(NDUHWDescriptionAttributeType())

	for k, v := range nduState {
		switch k {
		case "name":
			name = types.StringValue(v.(string))
		case "location":
			location = types.ObjectValueMust(LocationAttributeType(), LocationAttributeValue(v.(map[string]interface{})))
		case "nduAid":
			nduAid = types.StringValue(v.(string))
		case "contact":
			contact = types.StringValue(v.(string))
		case "lifecycleState":
			lifecycleState = types.StringValue(v.(string))
		case "managedBy":
			managedBy = types.StringValue(v.(string))
		case "polPowerCtrlMode":
			polPowerCtrlMode = types.StringValue(v.(string))
		case "labels":
			data := make(map[string]attr.Value)
			for k1, v1 := range v.(map[string]interface{}) {
				data[k1] = types.StringValue(v1.(string))
			}
			labels, _ = types.MapValue(types.StringType, data)
		case "connectivityState":
			connectivityState = types.StringValue(v.(string))
		case "restartAction":
			restartAction = types.StringValue(v.(string))
		case "factoryResetAction":
			factoryResetAction = types.BoolValue(v.(bool))
		case "hwDescription":
			hwDescription = types.ObjectValueMust(NDUHWDescriptionAttributeType(), NDUHWDescriptionAttributeValue(v.(map[string]interface{})))
		}
	}

	return map[string]attr.Value{
		"ndu_aid":              nduAid,
		"name":                 name,
		"location":             location,
		"contact":              contact,
		"lifecycle_state":      lifecycleState,
		"managed_by":           managedBy,
		"pol_power_ctrl_mode":  polPowerCtrlMode,
		"labels":               labels,
		"connectivity_state":   connectivityState,
		"restart_action":       restartAction,
		"factory_reset_action": factoryResetAction,
		"hw_description":       hwDescription,
	}
}

func NDUHWDescriptionAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"pi":            types.StringType,
		"mnfv":          types.StringType,
		"mnmn":          types.StringType,
		"mnmo":          types.StringType,
		"mnhw":          types.StringType,
		"mndt":          types.StringType,
		"serial_number": types.StringType,
		"clei":          types.StringType,
		"mac_address":   types.StringType,
		"piid":          types.StringType,
		"sv":            types.StringType,
		"icv":           types.StringType,
		"dmn":           types.ListType{ElemType: DMNObjectType()},
	}
}

func NDUHWDescriptionAttributeValue(moduleHWDescription map[string]interface{}) map[string]attr.Value {
	pi := types.StringNull()
	mnfv := types.StringNull()
	mnmn := types.StringNull()
	mnmo := types.StringNull()
	mnhw := types.StringNull()
	mndt := types.StringNull()
	serialNumber := types.StringNull()
	clei := types.StringNull()
	macAddress := types.StringNull()
	piid := types.StringNull()
	sv := types.StringNull()
	icv := types.StringNull()
	dmn := types.ListNull(DMNObjectType())

	for k, v := range moduleHWDescription {
		switch k {
		case "pi":
			pi = types.StringValue(v.(string))
		case "mnfv":
			mnfv = types.StringValue(v.(string))
		case "mnmn":
			mnmn = types.StringValue(v.(string))
		case "mnmo":
			mnmo = types.StringValue(v.(string))
		case "mnhw":
			mnhw = types.StringValue(v.(string))
		case "mndt":
			mndt = types.StringValue(v.(string))
		case "serialNumber":
			serialNumber = types.StringValue(v.(string))
		case "clei":
			clei = types.StringValue(v.(string))
		case "macAddress":
			macAddress = types.StringValue(v.(string))
		case "sv":
			sv = types.StringValue(v.(string))
		case "icv":
			icv = types.StringValue(v.(string))
		case "piid":
			piid = types.StringValue(v.(string))
		case "dmn":
			dmn = types.ListValueMust(DMNObjectType(), DMNObjectsValue(v.([]interface{})))
		}
	}
	return map[string]attr.Value{
		"pi":            pi,
		"mnfv":          mnfv,
		"mnmn":          mnmn,
		"mnmo":          mnmo,
		"mnhw":          mnhw,
		"mndt":          mndt,
		"serial_number": serialNumber,
		"clei":          clei,
		"mac_address":   macAddress,
		"sv":            sv,
		"icv":           icv,
		"piid":          piid,
		"dmn":           dmn,
	}
}

func DMNObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: DMNAttributeType(),
	}
}

func DMNObjectsValue(data []interface{}) []attr.Value {
	dmns := []attr.Value{}
	for _, v := range data {
		dmn := v.(map[string]interface{})
		if dmn != nil {
			dmns = append(dmns, types.ObjectValueMust(
				DMNAttributeType(),
				DMNAttributeValue(dmn)))
		}
	}
	return dmns
}

func DMNAttributeType() map[string]attr.Type {
	return map[string]attr.Type{
		"language": types.StringType,
		"value":    types.StringType,
	}
}

func DMNAttributeValue(moduleHWDescription map[string]interface{}) map[string]attr.Value {
	language := types.StringNull()
	value := types.StringNull()

	for k, v := range moduleHWDescription {
		switch k {
		case "language":
			language = types.StringValue(v.(string))
		case "value":
			value = types.StringValue(v.(string))
		}
	}
	return map[string]attr.Value{
		"language": language,
		"value":    value,
	}
}
