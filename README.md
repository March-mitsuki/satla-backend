# this is satla backend repo

overview here:
https://github.com/March-mitsuki/satla

frontend here:
https://github.com/March-mitsuki/satla-frontend

# how to dev

1. `docker compose -p satla -f ./_docker_dev/docker-compose.yaml up -d`
2. set up `DB_DSN`, `CORS_ORIGIN`, `PORT` in `.env.development` file
3. `go run .`

## how to use

1. clone this repository
1. `export GIN_MODE=release` export a environment variable to set gin mode
1. create file named `.env.production.local` in root dir
2. set up `DB_DSN`, `CORS_ORIGIN`, `PORT` in `.env.production.local` file
3. `go build -tags=jsoniter .` build a production version
4. `./vvvorld` run it

## todo

- [x] db 的 dsn 连接可以改为阅读 envconfig 而不是在 modle 内定义 var
- [x] 更改 user name 为 unique
- [ ] 把 api 内的 db 操作移到 db 里面
