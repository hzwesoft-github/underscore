TARGET1 = ubus_demo.out
# TARGET2 = notify_demo.out
# TARGET3 = event_demo.out

# change to yours
BUILD_DIR=/opt/neptune-gateway/build
OPENWRT_SDK=openwrt-sdk

STAGING_DIR=$(BUILD_DIR)/$(OPENWRT_SDK)/staging_dir
CC=$(BUILD_DIR)/$(OPENWRT_SDK)/staging_dir/toolchain-x86_64_gcc-8.4.0_musl/bin/x86_64-openwrt-linux-gcc

CGO_CFLAGS=-I$(BUILD_DIR)/include
CGO_LDFLAGS=-L$(BUILD_DIR)/lib -lubox -lblobmsg_json -lubus -luci -ljson-c

all: $(TARGET1)

$(TARGET1):
	# CC="$(CC)" CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	# go build -ldflags "-s -w" -o $(TARGET1)
	CC="$(CC)" CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -o $(TARGET1)

clean:
	rm -f $(TARGET1) *.o
