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
	@go mod tidy
