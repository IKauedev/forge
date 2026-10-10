output "state_bucket" {
  description = "Bucket do estado remoto: use em envs/<ambiente>/backend.hcl."
  value       = aws_s3_bucket.state.id
}

output "ecr_repository_url" {
  description = "URL do repositório ECR."
  value       = module.ecr.repository_url
}

output "deploy_role_arns" {
  description = "Valor da variável AWS_ROLE_ARN de cada Environment do GitHub (dev e prod)."
  value       = { for env, role in aws_iam_role.deploy : env => role.arn }
}

output "oidc_provider_arn" {
  description = "Provedor OIDC do GitHub usado pelas roles."
  value       = local.oidc_provider_arn
}

output "github_variables" {
  description = "Resumo do que configurar no GitHub (Settings > Environments > Variables)."
  value = {
    AWS_REGION = local.region
    AWS_ROLE_ARN = {
      for env, role in aws_iam_role.deploy : env => role.arn
    }
  }
}
