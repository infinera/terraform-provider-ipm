package actionservice

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"terraform-provider-ipm/internal/ipm_pf"
	common "terraform-provider-ipm/internal/provider/internal/common"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ActionsResource{}
	_ resource.ResourceWithConfigure   = &ActionsResource{}
	_ resource.ResourceWithImportState = &ActionsResource{}
)

// NewActionsResource is a helper function to simplify the provider implementation.
func NewActionsResource() resource.Resource {
	return &ActionsResource{}
}

type ActionsResource struct {
	client *ipm_pf.Client
}

type ResourceAction struct {
	Type           types.String  `tfsdk:"type"`
	Identifier    common.ResourceIdentifier `tfsdk:"identifier"`
	Action         types.String  `tfsdk:"action"`
	Parameter      types.String `tfsdk:"paremeter"`
	RawAction      types.Bool   `tfsdk:"raw_action"`
	Response       types.String `tfsdk:"response"`
	DelayBeforeApply types.Int64   `tfsdk:"delay_before_apply"`
	StopAtError    types.Bool  `tfsdk:"stop_at_error"`
}

type ActionsResourceData struct {
	Id             types.String  `tfsdk:"id"`
	ResourceActions   []ResourceAction  `tfsdk:"resource_actions"`
}

// Metadata returns the data source type name.
func (r *ActionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_actions"
}

// Schema defines the schema for the data source.
func (r *ActionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	//type ActionsResourceData struct
	resp.Schema = schema.Schema{
		Description: "Manages Action",
		Attributes:  actionsResourceSchemaAttributes(),
	}
}

// Configure adds the provider configured client to the data source.
func (r *ActionsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*ipm_pf.Client)
}

func (r ActionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ActionsResourceData

	diags := req.Config.Get(ctx, &data)

	tflog.Debug(ctx, "ActionsResource: Create - ", map[string]interface{}{"ActionsResourceData": data})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.apply(&data, ctx, &resp.Diagnostics)
	resp.State.Set(ctx, &data)

}

