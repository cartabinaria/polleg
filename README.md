# Polleg

Polleg is a web service for students to answer exam exercises directly on the
CartaBinaria website.

## Usage

```golang
go run cmd/polleg.go <config-file>
```

To generate the swagger documentation use

```shell
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseDependency -g cmd/polleg.go
swag fmt -g cmd/polleg.go
```
