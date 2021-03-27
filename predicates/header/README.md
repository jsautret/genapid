# header

The `header` predicate checks the value of a predicate and/or gets it value

## Options

| Option  | Required | Description        |
| ---     | ---      | ---                |
| `name`  | yes      | Name of the header |
| `value` |          | Value to match     |

## Results

| Field    | Type    | Description                                               |
| ---      | ---     | ---                                                       |
| `result` | boolean | true if value is not set, or true if value matches header |
| `value`  | string  | Value of header                                           |
