# fecha-automatico-leilao

Sistema de leilões em Go (Gin + MongoDB) com **fechamento automático**: ao ser
criado, cada leilão dispara uma goroutine em background que aguarda o tempo
configurado em `AUCTION_INTERVAL` e então atualiza seu status para `Closed`
no banco, sem qualquer intervenção manual.

## Como rodar o projeto (Docker / Docker Compose)

Pré-requisitos: Docker e Docker Compose.

```bash
docker compose up --build
```

Isso sobe dois serviços:

- `app`: a API HTTP, em `http://localhost:8080`
- `mongodb`: banco MongoDB, em `localhost:27017`

Endpoints principais:

| Método | Rota                          | Descrição                          |
|--------|-------------------------------|-------------------------------------|
| POST   | `/auction`                    | Cria um leilão                      |
| GET    | `/auction`                    | Lista leilões (filtros: `status`, `category`, `productName`) |
| GET    | `/auction/:auctionId`         | Busca um leilão por id              |
| GET    | `/auction/winner/:auctionId`  | Busca o lance vencedor de um leilão |
| POST   | `/bid`                        | Cria um lance                       |
| GET    | `/bid/:auctionId`              | Lista lances de um leilão           |
| GET    | `/user/:userId`               | Busca um usuário                    |

`status` do leilão: `0` = Active (aberto), `1` = Completed (fechado).

## Variáveis de ambiente

Definidas em [cmd/auction/.env](cmd/auction/.env) e carregadas tanto pela
aplicação quanto pelo `docker-compose.yml`:

| Variável                     | Descrição                                                                 | Exemplo   |
|-------------------------------|----------------------------------------------------------------------------|-----------|
| `AUCTION_INTERVAL`            | Duração de um leilão. Passado o tempo, ele é fechado automaticamente.     | `20s`     |
| `BATCH_INSERT_INTERVAL`       | Intervalo máximo de espera antes de gravar um lote de lances no banco.    | `20s`     |
| `MAX_BATCH_SIZE`               | Quantidade de lances acumulados que dispara a gravação em lote.           | `4`       |
| `MONGO_INITDB_ROOT_USERNAME`  | Usuário root do MongoDB (usado pela imagem oficial do Mongo).             | `admin`   |
| `MONGO_INITDB_ROOT_PASSWORD`  | Senha root do MongoDB.                                                    | `admin`   |
| `MONGODB_URL`                 | String de conexão usada pela aplicação para acessar o MongoDB.            | `mongodb://admin:admin@mongodb:27017/auctions?authSource=admin` |
| `MONGODB_DB`                  | Nome do banco usado pela aplicação.                                       | `auctions`|

`AUCTION_INTERVAL` e `BATCH_INSERT_INTERVAL` aceitam qualquer valor
compatível com [`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration)
(ex: `20s`, `5m`, `1h`). Se a variável estiver ausente ou inválida,
`AUCTION_INTERVAL` assume o padrão de 5 minutos.

Para alterar o tempo de duração dos leilões, edite `AUCTION_INTERVAL` em
`cmd/auction/.env` e reinicie a aplicação (`docker compose up --build`).

## Rodando localmente sem Docker

```bash
go mod download
go run cmd/auction/main.go
```

Requer um MongoDB acessível na URL definida em `MONGODB_URL` (por padrão
aponta para o hostname `mongodb`, válido dentro da rede do Docker Compose;
fora do Docker, aponte para `localhost:27017` com as mesmas credenciais).

## Testes

```bash
# sobe só o mongodb para o teste de integração do fechamento automático
docker compose up -d mongodb

go vet ./...
go test ./...
```

O teste
[`internal/infra/database/auction/create_auction_test.go`](internal/infra/database/auction/create_auction_test.go)
é um teste de integração que:

1. Configura `AUCTION_INTERVAL=2s` (curto, só para o teste rodar rápido).
2. Cria um leilão e confirma que nasce com status `Active`.
3. Aguarda o intervalo configurado.
4. Confirma que o status foi alterado para `Closed` automaticamente pela
   goroutine de fechamento, sem nenhuma chamada manual de atualização.

Ele se conecta por padrão a `mongodb://admin:admin@localhost:27017/auctions?authSource=admin`
(configurável via `TEST_MONGODB_URL`) e é pulado automaticamente (`SKIP`) se
nenhum MongoDB estiver acessível.
