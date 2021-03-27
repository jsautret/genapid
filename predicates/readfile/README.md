# readfile

The `readfile` predicate reads and parse the value of file.

## Options

Option | Required | Description
---|---|---
`json` | | Path to a JSON file
`yaml` | | Path to a YAML file

One of `json` or `yaml` must be present.

## Results

Field | Type | Description
---|---|---
`result` | boolean | true if file was read
`content` | according to file content | The parsed content of the file
