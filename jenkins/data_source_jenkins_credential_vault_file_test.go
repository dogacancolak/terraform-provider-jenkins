package jenkins

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJenkinsCredentialVaultFileDataSource_basic(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_credential_vault_file foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance tests %s"
					path = "secret/data/foo"
					file_name = "secret.txt"
					use_key = true
					vault_key = "my-key"
				}

				data jenkins_credential_vault_file foo {
					name   = jenkins_credential_vault_file.foo.name
					domain = "`+defaultCredentialDomain+`"
				}`, randString, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_file.foo", "id", "/tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "id", "/tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "name", "tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "description", "Terraform acceptance tests "+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "path", "secret/data/foo"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "file_name", "secret.txt"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "use_key", "true"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "vault_key", "my-key"),
				),
			},
		},
	})
}

func TestAccJenkinsCredentialVaultFileDataSource_basic_namespaced(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_credential_vault_file foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance tests %s"
					path = "secret/data/foo"
					file_name = "secret.txt"
					namespace = "my-namespace"
				}

				data jenkins_credential_vault_file foo {
					name   = jenkins_credential_vault_file.foo.name
					domain = "`+defaultCredentialDomain+`"
				}`, randString, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "name", "tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "namespace", "my-namespace"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "file_name", "secret.txt"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_file.foo", "use_key", "false"),
				),
			},
		},
	})
}
