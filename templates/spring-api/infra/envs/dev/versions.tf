terraform {
  required_version = ">= 1.10.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  # Bucket e região vêm de -backend-config=backend.hcl (veja backend.hcl.example).
  # O lock é um arquivo .tflock no próprio bucket (sem DynamoDB).
  backend "s3" {
    key          = "envs/dev/terraform.tfstate"
    use_lockfile = true
    encrypt      = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "starter-api"
      Environment = "dev"
      ManagedBy   = "terraform"
    }
  }
}
