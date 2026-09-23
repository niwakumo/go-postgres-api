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