# Pipeline (GitHub Actions)

- `ci.yml`: em PRs e em branches que não são a main: `npm ci`, lint, formatação, testes e build.
- `deploy.yml`: push na `main`. Valida e constrói uma vez, publica em **dev** e, depois da aprovação do
  Environment **prod**, publica o mesmo build em produção.
- `deploy-env.yml`: workflow reutilizável. Gera o `config.json` do ambiente, sincroniza o S3 e invalida o CloudFront.

Autenticação com a AWS por OIDC (role criada em `infra/bootstrap`); não há access keys.

## Cache

- Arquivos com hash no nome: `public,max-age=31536000,immutable`.
- `index.html` e `config.json`: `no-cache, no-store, must-revalidate`, publicados por último e invalidados no CloudFront.

## Configuração no GitHub

Crie os Environments `dev` e `prod` (em `prod`, ative _Required reviewers_) e, em cada um, as variáveis:

| Variável                     | Origem / uso                                                       |
| ---------------------------- | ------------------------------------------------------------------ |
| `AWS_REGION`                 | região do bucket                                                   |
| `AWS_ROLE_ARN`               | saída `deploy_role_arns` do bootstrap                              |
| `S3_BUCKET`                  | saída `bucket_name` de `infra/envs/<env>`                          |
| `CLOUDFRONT_DISTRIBUTION_ID` | saída `distribution_id` de `infra/envs/<env>`                      |
| `API_BASE_URL`               | URL da API do ambiente (vai para o `config.json`)                  |
| `AUTH_ENABLED`               | opcional, `true` ou `false` (padrão `false`)                       |
| `AUTH_ISSUER`                | emissor OIDC (obrigatório com `AUTH_ENABLED=true`)                 |
| `AUTH_CLIENT_ID`             | client ID do app no provedor (obrigatório com `AUTH_ENABLED=true`) |
| `AUTH_SCOPE`                 | opcional (padrão `openid profile email`)                           |
