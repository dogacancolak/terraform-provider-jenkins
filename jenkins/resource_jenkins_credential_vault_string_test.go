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

func TestAccJenkinsCredentialVaultString_basic(t *testing.T) {
	var cred VaultStringCredentials

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy:             testAccCheckJenkinsCredentialVaultStringDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource jenkins_credential_vault_string foo {
				  name = "test-vault-string"
				  path = "secret/data/foo"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_string.foo", "id", "/test-vault-string"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_string.foo", "vault_key", "secret"),
					testAccCheckJenkinsCredentialVaultStringExists("jenkins_credential_vault_string.foo", &cred),
				),
			},
			{
				// Update by adding description and vault_key
				Config: `
				resource jenkins_credential_vault_string foo {
				  name = "test-vault-string"
				  description = "new-description"
				  path = "secret/data/foo"
				  vault_key = "token"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJenkinsCredentialVaultStringExists("jenkins_credential_vault_string.foo", &cred),
					resource.TestCheckResourceAttr("jenkins_credential_vault_string.foo", "description", "new-description"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_string.foo", "vault_key", "token"),
				),
			},
			{
				ResourceName:      "jenkins_credential_vault_string.foo",
				ImportState:       true,
				ImportStateId:     defaultCredentialDomain + "/test-vault-string",
				ImportStateVerify: true,
				// ImportState writes folder as "" for a global credential where a
				// normal apply leaves it null. Shared by every credential type.
				ImportStateVerifyIgnore: []string{"folder"},
			},
		},
	})
}

func TestAccJenkinsCredentialVaultString_folder(t *testing.T) {
	var cred VaultStringCredentials
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckJenkinsCredentialVaultStringDestroy,
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

				resource jenkins_credential_vault_string foo {
				  name = "test-vault-string"
				  folder = jenkins_folder.foo_sub.id
				  path = "secret/data/foo"
				}`, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_string.foo", "id", "/job/tf-acc-test-"+randString+"/job/subfolder/test-vault-string"),
					resource.TestCheckResourceAttr("jenkins_credential_vault_string.foo", "vault_key", "secret"),
					testAccCheckJenkinsCredentialVaultStringExists("jenkins_credential_vault_string.foo", &cred),
				),
			},
		},
	})
}

func testAccCheckJenkinsCredentialVaultStringExists(resourceName string, cred *VaultStringCredentials) resource.TestCheckFunc {
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

func testAccCheckJenkinsCredentialVaultStringDestroy(s *terraform.State) error {
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jenkins_credential_vault_string" {
			continue
		}

		cred := VaultStringCredentials{}
		manager := testAccClient.Credentials()
		manager.Folder = formatFolderName(rs.Primary.Attributes["folder"])
		err := manager.GetSingle(ctx, rs.Primary.Attributes["domain"], rs.Primary.Attributes["name"], &cred)
		if err == nil {
			return fmt.Errorf("Credentials still exists: %s - %s", rs.Primary.Attributes["folder"], rs.Primary.Attributes["name"])
		}
	}

	return nil
}
