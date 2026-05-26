.PHONY: dev run install

install:
	go mod tidy
	npm install
	go install github.com/a-h/templ/cmd/templ
	go install github.com/air-verse/air@latest
	templ generate

dev:
	npx @tailwindcss/cli -i ./web/static/css/input.css -o ./web/static/css/output.css --watch &
	air

run:
	templ generate
	go build
	./sprint-tools
