package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mariadb-corporation/terraform-provider-skysql-v2/internal/skysql"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &VersionsDataSource{}

func NewVersionsDataSource() datasource.DataSource {
	return &VersionsDataSource{}
}

// VersionsDataSource defines the data source implementation.
type VersionsDataSource struct {
	client *skysql.Client
}

type versionDataSourceDataSourceModel struct {
	Versions []versionModel `tfsdk:"versions"`
}

type versionModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Version         types.String `tfsdk:"version"`
	Topology        types.String `tfsdk:"topology"`
	Product         types.String `tfsdk:"product"`
	DisplayName     types.String `tfsdk:"display_name"`
	IsMajor         types.Bool   `tfsdk:"is_major"`
	ReleaseDate     types.String `tfsdk:"release_date"`
	ReleaseNotesUrl types.String `tfsdk:"release_notes_url"`
}

func (d *VersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_versions"
}

func (d *VersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"versions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"version": schema.StringAttribute{
							Computed: true,
						},
						"topology": schema.StringAttribute{
							Computed: true,
						},
						"product": schema.StringAttribute{
							Computed: true,
						},
						"display_name": schema.StringAttribute{
							Computed: true,
						},
						"is_major": schema.BoolAttribute{
							Computed: true,
						},
						"release_date": schema.StringAttribute{
							Computed: true,
						},
						"release_notes_url": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *VersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*skysql.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *VersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state versionDataSourceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	versions, err := d.client.GetVersions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SkySQL versions", err.Error())
		return
	}

	for _, version := range versions {
		versionState := versionModel{
			Id:              types.StringValue(version.Id),
			Name:            types.StringValue(version.Name),
			Version:         types.StringValue(version.Version),
			Topology:        types.StringValue(version.Topology),
			Product:         types.StringValue(version.Product),
			DisplayName:     types.StringValue(version.DisplayName),
			IsMajor:         types.BoolValue(version.IsMajor),
			ReleaseDate:     types.StringValue(version.ReleaseDate.String()),
			ReleaseNotesUrl: types.StringValue(version.ReleaseNotesUrl),
		}
		state.Versions = append(state.Versions, versionState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
