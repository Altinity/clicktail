Golang `regexp` package uses the [RE2 syntax](https://github.com/google/re2/wiki/Syntax).

Example CLI usage (from honeytail root)
```
honeytail -p regex -k $HONEYTAIL_WRITEKEY \
  -f some/path/system.log \
  --dataset 'MY_TEST_DATASET' \
  --backfill \
  --regex.line_regex="(?P<time>\d{2}:\d{2}:\d{2}) (?P<field1>\w+)" \
  --regex.line_regex="(?P<foo>\w+)" \
  --regex.timefield="time" \
  --regex.time_format="%H:%M:%S"
```
