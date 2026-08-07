mount:
	sudo mount -t  drvfs F: /mnt/f
	sudo mount -t  drvfs N: /mnt/n
	sudo mount -t  drvfs D: /mnt/d
	sudo mount -t  drvfs E: /mnt/e
	sudo mount -t  drvfs X: /mnt/x
	sudo mount -t  drvfs J: /mnt/j

install:
	go get .

build:
	go build -o out/dedup-cli main.go

run:
	go run main.go

persist:
	go run main.go persist --path "/mnt/c/Users/fabien/Downloads/T,/mnt/c/Users/fabien/Downloads/O","/mnt/e/data/T","/mnt/f/T"

sorting:
# /mnt/c/Users/fabien/Downloads/T,
# /mnt/c/Users/fabien/Downloads/O,
# /mnt/e/O,
# /mnt/e/O/O6_under30,
# /mnt/e/T,
# /mnt/e/data/O,
# /mnt/e/data/T,
# /mnt/e/data/N,
# /mnt/e/data/N/_dedup,
# /mnt/j/N/_ALL
# /mnt/f/T,
# /mnt/d/_dedup,
#--uniqueSearch
	go run main.go sort --paths "/mnt/d/_dedup" --move=true --search=false

dedup_file:
	go run main.go dedupFile --paths "/mnt/e/data/N"

moving:
	go run main.go sort --paths "/mnt/e/data/N/_dedup" --move=true --search=false

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
