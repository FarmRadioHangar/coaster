VERSION=0.1.1
NAME=coaster_$(VERSION)
OUT_DIR=bin/linux_arm/coaster_$(VERSION)

all:$(OUT_DIR)/coaster
$(OUT_DIR)/coaster:main.go
	gox  \
		-output "bin/{{.Dir}}_$(VERSION)/{{.OS}}_{{.Arch}}/{{.Dir}}" \
		-osarch "linux/arm" github.com/FarmRadioHangar/coaster

tar:
	cd bin/ && tar -zcvf coaster_$(VERSION).tar.gz  coaster_$(VERSION)/