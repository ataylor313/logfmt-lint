# logfmt-lint

A lot of Go services (and tools like Heroku's router, Vault, Consul) log in
logfmt: space-separated `key=value` pairs, one line per event. It's easy to
read but there's usually nothing checking that the lines are well-formed or
that the fields you actually depend on downstream — a timestamp, a level,
a message — are present. A deploy that quietly drops `level=` from every
line doesn't fail until someone's alerting query comes up empty.

`logfmt-lint` parses each line, checks it against a small built-in schema
(`time` must be RFC3339, `level` must be one of `debug|info|warn|error|fatal`,
`msg` must be present), and reports the result. It exits non-zero if any
line failed, so it's usable as a CI check on log fixtures or a smoke test
against a sample of production output.

## Usage

```
go run . access.log
```

```
line 1: ok
line 2: invalid
    field level has unrecognized value: "verbose"
    raw: time=2024-01-15T10:31:02Z level=verbose msg="cache miss" key=user:42

2 lines, 1 valid, 1 invalid
```

Machine-readable output, for feeding into other tooling:

```
go run . --json access.log
```

```json
[
  {
    "line": 1,
    "raw": "time=2024-01-15T10:30:00Z level=info msg=\"request completed\" status=200",
    "valid": true,
    "fields": {
      "level": "info",
      "msg": "request completed",
      "status": "200",
      "time": "2024-01-15T10:30:00Z"
    }
  },
  {
    "line": 2,
    "raw": "time=2024-01-15T10:31:02Z level=verbose msg=\"cache miss\" key=user:42",
    "valid": false,
    "errors": [
      "field level has unrecognized value: \"verbose\""
    ],
    "fields": {
      "key": "user:42",
      "level": "verbose",
      "msg": "cache miss",
      "time": "2024-01-15T10:31:02Z"
    }
  }
]
```

With no file arguments it reads from stdin, so it composes with `tail -f`,
`grep`, or a CI job that pipes in a captured log sample.

## Building

Standard library only, no dependencies to fetch:

```
go build .
```

## Current limitations

The validation schema is hardcoded (see `validate.go`). Quoted values
follow Go string escaping via `strconv.Unquote`, which covers the common
case but isn't a full logfmt implementation. See the roadmap in the repo
description for what's planned next.

## License

MIT, see LICENSE.
