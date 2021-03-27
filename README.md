[![Go Report Card](https://goreportcard.com/badge/github.com/jsautret/genapid)](https://goreportcard.com/report/github.com/jsautret/genapid)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/jsautret/genapid)


# Generic API Daemon

genapid is an API server using YAML format to describe the API behavior.

It can be used to process Webhooks, add some custom commands to a Google Home or as an API broker between several API or IoT services.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [Generic API Daemon](#generic-api-daemon)
    - [Concept](#concept)
    - [Examples](#examples)
        - [Controlling Kodi with Google Assistant](#controlling-kodi-with-google-assistant)
    - [Install](#install)
        - [Compilation](#compilation)
        - [Ansible](#ansible)
    - [Run](#run)
    - [Configuration](#configuration)
        - [pipe](#pipe)
        - [Predicates](#predicates)
            - [Options](#options)
                - [`name`](#name)
                - [`result`](#result)
                - [`register`](#register)
                - [`when`](#when)
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

### Controlling Kodi with Google Assistant

* Control Kodi by voice using a Google Home:
[examples/kodi/](examples/kodi/)

## Install

### Compilation

Needs [go](https://golang.org/).

```
$ go get -u github.com/jsautret/genapid
```

### Ansible

If you use Ansible, you can adapt the role in [ansible/](ansible/) to
deploy genapid behind an Apache server.

## Run

``` shell
$ Usage of ./cmd/genapid/genapid:
  -config string
        Config file (default "api.yml")
  -loglevel string
        Log level (default "info")
  -port int
        Listening port (default 9110)
```

## Configuration

The API is described in a YAML file, which passed to genapid using the
`-config` option.

### pipe

The API description file is a list of `pipe` elements. Each `pipe` is
a list of predicates or sub-pipes. Only the `name` and `result`
options can be used on `pipe` (see below for the description of
options).

The result of a `pipe`is always true, unless `result` option is set.

### Predicates

The list of predicates is in [predicates/](predicates/). There is also
some additional predicates described below.

Each predicates has it own specific parameters described in it documentation.

#### Options

The following options can be set on predicates:

##### `name`

Type: string

Used for documenting and logs readability only.

##### `result`

Type: boolean

Force the value of the predicate

##### `register`

Type string

Store the results set by the predicate. This data can be accessed in
following predicates with the `R` map. For example, if you set option
`register: myresult`, the data set by the predicate can then be
accessed with `R.myresult` which is a map. The `result` key will
contain the boolean result of the predicate (real one, not the one set
with the `result`option). So `R.myresult.result` can be used to check
the result of the predicate. Some predicate may provide additional
fields described in their documentation.

##### `when`
Type: boolean

If false, the predicate evaluation is skipped.

#### Special predicates

##### `variable`
Used to set variables. The variables can be accessed in following
predicates with the `V` map. It takes a list of map as parameters.

Example:
``` yaml
variable:
 - variable1: value1
 - variable2: value2
```

##### `default`
Used to set default parameters for the following
predicates. Expressions are evaluated when the predicate is evaluated,
not when `default` is evaluated.

Example:
``` yaml
default:
  http:
    url: http://domain.com/api
    method: get
```

### Expressions

If the value of the parameter of a predicate starts with an = (equal
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
