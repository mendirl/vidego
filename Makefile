mount:
	sudo mount -t  drvfs F: /mnt/f
	sudo mount -t  drvfs N: /mnt/n

install:
	go get .

build:
	go build -o out/dedup-cli main.go

run:
	go run main.go

persist:
	go run main.go persist --path "/mnt/n/N"

sorting:
	go run main.go sort --paths "/mnt/n/O,/mnt/d/O,/mnt/g/O"

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