variable "name" {
  description = "Nome do repositório ECR (o mesmo nome da aplicação)."
  type        = string
}

variable "keep_last_images" {
  description = "Quantas imagens manter; as mais antigas expiram."
  type        = number
  default     = 30
}

variable "force_delete" {
  description = "Permite destruir o repositório mesmo com imagens."
  type        = bool
  default     = false
}
