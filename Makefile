compile:
	CGO_ENABLED=0 GOOS=linux GOARCH='amd64' go build -v -o ./websub ./cmd/websub/main.go

essentials:
	go install github.com/golang/mock/mockgen

mock: essentials
	mockgen -package=hub -source=pkg/hub/h.go -destination=pkg/hub/mock.go

unit: mock
	go test --race -gcflags=-l -v -coverprofile .coverage.out.tmp ./...
	cat .coverage.out.tmp | grep -v "mock.go" > .coverage.out
	rm -rf .coverage.out.tmp
	go tool cover -func .coverage.out

