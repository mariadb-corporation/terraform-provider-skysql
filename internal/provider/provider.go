package provider

import (
	"context"
	"errors"
	"os"

	"github.com/mariadb-corporation/terraform-provider-skysql/internal/skysql"
	"github.com/matryer/resync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure skySQLProvider satisfies various provider interfaces.
var _ provider.Provider = &skySQLProvider{}

var configureOnce resync.Once

// skySQLProvider defines the provider implementation.
type skySQLProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SkySQLProviderModel describes the provider data model.
type SkySQLProviderModel struct {
	BaseURL     types.String `tfsdk:"base_url"`
	AccessToken types.String `tfsdk:"access_token"`
}

func (p *skySQLProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "skysql"
	resp.Version = p.version
}

func (p *skySQLProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The SkySQL terraform provider",
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				MarkdownDescription: "SkySQL API access token",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Function to read environment with a default value
func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func (p *skySQLProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	accessToken := os.Getenv("TF_SKYSQL_API_ACCESS_TOKEN")
	baseURL := getEnv("TF_SKYSQL_API_BASE_URL", "https://api.mariadb.com")

	var data SkySQLProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if data.AccessToken.ValueString() != "" {
		accessToken = data.AccessToken.ValueString()
	}

	if data.BaseURL.ValueString() != "" {
		baseURL = data.BaseURL.ValueString()
	}

	if accessToken == "" {
		resp.Diagnostics.AddError(
			"Missing SkySQL Access Token Configuration",
			"While configuring the provider, the API access token was not found in "+
				"the TF_SKYSQL_API_ACCESS_TOKEN environment variable or provider "+
				"configuration block access_token attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	if baseURL == "" {
		resp.Diagnostics.AddError(
			"Missing Endpoint Configuration",
			"While configuring the provider, the endpoint was not found in "+
				"the TF_SKYSQL_API_BASE_URL environment variable or provider "+
				"configuration block base_url attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	client := skysql.New(baseURL, accessToken)

	configureOnce.Do(func() {
		_, err := client.GetVersions(ctx, skysql.WithPageSize(1))
		if err != nil {
			if errors.Is(err, skysql.ErrorUnauthorized) {
				resp.Diagnostics.AddError(
					"Unable to connect to SkySQL",
					"While configuring the provider, the API access token was not valid.",
				)
				return
			}
			resp.Diagnostics.AddError(
				"Unable to connect to SkySQL",
				"While configuring the provider, the API returns error: "+err.Error(),
			)
		}
	})

	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *skySQLProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServiceResource,
		NewServiceAllowListResource,
	}
}

func (p *skySQLProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectsDataSource,
		NewVersionsDataSource,
		NewServiceDataSource,
		NewCredentialsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &skySQLProvider{
			version: version,
		}
	}
}
