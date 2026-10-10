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

# forge:if auth!=none
variable "jwt_issuer_uri" {
  description = "Issuer do JWT (Cognito, Keycloak, Auth0...), valor da variável JWT_ISSUER_URI da aplicação."
  type        = string
}
# forge:end

# forge:if cache=redis
#~ variable "redis_host" {
#~   description = "Host de um Redis existente (ElastiCache ou outro). O forge não cria o Redis na AWS."
#~   type        = string
#~ }
# forge:end

# forge:if vpc=existing
#~ variable "vpc_id" {
#~   description = "VPC existente."
#~   type        = string
#~ }
#~
#~ variable "public_subnet_ids" {
#~   description = "Subnets públicas existentes (ALB)."
#~   type        = list(string)
#~ }
#~
#~ variable "private_subnet_ids" {
#~   description = "Subnets privadas existentes (tasks e banco)."
#~   type        = list(string)
#~ }
# forge:end
