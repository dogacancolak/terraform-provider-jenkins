resource "jenkins_credential_vault_string" "example" {
  name      = "example-vault-string"
  path      = "secret/data/example"
  vault_key = "token"
}
