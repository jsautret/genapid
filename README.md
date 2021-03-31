[![GitHub release](https://img.shields.io/github/release/jsautret/genapid.svg)](https://github.com/jsautret/genapid/releases)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/jsautret/genapid)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/jsautret/genapid)
[![Go Report Card](https://goreportcard.com/badge/github.com/jsautret/genapid)](https://goreportcard.com/report/github.com/jsautret/genapid)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/jsautret/genapid/test?label=tests)](https://github.com/jsautret/genapid/actions/workflows/test.yml)



# Generic API Daemon

genapid is an API server using YAML format to describe the API logic.

It can be used to process Webhooks, add some custom commands to a
Google Home or as an API broker between several API or IoT services.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [Generic API Daemon](#generic-api-daemon)
    - [Concept](#concept)
    - [Examples](#examples)
        - [Receive Github Webhook and send Pushbullet notification](#receive-github-webhook-and-send-pushbullet-notification)
        - [Controlling Kodi with Google Assistant](#controlling-kodi-with-google-assistant)
    - [Install](#install)
        - [Binary releases](#binary-releases)
        - [Docker](#docker)
            - [Run the genapi container](#run-the-genapi-container)
        - [Compile from sources](#compile-from-sources)
        - [Ansible](#ansible)
    - [Run](#run)
    - [Configuration](#configuration)
        - [pipe](#pipe)
        - [Predicates](#predicates)
            - [Options](#options)
            - [Special predicates](#special-predicates)
                - [`variable`](#variable)
                - [`default`](#default)
        - [Expressions](#expressions)
            - [`R`](#r)
            - [`V`](#v)
            - [`In`](#in)

<!-- markdown-toc end -->

## Concept

The API is described using **pipes** of **predicates**. All predicates
in a pipe are evaluated as long as they are true. When a predicate is
false, the predicates in the next pipe are evaluated. Some predicates
can have side effects, like calling an external API, to perform actual
actions.

## Examples

### Receive Github Webhook and send Pushbullet notification

The examples shows how to receive a Webhook event and call an external
API or use Google Home to get a voice notification:
[examples/github/](examples/github/)


### Controlling Kodi with Google Assistant

Control Kodi by voice using a Google Home and receive voice feedback:
[examples/kodi/](examples/kodi/)



## Install

### Binary releases

Binary execs for various platforms are available here:

https://github.com/jsautret/genapid/releases

### Docker

Docker containers are available on Docker Hub:

https://hub.docker.com/repository/docker/jsautret/genapid/general


#### Run the genapid container

The configuration file that describes your API must be named `api.yml`
and placed in a directory that you have to mount on `/conf` volume and
port 8080 must be mapped with a local port:

``` shell
CONF=/etc/genapid # must contain api.yml
PORT=9080
docker run --name genapid -d -p $PORT:8080 -v $CONF:/conf jsautret/genapid:latest

curl http://localhost:$PORT/test
docker logs genapid
```

### Compile from sources

Needs [go](https://golang.org/) 1.14 or later.

```
$ go get -u github.com/jsautret/genapid/cmd/genapid
```

Exec will be build here: `$GOPATH/bin/genapid`.

### Ansible

If you use Ansible, you can adapt the role in [ansible/](ansible/) to
deploy genapid behind an Apache server.

## Run

``` shell
$ ./cmd/genapid/genapid -h
Usage of ./cmd/genapid/genapid:
  -config string
        Config file (default "api.yml")
  -loglevel string
        Log level (default "info")
  -port int
        Listening port (default 9110)
```

The valid log levels are:
- panic
- fatal
- error
- warn
- info
- debug
- trace

## Configuration

The API is described in a YAML file, which is passed to genapid using
the `-config` option.

There is some examples in [examples/](examples/) directory.

The main structure is a list of [pipes](#pipe). Each pipe is evaluated for each incoming request.

### `init`

The first element of the top-level list can be `init`, which contains
a list of [predicates](#predicates). The predicates in `init` are
evaluated only once when genapid starts. `init` can be used to read
data from a file and populate the context once for all and not for
every request. See the beginning of
[`github.yml`](examples/github/github.yml) for an example.

### `pipe`

The API description file is a list of `pipe` elements. Each `pipe`
contains a list of predicates or sub-pipes. The `name` and `result`
options can be used on `pipe` (see below for the description of
these options).

The result of a `pipe`is always true, unless `result` option is set.

### Predicates

Each predicate in a pipe is evaluated for an give incoming
request. When a predicate returns false, the following predicates in
the pipe are ignored and the next pipe in the conf is evaluated.

The list of predicates types is in [predicates/](predicates/). There is also
some [additional predicates](#special-predicates) described below.

Each predicates has it own specific parameters described in its documentation.

#### Options

The following options can be set on predicates:

| Name     | Type    | Description                                     |
| ---      | ---     | ---                                             |
| `name`   | string  | Used for documenting and logs readability only. |
| `result` | boolean | Force the result value of the predicate         |
| `when`   | boolean | If false, the predicate evaluation is skipped.  |
| `register` | string  | Store the results set by the predicate. This data can be accessed in following predicates with the `R` map. For example, if you set option `register: myresult`, the data set by the predicate can then be accessed with `R.myresult` which is a map. The `result` key will contain the boolean result of the predicate (real one, not the one set with the `result`option). So `R.myresult.result` can be used to check the result of the predicate. Some predicate may provide additional fields described in their documentation. |

#### Special predicates

##### `variable`
Used to set variables. It takes a list of map as parameters.

The variables can be accessed in predicates with the `V`
map.

Example:
``` yaml
variable:
 - variable1: value1
 - variable2: '=  V.variable1 + "_suffix"
log:
  msg: '= "variable2: " + V.variable2 ' # will log "variable2: value1_suffix"
```

##### `default`
Used to set default parameters for predicates. Expressions are
evaluated when the predicate is evaluated, not when `default` is
evaluated.

Example:
``` yaml
default:
  http:
    url: http://domain.com/api
    method: get
```

### Expressions

If the value of the parameter of a predicate starts with an `=` (equal
sign), it will be evaluated as a [Gval
expression](https://github.com/PaesslerAG/gval). If the evaluation of
an expression fails, the predicate returns false.

The Following data is accessible in those expressions:

#### `R`

Map containing data stored using the `register` option.

#### `V`

Map containing variables set by the `variable` predicate.

#### `In`

Map containing information about the incoming request received by
genapid. It has the following fields:

* `URL.Params`: Map containing the URL query parameters.
* `Mime`: Content-Type
* `Req`
  * `Method`: HTTP method
  * `URL`
    * `Path`: URL path
    * `Host`: URL Host
    * `Scheme`: URL protocol
