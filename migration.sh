goose -dir ./migrations postgres "postgres://user:password@localhost:5432/db?sslmode=disable" status
goose -dir ./migrations postgres "postgres://user:password@localhost:5432/db?sslmode=disable" up