# sprint-tools

Scaffold lifted from the `blog` project — same tech stack, no content.

## Stack

- Go 1.24 + chi router
- a-h/templ for HTML templates
- Tailwind CSS v4 + DaisyUI
- alexedwards/scs sessions with Redis store
- libsql / Turso client (optional; only initialised when `TURSO_DATABASE` is set)
- Air for live reload

## Dev

```sh
go mod tidy
npm install
make dev
```

Serves on `http://localhost:7000`.

## Build

```sh
templ generate
npx @tailwindcss/cli -i ./web/static/css/input.css -o ./web/static/css/output.css
go build -o sprint-tools
```
