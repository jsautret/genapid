# body

The `body` predicate checks the Content-Type of the incoming request and/or parse its body.

## Options

| Option  | Required | Description                                                                                                  |
| ---     | ---      | ---                                                                                                          |
| `mime`  |          | If set, check that the Mime type of the Content-Type header value of the incoming request matches this value |
| `type`  |          | Can be `string` or `json`. If set, the body is read and parsed accordingly to this value.                    |
| `limit` |          | Max body size in bytes. If received body is larger, it will be truncated. Default is 1232896 (1Mb).          |


## Results

| Field     | Type              | Description                                                                                                                       |
| ---       | ---               | ---                                                                                                                               |
| `result`  | boolean           | false if `mime` was not matched (if set) or body cannot be parsed accordingly to `type` (if set) or method is GET, HEAD or DELETE |
| `payload` | depends on `type` | It `type` is set, contains the parsed body                                                                                        |

## Example:

``` yaml
- body:
    mime: application/json
    type: json
  register: body

- log:
    msg: body.payload.token
```
