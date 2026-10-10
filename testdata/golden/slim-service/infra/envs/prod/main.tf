locals {
  app  = "slim-service"
  env  = "prod"
  name = "${local.app}-${local.env}"
}

module "network" {
  source = "../../modules/network"

  name               = local.name
  single_nat_gateway = false
}

locals {
  vpc_id             = module.network.vpc_id
  public_subnet_ids  = module.network.public_subnet_ids
  private_subnet_ids = module.network.private_subnet_ids
}

locals {
  db_environment = {}
}

locals {
  db_secrets = {}
}

locals {
  db_secret_arns = []
}

locals {
  auth_environment = {}
}

locals {
  cache_environment = {}
}

module "ecs_service" {
  source = "../../modules/ecs-service"

  name               = local.name
  container_name     = local.app
  container_image    = var.container_image
  vpc_id             = local.vpc_id
  public_subnet_ids  = local.public_subnet_ids
  private_subnet_ids = local.private_subnet_ids
  certificate_arn    = var.certificate_arn

  cpu                     = 1024
  memory                  = 2048
  desired_count           = 2
  min_count               = 2
  max_count               = 6
  log_retention_days      = 90
  alb_deletion_protection = true

  environment = merge(
    { SPRING_PROFILES_ACTIVE = "prod" },
    local.db_environment,
    local.auth_environment,
    local.cache_environment,
  )
  secrets     = local.db_secrets
  secret_arns = local.db_secret_arns
}
