resource "jenkins_credential_vault_ssh" "example" {
  name = "example-vault-ssh"
  path = "secret/data/example/ssh"
}
