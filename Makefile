INTEGRATION_TEST_PATH?=./tests

ENV_LOCAL_TEST=\
  MONGO_DB=mongodb://dbtest:supra**@localhost:27017/?authSource=admin

docker.start:
	cd ./ && docker compose up -d

docker.stop:
	cd ./ && docker compose down -v

test.integration:
	$(ENV_LOCAL_TEST) \
    go test -tags=integration $(INTEGRATION_TEST_PATH) -count=1 -run=$(INTEGRATION_TEST_SUITE_PATH)

    # in future this command will trigger integration test with verbose mode
test.integration.debug:
	$(ENV_LOCAL_TEST) \
    go test -tags=integration $(INTEGRATION_TEST_PATH) -count=1 -v -run=$(INTEGRATION_TEST_SUITE_PATH)

run.tests:
	go test ./...

build-proxy:
	cd cmd/proxy && go build -o proxy && cd ../..

