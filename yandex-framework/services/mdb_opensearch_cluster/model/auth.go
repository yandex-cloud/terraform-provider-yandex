package model

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type AuthSettings struct {
	SAML types.Object `tfsdk:"saml"`
}

type SAML struct {
	Enabled                types.Bool   `tfsdk:"enabled"`
	IdpEntityID            types.String `tfsdk:"idp_entity_id"`
	IdpMetadataFileContent types.String `tfsdk:"idp_metadata_file_content"`
	SpEntityID             types.String `tfsdk:"sp_entity_id"`
	DashboardsUrl          types.String `tfsdk:"dashboards_url"`
	RolesKey               types.String `tfsdk:"roles_key"`
	SubjectKey             types.String `tfsdk:"subject_key"`
}

var AuthSettingsAttrTypes = map[string]attr.Type{
	"saml": types.ObjectType{AttrTypes: samlAttrTypes},
}

var samlAttrTypes = map[string]attr.Type{
	"enabled":                   types.BoolType,
	"idp_entity_id":             types.StringType,
	"idp_metadata_file_content": types.StringType,
	"sp_entity_id":              types.StringType,
	"dashboards_url":            types.StringType,
	"roles_key":                 types.StringType,
	"subject_key":               types.StringType,
}

func AuthSettingsToState(ctx context.Context, authSettings *opensearch.AuthSettings, state types.Object) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if authSettings == nil {
		return state, diags
	}

	stateAuthSettings, diags := AuthSettingsFromState(ctx, state)
	if diags.HasError() {
		return types.ObjectUnknown(AuthSettingsAttrTypes), diags
	}

	saml, diags := SAMLToState(ctx, authSettings.GetSaml(), stateAuthSettings)
	if diags.HasError() {
		return types.ObjectUnknown(AuthSettingsAttrTypes), diags
	}

	if saml.IsNull() {
		return state, diags
	}

	return types.ObjectValueFrom(ctx, AuthSettingsAttrTypes, AuthSettings{
		SAML: saml,
	})
}

func SAMLToState(ctx context.Context, saml *opensearch.SAMLSettings, state *AuthSettings) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if saml == nil || saml.GetIdpEntityId() == "" {
		return types.ObjectNull(samlAttrTypes), diags
	}

	fileContent := string(saml.GetIdpMetadataFile())

	if state == nil || state.SAML.IsNull() {
		return types.ObjectValueFrom(ctx, samlAttrTypes, SAML{
			Enabled:                types.BoolValue(saml.GetEnabled()),
			IdpEntityID:            types.StringValue(saml.GetIdpEntityId()),
			IdpMetadataFileContent: types.StringValue(fileContent),
			SpEntityID:             types.StringValue(saml.GetSpEntityId()),
			DashboardsUrl:          types.StringValue(saml.GetDashboardsUrl()),
			RolesKey:               defaultIfEmpty(saml.GetRolesKey(), types.StringNull()),
			SubjectKey:             defaultIfEmpty(saml.GetSubjectKey(), types.StringNull()),
		})
	}

	stateSaml, d := SAMLFromState(ctx, state.SAML)
	if d.HasError() {
		return types.ObjectUnknown(samlAttrTypes), d
	}

	rolesKey := defaultIfEmpty(saml.GetRolesKey(), stateSaml.RolesKey)
	subjectKey := defaultIfEmpty(saml.GetSubjectKey(), stateSaml.SubjectKey)

	return types.ObjectValueFrom(ctx, samlAttrTypes, SAML{
		Enabled:                types.BoolValue(saml.GetEnabled()),
		IdpEntityID:            types.StringValue(saml.GetIdpEntityId()),
		IdpMetadataFileContent: types.StringValue(fileContent),
		SpEntityID:             types.StringValue(saml.GetSpEntityId()),
		DashboardsUrl:          types.StringValue(saml.GetDashboardsUrl()),
		RolesKey:               rolesKey,
		SubjectKey:             subjectKey,
	})
}

func AuthSettingsFromState(ctx context.Context, state types.Object) (*AuthSettings, diag.Diagnostics) {
	res := &AuthSettings{}
	diags := state.As(ctx, &res, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diags
}

func SAMLFromState(ctx context.Context, state types.Object) (*SAML, diag.Diagnostics) {
	res := &SAML{}
	diags := state.As(ctx, &res, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diags
}

func defaultIfEmpty(value string, defaultValue basetypes.StringValue) basetypes.StringValue {
	if value == "" {
		return defaultValue
	}

	return types.StringValue(value)
}
