#!/bin/bash

rm -Rf root
mkdir root
cp init root/init
chmod 0755 root/init
mkdir root/bin
cp busybox-armv8l root/bin
chmod 0755 root/bin/busybox-armv8l
ln -s /bin/busybox-armv8l root/bin/sh
cp kexec root/bin
chmod 0755 root/bin/kexec

mkdir -p root/usr/share/udhcpc/
cp simple.script root/usr/share/udhcpc/default.script
chmod 0755 root/usr/share/udhcpc/default.script

(cd root; find | cpio -o -H newc) |gzip -9 -c > initramfs
