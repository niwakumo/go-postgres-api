# ハンズオン 1

## GoらしいテストとCIカバレッジレポート

```shell
go get github.com/stretchr/testify
go mod tidy
```

```shell
# 1. カバレッジプロファイル (coverage.out) を生成しながら全テストを実行
go test -v -coverprofile=coverage.out ./...

# 2. ターミナル上でパッケージごとのカバレッジ率を確認
go tool cover -func=coverage.out

# HTMLファイルを出力
go tool cover -html=coverage.out -o coverage.html

# Macのデフォルトブラウザで開く
open coverage.html
```

## golang-migrate によるスキーマ管理と自動適用

Testcontainersで読み込ませるスキーマ情報を別リポジトリにする

### go-postgres-schemaでの作業

```shell
# クローン (your-username をご自身のアカウント名に置き換えてください)
git clone https://github.com/<your-username>/go-postgres-schema.git
cd go-postgres-schema

# マイグレーション用フォルダの作成
mkdir -p migrations

git add migrations/
git commit -m "feat: add initial users schema migration"
git push origin main

```

### go-postgres-apiでの作業

Git Submoduleを追加

```shell
# internal/db 配下に submodule として追加
git submodule add https://github.com/<your-username>/go-postgres-schema.git internal/db/schema
```

golang-migrateパッケージの導入

```shell
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/pgx/v5
go get -u github.com/golang-migrate/migrate/v4/source/iofs
go mod tidy
```





