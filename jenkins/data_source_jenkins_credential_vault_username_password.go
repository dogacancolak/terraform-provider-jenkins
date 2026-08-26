package jenkins

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type credentialVaultUsernamePasswordDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Folder        types.String `tfsdk:"folder"`
	Description   types.String `tfsdk:"description"`
	Domain        types.String `tfsdk:"domain"`
	Scope         types.String `tfsdk:"scope"`
	Namespace     types.String `tfsdk:"namespace"`
	PrefixPath    types.String `tfsdk:"prefix_path"`
	Path          types.String `tfsdk:"path"`
	EngineVersion types.Int64  `tfsdk:"engine_version"`
	UsernameKey   types.String `tfsdk:"username_key"`
	PasswordKey   types.String `tfsdk:"password_key"`
}

type credentialVaultUsernamePasswordDataSource struct {
	*dataSourceHelper
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &credentialVaultUsernamePasswordDataSource{}

func newCredentialVaultUsernamePasswordDataSource() datasource.DataSource {
	return &credentialVaultUsernamePasswordDataSource{
		dataSourceHelper: newDataSourceHelper(),
	}
}

func (d *credentialVaultUsernamePasswordDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential_vault_username_password"
}

func (d *credentialVaultUsernamePasswordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get the attributes of a vault username-password credential within Jenkins.",
		Attributes: d.schemaCredential(map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The Vault namespace to read the secret from.",
				Computed:            true,
			},
			"prefix_path": schema.StringAttribute{
				MarkdownDescription: "The Vault mount prefix path prepended to `path`.",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The Vault path of the secret to read.",
				Computed:            true,
			},
			"engine_version": schema.Int64Attribute{
				MarkdownDescription: "The KV engine version of the Vault secrets engine.",
				Computed:            true,
			},
			"username_key": schema.StringAttribute{
				MarkdownDescription: "The key within the Vault secret the username is read from.",
				Computed:            true,
			},
			"password_key": schema.StringAttribute{
				MarkdownDescription: "The key within the Vault secret the password is read from.",
				Computed:            true,
			},
		}),
	}
}

func (d *credentialVaultUsernamePasswordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data credentialVaultUsernamePasswordDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm := d.client.Credentials()
	cm.Folder = formatFolderName(data.Folder.ValueString())

	if data.Domain.IsNull() {
		data.Domain = basetypes.NewStringValue(defaultCredentialDomain)
	}

	cred := VaultUsernamePasswordCredentials{}
	err := cm.GetSingle(ctx, data.Domain.ValueString(), data.Name.ValueString(), &cred)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while parsing the data source read response. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}

	data.ID = types.StringValue(generateCredentialID(data.Folder.ValueString(), cred.ID))
	data.Scope = types.StringValue(cred.Scope)
	data.Description = types.StringValue(cred.Description)
	data.Namespace = types.StringValue(cred.Namespace)
	data.PrefixPath = types.StringValue(cred.PrefixPath)
	data.Path = types.StringValue(cred.Path)
	data.EngineVersion = types.Int64Value(cred.EngineVersion)
	data.UsernameKey = types.StringValue(cred.UsernameKey)
	data.PasswordKey = types.StringValue(cred.PasswordKey)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
