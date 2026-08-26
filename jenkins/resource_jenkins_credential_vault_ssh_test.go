package jenkins

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccJenkinsCredentialVaultSSH_basic(t *testing.T) {
	var cred VaultSSHCredentials

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy:             testAccCheckJenkinsCredentialVaultSSHDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource jenkins_credential_vault_ssh foo {
				  name = "test-vault-ssh"
				  path = "secret/data/foo"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "id", "/test-vault-ssh"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "username_key", "username"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "private_key_key", "private_key"),
					testAccCheckJenkinsCredentialVaultSSHExists("jenkins_credential_vault_ssh.foo", &cred),
				),
			},
			{
				// Update by adding description and custom keys
				Config: `
				resource jenkins_credential_vault_ssh foo {
				  name = "test-vault-ssh"
				  description = "new-description"
				  path = "secret/data/foo"
				  username_key = "user"
				  private_key_key = "key"
				  passphrase_key = "phrase"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJenkinsCredentialVaultSSHExists("jenkins_credential_vault_ssh.foo", &cred),
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "description", "new-description"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "username_key", "user"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "private_key_key", "key"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "passphrase_key", "phrase"),
				),
			},
		},
	})
}

func testAccCheckJenkinsCredentialVaultSSHExists(resourceName string, cred *VaultSSHCredentials) resource.TestCheckFunc {
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

func testAccCheckJenkinsCredentialVaultSSHDestroy(s *terraform.State) error {
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jenkins_credential_vault_ssh" {
			continue
		} else if _, ok := rs.Primary.Meta["name"]; !ok {
			continue
		}

		cred := VaultSSHCredentials{}
		manager := testAccClient.Credentials()
		manager.Folder = formatFolderName(rs.Primary.Meta["folder"].(string))
		err := manager.GetSingle(ctx, rs.Primary.Meta["domain"].(string), rs.Primary.Meta["name"].(string), &cred)
		if err == nil {
			return fmt.Errorf("Credentials still exists: %s - %s", rs.Primary.Attributes["folder"], rs.Primary.Attributes["name"])
		}
	}

	return nil
}
