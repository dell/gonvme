
# These values should be set for running the entire test suite
# all must be valid
Portal="1.1.1.1"
Target="nqn.1988-11.com.mock:00:e6e2d5b871f1403E169D"


all:check int-test

mock-test:
	go clean -cache
	go test -v -coverprofile=c.out --run=TestMock

int-test:
	GONVME_PORTAL=$(Portal) GONVME_TARGET=$(Target)  \
		 go test -v -timeout 20m -coverprofile=c.out -coverpkg ./...

gocover:
	go tool cover -html=c.out

check:
	gofmt -d .
	golint -set_exit_status
	go vet
