resource "jenkins_credential_vault_string" "global" {
  name      = "global-vault-string"
  namespace = "baz"
  path      = "secret/data/global"
  vault_key = "global-key"
}

data "jenkins_credential_vault_string" "global" {
  depends_on = [jenkins_credential_vault_string.global]
  name       = "global-vault-string"
}

output "vault_string" {
  value = data.jenkins_credential_vault_string.global
}

resource "jenkins_credential_vault_string" "folder" {
  name   = "folder-vault-string"
  folder = jenkins_folder.example.id
  path   = "secret/data/folder"
}

resource "jenkins_credential_vault_string" "global-namespaced" {
  name      = "global-vault-string-namespaced"
  path      = "secret/data/global"
  namespace = "my-namespace"
}

resource "jenkins_credential_vault_string" "folder-namespaced" {
  name      = "folder-vault-string-namespaced"
  folder    = jenkins_folder.example.id
  path      = "secret/data/folder"
  namespace = "my-namespace"
}
