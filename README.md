# scurry
Golang interface to Scamper

## Dependencies

Other than the normal go module dependencies, you will also need to have built
[Scamper](https://www.caida.org/catalog/software/scamper/ (or installed it from
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

TODO

### Package

TODO
