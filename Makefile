install:
	go get .

build:
	go build -o out/dedup-cli main.go

run:
	go run main.go

persist:
	go run main.go persist --path "/mnt/d/O,/mnt/e/O,/mnt/f/O,/mnt/h/O,/mnt/n/O,/mnt/n/N"

dedup:
	go run main.go dedup

filtering:
	go run main.go filter --path "/mnt/f/T"

putback:
	go run main.go putback

delete:
	go run main.go delete

move:
	go run main.go move --source "/mnt/f/J" --destination "/mnt/e/J"

organizeDb:
	go run main.go organize

organize:
	go run main.go organize --path "/mnt/e/T"