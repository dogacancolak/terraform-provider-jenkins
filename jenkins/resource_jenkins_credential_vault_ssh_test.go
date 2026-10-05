package jenkins

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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
			{
				ResourceName:      "jenkins_credential_vault_ssh.foo",
				ImportState:       true,
				ImportStateId:     defaultCredentialDomain + "/test-vault-ssh",
				ImportStateVerify: true,
				// ImportState writes folder as "" for a global credential where a
				// normal apply leaves it null. Shared by every credential type.
				ImportStateVerifyIgnore: []string{"folder"},
			},
		},
	})
}

func TestAccJenkinsCredentialVaultSSH_folder(t *testing.T) {
	var cred VaultSSHCredentials
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckJenkinsCredentialVaultSSHDestroy,
			testAccCheckJenkinsFolderDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_folder foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance testing"

					lifecycle {
						ignore_changes = [template]
					}
				}

				resource jenkins_folder foo_sub {
					name = "subfolder"
					folder = jenkins_folder.foo.id
					description = "Terraform acceptance testing"

					lifecycle {
						ignore_changes = [template]
					}
				}

				resource jenkins_credential_vault_ssh foo {
				  name = "test-vault-ssh"
				  folder = jenkins_folder.foo_sub.id
				  path = "secret/data/foo"
				}`, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "id", "/job/tf-acc-test-"+randString+"/job/subfolder/test-vault-ssh"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "username_key", "username"),
					testAccCheckJenkinsCredentialVaultSSHExists("jenkins_credential_vault_ssh.foo", &cred),
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
		}

		cred := VaultSSHCredentials{}
		manager := testAccClient.Credentials()
		manager.Folder = formatFolderName(rs.Primary.Attributes["folder"])
		err := manager.GetSingle(ctx, rs.Primary.Attributes["domain"], rs.Primary.Attributes["name"], &cred)
		if err == nil {
			return fmt.Errorf("Credentials still exists: %s - %s", rs.Primary.Attributes["folder"], rs.Primary.Attributes["name"])
		}
	}

	return nil
}
