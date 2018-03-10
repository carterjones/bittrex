build: bindata
	go build ./...

test:
	go test -race ./...

cover:
	go test -coverprofile=c.out -covermode=atomic -race ./...

lint:
	$(GOPATH)/bin/gometalinter \
	--cyclo-over 12 \
	--disable gotype \
	--disable gotypex \
	--enable nakedret \
	--exclude "/usr/local/go/src/" \
	--exclude "internal/bindata.go" \
	--vendor \
	./...

bindata: bindata_install
	go-bindata -pkg internal -o ./internal/bindata.go test-fixtures
	go fmt ./internal/bindata.go

bindata_install:
	which bindata &>/dev/null || go get -u github.com/hashicorp/go-bindata/...

update:
	dep ensure -update
	git checkout -- vendor/github.com/gorilla # This is due to whitespace issues.
	pre-commit autoupdate