func (r ActionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ActionsResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ActionsResource: Read - ", map[string]interface{}{"ActionsResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ActionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ActionsResourceData

	diags := req.Plan.Get(ctx, &data)
	tflog.Debug(ctx, "ActionsResource: Update", map[string]interface{}{"ActionsResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.apply(&data, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r ActionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ActionsResourceData

	diags := req.State.Get(ctx, &data)

	tflog.Debug(ctx, "ActionsResource: Delete", map[string]interface{}{"ActionsResourceData": data})

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ActionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve Action ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func applyActions(ctx context.Context, client *ipm_pf.Client, resourceActions []ResourceAction) (err error) {
		for _, resourecAction := range resourceActions {
			command, err := getActionCommand(resourecAction)
			tflog.Debug(ctx, "applyActions ###########: ", map[string]interface{}{"command": command})
			if err != nil {
				_, err := client.ExecuteIPMHttpCommand("POST", command, nil)
				if err == nil {
					resourecAction.Response = types.StringValue("Failed. " + err.Error())
					if  !resourecAction.StopAtError.IsNull() && resourecAction.StopAtError.ValueBool() == true {
						break
					}
				} else {
					resourecAction.Response = types.StringValue("Send.")
				}
			} 
			if  !resourecAction.DelayBeforeApply.IsNull() {
				time.Sleep(time.Duration(resourecAction.DelayBeforeApply.ValueInt64()) * time.Second)
			}
		}
		return nil
}


func (r *ActionsResource) apply(plan *ActionsResourceData, ctx context.Context, diags *diag.Diagnostics) {

	tflog.Debug(ctx, "ActionsResource: Apply ## ", map[string]interface{}{"plan": plan})

	if len(plan.ResourceActions) == 0  {
		diags.AddError(
			"ActionsResource: create ##: Error Create ActionsResource",
			"Create: Could not create No Resource's action is specified",
		)
		return
	}

	applyActions(ctx,  r.client, plan.ResourceActions)

	if plan.Id.IsNull() {
		plan.Id = types.StringValue((uuid.New()).String())
	}

	tflog.Debug(ctx, "ActionsResource: create ##", map[string]interface{}{"plan": plan})
}


func (r *ActionsResource) read(state *ActionsResourceData, ctx context.Context, diags *diag.Diagnostics) {

	if len(state.ResourceActions) == 0  {
		diags.AddError(
			"ActionsResource: read ##: Error Read ActionsResource",
			"Read: Could not create No Resource's action is specified",
		)
		return
	}

	tflog.Debug(ctx, "ActionsResource: read ## ", map[string]interface{}{"plan": state})
}

func getActionCommand(resourceAction ResourceAction) (string, error) {
	if  resourceAction.RawAction.IsNull() || !resourceAction.RawAction.ValueBool() {
		resourceType := resourceAction.Type.ValueString()
		deviceId := resourceAction.Identifier.DeviceId.ValueString()
		action := resourceAction.Action.ValueString()
		if len(action) == 0 {
			return "", errors.New("Action is not specified.")
		}
		switch resourceType {
			case "NDU": 
				// '/ndus/{deviceId}/coldStart|warmStart|factoryReset|retry|adopt'
					return "/ndus/" + deviceId + "/" + action, nil;
			case "NDU Port": 
				// '/ndus/{nduId}/ports/{nduPortColId}/retry|adopt'
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.ParentColId.ValueString() + "/" + action, nil;
			case "NDU TOM": 
				// '/ndus/{nduId}/ports/{nduPortColId}/tom/{nduTomColId}/retry|adopt'
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.ParentColId.ValueString() + "/tom/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU XR": 
				// '/ndus/{nduId}/ports/{nduPortColId}/xr/{nduXrColId}/coldStart|warmStart|factoryReset|retry|adopt'
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.ParentColId.ValueString() + "/xr/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU EDFA": 
				// '/ndus/{nduId}/ports/{nduPortColId}/edfa/{nduEdfaColId}/retry|adopt'
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.ParentColId.ValueString() + "/edfa/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU VOA": 
				// '/ndus/{nduId}/ports/{nduPortColId}/voa/{nduVoaColId}/retry|adopt'
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.ParentColId.ValueString() + "/voa/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU Line PTP": 
				// '/ndus/{nduId}/ports/{nduPortColId}/linePtps/{nduLinePtpColId}/retry|adopt'
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.ParentColId.ValueString() + "/linePtps/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU Trib PTP": 
				// '/ndus/{nduId}/ports/{nduPortColId}/tripPtps/{nduTribPtpColId}/retry|adopt'
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.ParentColId.ValueString() + "/tribPtps/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU POL PTP": 
				// '/ndus/{nduId}/ports/{nduPortColId}/tom/{nduPolPtpColId}/retry|adopt'
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.ParentColId.ValueString() + "/polPtps/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU Carrier": 
				// '/ndus/{nduId}/ports/{nduPortColId}/linePtps/{nduLinePtpColId}/carrier/{carrierColId}/retry|adopt
				return "/ndus/" + deviceId  + "/ports/" + resourceAction.Identifier.GrandParentColId.ValueString() + "/linePtps/" + resourceAction.Identifier.ParentColId.ValueString() + "/carrier/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU OTU" :
				// '/ndus/{deviceId}/otus/{otuColId}/retry'
				return "/ndus/" + deviceId  + "/otus/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "NDU Ethernet Client":
				// '/modules/{deviceId}/ethernetClients/{ethernetColId}/clrLldpStats|flushLldpHostDb|retry';
				return "/ndus/" + deviceId  + "/ethernets/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "Module": 
				// '/modules/{deviceId}/coldStart|warmStart|factoryReset|retry'
					return "/modules/" + deviceId + "/" + action, nil;
			case "Ethernet Client":
				// '/modules/{deviceId}/ethernetClients/{ethernetColId}/clrLldpStats|flushLldpHostDb|retry';
				return "/modules/" + deviceId  + "/ethernetClients/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "OTU" :
				// '/modules/{deviceId}/otus/{otuColId}/retry'
				return "/modules/" + deviceId  + "/otus/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "ODU" :
				// '/modules/{deviceId}/otus/{otuColId}/odus/{oduColId}/retry'
				return "/modules/" + deviceId  + "/otus/" + resourceAction.Identifier.ParentColId.ValueString() + "/odus/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "Carrier" :
				// '/modules/{deviceId}/linePtps/{linePtpColId}/carriers/{carrierColId}/retry'
				return "/modules/" + deviceId  + "/linePtps/" + resourceAction.Identifier.ParentColId.ValueString() + "/carriers/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "DSC":
				// '/modules/{deviceId}/linePtps/{linePtpColId}/carriers/{carrierColId}/dscs/{dscColId}/retry'
				return "/modules/" + deviceId  + "/linePtps/" + resourceAction.Identifier.GrandParentColId.ValueString() + "/carriers/" + resourceAction.Identifier.ParentColId.ValueString() + "/dscs/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
			case "DSCG":
				// '/modules/{deviceId}/linePtps/{linePtpColId}/carriers/{carrierColId}/dsgs/{dsgColId}/retry'
				return "/modules/" + deviceId  + "/linePtps/" + resourceAction.Identifier.GrandParentColId.ValueString() + "/carriers/" + resourceAction.Identifier.ParentColId.ValueString() + "/dsgs/" + resourceAction.Identifier.ColId.ValueString() + "/" + action, nil;
	        default:
				return "", errors.New("Invalid Resource Type: " + resourceType)
		}
	} else {
		return resourceAction.Action.ValueString(), nil
	}
}

func actionsResourceSchemaAttributes() map[string]schema.Attribute { 
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of the Module.",
			Computed:    true,
		},
		"resource_actions":schema.ListNestedAttribute{
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "type",
						Optional:     true,
					},
					"identifier": common.ResourceIdentifierAttribute(),
					"action": schema.StringAttribute{
						Description: "Action",
						Optional:     true,
					},
					"paremeter": schema.StringAttribute{
						Description: "paremeter",
						Optional:     true,
					},
					"response": schema.StringAttribute{
						Description: "Response",
						Computed:     true,
					},
					"raw_action": schema.BoolAttribute{
						Description: "raw_action",
						Optional:     true,
					},
					"delay_before_apply": schema.Int64Attribute{
						Description: "delay_before_apply",
						Optional:     true,
					},
					"stop_at_error": schema.BoolAttribute{
						Description: "stop apply at first error.",
						Optional:    true,
					},
				},
			},
		},
	}
}