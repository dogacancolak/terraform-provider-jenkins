package jenkins

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJenkinsCredentialVaultStringDataSource_basic(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_credential_vault_string foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance tests %s"
					path = "secret/data/foo"
					vault_key = "my-key"
				}

				data jenkins_credential_vault_string foo {
					name   = jenkins_credential_vault_string.foo.name
					domain = "`+defaultCredentialDomain+`"
				}`, randString, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_string.foo", "id", "/tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "id", "/tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "name", "tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "description", "Terraform acceptance tests "+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "path", "secret/data/foo"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "vault_key", "my-key"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "engine_version", "2"),
				),
			},
		},
	})
}

func TestAccJenkinsCredentialVaultStringDataSource_basic_namespaced(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_credential_vault_string foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance tests %s"
					path = "secret/data/foo"
					namespace = "my-namespace"
					prefix_path = "my-prefix"
					engine_version = 1
				}

				data jenkins_credential_vault_string foo {
					name   = jenkins_credential_vault_string.foo.name
					domain = "`+defaultCredentialDomain+`"
				}`, randString, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "name", "tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "namespace", "my-namespace"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "prefix_path", "my-prefix"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "engine_version", "1"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_string.foo", "vault_key", "secret"),
				),
			},
		},
	})
}
