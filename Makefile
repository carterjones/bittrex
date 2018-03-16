build: bindata
	go generate ./...
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

bindata:
	which go-bindata &>/dev/null || go get -u github.com/kevinburke/go-bindata/...

update:
	dep ensure -update
	git checkout -- vendor/github.com/gorilla # This is due to whitespace issues.
	pre-commit autoupdate
