## HTTP API

All examples below are run from the terminal and use the [httpie](https://httpie.io/docs/cli) CLI tool to make HTTP requests.

**Get system time**

http "localhost:8080/api/v1/system/time"

```txt
1733233869
```

**Set system time**

http "localhost:8080/api/v1/system/time?value=1733233869"

```txt
OK
```
