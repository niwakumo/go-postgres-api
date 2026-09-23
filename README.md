# Prepare Environment

## Go install

```shell
brew install go
go version

echo 'export GOPATH=$HOME/go' >> ~/.zshrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.zshrc

source ~/.zshrc
cat ~/.zshrc | grep GO

go env GOARCH GOOS
```

## Postgres Setups

```shell
mkdir -p docker/init

# make docker-compose.yml &  .sql files for initializaiton.
docker compose up -d

# Ensure initialization.
 docker exec -it dev-postgres psql -U postgres -d devdb -c "SELECT * FROM users;"

```

## Go Setups

> [!IMPORTANT] ソースコードを修正したらこのコマンド
> `go mod tidy`

```shell
# モジュール初期化
go mod init go-postgres-api

# 軽量Webルーター (chi) のインストール
go get -u github.com/go-chi/chi/v5

# PostgreSQL ドライバ (pgx) のインストール
go get -u github.com/jackc/pgx/v5/stdlib

# go-postgres-api
mkdir -p internal/handler internal/service internal/repository
```

```shell
go run main.go

# another terminal
curl -i http://localhost:8080/users/1

curl -i -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Charlie", "email": "charlie@example.com"}'

# error handling
curl -i http://localhost:8080/users/999

```

## Testcontainers

```shell
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres

# ソースコードを修正したらこのコマンド
go mod tidy
```

> [!IMPORTANT] Rancher Desktopを起動したらこの設定を入れる

```shell
# 設定確認
docker context inspect --format '{{.Endpoints.docker.Host}}'
# unix:///Users/{ユーザー名}/.rd/docker.sock の場合は以下
export DOCKER_HOST=$(docker context inspect --format '{{.Endpoints.docker.Host}}')
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock

# 設定を反映
source ~/.zshrc

# テスト実行
go test -v ./internal/repository/...

```

## GitHub　repository

```shell
# 1. Git初期化
git init

# 2. デフォルトブランチを main に設定
git branch -M main

# 3. ファイルをステージング
git add .

# 4. 初回コミット
git commit -m "feat: initial commit with API and testcontainers"
```
## GitHub Actions

```shell
mkdir -p .github/workflows
```

