resource "jenkins_credential_vault_file" "example" {
  name      = "example-vault-file"
  path      = "secret/data/example"
  file_name = "private-key.pem"
  use_key   = true
  vault_key = "private-key"
}
