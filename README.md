# Goportunitties

Painel de vagas de tecnologia: uma API REST em Go (Gin + GORM + SQLite) e uma
interface web em React + TypeScript para acompanhar, publicar e organizar
oportunidades.

O build de produção é **um único binário**: o frontend compilado é embutido com
`embed.FS` e servido pelo próprio servidor Go, na mesma origem da API.

O projeto nasceu do conteúdo do canal
[Arthur404dev](https://www.youtube.com/@Arthur404dev) — os créditos pela base e
pela didática são dele. A partir daí ele vem sendo refatorado e expandido como
exercício de arquitetura.

## Estrutura do repositório

A raiz é o módulo Go (é o que define o caminho de import e o que toda a
ferramentaria da linguagem espera encontrar ali). O frontend é um projeto npm
próprio, com o seu próprio `package.json` e build — por isso vive numa pasta
irmã, e não espalhado pela raiz.

```
cmd/            binários (hoje só a API; o worker de ingestão entra aqui)
internal/       o código do servidor, uma pasta por camada
frontend/       a aplicação React
docs/           coleção do Postman
data/           o arquivo SQLite — criado em runtime, fora do versionamento
```

A única fronteira que os dois lados cruzam é `internal/web/dist`: o Vite compila
para lá porque `embed.FS` só embute o que está dentro do pacote Go. É o preço do
binário único.

## Arquitetura

```
cmd/api/            entrypoint: carrega config, monta as dependências e sobe o servidor
internal/
  config/           configuração por variáveis de ambiente (único lugar que lê o ambiente)
  logger/           log estruturado com log/slog
  database/         conexão SQLite (WAL + busy_timeout) e migração do schema
  model/            structs persistidas — sem tags json, o banco não é o contrato
  dto/              contratos de entrada/saída, validação e envelope de resposta
  repository/       acesso a dados via GORM (única camada que fala com o banco)
  service/          regras de negócio (não conhece Gin nem GORM)
  handler/          HTTP: lê a requisição, chama o service, mapeia o resultado no status
  middleware/       CORS, request ID e log de acesso
  router/           registro das rotas e fallback da SPA
  web/              frontend compilado, embutido no binário
```

As dependências apontam sempre para dentro: o `service` recebe um
`repository.OpeningRepository` (interface) pelo construtor, e o erro de domínio
`ErrOpeningNotFound` é traduzido em 404 apenas no `handler` — nenhuma camada
acima do repositório conhece o `gorm.ErrRecordNotFound`. Todas as operações
recebem `context.Context`, então uma requisição cancelada aborta a query.

O frontend segue a mesma ideia de separação:

```
frontend/src/
  api/            cliente HTTP e endpoints (único lugar que conhece a API)
  state/          contextos: coleção de vagas + modais da aplicação
  hooks/          consultas à API, debounce e tema
  components/
    ui/           primitivos genéricos (Button, Modal, Field, Toast, Pagination, ...)
    openings/     componentes de domínio (card, detalhe, formulário, filtros)
    dashboard/    gráfico e agregação por mês
    layout/       cabeçalho e rodapé
  pages/          Home, Vagas e Painel — leem o estado por hook, não por props
  styles/         design tokens e estilos globais
  types/          tipos espelhando os DTOs do Go
  utils/          formatação (moeda, datas, iniciais)
```

Busca, filtros, ordenação e paginação acontecem **no servidor**, em SQL. O
frontend só descreve a consulta.

## Como rodar

Pré-requisitos: **Go 1.22+** e **Node 22+**.

### Desenvolvimento (dois processos)

```bash
go run ./cmd/api
```

```bash
cd frontend && npm install && npm run dev
```

Abra <http://localhost:5173>. O Vite faz proxy de `/api` para o backend, então
não é preciso configurar mais nada. O banco SQLite é criado automaticamente em
`data/main.db` na primeira execução.

### Produção (um binário)

```bash
make build && ./goportunitties
```

Abra <http://localhost:8080> — a mesma porta serve a API e a interface.

### Docker

```bash
docker compose up --build
```

A imagem final é `distroless/static` rodando como usuário sem privilégios, com
o SQLite em um volume nomeado.

## Testes

```bash
make test
```

| Camada | Ferramenta | O que cobre |
|---|---|---|
| `repository` | `go test` + SQLite temporário | filtros, ordenação, paginação, facetas e agregações reais em SQL |
| `service` | `go test` + repositório fake | regras de negócio sem tocar no banco |
| `handler` | `go test` + `httptest` | status, validação e o que não pode vazar num 500 |
| `router` | `go test` + `httptest` | o contrato inteiro ponta a ponta, incluindo CORS e fallback da SPA |
| `frontend` | Vitest + Testing Library | cliente HTTP, tradução de filtros, paginação e o formulário |

## Variáveis de ambiente

Todas têm valor padrão — o projeto roda sem configurar nenhuma.

| Variável | Padrão | Descrição |
|---|---|---|
| `APP_ENV` | `development` | `production` liga log JSON e o modo release do Gin |
| `PORT` | `8080` | Porta da API |
| `DB_PATH` | `./data/main.db` | Caminho do arquivo SQLite |
| `CORS_ORIGINS` | `http://localhost:5173,http://127.0.0.1:5173` | Origens autorizadas (separadas por vírgula) |
| `SHUTDOWN_TIMEOUT` | `10s` | Tempo dado às requisições em andamento no encerramento |
| `DEFAULT_PAGE_SIZE` | `12` | Tamanho de página quando o cliente não pede outro |
| `MAX_PAGE_SIZE` | `100` | Teto do tamanho de página |

No frontend, `VITE_API_URL` sobrescreve a URL da API caso o front seja
publicado separadamente do backend.

## Endpoints

Base: `/api/v1`

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/openings` | Lista as vagas — filtrada, ordenada e paginada |
| `GET` | `/openings/facets` | Contagem de cada opção de filtro para a consulta atual |
| `GET` | `/openings/stats` | Agregados do índice inteiro (painel) |
| `POST` | `/openings` | Cria uma vaga — responde **201** com o header `Location` |
| `GET` | `/openings/:id` | Detalha uma vaga |
| `PUT` | `/openings/:id` | Substitui a vaga inteira (todos os campos obrigatórios) |
| `PATCH` | `/openings/:id` | Atualiza só os campos enviados |
| `DELETE` | `/openings/:id` | Remove uma vaga (soft delete) |

Fora do versionamento: `GET /healthz` (o processo está vivo) e `GET /readyz`
(o banco responde).

### Parâmetros da listagem

| Parâmetro | Valores | Padrão |
|---|---|---|
| `search` | texto (cargo ou empresa) | — |
| `location` | localidade exata | todas |
| `remote` | `true` / `false` | ambas |
| `minSalary` | inteiro ≥ 0 | `0` |
| `sort` | `recent`, `salary-desc`, `salary-asc`, `role`, `updated` | `recent` |
| `page` / `pageSize` | inteiros | `1` / `12` |

Filtro inválido é **422**; janela de paginação fora de faixa é ajustada para o
limite mais próximo, e a janela efetiva volta em `meta`.

### Envelope

```json
{
  "message": "list-openings successful",
  "data": [{ "id": 1, "role": "..." }],
  "meta": { "page": 1, "pageSize": 12, "total": 57, "totalPages": 5 }
}
```

```json
{
  "message": "validation failed",
  "error": "há campos inválidos na requisição",
  "fields": { "link": "informe uma URL começando com http:// ou https://" }
}
```

`fields` é o que permite ao formulário apontar o erro no campo certo em vez de
mostrar uma mensagem genérica.

### Corpo de criação

```json
{
  "role": "Desenvolvedor Back-end Go",
  "company": "Acme Corp",
  "location": "São Paulo, SP",
  "remote": true,
  "link": "https://acme.com/vagas/1",
  "salary": 15000
}
```

`salary` aceita `0` (a combinar). `link` só aceita `http://` ou `https://` —
esse valor vira o `href` de um link na interface.

## Testando a API

Há uma coleção pronta para importar no Postman em
`docs/Goportunitties.postman_collection.json`, com os casos de sucesso e os de erro
(422 de validação, 404 de registro inexistente, 400 de id inválido).

## Interface

A interface cobre o CRUD completo: listagem paginada com busca, filtros por
modalidade, salário e localidade — com contagem por opção calculada no
servidor —, ordenação, painel de detalhe, formulário com validação espelhando a
do backend (e exibindo os erros que só o servidor consegue julgar), confirmação
de exclusão, estados de carregamento e vazio, notificações, tema claro/escuro e
layout responsivo.

## Próximo passo

O projeto está preparado para deixar de ser um CRUD e virar um agregador: um
worker concorrente que busca vagas em fontes públicas, normaliza, deduplica e
alimenta o índice sozinho. A paginação, os índices do banco, o WAL e o
`context.Context` propagado existem para que essa etapa não precise reescrever
o que já está aqui.
