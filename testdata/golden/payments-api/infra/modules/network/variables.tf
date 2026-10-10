variable "name" {
  description = "Prefixo dos recursos (<app>-<ambiente>)."
  type        = string
}

variable "cidr_block" {
  description = "CIDR da VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "az_count" {
  description = "Quantidade de zonas de disponibilidade (2 ou mais)."
  type        = number
  default     = 2

  validation {
    condition     = var.az_count >= 2 && var.az_count <= 6
    error_message = "Use entre 2 e 6 zonas de disponibilidade."
  }
}

variable "single_nat_gateway" {
  description = "true usa um único NAT Gateway (mais barato, sem redundância); false usa um por zona."
  type        = bool
  default     = true
}
