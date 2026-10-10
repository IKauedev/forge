output "site_url" {
  description = "URL pública do site."
  value       = module.site.site_url
}

output "bucket_name" {
  description = "Bucket do site (variável S3_BUCKET do Environment no GitHub)."
  value       = module.site.bucket_name
}

output "distribution_id" {
  description = "ID da distribuição (variável CLOUDFRONT_DISTRIBUTION_ID do Environment no GitHub)."
  value       = module.site.distribution_id
}
