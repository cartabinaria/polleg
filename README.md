# `polleg`

Progetto che permette agli studenti di rispondere agli esercizi delle prove
d'esame direttamente sul sito di CartaBinaria.

## Todo

- gestire l'upload di immagini
- migliorare la documentazione delle api con swaggo
- implementare per bene il sistema delle proposte
- aggiungere un api post per la modifica di una risposta

## Usage

```golang
go run cmd/polleg.go
```

To generate the swagger documentation use

```shell
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseDependency -g cmd/polleg.go
swag fmt -g cmd/polleg.go
```
