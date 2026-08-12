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
cmd/api/        o servidor HTTP
cmd/worker/     o worker de ingestão
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
  ingestion/        leitura dos portais de vagas (uma fonte por arquivo)
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

### Ingestão

```bash
make ingest
```

Roda uma passada e sai. Para deixá-lo rodando na periodicidade configurada:

```bash
go run ./cmd/worker
```

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
| `ingestion` | `go test` + `httptest` | mapeamento de cada portal, concorrência, timeout e falha isolada |
| `frontend` | Vitest + Testing Library | cliente HTTP, tradução de filtros, paginação, procedência e o formulário |

## Variáveis de ambiente

Todas têm valor padrão — o projeto roda sem configurar nenhuma, **exceto
`ADMIN_KEY` em produção**, que é obrigatória.

| Variável | Padrão | Descrição |
|---|---|---|
| `APP_ENV` | `development` | `production` liga log JSON e o modo release do Gin |
| `PORT` | `8080` | Porta da API |
| `DB_PATH` | `./data/main.db` | Caminho do arquivo SQLite |
| `CORS_ORIGINS` | `http://localhost:5173,http://127.0.0.1:5173` | Origens autorizadas (separadas por vírgula) |
| `ADMIN_KEY` | vazio | Chave que libera criar/editar/excluir vagas — ver abaixo |
| `SHUTDOWN_TIMEOUT` | `10s` | Tempo dado às requisições em andamento no encerramento |
| `DEFAULT_PAGE_SIZE` | `12` | Tamanho de página quando o cliente não pede outro |
| `MAX_PAGE_SIZE` | `100` | Teto do tamanho de página |
| `INGESTION_SOURCES` | `backend-br,frontend-br,remotive` | Portais a ler, separados por vírgula |
| `INGESTION_INTERVAL` | `6h` | Intervalo entre passadas; `0` roda uma vez e sai |
| `INGESTION_TIMEOUT` | `30s` | Tempo máximo por portal |
| `INGESTION_MAX_PER_SOURCE` | `100` | Teto de vagas que um portal contribui por passada |
| `INGESTION_USER_AGENT` | identificação do projeto | User-Agent enviado aos portais |
| `INGESTION_GITHUB_TOKEN` | vazio | Opcional; eleva o limite do GitHub de 60 para 5000 req/h |

No frontend, `VITE_API_URL` sobrescreve a URL da API caso o front seja
publicado separadamente do backend.

## Endpoints

