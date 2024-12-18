## Introduction

This guide covers the steps required to setup device-hub on the Raspberry Pi.

## Install

### RPi OS

Install Raspberry Pi OS using [Raspberry Pi Imager](https://www.raspberrypi.com/software/). It's possible to use any Linux-based OS. The following guide uses Raspberry Pi OS Lite 64-bit. During the installation make sure SSH is enabled and "allow public-key authentication only" option is set. It will be required later to perform the system setup. Next ensure that Wireless LAN is configured with router's SSID and password, or it can be skipped if the Ethernet connection will be used. After the OS is installed, it's requirde to connect to RPi to perform the required settings. For the RPi 4 it's recommended to update the bootloader right after the OS is installed. It can be done with the following command:

```bash
sudo rpi-eeprom-update
sudo reboot
```

After the OS is installed, power on the RPi and check if the mDNS works properly:

```bash
# Replace `device-hub-rpi` with configured hostname.
ping device-hub-rpi.local
```

The result can be as follows:

```
64 bytes from 192.168.1.169 (192.168.1.169): icmp_seq=1 ttl=64 time=8.86 ms
64 bytes from 192.168.1.169 (192.168.1.169): icmp_seq=2 ttl=64 time=8.50 ms
64 bytes from 192.168.1.169 (192.168.1.169): icmp_seq=3 ttl=64 time=12.4 ms
```

For some reason my Pi doesn't want to boot until I connect it to the external display. After it is connected to the display for the first time, it boots normally and I can continue to use it without any issues.

If the Pi operates normally it's time to connect to it via SSH and setup the device-hub software.

## Device-Hub

- During the OS installation it was required to enable SSH. Now it's time to add private key to the ssh-agent, so that it will be possible to connect to Pi without manually specifying path to the private key:

```bash
# Replace rpi3b with the required file name.
ssh-add ~/.ssh/rpi3b
```

After that it should be possible to connect to Pi as follows:

```bash
# Replace dshil with the configured user name
# Replace device-hub-rpi with the configured hostname
ssh dshil@device-hub-rpi.local
```

Now it's time to install the required packages:

- Install Docker

```bash
for pkg in docker.io docker-doc docker-compose podman-docker containerd runc; do sudo apt-get remove $pkg; done

# Add Docker's official GPG key:
sudo apt-get update
sudo apt-get install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

- Install git

```bash
sudo apt-get install git
```

- Clone the device-hub repository

```bash
git clone https://github.com/open-control-systems/device-hub.git
```

- Run influxdb service:

```bash
# Replace <username>, <password>, <admin>, <bucket>, <org> with the required credentials.
# See also https://docs.influxdata.com/influxdb/cloud/reference/key-concepts/data-elements/.
cd device-hub/projects/main
INFLUXDB_USERNAME="<username>" \
INFLUXDB_PASSWORD="<password>" \
INFLUXDB_ADMIN_TOKEN="<admin_token>" \
INFLUXDB_BUCKET="<bucket>" \
INFLUXDB_ORG="<org>" \
docker compose up influxdb -d
```

- Ensure influxdb works as expected. Use ssh port forwarding to access influxdb on the local PC:

```bash
# 8086 - local PC port.
# localhost:8086 - target RPi port.
ssh -L 8086:localhost:8086 dshil@device-hub-rpi.local
```

After port forwarding is enabled, access localhost:8086 in the web-browser, and enter influxdb credentials recently used to run the docker compose service.

- Determine `org` identifier, will be used later to create API tokens:

```bash
curl http://localhost:8086/api/v2/orgs -H "Authorization: Token <admin_token>"
```

- Retrieve influxdb API token for device-hub software

```bash
# See also: https://docs.influxdata.com/influxdb/cloud/admin/tokens/create-token/
curl http://localhost:8086/api/v2/authorizations \
  -H "Authorization: Token <admin_token>" \
  -H 'Content-type: application/json' \
  --data '{
  "status": "active",
  "description": "device-hub API token",
  "orgId": "<orgId>",
  "permissions": [
    {
      "action": "write",
      "resource": {
        "type": "buckets"
      }
    }
  ]
}'
```

Save "token" field for later usage.

- Run the device-hub software

```bash
cd device-hub/projects/main

# Create log directory.
sudo mkdir -p /var/log/device-hub

INFLUXDB_API_TOKEN="<api_token>" \
INFLUXDB_BUCKET="<bucket>" \
INFLUXDB_ORG="<org>" \
DEVICE_HUB_LOG_PATH="/var/log/device-hub/app.log" \
docker compose up device-hub -d
```

After everything is properly configured, the services can be started with the `docker.sh` script as follows:

```bash
INFLUXDB_USERNAME="<username>" \
INFLUXDB_PASSWORD="<password>" \
INFLUXDB_ADMIN_TOKEN="<admin_token>" \
INFLUXDB_ORG="<org>" \
INFLUXDB_API_TOKEN="<api_token>" \
DEVICE_HUB_LOG_PATH="/var/log/device-hub/app.log" \
./docker.sh
```

## Usage

After all services are started it's time to connect to the device's WiFi AP to ensure the device-hub software is able to get the data from the real device:

## References

- RPi official getting started [documentation](https://www.raspberrypi.com/documentation/computers/getting-started.html)
