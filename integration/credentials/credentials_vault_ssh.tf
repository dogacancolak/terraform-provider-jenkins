resource "jenkins_credential_vault_ssh" "global" {
  name            = "global-vault-ssh"
  namespace       = "baz"
  path            = "secret/data/global"
  username_key    = "global-username"
  private_key_key = "global-private-key"
  passphrase_key  = "global-passphrase"
}

data "jenkins_credential_vault_ssh" "global" {
  depends_on = [jenkins_credential_vault_ssh.global]
  name       = "global-vault-ssh"
}

output "vault_ssh" {
  value = data.jenkins_credential_vault_ssh.global
}

resource "jenkins_credential_vault_ssh" "folder" {
  name   = "folder-vault-ssh"
  folder = jenkins_folder.example.id
  path   = "secret/data/folder"
}

resource "jenkins_credential_vault_ssh" "global-namespaced" {
  name      = "global-vault-ssh-namespaced"
  path      = "secret/data/global"
  namespace = "my-namespace"
}

resource "jenkins_credential_vault_ssh" "folder-namespaced" {
  name      = "folder-vault-ssh-namespaced"
  folder    = jenkins_folder.example.id
  path      = "secret/data/folder"
  namespace = "my-namespace"
}
