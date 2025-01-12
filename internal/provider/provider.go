// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	exampledatasource "github.com/digidaniel-tech/terraform-provider-truenas-scale/internal/data_sources/example_data_source"
	examplefunction "github.com/digidaniel-tech/terraform-provider-truenas-scale/internal/functions/example_function"
	"github.com/digidaniel-tech/terraform-provider-truenas-scale/internal/resources/app_resource"
)

var _ provider.Provider = &TruenasScaleProvider{}
var _ provider.ProviderWithFunctions = &TruenasScaleProvider{}

type TruenasScaleProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type TruenasScaleProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *TruenasScaleProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription:    "Host to Truenas",
                Required:               true,
			},
            "username": schema.StringAttribute{
                MarkdownDescription:    "Username used to authenticate towards truenas",
                Required:               true,
            },
            "password": schema.StringAttribute{
                MarkdownDescription:    "Password used to authenticate towards truenas",
                Required:               true,
            },
		},
	}
}

func (p *TruenasScaleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TruenasScaleProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TruenasScaleProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		exampledatasource.NewExampleDataSource,
	}
}

func (p *TruenasScaleProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		examplefunction.NewExampleFunction,
	}
}

func (p *TruenasScaleProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		appresource.NewAppResource,
	}
}

func (p *TruenasScaleProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scaffolding"
	resp.Version = p.version
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TruenasScaleProvider{
			version: version,
		}
	}
}
