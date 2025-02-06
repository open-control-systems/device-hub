## Guide

Install [logrotate](https://linux.die.net/man/8/logrotate).

```bash
sudo apt install logrotate
```

Create log rotation configuration file: `sudo vi /etc/logrotate.d/device-hub`

```
# Replace /var/log/device-hub with the appropriate log directory that was previously configured.
/var/log/device-hub/*.log {
    # Rotate logs daily
    daily

    # Rotate if file exceeds 50Mb
    size 50M

    # Keep 7 old log files
    rotate 7

    # Compress old logs
    compress

    # Delay compression to avoid conflicts
    delaycompress

    # Ignore errors if file is missing
    missingok

    # Skip empty log files
    notifempty

    # Truncate the original file after copying
    copytruncate
}
```

Make sure the `cron` job is set for `logrotate`:

```bash
ls /etc/cron.daily/
apt-compat  dpkg  logrotate  man-db
```
