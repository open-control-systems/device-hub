## Device Data Storage

The device-hub can store telemetry data from the IoT devices in the persistent storage.

For the influxdb database, see the following device-hub CLI options:

```
--storage-influxdb-api-token string                influxdb API token
--storage-influxdb-bucket string                   influxdb bucket
--storage-influxdb-org string                      influxdb Org
--storage-influxdb-url string                      influxdb URL
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

If the UNIX time setup fails for any reason, check the following:
- Automatic NTP synchronization should be disabled: `timedatectl set-ntp false`
- Docker container should be provided with the appropriate capabilities for the UNIX time modification: `docker run --cap-add=SYS_TIME`

**remote current timestamp** is relied on the UNIX time received from the device. The device-hub expectes the device to implement the following HTTP API for the UNIX time configuration:

```
GET /system/time - get UNIX time, return -1 if the timestamp is invalid or unknown
GET /system/time?value=123 - set UNIX time
```

The device-hub can automatically compensate system clock drift for the remote device. See the following configuration options:

```
--device-time-sync-drift-interval string           Maximum allowed time drift between local and device UNIX time (empty to disable drift check) (default "5s")
```

It's possible to disable automatic device time synchronization. See the following configuration options:

```
--device-time-sync-disable      Disable automatic device time synchronization
```

## Inactive Device Monitoring

The device-hub can automatically monitor the inactivity of added devices. If the device is inactive for the configured interval, it is automatically removed from the device-hub.

For more advanced configuration, see the following device-hub CLI options:

```
--device-monitor-inactive-disable                  Disable device inactivity monitoring
--device-monitor-inactive-max-interval string      How long it's allowed for a device to be inactive (default "2m")
--device-monitor-inactive-update-interval string   How often to check for a device inactivity (default "10s")
```

## mDNS Server

The device-hub has a bult-in mDNS server. This allows to assign a memorable hostname to the device-hub and use it instead of an explicit IP address, which can be changed from time to time.

For more advanced configuration, see the following device-hub CLI options:

```
--mdns-server-disable               Disable mDNS server
--mdns-server-hostname string       mDNS server hostname (default "device-hub")
--mdns-server-iface string          Comma-separated list of network interfaces for the mDNS server (empty for all interfaces)
```

Once the mDNS server is properly configured, it should be possible to access the device-hub as follows:

```
# device-hub is a hostname.
# 8081 is a HTTP port on which HTTP server is running.
curl device-hub.local:8081/api/v1/system/time
```

## mDNS Browser

The device-hub has a bult-in mDNS browser. This allows the device-hub to reach the device in the local network, without explicitly specifying an IP address of the device. For example, it's possible to add the following device to the device-hub:

```
# URI - http://bonsai-growlab.local:80/api/v1
# Description - home-plant
curl device-hub.local/api/v1/device/add?uri=http://bonsai-growlab.local:80/api/v1&desc=home-plant
```

`bonsai-growlab.local` is the mDNS hostname of the device. device-hub can automatically resolve it to the actual IP address. If the IP address of the device changes in the future, the device-hub will automatically handle it.

For more advanced configuration, see the following device-hub CLI options:

```
--mdns-browse-iface string          Comma-separated list of network interfaces for the mDNS lookup (empty for all interfaces)
--mdns-browse-interval string       How often to perform mDNS lookup over local network (default "1m")
--mdns-browse-timeout string        How long to perform a single mDNS lookup over local network (default "30s")
```

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

The device can now be added to the device-hub automatically. For more advanced configuration, see the following device-hub CLI options:

```
--mdns-autodiscovery-disable                       Disable automatic device discovery on the local network
--mdns-browse-interval string                      How often to perform mDNS lookup over local network (default "1m")
--mdns-browse-timeout string                       How long to perform a single mDNS lookup over local network (default "30s")
```
