# Kubernetes (Kustomize: dev, hml e prd)

Manifests para rodar a aplicação em um cluster Kubernetes (testado contra o EKS, mas nada aqui é exclusivo dele).
A imagem é a do `Dockerfile` e **a mesma nos três ambientes**: o que muda é só a configuração do overlay.

## Estrutura

```
k8s/
  base/                       recursos comuns
    deployment.yaml           pods não-root, root filesystem somente leitura, probes, preStop, spread por nó
    service.yaml              ClusterIP :80 -> :8080
    ingress.yaml              nginx + TLS (host de exemplo: kube-service.example.com)
    hpa.yaml                  autoscaling por CPU (70%)
    pdb.yaml                  drenar um nó derruba no máximo 1 pod por vez
    networkpolicy.yaml        nega tudo e libera só o necessário
    serviceaccount.yaml       sem token da API montado
    kustomization.yaml        ConfigMap gerado (variáveis não sensíveis)
  overlays/
    dev/  hml/  prd/          namespace, réplicas, recursos, host e endereços de cada ambiente
```

| Ambiente | Namespace | Réplicas (HPA) | Host |
|---|---|---|---|
| dev | `kube-service-dev` | 1 a 2 | `kube-service.dev.example.com` |
| hml | `kube-service-hml` | 2 a 3 | `kube-service.hml.example.com` |
| prd | `kube-service-prd` | 3 a 10 | `kube-service.example.com` |

Cada namespace nasce com o Pod Security `restricted`; os pods do template já cumprem.

## O que o cluster precisa ter

- **ingress-nginx** (namespace `ingress-nginx`): o `Ingress` usa `ingressClassName: nginx` e a NetworkPolicy só aceita
  tráfego vindo desse namespace. Outro controller? Ajuste as duas coisas.
- **metrics-server**: sem ele o HPA não enxerga a CPU.
- **Certificado TLS**: o `Ingress` espera o Secret `kube-service-tls`. Com o cert-manager, descomente a annotation
  `cert-manager.io/cluster-issuer` em `base/ingress.yaml`.
- Um CNI que aplique NetworkPolicy (Calico, Cilium ou o *network policy* do VPC CNI). Sem isso as regras são ignoradas.

## Antes do primeiro deploy

1. **Troque os valores de exemplo** (terminam em `example.com`, `example.internal` ou `example.invalid`): hosts em
   `overlays/*/kustomization.yaml` e `base/ingress.yaml`, `DB_HOST` e `JWT_ISSUER_URI` quando existirem.
2. **Segredo do banco** (só se a aplicação tem banco): crie o Secret em cada namespace; a senha nunca vai para o git.
   Com External Secrets, Sealed Secrets ou o Secrets Store CSI Driver, crie um Secret com o mesmo nome e as mesmas chaves.

   ```sh
   kubectl create namespace kube-service-dev   # ou aplique o overlay uma vez antes
   kubectl -n kube-service-dev create secret generic kube-service-db \
     --from-literal=username=kube_service --from-literal=password='<senha>'
   ```

   O pod lê `DB_USERNAME` e `DB_PASSWORD` desse Secret; `DB_URL` é montado a partir de `DB_HOST` (ConfigMap do overlay).
3. **Repositório de imagens**: o pipeline publica no ECR (`kube-service`). Crie o repositório
   (`aws ecr create-repository --repository-name kube-service`) ou aponte `ECR_REPOSITORY` para outro.

## Aplicar na mão

```sh
kubectl kustomize k8s/overlays/dev          # só renderiza, não toca no cluster
kubectl diff -k k8s/overlays/dev            # o que mudaria

# A base não fixa imagem. Defina uma antes de aplicar:
cd k8s/overlays/dev
kustomize edit set image kube-service=<registro>/kube-service:<tag>
kubectl apply -k .
```

Não commite o `kustomization.yaml` com o bloco `images` que o `kustomize edit` acrescenta: o pipeline faz isso a cada deploy.

## Pipeline (GitHub Actions, sem access key)

`deploy-k8s.yml` constrói a imagem uma vez, publica no ECR e promove **a mesma imagem** por dev, hml e prd com
`deploy-k8s-env.yml` (que roda `kubectl apply -k` e espera o rollout). A autenticação na AWS é por **OIDC**, e o acesso
ao cluster vem de uma *access entry* do EKS para a role de deploy.

| Gatilho | O que acontece |
|---|---|
| Manual (`workflow_dispatch`) na `main` | build, push e deploy em dev -> hml -> prd |
| Push na `main` | o mesmo, **somente** se a variável de repositório `K8S_DEPLOY_ENABLED` for `true` |

A trava existe porque `deploy.yml` (ECS) também dispara no push da `main`. Use um dos dois pipelines por padrão.
Se só usa Kubernetes, apague `deploy.yml`, `deploy-env.yml` e a pasta `infra/` do ECS e defina `K8S_DEPLOY_ENABLED=true`.

Em **Settings -> Environments** crie `dev`, `hml` e `prd`. Recomendado: **Required reviewers** em `prd` (e, se quiser, em `hml`).

| Variável | Onde | Obrigatória | Exemplo / padrão |
|---|---|---|---|
| `AWS_REGION` | repositório ou Environment | sim | `us-east-1` |
| `AWS_ROLE_ARN` | **por Environment** | sim | `arn:aws:iam::123456789012:role/kube-service-dev-deploy` |
| `EKS_CLUSTER` | Environment | não | padrão `kube-service-<ambiente>` |
| `ECR_REPOSITORY` | repositório | não | padrão `kube-service` |
| `K8S_DEPLOY_ENABLED` | repositório | não | `true` liga o deploy no push da `main` |

A role de cada ambiente confia no OIDC do GitHub (restrinja repositório **e** Environment), pode `eks:DescribeCluster`
e precisa de permissão no cluster: crie uma *access entry* para ela e associe uma política de acesso restrita ao
namespace do ambiente (`AmazonEKSEditPolicy` com escopo `kube-service-dev`, por exemplo). O primeiro deploy também cria o
namespace, então esse escopo precisa permitir isso ou o namespace deve existir antes.

O `ci.yml` renderiza os três overlays a cada PR: um patch quebrado derruba o CI, não o deploy.

## Segurança

- Pods como UID `10001`, sem escalar privilégios, sem capabilities, `seccomp` `RuntimeDefault` e *root filesystem*
  somente leitura (o `/tmp` é um `emptyDir`).
- Sem token da API do Kubernetes montado no pod.
- NetworkPolicy: entrada só do ingress-nginx; saída para DNS, HTTPS (443) e, quando há banco ou Redis, a porta deles.
  Para restringir o destino, acrescente `to: - ipBlock: cidr: ...` em `base/networkpolicy.yaml`.
- Nenhum segredo no repositório; o limite de memória do pod acomoda a JVM (`MaxRAMPercentage=75`).
