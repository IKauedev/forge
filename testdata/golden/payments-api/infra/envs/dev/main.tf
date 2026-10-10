locals {
  app  = "payments-api"
  env  = "dev"
  name = "${local.app}-${local.env}"
}

module "network" {
  source = "../../modules/network"

  name               = local.name
  single_nat_gateway = true
}

locals {
  vpc_id             = module.network.vpc_id
  public_subnet_ids  = module.network.public_subnet_ids
  private_subnet_ids = module.network.private_subnet_ids
}

module "database" {
  source = "../../modules/database"

  name                      = local.name
  vpc_id                    = local.vpc_id
  subnet_ids                = local.private_subnet_ids
  allowed_security_group_id = module.ecs_service.security_group_id
  db_name                   = "payments_api"

  engine                 = "postgres"
  engine_version         = "17"
  parameter_group_family = "postgres17"
  port                   = 5432
  jdbc_scheme            = "postgresql"
  log_exports            = ["postgresql", "upgrade"]

  instance_class          = "db.t4g.micro"
  multi_az                = false
  backup_retention_period = 1
  deletion_protection     = false
  skip_final_snapshot     = true
}

locals {
  db_environment = {
    DB_URL = module.database.jdbc_url
  }
}

locals {
  db_secrets = {
    DB_USERNAME = "${module.database.master_user_secret_arn}:username::"
    DB_PASSWORD = "${module.database.master_user_secret_arn}:password::"
  }
}

locals {
  db_secret_arns = [module.database.master_user_secret_arn]
}

locals {
  auth_environment = {
    JWT_ISSUER_URI = var.jwt_issuer_uri
  }
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

  cpu                     = 512
  memory                  = 1024
  desired_count           = 1
  min_count               = 1
  max_count               = 2
  log_retention_days      = 30
  alb_deletion_protection = false

  environment = merge(
    { SPRING_PROFILES_ACTIVE = "dev" },
    local.db_environment,
    local.auth_environment,
    local.cache_environment,
  )
  secrets     = local.db_secrets
  secret_arns = local.db_secret_arns
}
