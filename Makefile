.PHONY: docker
docker:
	@cd cmd
	@rm webook || true
	@set GOOS=linux
	@set GOARCH=arm
	@go build -o webook .
	@docker rmi -t flycash/webook:v0.0.1 .
	@docker build -t flycash/webook:v0.0.1 .