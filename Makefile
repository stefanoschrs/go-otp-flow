build:
	go-assets-builder templates -o templates.go
	go build -o app
