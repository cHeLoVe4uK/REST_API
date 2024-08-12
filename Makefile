run:
	go run cmd/restapi/main.go

lint:
	golangci-lint run --disable-all -E unused -E gofumpt -E govet -E errcheck

fix:
	golangci-lint run --disable-all -E gofumpt --fix

# migrate -path migrations -database "postgres://localhost:port/db_name?sslmode=disable&user=user_name&password=your_password" down or up