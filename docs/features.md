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
