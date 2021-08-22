# scurry
Golang interface to
[Scamper](https://www.caida.org/catalog/software/scamper/).

**Note:** This is still very much a work-in-progress. The measurement
structures are only partially implemented.

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
# scamper -p 1000 -U /tmp/scamper.sock
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

Traceroute to `8.8.8.8`
```json
scurry -s /tmp/scamper.sock trace -t 8.8.8.8 | jq
2021-08-22T10:39:25-07:00 INF Scurrying! cfg={"LogLevel":"info","Ping":{},"ScamperURL":"/tmp/scamper.sock","Target":["8.8.8.8"],"Trace":{}}
2021-08-22T10:39:25-07:00 INF Waiting for remaining measurements to complete linger=60000 module=controller outstanding=1 package=scurry
2021-08-22T10:39:35-07:00 INF Finished receiving results total=1
{
  "type": "trace",
  "target": "8.8.8.8",
  "options": {
    "ping": {},
    "trace": {}
  },
  "result": {
    "type": "trace",
    "version": "0.1",
    "method": "udp-paris",
    "src": "10.250.100.2",
    "dst": "8.8.8.8",
    "start": {
      "sec": 1629653965,
      "usec": 261318
    },
    "ping_sent": 0,
    "probe_size": 44,
    "userid": 1,
    "ttl": 0,
    "wait": 5,
    "timeout": 0
  }
}
```

### Package

#### Controller

The `Controller` type is currently the easiest way to drive
scamper. It accepts [`Measurement`](./measurement.go) objects over a
channel (`Controller.MeasurementQueue()`), and (asynchronously)
returns the same objects populated with a scamper result object over
another channel (`Controller.ResultQueue()`).

See the `main()` function of the [scurry CLI](./cmd/scurry/main.go)
for a worked example of using the Controller.

#### ScAttach

The [`ScAttach`](./attach.go) type is a low-level Scamper "attach"
driver. It connects to an already-running Scamper daemon (either via
TCP or unix domain socket), attaches using the (as-yet undocumented)
`attach format json` command to request results be returned in JSON
format rather than uuencoded warts binary.

ScAttach exposes three channels:
 - `CommandQueue() chan string`
 - `ResultQueue() chan string`
 - `ErrorQueue() chan string`

Scamper command strings can be sent directly to the buffered CommandQueue
channel. For example: `attach.CommandQueue() <- "ping 8.8.8.8"`

Results and/or errors received from Scamper will be returned over the
`ResultQueue()` and `ErrorQueue()` channels respectively. These
channels _must_ be serviced otherwise ScAttach will deadlock once the
channel buffers fill up.

## TODOs

The ultimate goal would be for Scamper to support usage as library,
in which case we could use cgo to drive scamper directly rather than
needing to connect to a stand-alone instance.

In the meantime:
 - Tests!!
 - Finish `ScResult` implementation.
 - Better CLI measurement building (see note for initMeasurement in main.go)
 - Automatically start scamper daemon just in time for ScAttach to
   connect to it.
 - Transparently manage a pool of scamper instances (using ^^) to
   allow high-volume probing (round-robin between instances). Could
   also steer measurements based on target to try and work-around
   scamper's one-concurrent-measurement-per-target limitation.
