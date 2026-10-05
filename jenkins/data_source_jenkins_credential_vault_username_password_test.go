package jenkins

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJenkinsCredentialVaultUsernamePasswordDataSource_basic(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_credential_vault_username_password foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance tests %s"
					path = "secret/data/foo"
					username_key = "my-username"
					password_key = "my-password"
				}

				data jenkins_credential_vault_username_password foo {
					name   = jenkins_credential_vault_username_password.foo.name
					domain = "`+defaultCredentialDomain+`"
				}`, randString, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_username_password.foo", "id", "/tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "id", "/tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "name", "tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "description", "Terraform acceptance tests "+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "path", "secret/data/foo"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "username_key", "my-username"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "password_key", "my-password"),
				),
			},
		},
	})
}

func TestAccJenkinsCredentialVaultUsernamePasswordDataSource_basic_namespaced(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_credential_vault_username_password foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance tests %s"
					path = "secret/data/foo"
					namespace = "my-namespace"
				}

				data jenkins_credential_vault_username_password foo {
					name   = jenkins_credential_vault_username_password.foo.name
					domain = "`+defaultCredentialDomain+`"
				}`, randString, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "name", "tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "namespace", "my-namespace"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "username_key", "username"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_username_password.foo", "password_key", "password"),
				),
			},
		},
	})
}
