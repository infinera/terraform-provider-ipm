package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IpmError struct {
	Code    types.String `tfsdk:"code"`
	Message types.String `tfsdk:"message"`
}

type LifecycleStateCause struct {
	Action    types.Int64  `tfsdk:"action"`
	Timestamp types.String `tfsdk:"timestamp"`
	TraceId   types.String `tfsdk:"trace_id"`
	Errors    []IpmError   `tfsdk:"errors"`
}
