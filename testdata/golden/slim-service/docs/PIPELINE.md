# Pipeline de CI/CD (GitHub Actions → AWS)

Autenticação na AWS **somente por OIDC**: não existe access key nem secret da AWS neste repositório.

## Fluxo

| Evento | Workflow | O que acontece |
|---|---|---|
| Pull request ou push em branch ≠ `main` | `ci.yml` | `./mvnw verify` (testes, Spotless, ArchUnit, JaCoCo), Hadolint no `Dockerfile` e Trivy na imagem |
| Push na `main` | `deploy.yml` | build da imagem, push no ECR (tag = SHA do commit), deploy em **dev**, depois deploy em **prod** após aprovação |
| Manual (`workflow_dispatch`) na `main` | `deploy.yml` | repete o fluxo acima para o commit atual |

A imagem é construída **uma vez**. Dev e prod recebem exatamente a mesma imagem (mesma tag).
`deploy-env.yml` é um workflow reutilizável: baixa a task definition atual do ECS, troca só a imagem e faz o
deploy esperando o serviço estabilizar (`wait-for-service-stability`).

O Trivy reprova a imagem com vulnerabilidades HIGH/CRITICAL **que já têm correção**. Se um CVE novo quebrar o CI,
atualize a dependência (veja os overrides de versão em `pom.xml`) ou a imagem base.

## O que configurar no GitHub

Em **Settings → Environments**, crie dois ambientes:

| Environment | Proteção |
|---|---|
| `dev` | nenhuma |
| `prod` | **Required reviewers** (a aprovação manual do deploy em produção) e, se quiser, restrição à branch `main` |

Em **Settings → Secrets and variables → Actions → Variables** (não são secrets):

| Variável | Onde | Obrigatória | Exemplo / padrão |
|---|---|---|---|
| `AWS_REGION` | repositório ou Environment | sim | `us-east-1` |
| `AWS_ROLE_ARN` | **por Environment** (`dev`, `prod`) | sim | `arn:aws:iam::123456789012:role/slim-service-dev-deploy` |
| `ECR_REPOSITORY` | repositório | não | `slim-service` |
| `ECS_CLUSTER` | Environment | não | `slim-service-dev` |
| `ECS_SERVICE` | Environment | não | `slim-service-dev` |
| `ECS_TASK_FAMILY` | Environment | não | `slim-service-dev` |

Sem as opcionais, os nomes seguem a convenção `<app>-<ambiente>` (`slim-service-dev`, `slim-service-prod`). A infra
Terraform do forge (planejada) seguirá a mesma convenção, então normalmente não será preciso defini-las.

## Roles na AWS (OIDC)

Cada role de deploy precisa confiar no provedor OIDC do GitHub e restringir **repositório e ambiente**:

```json
{
  "Effect": "Allow",
  "Principal": { "Federated": "arn:aws:iam::<conta>:oidc-provider/token.actions.githubusercontent.com" },
  "Action": "sts:AssumeRoleWithWebIdentity",
  "Condition": {
    "StringEquals": {
      "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
      "token.actions.githubusercontent.com:sub": "repo:<org>/<repo>:environment:dev"
    }
  }
}
```

Use `environment:prod` na role de produção. Permissões mínimas: ECR (login, push/pull da imagem do repositório),
`ecs:DescribeTaskDefinition`, `ecs:RegisterTaskDefinition`, `ecs:UpdateService`, `ecs:DescribeServices` e
`iam:PassRole` apenas para as roles de execução e da task. Esta versão do projeto ainda não traz o Terraform
(`infra/`): até lá, crie o provedor OIDC e as roles manualmente ou com a sua própria infra.

## Primeiro deploy

O workflow não cria infraestrutura. Antes do primeiro push na `main` precisam existir o repositório ECR, o cluster
ECS, o serviço e a task definition inicial (com um container chamado `slim-service`), cada um com o nome da
convenção acima. Sem isso o passo "Baixar a task definition atual" falha com a mensagem
`Task definition ... não encontrada`.

## Conta separada para produção

O padrão usa uma conta AWS e um repositório ECR compartilhado. Para prod em outra conta: configure replicação do
ECR (ou permissão de pull entre contas), aponte `ECR_REPOSITORY`/`AWS_ROLE_ARN` do Environment `prod` para a outra
conta e mantenha a mesma tag de imagem.
