mount:
	sudo mount -t  drvfs F: /mnt/f
	sudo mount -t  drvfs N: /mnt/n
	sudo mount -t  drvfs D: /mnt/d
	sudo mount -t  drvfs E: /mnt/e

install:
	go get .

build:
	go build -o out/dedup-cli main.go

run:
	go run main.go

persist:
	go run main.go persist --path "/mnt/f/O"

sorting:
	go run main.go sort --paths "/mnt/f/T" --move=false

#,/mnt/c/Users/fabien/Downloads/N,/mnt/c/Users/fabien/Downloads/T

sortingAll:
	go run main.go sort --paths "/mnt/c/Users/fabien/Downloads/T,/mnt/c/Users/fabien/Downloads/O,/mnt/c/Users/fabien/Downloads/N,/mnt/d/T,/mnt/d/N,/mnt/d/O,/mnt/e/T,/mnt/e/N" --move=false

dedup:
	go run main.go sort --paths "/mnt/d/O/O15_over70,/mnt/d/O/O16_over90,/mnt/d/O/O14_under70,/mnt/d/O/O13_under65" --move=false
	go run main.go dedup C:\Users\fabien\Downloads

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
