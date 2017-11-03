build:
	mkdir -p ./bin
	go build -o ./bin/producer ./producer
	go build -o ./bin/consumer ./consumer