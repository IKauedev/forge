terraform {
  required_version = ">= 1.10.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  # Estado local de propósito: este diretório cria o bucket que guarda o estado dos ambientes.
  # Depois do primeiro apply, guarde o terraform.tfstate em lugar seguro (ou migre para o bucket criado).
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "orders-service"
      ManagedBy = "terraform"
      Stack     = "bootstrap"
    }
  }
}
