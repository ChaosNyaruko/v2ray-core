.PHONY: run build clean debug


CONFIG := $(V2CONFIG)
run: 
	@echo "config loaded: "$(CONFIG)
	cp ./*.dat ./output/
	./build.sh
	./output/v2ray-app run -config=$(CONFIG)
build:
	./build.sh
clean:
	rm -r output/
debug:
	./build.sh
	V2RAY_LOCATION_ASSET=/opt/homebrew/share/v2ray ./output/v2ray-app run -config=copy.json

