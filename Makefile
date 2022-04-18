
all:check int-test

mock-test:
	go clean -cache
	go test -v -coverprofile=c.out --run=TestMock

int-test:
	GONVMETCP_PORTAL=$(TCPPortal) GONVME_TARGET=$(Target) GONVMEFC_PORTAL=$(FCPortal) GONVMEFC_HostAddress=$(FCHostAddress) \
		 go test -v -timeout 20m -coverprofile=c.out -coverpkg ./...

gocover:
	go tool cover -html=c.out

check:
	gofmt -d .
	golint -set_exit_status
	go vet
