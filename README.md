# vvvorld
vvvorld backend

frontend here:

https://github.com/March-mitsuki/vvvorld-frontend

## how to use
1. `export GIN_MODE=release` export a environment variable to set gin mode
1. create file named `.env.production.local` in root dir
1. set up `DB_DSN` and `CORS_ORIGIN` in `.env.production.local` file
1. `go build -tags=jsoniter .` build a production version
1. `./vvvorld` run it

## todo
- [ ] db的dsn连接可以改为阅读envconfig而不是在modle内定义var
- [x] 更改user name为unique