Base: `/api/v1`

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/openings` | Lista as vagas — filtrada, ordenada e paginada |
| `GET` | `/openings/facets` | Contagem de cada opção de filtro para a consulta atual |
| `GET` | `/openings/stats` | Agregados do índice inteiro (painel) |
| `POST` 🔒 | `/openings` | Cria uma vaga — responde **201** com o header `Location` |
| `GET` | `/openings/:id` | Detalha uma vaga |
| `PUT` 🔒 | `/openings/:id` | Substitui a vaga inteira (todos os campos obrigatórios) |
| `PATCH` 🔒 | `/openings/:id` | Atualiza só os campos enviados |
| `DELETE` 🔒 | `/openings/:id` | Remove uma vaga (soft delete) |

Fora do versionamento: `GET /healthz` (o processo está vivo) e `GET /readyz`
(o banco responde).

### Rotas de escrita (🔒)

Toda leitura é pública; toda escrita exige o header `X-Admin-Key` batendo com
`ADMIN_KEY`. Sem isso, qualquer visitante do site conseguiria criar vagas
falsas, editar ou apagar as reais que a ingestão traz — não é hipotético, é o
primeiro teste que qualquer pessoa faz ao abrir um CRUD público.

Em desenvolvimento, sem `ADMIN_KEY` configurada, a trava vira um no-op — é o
que mantém `go run ./cmd/api` funcionando sem configuração nenhuma. Em
produção (`APP_ENV=production`), o binário **recusa subir** sem a variável
definida — a ausência falha alto, no boot, em vez de silenciosamente deixar
tudo aberto.

Para escrever pelo navegador como administrador, defina a chave uma vez no
console (`localStorage.setItem('goportunitties:admin-key', 'a-mesma-chave-do-servidor')`)
— não existe tela de login, o projeto tem um operador, não usuários.

### Parâmetros da listagem

| Parâmetro | Valores | Padrão |
|---|---|---|
| `search` | texto (cargo ou empresa) | — |
| `location` | localidade exata | todas |
| `source` | `manual` ou o slug de um portal | todas |
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

## Ingestão

O worker lê portais públicos e alimenta o índice sozinho. Cada portal é um
arquivo em `internal/ingestion/` que implementa uma interface de três métodos;
o orquestrador não conhece portal nenhum.

| Portal | O que traz |
|---|---|
| `backend-br` | vagas brasileiras de back-end ([backend-br/vagas](https://github.com/backend-br/vagas)) |
| `frontend-br` | vagas brasileiras de front-end ([frontendbr/vagas](https://github.com/frontendbr/vagas)) |
| `remotive` | vagas remotas internacionais |
| `remoteok` | disponível, mas **desligado por padrão** — ver abaixo |

Os dois primeiros são quadros da comunidade que rodam sobre **GitHub Issues**:
uma issue é uma vaga. Isso dá um feed público, gratuito e sem chave de vagas
brasileiras reais — que os portais brasileiros tradicionais (Gupy, Vagas.com,
Catho) não oferecem. O custo é que o título é prosa, não campos: o adaptador lê
uma convenção humana (`[Modalidade - Cidade] Cargo - Empresa`) e **descarta** a
vaga cujo empregador não consegue identificar, em vez de gravar um palpite.

O `remoteok` continua implementado e testado, mas fora do conjunto padrão: o
feed gratuito dele mediu cinco vagas de tecnologia em cem entradas, sendo o
resto trabalho de varejo e páginas de erro servidas como se fossem vagas
("Page Not Found", "YOUR JOB DESCRIPTION HERE"). Para religá-lo, basta
adicioná-lo a `INGESTION_SOURCES`.

Os portais são consultados em paralelo — é I/O contra serviços independentes —
mas a escrita passa por um funil único, porque o SQLite serializa escritores de
qualquer forma e um segundo só trocaria paralelismo por disputa de lock. Cada
lote vai numa transação própria: um portal que falha no meio não deixa nada
para trás.

Rodar duas vezes não duplica nada. A chave `(source, external_id)` faz a
segunda passada atualizar as linhas em vez de inseri-las, e disso decorrem duas
garantias: a ingestão não alcança vaga cadastrada à mão, e vaga que você
excluiu não volta.

### Uma limitação conhecida

Vagas importadas entram como **"a combinar"**. Os portais publicam salário em
dólar por ano ou como texto livre (`"$50k - $70k"`, `"competitive"`), e esta
aplicação guarda um inteiro em reais por mês. Converter exigiria inventar
câmbio e jornada, e o número inventado ficaria indistinguível dos reais nas
estatísticas. Por isso a mediana e a média já ignoram zeros.

Resolver isso de verdade pede `currency` e `period` no modelo — a mudança
atravessa DTO, filtro de salário mínimo, facetas e formatação no frontend, e
por isso ficou de fora daqui.

## Próximo passo

Três pontos, em ordem de impacto.

**Vagas expiradas não saem do índice.** O upsert atualiza o que a fonte ainda
publica, mas uma vaga que sai do ar simplesmente para de ser atualizada e fica
no índice para sempre. Marcar como encerrada o que não aparece há N passadas
resolve.

**A busca não escala.** `LIKE '%termo%'` não usa índice: é varredura completa.
Com centenas de vagas é irrelevante, com dezenas de milhares vira o gargalo.
FTS5 resolve.

**A faceta de localidade explode.** Ela lista todos os valores distintos, o que
era razoável num índice curado e deixa de ser com dados importados. Limitar às
localidades mais frequentes resolve.
