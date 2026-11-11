package mdb_redis_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// valkeySearchValidator validates that reader_threads and writer_threads are only set when enabled is true
type valkeySearchValidator struct{}

// Description returns a human-readable description of the validator
func (v valkeySearchValidator) Description(_ context.Context) string {
	return "Validates that reader_threads and writer_threads are only set when enabled is true"
}

// MarkdownDescription returns a markdown description of the validator
func (v valkeySearchValidator) MarkdownDescription(_ context.Context) string {
	return "Validates that reader_threads and writer_threads are only set when enabled is true"
}

// Validate performs the validation
func (v valkeySearchValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// If the attribute is not configured, skip validation
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Convert the object to our model
	var valkeySearch ValkeySearch
	diags := req.ConfigValue.As(ctx, &valkeySearch, baseOptions)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// If enabled is false, reader_threads and writer_threads should not be set
	if !valkeySearch.Enabled.ValueBool() {
		if !valkeySearch.ReaderThreads.IsNull() && !valkeySearch.ReaderThreads.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("reader_threads"),
				"Invalid valkey_search configuration",
				"reader_threads can only be set when valkey_search.enabled is true",
			)
		}
		if !valkeySearch.WriterThreads.IsNull() && !valkeySearch.WriterThreads.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("writer_threads"),
				"Invalid valkey_search configuration",
				"writer_threads can only be set when valkey_search.enabled is true",
			)
		}
	}
}

// ValkeySearchValidator returns a new validator for the valkey_search block
func ValkeySearchValidator() validator.Object {
	return valkeySearchValidator{}
}
