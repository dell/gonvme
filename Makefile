
# These values should be set for running the entire test suite
# all must be valid
TCPPortal="1.1.1.1"
FCPortal="nn-0x11aaa111a1111a1a:pn-0x11aaa11111111a1a"
Target="nqn.1988-11.com.mock:00:e6e2d5b871f1403E169D"
FCHostAddress="nn-0x11aaa111a1111a1a:pn-0x11aaa11111111a1a"


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
