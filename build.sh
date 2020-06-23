#!/bin/bash

rm -Rf root
mkdir root
cp init root/init
chmod 0755 root/init
mkdir root/bin
cp busybox-armv8l root/bin
chmod 0755 root/bin/busybox-armv8l
ln -s /bin/busybox-armv8l root/bin/sh
(cd root; find | cpio -o -H newc) |gzip -9 -c > initramfs
