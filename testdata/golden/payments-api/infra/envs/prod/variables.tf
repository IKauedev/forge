variable "aws_region" {
  description = "Região AWS do ambiente."
  type        = string
}

variable "container_image" {
  description = "Imagem do primeiro deploy (registro/repositório:tag), já enviada ao ECR. Depois o pipeline assume."
  type        = string
}

variable "certificate_arn" {
  description = "Certificado ACM do ALB. Obrigatório em produção: habilita HTTPS e redireciona a porta 80."
  type        = string

  validation {
    condition     = startswith(var.certificate_arn, "arn:")
    error_message = "Informe o ARN do certificado ACM para publicar produção com HTTPS."
  }
}

variable "jwt_issuer_uri" {
  description = "Issuer do JWT (Cognito, Keycloak, Auth0...), valor da variável JWT_ISSUER_URI da aplicação."
  type        = string
}

