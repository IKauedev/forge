variable "aws_region" {
  description = "Região AWS dos recursos."
  type        = string
}

variable "github_repository" {
  description = "Repositório GitHub que pode assumir as roles de deploy, no formato <org>/<repo>."
  type        = string

  validation {
    condition     = can(regex("^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$", var.github_repository))
    error_message = "Use o formato <org>/<repo>."
  }
}

variable "create_oidc_provider" {
  description = "Cria o provedor OIDC do GitHub. Use false se a conta já tem um (só pode existir um por conta)."
  type        = bool
  default     = true
}

variable "state_bucket_force_destroy" {
  description = "Permite destruir o bucket de state mesmo com objetos. Deixe false."
  type        = bool
  default     = false
}
