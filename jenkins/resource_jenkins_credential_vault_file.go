package jenkins

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VaultFileCredentials struct representing a Vault secret file credential, where the file
// content is resolved at runtime from the referenced Vault path (optionally a single key)
// rather than stored within Jenkins.
type VaultFileCredentials struct {
	XMLName       xml.Name `xml:"com.datapipe.jenkins.vault.credentials.common.VaultFileCredentialImpl"`
	ID            string   `xml:"id"`
	Scope         string   `xml:"scope"`
	Description   string   `xml:"description"`
	Namespace     string   `xml:"namespace"`
	PrefixPath    string   `xml:"prefixPath"`
	Path          string   `xml:"path"`
	EngineVersion int64    `xml:"engineVersion"`
	FileName      string   `xml:"fileName"`
	UseKey        bool     `xml:"useKey"`
	VaultKey      string   `xml:"vaultKey"`
}

type credentialVaultFileResourceModel struct {
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
	FileName      types.String `tfsdk:"file_name"`
	UseKey        types.Bool   `tfsdk:"use_key"`
	VaultKey      types.String `tfsdk:"vault_key"`
}

type credentialVaultFileResource struct {
	*resourceHelper
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &credentialVaultFileResource{}

func newCredentialVaultFileResource() resource.Resource {
	return &credentialVaultFileResource{
		resourceHelper: newResourceHelper(),
	}
}

// Metadata should return the full name of the resource.
func (r *credentialVaultFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential_vault_file"
}

// Schema should return the schema for this resource.
func (r *credentialVaultFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages a Vault secret file credential within Jenkins. The file content is resolved from Vault at runtime, so no secret material is stored within Jenkins or in Terraform state.

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
				MarkdownDescription: "The KV engine version of the Vault secrets engine. Must be either `1` or `2`. Defaults to `2`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2),
			},
			"file_name": schema.StringAttribute{
				MarkdownDescription: "The file name presented to the job for the resolved secret. If not set, the Vault plugin generates a random name.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"use_key": schema.BoolAttribute{
				MarkdownDescription: "When `true`, the file content is the value of `vault_key` within the secret. When `false`, the file content is the whole secret serialized as JSON. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"vault_key": schema.StringAttribute{
				MarkdownDescription: "The key within the Vault secret to read the file content from. Only used when `use_key` is `true`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
		}),
	}
}

// Create is called when the provider must create a new resource. Config
// and planned state values should be read from the
// CreateRequest and new state values set on the CreateResponse.
func (r *credentialVaultFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data credentialVaultFileResourceModel

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

	cred := VaultFileCredentials{
		ID:            data.Name.ValueString(),
		Scope:         data.Scope.ValueString(),
		Description:   data.Description.ValueString(),
		Namespace:     data.Namespace.ValueString(),
		PrefixPath:    data.PrefixPath.ValueString(),
		Path:          data.Path.ValueString(),
		EngineVersion: data.EngineVersion.ValueInt64(),
		FileName:      data.FileName.ValueString(),
		UseKey:        data.UseKey.ValueBool(),
		VaultKey:      data.VaultKey.ValueString(),
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

	// The Vault plugin generates a random file name if none is supplied; read it back so
	// state reflects the value Jenkins persisted rather than an empty string.
	if data.FileName.IsUnknown() || data.FileName.ValueString() == "" {
		stored := VaultFileCredentials{}
		if err := cm.GetSingle(ctx, data.Domain.ValueString(), data.Name.ValueString(), &stored); err == nil {
			data.FileName = types.StringValue(stored.FileName)
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is called when the provider must read resource values in order
// to update state. Planned state values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (r *credentialVaultFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data credentialVaultFileResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm := r.client.Credentials()
	cm.Folder = formatFolderName(data.Folder.ValueString())

	cred := VaultFileCredentials{}
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
	data.FileName = types.StringValue(cred.FileName)
	data.UseKey = types.BoolValue(cred.UseKey)
	data.VaultKey = types.StringValue(cred.VaultKey)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource. Config, planned
// state, and prior state values should be read from the
// UpdateRequest and new state values set on the UpdateResponse.
func (r *credentialVaultFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data credentialVaultFileResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm := r.client.Credentials()
	cm.Folder = formatFolderName(data.Folder.ValueString())

	cred := VaultFileCredentials{
		ID:            data.Name.ValueString(),
		Scope:         data.Scope.ValueString(),
		Description:   data.Description.ValueString(),
		Namespace:     data.Namespace.ValueString(),
		PrefixPath:    data.PrefixPath.ValueString(),
		Path:          data.Path.ValueString(),
		EngineVersion: data.EngineVersion.ValueInt64(),
		FileName:      data.FileName.ValueString(),
		UseKey:        data.UseKey.ValueBool(),
		VaultKey:      data.VaultKey.ValueString(),
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
func (r *credentialVaultFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data credentialVaultFileResourceModel

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
