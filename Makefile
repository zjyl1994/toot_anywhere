APP_NAME=tootanywhere
APP_VERSION=alpha
APP_BUILDVER=$(shell date "+%Y%m%d%H%M%S")
build:
	docker build . -t $(APP_NAME):$(APP_VERSION)-$(APP_BUILDVER)