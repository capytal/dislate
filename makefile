PORT?=8080

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1 run

build:
	go build -o bin/dislate

dev/watch:
	go run github.com/joho/godotenv/cmd/godotenv@v1.5.1 \
		go run github.com/air-verse/air@v1.52.2 \
			--build.cmd "go build -o tmp/bin/dislate" \
			--build.bin "tmp/bin/dislate" \
			--build.include_ext "go" \
			--build.stop_on_error "false" \
			--build.send_interrupt "true" \
			--misc.clean_on_exit true \
			-- -p $(PORT) -d

dev:
	go run github.com/joho/godotenv/cmd/godotenv@v1.5.1 \
		go run .

run: build
	./bin/dislate

clean:
	if [[ -d "dist" ]]; then rm -r ./dist; fi
	if [[ -d "tmp" ]]; then rm -r ./tmp; fi
	if [[ -d "bin" ]]; then rm -r ./bin; fi
