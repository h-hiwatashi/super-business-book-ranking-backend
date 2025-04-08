# alpineはバージョンを実質的に固定にできないためDebianベースを使用する
# Builder
FROM golang:1.24-bullseye as builder
# アップデートとgitのインストール
# goを扱うにはgitが必須であるがalpyneやDebianは軽量化のためgitが入っていない
RUN apt update && apt install -y git

WORKDIR /app

# Go Modulesの依存関係を事前にインストール
# COPY go.mod go.sum ./
# RUN go mod download && go mod verify

# # ソースコードをコピー
# COPY . .

# # アプリケーションをビルド
# RUN go build -o main .

# # 実行時環境変数を設定
# ENV GO_ENV=production

# # アプリケーションを実行
# CMD ["./main"]