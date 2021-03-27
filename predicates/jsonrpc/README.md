# jsonrpc

The `jsonrpc` predicate does an JSONRPC request.

## Options

| Option       | Required | Description                                   |
| ---          | ---      | ---                                           |
| `url`        | yes      | URL of API                                    |
| `procedure`  | yes      | JSONRPC procedure                             |
| `params`     |          | params of the procedure                       |
| `basic_auth` |          | set basic_auth.username & basic_auth.password |


## Results

| Field      | Type    | Description                           |
| ---        | ---     | ---                                   |
| `result`   | boolean | true if request was done successfully |
| `response` | struct  | result of the procedure               |

## Example:

``` yaml
jsonrpc:
  url: http://domain.com/jsonrpc
  procedure: test1
  params:
    param1: value1
    param2:
      - 8
      - value2
  basic_auth:
    username: USER1
    password: passwd1
```
