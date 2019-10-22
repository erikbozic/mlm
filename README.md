# MLM - Mesos Log Monitor [![Actions Status](https://github.com/erikbozic/mlm/workflows/go-build/badge.svg)](https://github.com/erikbozic/mlm/actions)

**Work in progress!**    
Command line tool for monitoring mesos task logs

## Build

Build with `go build -o mlm mlm/cmd/monitor`. This will produce the mlm executable.

We're building with go 1.12 and 1.13.

## Usage

Run this binary with `-m` flag and pass in the http url for mesos master.

Example:
`` ./mlm -m http://localhost:5050 ``  

This will fetch all tasks known to master and will let you specify those that
you want to monitor logs from.

Mesos master url will be saved into a configuration file and you won't have to
provide it next time you use this binary.

## Commands

 - `:b` issue this command to stop listening and return to the task selection  
 - `:q` issue this command to stop listening and quit the program  
 - `:f {filterString}` issue this command to start filtering log messages on all listeners using the provided filter string
 - `:p` issue this command to pause all listeners
 - `:u` issue  this command to unpause all listeners
 
## Preview
<a href="https://asciinema.org/a/EXQZXXC60O83cJrLMnVTnfoiN?autoplay=1"><img src="https://asciinema.org/a/EXQZXXC60O83cJrLMnVTnfoiN.png" width="720"/></a>

## Development
To get started you can setup mesos on you machine easily with the docker image: [mesos/mesos-mini](https://hub.docker.com/r/mesos/mesos-mini/).  
`docker run --rm --privileged -p 5050:5050 -p 5051:5051 -p 8080:8080 mesos/mesos-mini`

Deploy an application through marathon and you're good to Go!

 - CLI code lives in the `/cmd/monitor` directory
 - library code lives in the `/pkg/monitor` directory
