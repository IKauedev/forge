output "bucket_name" {
  description = "Bucket do site: destino do aws s3 sync."
  value       = aws_s3_bucket.site.id
}

output "bucket_arn" {
  description = "ARN do bucket do site."
  value       = aws_s3_bucket.site.arn
}

output "distribution_id" {
  description = "ID da distribuição CloudFront: alvo da invalidação."
  value       = aws_cloudfront_distribution.site.id
}

output "distribution_arn" {
  description = "ARN da distribuição CloudFront."
  value       = aws_cloudfront_distribution.site.arn
}

output "site_url" {
  description = "URL pública do site."
  value       = local.has_domain ? "https://${var.domain_name}" : "https://${aws_cloudfront_distribution.site.domain_name}"
}
