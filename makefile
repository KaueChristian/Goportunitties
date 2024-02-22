.PHON : deafult run build test docs clean

APP_NAME=goportunitties

deafult: run 

run:
	@go run main.go
build:
	@go build -o $(APP_NAME) main.go
test:
	@go test ./ ...
docs:
	@swag init
clean:
	@rm -f $(APP_NAME)
	@rm -tf ./docs