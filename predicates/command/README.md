# command

The `command` predicate runs an external command.

## Options

| Option      | Required | Description                                           |
| ---         | ---      | ---                                                   |
| `cmd`       | yes      | Name of the command to run                            |
| `chdir`     |          | Change into this directory before running the command |
| `Args`      |          | List of command line args                             |
| `Stdin`     |          | String to pass to command stdin                       |
| `Backgound` |          | If true, don't wait for the command return            |

## Results

| Field    | Type    | Description                                                            |
| ---      | ---     | ---                                                                    |
| `result` | boolean | false if command return code is not 0. Use `result` option to override |
| `rc`     | int     | Return code of the command                                             |
| `stdout` | string  | standard output                                                        |
| `stderr` | string  | error output                                                           |

## Example:

``` yaml
command:
  cmd: tr
  args:
    - A-Z
    - a-z
  stdin: "Hello World!"
register: lower
log:
  msg: R.lower.stdout
```
