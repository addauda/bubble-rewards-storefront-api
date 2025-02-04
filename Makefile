build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/validate validate/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/redeem redeem/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/heartbeat heartbeat/main.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

.PHONY: deploy
deploy: clean build
	sls deploy --verbose
