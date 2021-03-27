# http

The `http` predicate does an HTTP request.

## Options

Option | Required | Description
---|---|---
`url` | yes | URL to call
`method` | | HTTP method (default to GET)
`headers` | | headers to set
`params` | | URL query params
`body` | | Use `body.string` to send a text or `body.json` to send json.
`response` | | set to `JSON` to parse the response.
`basic_auth` | | set basic_auth.username & basic_auth.password


## Results

Field | Type | Description
---|---|---
`result` | boolean | true if request was done
`response` | | response as string or struct, depending of the `response` option
`type` | string | Content-Type
`code` | int | returned HTTP code

## Example:

``` yaml
http:
  url: http://test/params
  method: post
  body:
    json:
      k1: v1
      k2: v2
  response: json
  basic_auth:
    username: myuser
    password: mypasswd
  params:
    k1: v1
    k2: v2
  headers:
    h1: v1
    h2: v2
```
