# Infra (Terraform)

Tudo em `infra/`. O forge não executa `apply`; você roda os passos abaixo.

## Ordem

1. `infra/bootstrap`: roda uma vez, com state local. Cria o bucket de state, o provedor OIDC do GitHub
   (`create_oidc_provider = false` se a conta já tem um) e as roles de deploy, restritas ao repositório
   (`github_repository`) e ao Environment. Copie `terraform.tfvars.example` para `terraform.tfvars`.
2. `infra/envs/dev` e `infra/envs/prod`: copie `backend.hcl.example` para `backend.hcl` e
   `terraform.tfvars.example` para `terraform.tfvars`; depois
   `terraform init -backend-config=backend.hcl`, `terraform plan`, `terraform apply`.
3. Configure no GitHub, por Environment, as variáveis `AWS_REGION` e `AWS_ROLE_ARN` (ver `docs/PIPELINE.md`).

## Decisões

- **Lock do state**: `use_lockfile = true` no backend S3 (lock nativo); sem DynamoDB.
- **Banco** (`--db != none`): RDS com `manage_master_user_password = true`; a senha fica no Secrets Manager
  e entra na task pelo bloco `secrets`. Nenhuma senha em variável ou arquivo. SG do banco aceita só o SG do ECS.
- **Rede** (`--vpc`): `new` cria VPC, subnets e NAT (dev: um NAT; prod: um por AZ). `existing` pede
  `vpc_id` e subnets públicas/privadas nas variáveis.
- **dev**: instância pequena. **prod**: Multi-AZ, backup e `deletion_protection`.
- Nomes seguem `<app>-<env>` (convenção do pipeline) e respeitam os limites da AWS.

## Validação local

```sh
terraform fmt -check -recursive
cd envs/dev && terraform init -backend=false && terraform validate
```
