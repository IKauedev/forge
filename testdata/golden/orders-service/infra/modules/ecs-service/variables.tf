variable "name" {
  description = "Nome do ambiente (<app>-<ambiente>): cluster, serviço, família, ALB e target group."
  type        = string
}

variable "container_name" {
  description = "Nome do container (o mesmo da aplicação; o pipeline usa esse nome)."
  type        = string
}

variable "container_image" {
  description = "Imagem inicial (registro/repositório:tag). As próximas versões vêm do pipeline."
  type        = string
}

variable "container_port" {
  description = "Porta HTTP da aplicação."
  type        = number
  default     = 8080
}

variable "health_check_path" {
  description = "Caminho de health check usado pelo ALB e pelo container."
  type        = string
  default     = "/actuator/health/readiness"
}

variable "vpc_id" {
  description = "VPC do serviço."
  type        = string
}

variable "public_subnet_ids" {
  description = "Subnets públicas do ALB."
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "Subnets privadas das tasks."
  type        = list(string)
}

variable "certificate_arn" {
  description = "Certificado ACM do ALB. Vazio mantém só HTTP; preenchido habilita HTTPS e redireciona a porta 80."
  type        = string
  default     = ""
}

variable "alb_deletion_protection" {
  description = "Protege o ALB contra exclusão acidental."
  type        = bool
  default     = false
}

variable "cpu" {
  description = "CPU da task Fargate (unidades)."
  type        = number
  default     = 512
}

variable "memory" {
  description = "Memória da task Fargate (MiB)."
  type        = number
  default     = 1024
}

variable "desired_count" {
  description = "Número inicial de tasks (depois o autoscaling assume)."
  type        = number
  default     = 1
}

variable "min_count" {
  description = "Mínimo de tasks do autoscaling."
  type        = number
  default     = 1
}

variable "max_count" {
  description = "Máximo de tasks do autoscaling."
  type        = number
  default     = 2
}

variable "cpu_target" {
  description = "Utilização média de CPU alvo do autoscaling (%)."
  type        = number
  default     = 70
}

variable "log_retention_days" {
  description = "Retenção dos logs no CloudWatch."
  type        = number
  default     = 30
}

variable "environment" {
  description = "Variáveis de ambiente do container (valores não secretos)."
  type        = map(string)
  default     = {}
}

variable "secrets" {
  description = "Segredos do container: nome da variável => valueFrom (ARN do Secrets Manager, com chave JSON)."
  type        = map(string)
  default     = {}
}

variable "secret_arns" {
  description = "ARNs de segredos que a role de execução pode ler (os usados em secrets)."
  type        = list(string)
  default     = []
}
