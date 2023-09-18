.PHONY: docker
docker:
	@cd cmd
	@rm webook || true
	@set GOOS=linux
	@set GOARCH=arm
	@go build -o webook .
	@docker rmi -t flycash/webook:v0.0.1 .
	@docker build -t flycash/webook:v0.0.1 .

.PHONY: mock
mock:
	@mockgen -source=internal/service/user.go -package=svcmocks -destination=internal/service/mocks/user_gen.go
	@mockgen -source=internal/service/code.go -package=svcmocks -destination=internal/service/mocks/code_gen.go
	@mockgen -source=internal/repository/user.go -package=repomocks -destination=internal/repository/mocks/user_gen.go
	@mockgen -source=internal/repository/code.go -package=repomocks -destination=internal/repository/mocks/code_gen.go
	@mockgen  -package=redismocks -destination=internal/repository/redismocks/code_gen.go github.com/redis/go-redis/v9 Cmdable
	@go mod tidy
