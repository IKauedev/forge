variable "name" {
  description = "Identificador do banco e prefixo dos recursos (<app>-<ambiente>)."
  type        = string
}

variable "vpc_id" {
  description = "VPC onde o banco fica."
  type        = string
}

variable "subnet_ids" {
  description = "Subnets privadas do banco."
  type        = list(string)
}

variable "allowed_security_group_id" {
  description = "Security group do serviço ECS, o único autorizado a falar com o banco."
  type        = string
}

variable "engine" {
  description = "Engine do RDS (vem do catálogo de bancos do forge)."
  type        = string
}

variable "engine_version" {
  description = "Versão maior do engine; o RDS escolhe o minor."
  type        = string
}

variable "parameter_group_family" {
  description = "Family do parameter group (vem do catálogo de bancos do forge)."
  type        = string
}

variable "port" {
  description = "Porta do banco."
  type        = number
}

variable "jdbc_scheme" {
  description = "Esquema JDBC do engine (vem do catálogo de bancos do forge)."
  type        = string
}

variable "log_exports" {
  description = "Logs do engine exportados para o CloudWatch."
  type        = list(string)
  default     = []
}

variable "db_name" {
  description = "Nome do banco criado na instância (snake_case)."
  type        = string
}

variable "username" {
  description = "Usuário mestre. A senha é gerada pelo RDS e guardada no Secrets Manager."
  type        = string
  default     = "app_master"
}

variable "instance_class" {
  description = "Classe da instância."
  type        = string
}

variable "allocated_storage" {
  description = "Armazenamento inicial em GiB."
  type        = number
  default     = 20
}

variable "max_allocated_storage" {
  description = "Limite do autoscaling de armazenamento em GiB."
  type        = number
  default     = 100
}

variable "multi_az" {
  description = "Réplica síncrona em outra zona."
  type        = bool
  default     = false
}

variable "backup_retention_period" {
  description = "Dias de retenção de backup automático."
  type        = number
  default     = 7
}

variable "deletion_protection" {
  description = "Impede a exclusão acidental da instância."
  type        = bool
  default     = false
}

variable "skip_final_snapshot" {
  description = "Pula o snapshot final ao destruir (use true só fora de produção)."
  type        = bool
  default     = true
}
