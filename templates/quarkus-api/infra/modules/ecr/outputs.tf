output "repository_url" {
  description = "URL do repositório (registro/nome)."
  value       = aws_ecr_repository.this.repository_url
}

output "repository_arn" {
  description = "ARN do repositório."
  value       = aws_ecr_repository.this.arn
}

output "repository_name" {
  description = "Nome do repositório."
  value       = aws_ecr_repository.this.name
}
