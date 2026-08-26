resource "jenkins_credential_vault_username_password" "example" {
  name = "example-vault-username-password"
  path = "secret/data/example/creds"
}
