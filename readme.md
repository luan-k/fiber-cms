# go module

- first we need to initialize our module. run `go mod init github.com/go-live-cms/go-live-cms`

- then we run `go mod tidy`

- after that, we should be able to run `go run main.go`

# migrate

- first we need to use this https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#linux-deb-package

- I have it installed on both windows and wsl

- to install on wsl I used these commands
  - `curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz`
  - `sudo mv migrate.linux-amd64 $GOPATH/bin/migrate`
  - then run `migrate --version` to see if it installed correctly

# DB queries using [sqlc](https://docs.sqlc.dev/en/stable/overview/install.html)

- run `sudo snap install sqlc` on your wsl, sqlc is not available on windows

- by running `sqlc init` we can initialize sqlc

- run `make sqlc` to generate new sql code
