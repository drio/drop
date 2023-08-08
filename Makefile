PROD_SERVER:=bubbles
PRJ_NAME=drop
HOST?=$(shell hostname)
DROP_PORT?=9191
APP_NAME=drio-drop

# Load env variables when in prod
ifeq ($(HOST), bubbles)
include .env.prod
export $(shell sed 's/=.*//' .env.prod)
else
include .env.dev
export $(shell sed 's/=.*//' .env.dev)
endif

.PHONY: run open lint
lint:
	golangci-lint run

test:
	@ls *.go | entr -c -s 'go test -failfast -v ./*.go && notify "💚" || notify "🛑"'

single-run-test:
	go test -failfast -v *.go

coverage:
	@go test -cover ./...

coverage/html:
	go test -v -cover -coverprofile=c.out
	go tool cover -html=c.out

deploy:
	fly deploy -a $(APP_NAME)

.PHONY: check-vars
check-vars:
	echo u:$$KAE_USER p:$$KAE_PASS db:$$KAE_DB ds:$$KAE_DELAY_SECS p:$$PORT

air:
	air

run:
	go run .

main-linux-amd64:
	GOARCH=amd64 GOOS=linux go build -ldflags "-w" -o $@ .

clean: 
	rm -f $(PRJ_NAME) main-linux-amd64 main

docker/rm:
	docker image rm --force $(PRJ_NAME)

docker/build: main-linux-amd64
	docker build -t $(PRJ_NAME) .

docker/run:
	docker run -p $(DROP_PORT):$(DROP_PORT) drop
