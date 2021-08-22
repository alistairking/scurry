# scurry
Golang interface to [Scamper](https://www.caida.org/catalog/software/scamper/).

## Dependencies

Other than the normal go module dependencies, you will also need to have built
[Scamper](https://www.caida.org/catalog/software/scamper/) (or installed it from
a package).

## Usage

Current this module is just a very basic scamper control socket client. It can
be used either as a basic CLI tool, or as a package.

In either mode, you will first need to start scamper and have it listen on a
control socket:
```bash
# scamper -p 1000 -U /tmp/scamper-scurry.sock
```

### CLI

The scurry CLI is just a simple wrapper around the package. It can
execute parallel measurements of a single type against a set of
targets.

#### Usage

```
scurry --help
Usage: scurry --target=TARGET,... --scamper-url=STRING <command>

Flags:
  -h, --help                  Show context-sensitive help.
  -t, --target=TARGET,...     IP to execute measurements towards
  -s, --scamper-url=STRING    URL to connect to scamper on (host:port or unix
                              domain socket)
      --log-level="info"      Log level

Commands:
  ping --target=TARGET,... --scamper-url=STRING
    Ping measurements

  traceroute --target=TARGET,... --scamper-url=STRING
    Traceroute measurements

Run "scurry <command> --help" for more information on a command.
```

#### Examples

Ping `8.8.8.8`
```json
$ scurry -s /tmp/scamper.sock ping -t 8.8.8.8 | jq
2021-08-22T10:02:55-07:00 INF Scurrying! cfg={"LogLevel":"info","Ping":{},"ScamperURL":"/tmp/scamper.sock","Target":["8.8.8.8"],"Traceroute":{}}
2021-08-22T10:02:55-07:00 INF Waiting for remaining measurements to complete linger=60000 module=controller outstanding=1 package=scurry
2021-08-22T10:02:59-07:00 INF Finished receiving results total=1
{
  "type": "ping",
  "target": "8.8.8.8",
  "options": {
    "ping": {},
    "traceroute": {}
  },
  "result": {
    "type": "ping",
    "version": "0.4",
    "method": "icmp-echo",
    "src": "10.250.100.2",
    "dst": "8.8.8.8",
    "start": {
      "sec": 1629651775,
      "usec": 474003
    },
    "ping_sent": 4,
    "probe_size": 84,
    "userid": 1,
    "ttl": 64,
    "wait": 1,
    "timeout": 1
  }
}
```

### Package

TODO. See the [CLI implementation](./cmd/scurry/main.go) for an
example.
