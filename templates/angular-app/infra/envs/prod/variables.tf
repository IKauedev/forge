variable "aws_region" {
  description = "Região AWS do bucket."
  type        = string
}

variable "domain_name" {
  description = "Domínio do site (ex.: app.example.com). Vazio usa o domínio *.cloudfront.net."
  type        = string
  default     = ""
}

variable "hosted_zone_id" {
  description = "Zona do Route 53 do domínio. Obrigatória com domain_name."
  type        = string
  default     = ""
}
