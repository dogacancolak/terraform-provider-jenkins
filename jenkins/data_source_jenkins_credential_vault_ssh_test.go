package jenkins

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJenkinsCredentialVaultSSHDataSource_basic(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_credential_vault_ssh foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance tests %s"
					path = "secret/data/foo"
					username_key = "my-username"
					private_key_key = "my-private-key"
					passphrase_key = "my-passphrase"
				}

				data jenkins_credential_vault_ssh foo {
					name   = jenkins_credential_vault_ssh.foo.name
					domain = "`+defaultCredentialDomain+`"
				}`, randString, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("jenkins_credential_vault_ssh.foo", "id", "/tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "id", "/tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "name", "tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "description", "Terraform acceptance tests "+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "path", "secret/data/foo"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "username_key", "my-username"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "private_key_key", "my-private-key"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "passphrase_key", "my-passphrase"),
				),
			},
		},
	})
}

func TestAccJenkinsCredentialVaultSSHDataSource_basic_namespaced(t *testing.T) {
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource jenkins_credential_vault_ssh foo {
					name = "tf-acc-test-%s"
					description = "Terraform acceptance tests %s"
					path = "secret/data/foo"
					namespace = "my-namespace"
				}

				data jenkins_credential_vault_ssh foo {
					name   = jenkins_credential_vault_ssh.foo.name
					domain = "`+defaultCredentialDomain+`"
				}`, randString, randString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "name", "tf-acc-test-"+randString),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "namespace", "my-namespace"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "username_key", "username"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "private_key_key", "private_key"),
					resource.TestCheckResourceAttr("data.jenkins_credential_vault_ssh.foo", "passphrase_key", "passphrase"),
				),
			},
		},
	})
}
