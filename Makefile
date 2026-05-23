.PHONY: dev run

dev:
	npx @tailwindcss/cli -i ./web/static/css/input.css -o ./web/static/css/output.css --watch &
	air

run:
	templ generate
	go build
	./sprint-tools
