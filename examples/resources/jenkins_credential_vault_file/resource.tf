resource "jenkins_credential_vault_file" "example" {
  name      = "example-vault-file"
  path      = "secret/data/example"
  use_key   = true
  vault_key = "private-key"
}
