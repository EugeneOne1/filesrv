.POSIX:

COMMIT = $$( git rev-parse --short HEAD )
GO.MACRO = $${GO:-go}
VERBOSE.MACRO = $${VERBOSE:-0}
RACE = 0
VERSION = v0.0.0
THEME_PATH=$${THEME_PATH:-./internal/dirs/themes}

ENV = env\
	COMMIT='$(COMMIT)'\
	GO="$(GO.MACRO)"\
	PATH="$${PWD}/bin:$$( "$(GO.MACRO)" env GOPATH )/bin:$${PATH}"\
	RACE='$(RACE)'\
	VERBOSE="$(VERBOSE.MACRO)"\
	VERSION='$(VERSION)'\
	THEME_PATH="$(THEME_PATH)"

run: build
	$(ENV) ./srv

build: go-build

go-build: ; $(ENV) "$(SHELL)" ./scripts/make/go-build.sh
go-deps:  ; $(ENV) "$(SHELL)" ./scripts/make/go-deps.sh
go-lint:  ; $(ENV) "$(SHELL)" ./scripts/make/go-lint.sh
go-tools: ; $(ENV) "$(SHELL)" ./scripts/make/go-tools.sh

go-test:  ; $(ENV) RACE='1' "$(SHELL)" ./scripts/make/go-test.sh
