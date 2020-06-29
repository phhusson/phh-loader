#!/bin/sh

export PATH=$PATH:/

set -xe

mdev -d
mdev -s


#wget http://192.168.1.53:8080/test.arm64
#wget http://192.168.1.53:8080/boot-debug.img
#wget http://192.168.1.53:8080/vendor_boot-debug.img
#chmod 0755 test.arm64
#./test.arm64

wget http://192.168.1.53:8080/Image
wget http://192.168.1.53:8080/ramdisk.img
wget http://192.168.1.53:8080/sfdisk
wget http://192.168.1.53:8080/gdisk
wget http://192.168.1.53:8080/fdisk
chmod 0755 sfdisk

wget http://192.168.1.53:8080/dtb-mainline -O dtb-to-boot
wget http://192.168.1.53:8080/dtc
wget http://192.168.1.53:8080/fdtput
wget http://192.168.1.53:8080/fdtget
chmod 0755 dtc fdtput fdtget

dtc -I fs -O dtb /sys/firmware/devicetree/base -o dtb-bootloader

serialNumber=$(fdtget -t s dtb-bootloader / serial-number)
if [ -n "$serialNumber" ];then
    echo "Setting serial number to $serialNumber"
    fdtput -t s dtb-to-boot / serial-number "$serialNumber"
fi

memreserve=$(fdtget -t x dtb-bootloader / memreserve)
if [ -n "$memreserve" ];then
    echo "Setting memreserve to $memreserve"
    #No quotes! fdtput requires multiple args if multiple longs
    fdtput -t x dtb-to-boot / memreserve $memreserve
fi

#We could also add `display0` here, but we won't copy framebuffer not, so meh
for al in uart0 uart1;do
    previous="$(fdtget -t s dtb-to-boot /aliases $al || true)"
    blValue="$(fdtget -t s dtb-bootloader /aliases $al)"
    if [ -n "$blValue" ] && [ -z "$previous" ];then
        echo "Setting alias $al to $blValue"
        fdtput -t s dtb-to-boot /aliases $al "$(fdtget -t s dtb-bootloader /aliases $al)"
    fi
done

rpiBoardrevExt=$(fdtget -t x dtb-bootloader /chosen rpi-boardrev-ext)
if [ -n "$rpiBoardrevExt" ];then
    echo "Setting rpi board rev to $rpiBoardrevExt"
    fdtput -t x dtb-to-boot /chosen rpi-boardrev-ext $rpiBoardrevExt
fi

memory=$(fdtget -t x dtb-bootloader /memory@0 reg)
fdtput -t x dtb-to-boot /memory@0 reg $memory

cmdline=$(cat /proc/cmdline)
cmdline="$cmdline androidboot.hardware=rpi4"
#cmdline="$cmdline androidboot.super_partition=mmcblk1p2"
cmdline="$cmdline androidboot.boot_devices=emmc2bus/fe340000.emmc2"
#cmdline="$cmdline androidboot.slot_suffix=_a"
cmdline="$cmdline androidboot.selinux=permissive"
cmdline="$cmdline androidboot.serialno=$serialNumber"

# d
# 2

# n
# 2
# 532480
# 598015
# c
# 2
# metadata

# n
# 3
# 598016
# +10G
# c
# 3
# super

#simg2img super.img mmcblk1p3
#mkfs.ext4 mmcblk1p4

sh

kexec -f --dtb=dtb-to-boot --command-line="$cmdline" --initrd=ramdisk.img Image
