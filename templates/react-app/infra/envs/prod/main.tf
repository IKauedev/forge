data "aws_caller_identity" "current" {}

locals {
  app = "starter-app"
  env = "prod"
  # O id da conta torna o nome do bucket único no mundo todo (limite de 63 caracteres).
  bucket_name = "${local.app}-${local.env}-${data.aws_caller_identity.current.account_id}"
}

module "site" {
  source = "../../modules/static-site"

  providers = {
    aws           = aws
    aws.us_east_1 = aws.us_east_1
  }

  name           = "${local.app}-${local.env}"
  bucket_name    = local.bucket_name
  domain_name    = var.domain_name
  hosted_zone_id = var.hosted_zone_id

  price_class          = "PriceClass_200"
  bucket_force_destroy = false
}
