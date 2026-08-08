# Goportunitties

Painel de vagas de tecnologia: uma API REST em Go (Gin + GORM + SQLite) e uma
interface web em React + TypeScript para acompanhar, publicar e organizar
oportunidades.

O projeto nasceu do conteúdo do canal
[Arthur404dev](https://www.youtube.com/@Arthur404dev) — os créditos pela base e
pela didática são dele. A partir daí ele vem sendo refatorado e expandido como
exercício de arquitetura.

## Arquitetura

O backend é dividido em camadas, cada uma com uma responsabilidade:

```
handler/     HTTP: lê a requisição, chama o service, devolve a resposta
service/     regras de negócio (não conhece Gin nem GORM)
repository/  acesso a dados via GORM (a única camada que fala com o banco)
dto/         contratos de entrada/saída da API + envelope de resposta
schemas/     structs persistidas (modelo do banco)
middleware/  CORS
config/      configuração por variáveis de ambiente, logger e conexão SQLite
router/      registro das rotas
```

O frontend segue a mesma ideia de separação:

```
frontend/src/
  api/         cliente HTTP e endpoints (único lugar que conhece a API)
  hooks/       estado e efeitos (useOpenings, useTheme)
  components/
    ui/        primitivos genéricos (Button, Modal, Field, Toast, ...)
    openings/  componentes de domínio (card, detalhe, formulário, filtros)
    layout/    cabeçalho
  styles/      design tokens e estilos globais
  types/       tipos espelhando os DTOs do Go
  utils/       formatação (moeda, datas, iniciais)
```

## Como rodar

Pré-requisitos: **Go 1.22+** e **Node 20+**.

### 1. Backend (porta 8080)

```bash
go mod tidy && go run .
```

O banco SQLite é criado automaticamente em `database/main.db` na primeira
execução.

### 2. Frontend (porta 5173)

```bash
cd frontend && npm install && npm run dev
```

Abra <http://localhost:5173>. O Vite faz proxy de `/api` para o backend, então
não é preciso configurar mais nada em desenvolvimento.

## Variáveis de ambiente

Todas têm valor padrão — o projeto roda sem configurar nenhuma.

| Variável | Padrão | Descrição |
|---|---|---|
| `PORT` | `8080` | Porta da API |
| `DB_PATH` | `./database/main.db` | Caminho do arquivo SQLite |
| `CORS_ORIGINS` | `http://localhost:5173,http://127.0.0.1:5173` | Origens autorizadas (separadas por vírgula) |

No frontend, `VITE_API_URL` sobrescreve a URL da API caso o front seja
publicado separadamente do backend.

## Endpoints

Base: `/api/v1`

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/openings` | Lista todas as vagas |
| `GET` | `/opening/:id` | Detalha uma vaga |
| `POST` | `/opening` | Cria uma vaga |
| `PUT` | `/opening/:id` | Atualiza parcialmente uma vaga |
| `DELETE` | `/opening/:id` | Remove uma vaga (soft delete) |

Toda resposta usa o mesmo envelope:

```json
{ "message": "create-opening successful", "data": { "id": 1, "role": "..." } }
{ "message": "operation failed", "error": "opening not found" }
```

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

`PUT` aceita os mesmos campos, todos opcionais — ao menos um precisa ser enviado.

## Testando a API

Há uma coleção pronta para importar no Postman em
`Goportunitties.postman_collection.json`, com os casos de sucesso e de erro
(400 de validação, 404 de registro inexistente).

## Interface

A interface cobre o CRUD completo: listagem com busca, filtro por modalidade
(remoto/presencial) e localidade, ordenação, painel de detalhe, formulário de
criação/edição com validação espelhando a do backend, confirmação de exclusão,
estados de carregamento e vazio, notificações, tema claro/escuro e layout
responsivo.

## Build de produção

```bash
go build -o goportunitties . && cd frontend && npm run build
```

O frontend gera os arquivos estáticos em `frontend/dist/`.
