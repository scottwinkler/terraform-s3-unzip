# Go parameters
BINARY_NAME=deployment
all: clean deps build package
build: 
	cd golang && GOOS=linux go build -o ${BINARY_NAME}
package:
	cd golang && zip deployment.zip ${BINARY_NAME}
clean: 
	go clean
	rm -f golang/$(BINARY_NAME)
	rm -f golang/$(BINARY_NAME).zip
deps:
	cd golang && dep ensure