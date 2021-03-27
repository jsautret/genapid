# match

The `match` predicate checks a string is equal to another string or match a given regexp. Regexp uses [go syntax](https://golang.org/pkg/regexp/syntax/).

## Options

| Option   | Required | Description            |
| ---      | ---      | ---                    |
| `sting`  | yes      | The string to match.   |
| `value`  |          | The string to compare. |
| `regexp` |          | The regexp to match .  |

One of `value` or `regexp` must be present.

## Results

| Field     | Type            | Description                                                                                                                                                                                                                                                                                               |
| ---       | ---             | ---                                                                                                                                                                                                                                                                                                       |
| `result`  | boolean         | true if the value of `string` is equal to the value of `value` or if the value of `string` matches the regexp set by `regexp`.                                                                                                                                                                            |
| `matches` | list of strings | (only when `regexp` option is set) List of strings containing capturing groups within the regular expression, numbered from left to right in order of opening parenthesis. First element is the match of the entire expression, submatch 1 the match of the first parenthesized subexpression, and so on. |
| `named`   | map of strings  | (only when `regexp` option is set) Map of strings containing named groups (?<name>expr) in the regexp.                                                                                                                                                                                                    |
