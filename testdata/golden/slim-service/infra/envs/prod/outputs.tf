output "alb_dns_name" {
  description = "DNS do ALB do ambiente."
  value       = module.ecs_service.alb_dns_name
}

output "ecs_cluster" {
  description = "Cluster ECS (variável ECS_CLUSTER do GitHub, se não seguir a convenção)."
  value       = module.ecs_service.cluster_name
}

output "ecs_service" {
  description = "Serviço ECS."
  value       = module.ecs_service.service_name
}

output "ecs_task_family" {
  description = "Família da task definition."
  value       = module.ecs_service.task_definition_family
}
