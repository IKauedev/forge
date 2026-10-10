output "address" {
  description = "Endereço (host) do banco."
  value       = aws_db_instance.this.address
}

output "port" {
  description = "Porta do banco."
  value       = aws_db_instance.this.port
}

output "jdbc_url" {
  description = "URL JDBC para a variável DB_URL da aplicação."
  value       = "jdbc:${var.jdbc_scheme}://${aws_db_instance.this.address}:${aws_db_instance.this.port}/${var.db_name}"
}

output "master_user_secret_arn" {
  description = "ARN do segredo (JSON com username e password) gerado pelo RDS."
  value       = aws_db_instance.this.master_user_secret[0].secret_arn
}

output "security_group_id" {
  description = "Security group do banco."
  value       = aws_security_group.this.id
}
