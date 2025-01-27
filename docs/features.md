## mDNS Auto Discovery

The device-hub can automatically add devices based on the mDNS txt records.

A device is required to have the following in its mDNS txt record:
- `autodiscovery_uri` - device URI, how device can be reached.
- `autodiscovery_desc` - human readable device description.
- `autodiscovery_mode` - auto-discovery mode, use `1` to add the device automatically.

URI examples:
- `http://bonsai-growlab.local/api/v1` - HTTP API over mDNS
- `http://192.168.4.1:17321/api/v1` - HTTP API over static IP

Desc examples:
- `room-plant-zamioculcas`
- `living-room-light-bulb`

Let's explore an example of the device, that correctly provides the required txt records. The following steps assume that [bonsai firmware](https://github.com/open-control-systems/bonsai-firmware) is installed on the device. Due to specific `bonsai-firmware` settings it's necessary for the device-hub to connect to the `bonsai-firmware` WiFi AP to ensure that device-hub can get the data from the device.

```bash
avahi-browse -r _http._tcp

+ wlp2s0 IPv4 Bonsai GrowLab Firmware                                                Web Site
 local
= wlp2s0 IPv4 Bonsai GrowLab Firmware                                                Web Site
 local
   hostname = [bonsai-growlab.local]
   address = [192.168.4.1]
   port = [80]
   txt = ["api_base_path=/api/" "api_versions=v1" "autodiscovery_uri=http://bonsai-growlab.local/api/v1"
"autodiscovery_desc=Bonsai GrowLab Firmware" "autodiscovery_mode=1"]
```

The device can now be added to the device-hub automatically. For more advanced configuration, see the device-hub CLI options:

```
--mdns-autodiscovery-disable                       Disable automatic device discovery on the local network
--mdns-browse-interval string                      How often to perform mDNS lookup over local network (default "1m")
--mdns-browse-timeout string                       How long to perform a single mDNS lookup over local network (default "30s")
```

## System Time Synchronization

The device-hub can automatically synchronize the UNIX time for the remote device.

The synchronization mechanism is based on the 3 UNIX timestamps:
- local timestamp - UNIX time of the local machine on which the device-hub is running
- remote current timestamp - latest UNIX time received from a device
- remote last timestamp - last UNIX time stored in the persistent storage, retrieved automatically, no action required

**local timestamp** can be configured as follows:

- If NTP service is running, it will be configured automatically, no action required

- Use `timedatectl` or `date` utility to manually configure the UNIX time

- Use HTTP API provided by the device-hub

In the device-hub log, look for the line "starting HTTP server":

```
inf:2025/01/14 07:40:11.355871 server_pipeline.go:40: http-server-pipeline: starting HTTP server: URL=htt
p://[::]:38807
```

Now the UNIX time can be get/set with the following API:

```
# Get UNIX time, return -1 if the timestamp is invalid or unknown
curl localhost:38807/api/v1/system/time

# Set UNIX time
curl localhost:38807/api/v1/system/time?value=123
```

If the UNIX setup fails for any reason, check the following:
- Automatic NTP synchronization should be disabled: `timedatectl set-ntp false`
- Docker container should be provided with the appropriate capabilities for the UNIX time modification: `docker run --cap-add=SYS_TIME`

**remote current timestamp** is relied on the UNIX time received from the device. The device-hub expectes the device to implement the following HTTP API for the UNIX time configuration:

```
GET /system/time - get UNIX time, return -1 if the timestamp is invalid or unknown
GET /system/time?value=123 - set UNIX time
```
