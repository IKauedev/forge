variable "name" {
  description = "Prefixo dos recursos (<app>-<env>)."
  type        = string
}

variable "bucket_name" {
  description = "Nome do bucket S3 (globalmente único, até 63 caracteres)."
  type        = string

  validation {
    condition     = length(var.bucket_name) >= 3 && length(var.bucket_name) <= 63 && can(regex("^[a-z0-9][a-z0-9.-]*[a-z0-9]$", var.bucket_name))
    error_message = "O nome do bucket deve ter de 3 a 63 caracteres, em minúsculas, números, pontos e hífens."
  }
}

variable "domain_name" {
  description = "Domínio do site (ex.: app.example.com). Vazio usa o domínio *.cloudfront.net."
  type        = string
  default     = ""
}

variable "hosted_zone_id" {
  description = "Zona do Route 53 que valida o certificado e recebe o alias. Obrigatória com domain_name."
  type        = string
  default     = ""

  validation {
    condition     = var.domain_name == "" || var.hosted_zone_id != ""
    error_message = "hosted_zone_id é obrigatório quando domain_name é informado."
  }
}

variable "price_class" {
  description = "Classe de preço do CloudFront."
  type        = string
  default     = "PriceClass_100"
}

variable "bucket_force_destroy" {
  description = "Permite destruir o bucket mesmo com objetos (use true só em dev)."
  type        = bool
  default     = false
}

variable "content_security_policy" {
  description = "Cabeçalho Content-Security-Policy. Ajuste connect-src para a URL da sua API."
  type        = string
  default     = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' https:; frame-ancestors 'none'; base-uri 'self'"
}
