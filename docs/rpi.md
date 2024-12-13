## Introduction

This guide covers the steps required to setup device-hub on the Raspberry Pi.

## Installation Instructions

Install Raspberry Pi OS using [Raspberry Pi Imager](https://www.raspberrypi.com/software/). It's possible to use any Linux-based OS. The following guide uses Raspberry Pi OS Lite 64-bit. During the installation make sure SSH is enabled, it will be required later to perform the system setup. Next ensure that Wireless LAN is configured with router's SSID and password, or it can be skipped if the Ethernet connection will be used. After the OS is installed, it's requirde to connect to RPi to perform the required settings. For the RPi 4 it's recommended to update the bootloader right after the OS is installed. It can be done with the following command:

```bash
sudo rpi-eeprom-update
sudo reboot
```

After the OS is installed, check the SSH connectivity: ping device-hub-rpi.local (replace `device-hub-rpi` with configured hostname). The result can be as follows:

```
64 bytes from 192.168.1.169 (192.168.1.169): icmp_seq=1 ttl=64 time=8.86 ms
64 bytes from 192.168.1.169 (192.168.1.169): icmp_seq=2 ttl=64 time=8.50 ms
64 bytes from 192.168.1.169 (192.168.1.169): icmp_seq=3 ttl=64 time=12.4 ms
```

For some reason my Pi doesn't want to boot until I connect it to the external display. After it is connected to the display for the first time, it boots normally and I can continue to use it without any issues.

## References

- RPi official getting started [documentation](https://www.raspberrypi.com/documentation/computers/getting-started.html)
