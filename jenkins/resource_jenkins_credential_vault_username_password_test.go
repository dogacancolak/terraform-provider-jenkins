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
			{
				ResourceName:      "jenkins_credential_vault_username_password.foo",
				ImportState:       true,
				ImportStateId:     defaultCredentialDomain + "/test-vault-username-password",
				ImportStateVerify: true,
				// ImportState writes folder as "" for a global credential where a
				// normal apply leaves it null. Shared by every credential type.
				ImportStateVerifyIgnore: []string{"folder"},
			},
		},
	})
}

func TestAccJenkinsCredentialVaultUsernamePassword_folder(t *testing.T) {
	var cred VaultUsernamePasswordCredentials
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckJenkinsCredentialVaultUsernamePasswordDestroy,
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

				resource jenkins_credential_vault_username_password foo {
				  name = "test-vault-username-password"
				  folder = jenkins_folder.foo_sub.id
				  path = "secret/data/foo"
				}`, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "id", "/job/tf-acc-test-"+randString+"/job/subfolder/test-vault-username-password"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "password_key", "password"),
					testAccCheckJenkinsCredentialVaultUsernamePasswordExists("jenkins_credential_vault_username_password.foo", &cred),
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
		}

		cred := VaultUsernamePasswordCredentials{}
		manager := testAccClient.Credentials()
		manager.Folder = formatFolderName(rs.Primary.Attributes["folder"])
		err := manager.GetSingle(ctx, rs.Primary.Attributes["domain"], rs.Primary.Attributes["name"], &cred)
		if err == nil {
			return fmt.Errorf("Credentials still exists: %s - %s", rs.Primary.Attributes["folder"], rs.Primary.Attributes["name"])
		}
	}

	return nil
}
