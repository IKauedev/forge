# Kubernetes (Kustomize: dev, hml e prd)

Manifests para servir a aplicação (Nginx sem privilégios, porta 8080) em um cluster Kubernetes (testado contra o EKS,
mas nada aqui é exclusivo dele). A imagem é a do `Dockerfile` e **a mesma nos três ambientes**: o `config.json` é
gerado na subida do container a partir do ConfigMap do overlay.

## Estrutura

```
k8s/
  base/                       recursos comuns
    deployment.yaml           pods não-root (UID 101), probes, preStop, spread por nó
    service.yaml              ClusterIP :80 -> :8080
    ingress.yaml              nginx + TLS (host de exemplo: starter-app.example.com)
    hpa.yaml                  autoscaling por CPU (70%)
    pdb.yaml                  drenar um nó derruba no máximo 1 pod por vez
    networkpolicy.yaml        entrada só do ingress; saída só para o DNS
    serviceaccount.yaml       sem token da API montado
    kustomization.yaml        ConfigMap gerado (vira o config.json)
  overlays/
    dev/  hml/  prd/          namespace, réplicas, host e URL da API de cada ambiente
```

| Ambiente | Namespace         | Réplicas (HPA) | Host                          | API (`API_BASE_URL`)          |
| -------- | ----------------- | -------------- | ----------------------------- | ----------------------------- |
| dev      | `starter-app-dev` | 1 a 2          | `starter-app.dev.example.com` | `https://api.dev.example.com` |
| hml      | `starter-app-hml` | 2 a 3          | `starter-app.hml.example.com` | `https://api.hml.example.com` |
| prd      | `starter-app-prd` | 3 a 10         | `starter-app.example.com`     | `https://api.example.com`     |

Cada namespace nasce com o Pod Security `restricted`; os pods do template já cumprem.

## Configuração em runtime

O ConfigMap define `API_BASE_URL`, `AUTH_ENABLED`, `AUTH_ISSUER`, `AUTH_CLIENT_ID` e `AUTH_SCOPE`, as mesmas variáveis
do container (`docker/40-config.sh`). **O `config.json` é público**: nunca ponha segredo no ConfigMap. Para ligar o
login, defina por overlay `AUTH_ENABLED=true`, `AUTH_ISSUER` e `AUTH_CLIENT_ID`.

## O que o cluster precisa ter

- **ingress-nginx** (namespace `ingress-nginx`): o `Ingress` usa `ingressClassName: nginx` e a NetworkPolicy só aceita
  tráfego vindo desse namespace. Outro controller? Ajuste as duas coisas.
- **metrics-server**: sem ele o HPA não enxerga a CPU.
- **Certificado TLS**: o `Ingress` espera o Secret `starter-app-tls`. Com o cert-manager, descomente a annotation
  `cert-manager.io/cluster-issuer` em `base/ingress.yaml`.
- Um CNI que aplique NetworkPolicy (Calico, Cilium ou o _network policy_ do VPC CNI). Sem isso as regras são ignoradas.

## Antes do primeiro deploy

1. **Troque os valores de exemplo** (hosts e URLs de API em `example.com`) em `overlays/*/kustomization.yaml` e
   `base/ingress.yaml`.
2. **Repositório de imagens**: o pipeline publica no ECR (`starter-app`). Crie o repositório
   (`aws ecr create-repository --repository-name starter-app`) ou aponte `ECR_REPOSITORY` para outro.

## Aplicar na mão

```sh
kubectl kustomize k8s/overlays/dev          # só renderiza, não toca no cluster
kubectl diff -k k8s/overlays/dev            # o que mudaria

# A base não fixa imagem. Defina uma antes de aplicar:
cd k8s/overlays/dev
kustomize edit set image starter-app=<registro>/starter-app:<tag>
kubectl apply -k .
```

Não commite o `kustomization.yaml` com o bloco `images` que o `kustomize edit` acrescenta: o pipeline faz isso a cada deploy.

## Pipeline (GitHub Actions, sem access key)

`deploy-k8s.yml` constrói a imagem uma vez, publica no ECR e promove **a mesma imagem** por dev, hml e prd com
`deploy-k8s-env.yml` (que roda `kubectl apply -k` e espera o rollout). A autenticação na AWS é por **OIDC**, e o acesso
ao cluster vem de uma _access entry_ do EKS para a role de deploy.

| Gatilho                                | O que acontece                                                                    |
| -------------------------------------- | --------------------------------------------------------------------------------- |
| Manual (`workflow_dispatch`) na `main` | build, push e deploy em dev -> hml -> prd                                         |
| Push na `main`                         | o mesmo, **somente** se a variável de repositório `K8S_DEPLOY_ENABLED` for `true` |

A trava existe porque `deploy.yml` (S3 + CloudFront) também dispara no push da `main`. Use um dos dois pipelines por
padrão. Se só usa Kubernetes, apague `deploy.yml`, `deploy-env.yml` e a pasta `infra/` e defina `K8S_DEPLOY_ENABLED=true`.

Em **Settings -> Environments** crie `dev`, `hml` e `prd`. Recomendado: **Required reviewers** em `prd` (e, se quiser, em `hml`).

| Variável             | Onde                       | Obrigatória | Exemplo / padrão                                        |
| -------------------- | -------------------------- | ----------- | ------------------------------------------------------- |
| `AWS_REGION`         | repositório ou Environment | sim         | `us-east-1`                                             |
| `AWS_ROLE_ARN`       | **por Environment**        | sim         | `arn:aws:iam::123456789012:role/starter-app-dev-deploy` |
| `EKS_CLUSTER`        | Environment                | não         | padrão `starter-app-<ambiente>`                         |
| `ECR_REPOSITORY`     | repositório                | não         | padrão `starter-app`                                    |
| `K8S_DEPLOY_ENABLED` | repositório                | não         | `true` liga o deploy no push da `main`                  |

A role de cada ambiente confia no OIDC do GitHub (restrinja repositório **e** Environment), pode `eks:DescribeCluster`
e precisa de permissão no cluster: crie uma _access entry_ para ela e associe uma política de acesso restrita ao
namespace do ambiente (`AmazonEKSEditPolicy` com escopo `starter-app-dev`, por exemplo). O primeiro deploy também cria o
namespace, então esse escopo precisa permitir isso ou o namespace deve existir antes.

O `ci.yml` renderiza os três overlays a cada PR: um patch quebrado derruba o CI, não o deploy.

## Segurança

- Pods como UID `101` (Nginx sem privilégios), sem escalar privilégios, sem capabilities e `seccomp` `RuntimeDefault`.
  O sistema de arquivos é gravável porque o entrypoint escreve o `config.json` em `/usr/share/nginx/html`.
- Sem token da API do Kubernetes montado no pod.
- NetworkPolicy: entrada só do ingress-nginx, saída só para o DNS.
- Nenhum segredo no repositório.
