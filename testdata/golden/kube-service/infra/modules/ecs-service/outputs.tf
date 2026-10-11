output "cluster_name" {
  description = "Nome do cluster ECS."
  value       = aws_ecs_cluster.this.name
}

output "service_name" {
  description = "Nome do serviço ECS."
  value       = aws_ecs_service.this.name
}

output "task_definition_family" {
  description = "Família da task definition."
  value       = aws_ecs_task_definition.this.family
}

output "alb_dns_name" {
  description = "DNS do ALB (aponte o seu domínio para ele)."
  value       = aws_lb.this.dns_name
}

output "security_group_id" {
  description = "Security group das tasks (autorizado pelo banco)."
  value       = aws_security_group.service.id
}

output "execution_role_arn" {
  description = "Role de execução da task."
  value       = aws_iam_role.execution.arn
}

output "task_role_arn" {
  description = "Role da aplicação."
  value       = aws_iam_role.task.arn
}
