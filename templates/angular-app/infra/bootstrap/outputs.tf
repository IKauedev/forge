output "state_bucket" {
  description = "Bucket do estado remoto: use em envs/<ambiente>/backend.hcl."
  value       = aws_s3_bucket.state.id
}

output "deploy_role_arns" {
  description = "Valor da variável AWS_ROLE_ARN de cada Environment do GitHub (dev e prod)."
  value       = { for env, role in aws_iam_role.deploy : env => role.arn }
}

output "oidc_provider_arn" {
  description = "Provedor OIDC do GitHub usado pelas roles."
  value       = local.oidc_provider_arn
}

output "site_bucket_names" {
  description = "Buckets dos sites (variável S3_BUCKET de cada Environment)."
  value       = local.site_buckets
}
