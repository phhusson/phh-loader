// This file is
//  Copyright 2020 Pierre-Hugues Husson <phh@phh.me>
// It is redistributed under the GPLv3 license, that you can see at https://www.gnu.org/licenses/gpl-3.0.txt

package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "log"
    "os"
    "os/exec"
    "unsafe"
)

type BootImgV3Header struct {
    Magic [8]byte //ANDROID!

    KernelSize uint32
    RamdiskSize uint32
    OsVersion uint32
    HeaderSize uint32
    Reserved [4]uint32
    HeaderVersion uint32
    Cmdline [1024+512]uint8
}

type VendorBootImgV3Header struct {
    Magic [8]byte //VNDRBOOT

    HeaderVersion uint32

    PageSize uint32
    KernelAddr uint32
    RamdiskAddr uint32

    VendorRamdiskSize uint32
    Cmdline [2048]uint8
    TagsAddr uint32
    Name [16]uint8
    HeaderSize uint32
    DtbSize uint32
    DtbAddr uint64
}

//Checks only 8 chars magics
func checkMagic(f *os.File, magic string) {
    magicBytes := make([]byte, 8)
    _, err := f.Read(magicBytes)
    if err != nil {
        log.Fatal("Failed reading magic", err)
    }

    f.Seek(0, 0)
    if string(magicBytes) != magic {
        log.Fatal("Expected boot-debug.img to be a boot image")
    }
}

func dumpTo(in *os.File, out *os.File, offset int64, length uint64) {
    bytes := make([]byte, length)
    _, err := in.ReadAt(bytes, offset)
    if err != nil {
        log.Fatal("Failed reading section")
    }
    _, err = out.Write(bytes)
}

func main() {
    gki, err := os.Open("boot-debug.img")
    if err != nil {
        log.Fatal("Failed opening gki", err)
    }
    defer gki.Close()

    checkMagic(gki, "ANDROID!")

    gkiHeader := BootImgV3Header {}
    {
        bufferByteArray := make([]byte, unsafe.Sizeof(gkiHeader))
        _, err = gki.Read(bufferByteArray)
        if err != nil {
            log.Fatal("Failed reading gki header")
        }
        buffer := bytes.NewBuffer(bufferByteArray)
        err = binary.Read(buffer, binary.LittleEndian, &gkiHeader)
        fmt.Printf("Got gki version: %d\n", gkiHeader.HeaderVersion)
        fmt.Printf("Got kernel size: %dMB\n", gkiHeader.KernelSize/(1024*1024))
    }

    vndrBoot, err := os.Open("vendor_boot-debug.img")
    if err != nil {
        log.Fatal("Failed opening vendor_boot")
    }
    defer vndrBoot.Close()
    checkMagic(vndrBoot, "VNDRBOOT")

    vendorbootHeader  := VendorBootImgV3Header {}
    {
        bufferByteArray := make([]byte, unsafe.Sizeof(vendorbootHeader))
        _, err = vndrBoot.Read(bufferByteArray)
        if err != nil {
            log.Fatal("Failed reading gki header")
        }
        buffer := bytes.NewBuffer(bufferByteArray)
        err = binary.Read(buffer, binary.LittleEndian, &vendorbootHeader)
        fmt.Printf("Got vendor boot version: %d\n", vendorbootHeader.HeaderVersion)
        fmt.Printf("Got vendor ramdisk size: %dMB\n", vendorbootHeader.VendorRamdiskSize/(1024*1024))
    }

    //Concatanated ramdisk to pass to kernel
    globalRamdisk, err := os.OpenFile("global-ramdisk", os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        log.Fatal("Failed creating file 'kernel'")
    }

    //Extracts from gki
    {
        {
            kernel, err := os.OpenFile("kernel", os.O_WRONLY|os.O_CREATE, 0644)
            if err != nil {
                log.Fatal("Failed creating file 'kernel'")
            }
            defer kernel.Close();
            dumpTo(gki, kernel, 4096, uint64(gkiHeader.KernelSize))
        }
        kernelSpace := ((gkiHeader.KernelSize + 4095) / 4096) * 4096

        {
            out, err := os.OpenFile("ramdisk", os.O_WRONLY|os.O_CREATE, 0644)
            if err != nil {
                log.Fatal("Failed creating file 'ramdisk'")
            }
            defer out.Close();
            dumpTo(gki, out, 4096 + int64(kernelSpace), uint64(gkiHeader.RamdiskSize))
            dumpTo(gki, globalRamdisk, 4096 + int64(kernelSpace), uint64(gkiHeader.RamdiskSize))
        }
    }

    //Extracts from vendor_boot
    {
        headerSpace := ((2112 + vendorbootHeader.PageSize - 1) / vendorbootHeader.PageSize) * vendorbootHeader.PageSize
        {
            out, err := os.OpenFile("vendor_ramdisk", os.O_WRONLY|os.O_CREATE, 0644)
            if err != nil {
                log.Fatal("Failed creating 'vendor_ramdisk'")
            }
            defer out.Close()
            dumpTo(vndrBoot, out, int64(headerSpace), uint64(vendorbootHeader.VendorRamdiskSize))
        }
        vendorRamdiskSpace := ((vendorbootHeader.VendorRamdiskSize + vendorbootHeader.PageSize - 1)/vendorbootHeader.PageSize) * vendorbootHeader.PageSize
        {
            out, err := os.OpenFile("dtb", os.O_WRONLY|os.O_CREATE, 0644)
            if err != nil {
                log.Fatal("Failed creating 'vendor_ramdisk'")
            }
            defer out.Close()
            dumpTo(vndrBoot, out, int64(headerSpace + vendorRamdiskSpace), uint64(vendorbootHeader.DtbSize))
            dumpTo(vndrBoot, globalRamdisk, int64(headerSpace + vendorRamdiskSpace), uint64(vendorbootHeader.DtbSize))
        }
    }
    cmd := exec.Command("kexec", "-f", "--dtb=dtb", "--initrd=global-ramdisk", "kernel")
    err = cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
}
