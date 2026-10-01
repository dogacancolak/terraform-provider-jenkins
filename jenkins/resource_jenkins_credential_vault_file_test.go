package jenkins

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccJenkinsCredentialVaultFile_basic(t *testing.T) {
	var cred VaultFileCredentials

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy:             testAccCheckJenkinsCredentialVaultFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource jenkins_credential_vault_file foo {
				  name = "test-vault-file"
				  path = "secret/data/foo"
				  file_name = "secret.txt"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "id", "/test-vault-file"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "use_key", "false"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "file_name", "secret.txt"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "engine_version", "2"),
					testAccCheckJenkinsCredentialVaultFileExists("jenkins_credential_vault_file.foo", &cred),
				),
			},
			{
				// Update by adding description and switching to use_key
				Config: `
				resource jenkins_credential_vault_file foo {
				  name = "test-vault-file"
				  description = "new-description"
				  path = "secret/data/foo"
				  file_name = "renamed.txt"
				  use_key = true
				  vault_key = "private-key"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJenkinsCredentialVaultFileExists("jenkins_credential_vault_file.foo", &cred),
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "description", "new-description"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "file_name", "renamed.txt"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "use_key", "true"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "vault_key", "private-key"),
				),
			},
		},
	})
}

func testAccCheckJenkinsCredentialVaultFileExists(resourceName string, cred *VaultFileCredentials) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := context.Background()

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return errors.New(resourceName + " not found")
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		manager := testAccClient.Credentials()
		manager.Folder = formatFolderName(rs.Primary.Attributes["folder"])
		err := manager.GetSingle(ctx, rs.Primary.Attributes["domain"], rs.Primary.Attributes["name"], cred)
		if err != nil {
			return fmt.Errorf("Unable to retrieve credentials for %s - %s: %w", rs.Primary.Attributes["folder"], rs.Primary.Attributes["name"], err)
		}

		return nil
	}
}

func testAccCheckJenkinsCredentialVaultFileDestroy(s *terraform.State) error {
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jenkins_credential_vault_file" {
			continue
		} else if _, ok := rs.Primary.Meta["name"]; !ok {
			continue
		}

		cred := VaultFileCredentials{}
		manager := testAccClient.Credentials()
		manager.Folder = formatFolderName(rs.Primary.Meta["folder"].(string))
		err := manager.GetSingle(ctx, rs.Primary.Meta["domain"].(string), rs.Primary.Meta["name"].(string), &cred)
		if err == nil {
			return fmt.Errorf("Credentials still exists: %s - %s", rs.Primary.Attributes["folder"], rs.Primary.Attributes["name"])
		}
	}

	return nil
}
