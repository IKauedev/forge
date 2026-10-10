# Infra (Terraform)

Tudo em `infra/`. O forge não executa `apply`; você roda os passos abaixo.

## Ordem

1. `infra/bootstrap`: roda uma vez, com state local. Cria o bucket de state, o provedor OIDC do GitHub
   (`create_oidc_provider = false` se a conta já tem um) e as roles de deploy, restritas ao repositório
   (`github_repository`) e ao Environment. Copie `terraform.tfvars.example` para `terraform.tfvars`.
2. `infra/envs/dev` e `infra/envs/prod`: copie `backend.hcl.example` para `backend.hcl` e
   `terraform.tfvars.example` para `terraform.tfvars`; depois
   `terraform init -backend-config=backend.hcl`, `terraform plan`, `terraform apply`.
3. Configure as variáveis dos Environments no GitHub (ver `docs/PIPELINE.md`).

## O que é criado por ambiente (`modules/static-site`)

- Bucket S3 privado (sem acesso público, TLS obrigatório, criptografia SSE-S3, versionamento).
- CloudFront com OAC: só a distribuição lê o bucket.
- Fallback de SPA: 403 e 404 respondem `/index.html` com status 200, então refresh em rota profunda funciona.
- `index.html` e `config.json` sem cache no CloudFront; demais arquivos com a política CachingOptimized.
- Response headers policy: HSTS, `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy` e CSP
  (ajuste `content_security_policy` para a URL da sua API).
- Domínio próprio (opcional): informe `domain_name` e `hosted_zone_id`. O certificado ACM é criado em
  us-east-1 (provider alias `us_east_1`) e validado por DNS no Route 53.

## Decisões

- **Lock do state**: `use_lockfile = true` no backend S3 (lock nativo); sem DynamoDB.
- **Nomes**: bucket `<app>-<env>-<conta>` (limite de 63 caracteres, validado no módulo).
- **dev**: `PriceClass_100` e bucket destrutível. **prod**: `PriceClass_200`, bucket protegido.
- A role de deploy invalida as distribuições da conta (o ID só existe após o apply). Depois de criar o
  ambiente, você pode restringir `InvalidaOCache` em `bootstrap/main.tf` ao ID da distribuição.

## Validação local

```sh
terraform fmt -check -recursive
cd envs/dev && terraform init -backend=false && terraform validate
```
