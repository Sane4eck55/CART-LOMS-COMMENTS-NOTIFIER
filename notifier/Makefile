BINDIR=${CURDIR}/bin
PACKAGE=route256/notifier

bindir:
	mkdir -p ${BINDIR}

build: bindir
	echo "build notifier"

.PHONY: run
run:
	echo "starting build"
	docker-compose up --build
