# this is satla backend repo

overview here:
https://github.com/March-mitsuki/satla

frontend here:
https://github.com/March-mitsuki/satla-frontend

## how to use

1. `export GIN_MODE=release` export a environment variable to set gin mode
1. create file named `.env.production.local` in root dir
1. set up `DB_DSN` and `CORS_ORIGIN` in `.env.production.local` file
1. `go build -tags=jsoniter .` build a production version
1. `./vvvorld` run it

## todo

- [ ] db 的 dsn 连接可以改为阅读 envconfig 而不是在 modle 内定义 var
- [x] 更改 user name 为 unique
