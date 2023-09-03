.PHONY: docker
docker:
	@if exist webook (del webook)
	@SET GOOS=linux&&SET GOARCH=arm&&go build -tags=k8s -o webook ./cmd/
	@docker rmi -f mokou/webook:v0.0.1 || true
	@docker build -t mokou/webook:v0.0.1 .
