resource "jenkins_credential_vault_username_password" "global" {
  name         = "global-vault-username-password"
  namespace    = "baz"
  path         = "secret/data/global"
  username_key = "global-username"
  password_key = "global-password"
}

data "jenkins_credential_vault_username_password" "global" {
  depends_on = [jenkins_credential_vault_username_password.global]
  name       = "global-vault-username-password"
}

output "vault_username_password" {
  value = data.jenkins_credential_vault_username_password.global
}

resource "jenkins_credential_vault_username_password" "folder" {
  name   = "folder-vault-username-password"
  folder = jenkins_folder.example.id
  path   = "secret/data/folder"
}

resource "jenkins_credential_vault_username_password" "global-namespaced" {
  name      = "global-vault-username-password-namespaced"
  path      = "secret/data/global"
  namespace = "my-namespace"
}

resource "jenkins_credential_vault_username_password" "folder-namespaced" {
  name      = "folder-vault-username-password-namespaced"
  folder    = jenkins_folder.example.id
  path      = "secret/data/folder"
  namespace = "my-namespace"
}
