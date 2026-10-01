package jenkins

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VaultUsernamePasswordCredentials struct representing a Vault username-password credential,
// where the username and password are resolved at runtime from the referenced Vault path/keys
// rather than stored within Jenkins.
type VaultUsernamePasswordCredentials struct {
	XMLName       xml.Name `xml:"com.datapipe.jenkins.vault.credentials.common.VaultUsernamePasswordCredentialImpl"`
	ID            string   `xml:"id"`
	Scope         string   `xml:"scope"`
	Description   string   `xml:"description"`
	Namespace     string   `xml:"namespace"`
	PrefixPath    string   `xml:"prefixPath"`
	Path          string   `xml:"path"`
	EngineVersion int64    `xml:"engineVersion"`
	UsernameKey   string   `xml:"usernameKey"`
	PasswordKey   string   `xml:"passwordKey"`
}

type credentialVaultUsernamePasswordResourceModel struct {
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

type credentialVaultUsernamePasswordResource struct {
	*resourceHelper
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &credentialVaultUsernamePasswordResource{}

func newCredentialVaultUsernamePasswordResource() resource.Resource {
	return &credentialVaultUsernamePasswordResource{
		resourceHelper: newResourceHelper(),
	}
}

// Metadata should return the full name of the resource.
func (r *credentialVaultUsernamePasswordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential_vault_username_password"
}

// Schema should return the schema for this resource.
func (r *credentialVaultUsernamePasswordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages a Vault username-password credential within Jenkins. The username and password are resolved from Vault at runtime, so no secret material is stored within Jenkins or in Terraform state.

~> The Jenkins installation that uses this resource is expected to have the [Hashicorp Vault Plugin](https://plugins.jenkins.io/hashicorp-vault-plugin/) installed in their system.`,
		Attributes: r.schemaCredential(map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The Vault namespace to read the secret from.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"prefix_path": schema.StringAttribute{
				MarkdownDescription: "The Vault mount prefix path to prepend to `path`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The Vault path of the secret to read.",
				Required:            true,
			},
			"engine_version": schema.Int64Attribute{
				MarkdownDescription: "The KV engine version of the Vault secrets engine. Must be either `1` or `2`. Defaults to `2`, which pins the credential rather than inheriting the Jenkins global or folder Vault configuration.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2),
				Validators: []validator.Int64{
					int64validator.OneOf(1, 2),
				},
			},
			"username_key": schema.StringAttribute{
				MarkdownDescription: "The key within the Vault secret to read the username from. Defaults to `username`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("username"),
			},
			"password_key": schema.StringAttribute{
				MarkdownDescription: "The key within the Vault secret to read the password from. Defaults to `password`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("password"),
			},
		}),
	}
}

// Create is called when the provider must create a new resource. Config
// and planned state values should be read from the
// CreateRequest and new state values set on the CreateResponse.
func (r *credentialVaultUsernamePasswordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data credentialVaultUsernamePasswordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm := r.client.Credentials()
	cm.Folder = formatFolderName(data.Folder.ValueString())

	// Validate that the folder exists
	if err := folderExists(ctx, r.client, cm.Folder); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Folder",
			fmt.Sprintf("An invalid folder name %q was specified. ", cm.Folder)+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}

	cred := VaultUsernamePasswordCredentials{
		ID:            data.Name.ValueString(),
		Scope:         data.Scope.ValueString(),
		Description:   data.Description.ValueString(),
		Namespace:     data.Namespace.ValueString(),
		PrefixPath:    data.PrefixPath.ValueString(),
		Path:          data.Path.ValueString(),
		EngineVersion: data.EngineVersion.ValueInt64(),
		UsernameKey:   data.UsernameKey.ValueString(),
		PasswordKey:   data.PasswordKey.ValueString(),
	}

	err := cm.Add(ctx, data.Domain.ValueString(), cred)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while creating the resource. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.ID = types.StringValue(generateCredentialID(data.Folder.ValueString(), cred.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is called when the provider must read resource values in order
// to update state. Planned state values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (r *credentialVaultUsernamePasswordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data credentialVaultUsernamePasswordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm := r.client.Credentials()
	cm.Folder = formatFolderName(data.Folder.ValueString())

	cred := VaultUsernamePasswordCredentials{}
	err := cm.GetSingle(ctx, data.Domain.ValueString(), data.Name.ValueString(), &cred)
	if err != nil {
		if strings.HasSuffix(err.Error(), "404") {
			// Job does not exist
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			"An unexpected error occurred while parsing the resource read response. "+
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource. Config, planned
// state, and prior state values should be read from the
// UpdateRequest and new state values set on the UpdateResponse.
func (r *credentialVaultUsernamePasswordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data credentialVaultUsernamePasswordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm := r.client.Credentials()
	cm.Folder = formatFolderName(data.Folder.ValueString())

	cred := VaultUsernamePasswordCredentials{
		ID:            data.Name.ValueString(),
		Scope:         data.Scope.ValueString(),
		Description:   data.Description.ValueString(),
		Namespace:     data.Namespace.ValueString(),
		PrefixPath:    data.PrefixPath.ValueString(),
		Path:          data.Path.ValueString(),
		EngineVersion: data.EngineVersion.ValueInt64(),
		UsernameKey:   data.UsernameKey.ValueString(),
		PasswordKey:   data.PasswordKey.ValueString(),
	}

	err := cm.Update(ctx, data.Domain.ValueString(), data.Name.ValueString(), &cred)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete is called when the provider must delete the resource. Config
// values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically
// call DeleteResponse.State.RemoveResource(), so it can be omitted
// from provider logic.
func (r *credentialVaultUsernamePasswordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data credentialVaultUsernamePasswordResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	cm := r.client.Credentials()
	cm.Folder = formatFolderName(data.Folder.ValueString())

	err := cm.Delete(ctx, data.Domain.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while deleting the resource. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return
	}
}
