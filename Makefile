.PHONY: api/build
api/build:
	@go build -o bin/api cmd/api/*.go

.PHONY: api/run
api/run: api/build
	@bin/api -dev

.PHONY: api/watch
api/watch:
	reflex -r '\.go$$|\.env$$' -d none -s make api/run
