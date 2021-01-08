## trailsc - Trails config compiler and cli

[![Go Report Card](https://goreportcard.com/badge/github.com/open-osquery/trailsc)](https://goreportcard.com/report/github.com/open-osquery/trailsc)

### What is `trailsc`
`trailsc` is a config manager for osquery and the trails extension. It's build
to be modular and facilitate development and testing of extension(s) and
osquery.

### How to use `trailsc`
The usage for `trailsc`
```
Manage and build configuration for open-osquery

Usage:
  trailsc [command]

Available Commands:
  help        Help about any command
  serve       Serve the osquery config bundle over http

Flags:
  -h, --help      help for trailsc
  -v, --verbose   Increase verbosity

Use "trailsc [command] --help" for more information about a command.
```

### Serving the config over and http endpoint for trails extension
```
Serve the osquery config bundle over http

Usage:
  trailsc serve [flags]

Flags:
  -a, --addr string        IP:PORT to serve the config on (default "localhost:9000")
  -c, --cert string        Config signer key and certificate (default "cert.pem")
  -n, --container string   The config bundle container name (default "trails-config")
  -d, --dir string         Directory to serve from (default "/home/user/trailsc")
  -h, --help               help for serve
  -r, --raw                Serve raw directory

Global Flags:
  -v, --verbose   Increase verbosity
```
Example,
```bash
trailsc --dir /tmp/trails-config
```
Then config can be then fetched as
```
wget http://localhost:9000/qa/trails-config.tar.gz
```
