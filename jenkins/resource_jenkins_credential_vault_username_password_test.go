package jenkins

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccJenkinsCredentialVaultUsernamePassword_basic(t *testing.T) {
	var cred VaultUsernamePasswordCredentials

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy:             testAccCheckJenkinsCredentialVaultUsernamePasswordDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource jenkins_credential_vault_username_password foo {
				  name = "test-vault-username-password"
				  path = "secret/data/foo"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "id", "/test-vault-username-password"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "username_key", "username"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "password_key", "password"),
					testAccCheckJenkinsCredentialVaultUsernamePasswordExists("jenkins_credential_vault_username_password.foo", &cred),
				),
			},
			{
				// Update by adding description and custom keys
				Config: `
				resource jenkins_credential_vault_username_password foo {
				  name = "test-vault-username-password"
				  description = "new-description"
				  path = "secret/data/foo"
				  username_key = "user"
				  password_key = "pass"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJenkinsCredentialVaultUsernamePasswordExists("jenkins_credential_vault_username_password.foo", &cred),
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "description", "new-description"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "username_key", "user"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "password_key", "pass"),
				),
			},
		},
	})
}

func testAccCheckJenkinsCredentialVaultUsernamePasswordExists(resourceName string, cred *VaultUsernamePasswordCredentials) resource.TestCheckFunc {
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

func testAccCheckJenkinsCredentialVaultUsernamePasswordDestroy(s *terraform.State) error {
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jenkins_credential_vault_username_password" {
			continue
		} else if _, ok := rs.Primary.Meta["name"]; !ok {
			continue
		}

		cred := VaultUsernamePasswordCredentials{}
		manager := testAccClient.Credentials()
		manager.Folder = formatFolderName(rs.Primary.Meta["folder"].(string))
		err := manager.GetSingle(ctx, rs.Primary.Meta["domain"].(string), rs.Primary.Meta["name"].(string), &cred)
		if err == nil {
			return fmt.Errorf("Credentials still exists: %s - %s", rs.Primary.Attributes["folder"], rs.Primary.Attributes["name"])
		}
	}

	return nil
}
