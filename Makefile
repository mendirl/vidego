install:
	go get .

build:
	go build -o out/dedup-cli main.go

run:
	go run main.go


persist:
	go run main.go persist --path "/mnt/f/O"

dedup:
	go run main.go dedup


filter:
	go run main.go filter --path "/mnt/n/P/T" --out ""

putback:
	go run main.go putback

move:
	go run main.go move --source "/mnt/f/J" --destination "/mnt/e/J"