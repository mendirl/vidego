install:
	go get .

build:
	go build -o out/dedup-cli main.go

run:
	go run main.go

persist:
	go run main.go persist --path "/mnt/d,/mnt/e,/mnt/f,/mnt/n"

dedup:
	go run main.go dedup

filter:
	go run main.go filter --path "/mnt/f/T"

putback:
	go run main.go putback

delete:
	go run main.go delete

move:
	go run main.go move --source "/mnt/f/J" --destination "/mnt/e/J"