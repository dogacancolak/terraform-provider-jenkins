resource "jenkins_credential_vault_file" "global" {
  name      = "global-vault-file"
  namespace = "baz"
  path      = "secret/data/global"
  file_name = "global-vault-file.txt"
  use_key   = true
  vault_key = "global-key"
}

data "jenkins_credential_vault_file" "global" {
  depends_on = [jenkins_credential_vault_file.global]
  name       = "global-vault-file"
}

output "vault_file" {
  value = data.jenkins_credential_vault_file.global
}

resource "jenkins_credential_vault_file" "folder" {
  name      = "folder-vault-file"
  folder    = jenkins_folder.example.id
  path      = "secret/data/folder"
  file_name = "folder-vault-file.txt"
}

resource "jenkins_credential_vault_file" "global-namespaced" {
  name      = "global-vault-file-namespaced"
  path      = "secret/data/global"
  file_name = "global-vault-file-namespaced.txt"
  namespace = "my-namespace"
}

resource "jenkins_credential_vault_file" "folder-namespaced" {
  name      = "folder-vault-file-namespaced"
  folder    = jenkins_folder.example.id
  path      = "secret/data/folder"
  file_name = "folder-vault-file-namespaced.txt"
  namespace = "my-namespace"
}
