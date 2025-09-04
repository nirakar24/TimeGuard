# TimeGuard

TimeGuard is a cross-platform CLI tool for handling tricky time concepts: timezone conversions, DST transitions, leap seconds, leap smearing, and log timestamp validation.

## Commands

| Command | Purpose | Example |
|---------|---------|---------|
| convert | Convert a local datetime from one IANA TZ to another | `timeguard convert "2025-09-04 14:00" --from Asia/Kolkata --to America/New_York` |
| dst-check | Detect gap/overlap around DST transitions | `timeguard dst-check "2025-03-30 02:30" --zone Europe/Berlin` |
| leap-check | Check if supplied UTC second is a defined leap second | `timeguard leap-check 2016-12-31T23:59:60Z` |
| smear | Simulate Google-style 24h leap smear | `timeguard smear 2016-12-31 --method google` |
| now | Show current time + offset + DST flag in zone | `timeguard now --zone Australia/Adelaide` |
| validate-logs | Scan log file for timestamp issues | `timeguard validate-logs ./server.log` |

### Log Validation Heuristics
Flags:
1. parse-error – timestamp failed RFC3339 parsing
2. time-regression – backward jump >5 minutes
3. missing-timezone-context – line lacks explicit TZ hint (TZ= or zone=)

Planned: ambiguous-overlap, non-existent-gap detection, mixed-format detection.

## Build

```
go build -o timeguard .
```

## Examples

```
$ timeguard convert "2025-09-04 14:00" --from Asia/Kolkata --to America/New_York
$ timeguard dst-check "2025-03-30 02:30" --zone Europe/Berlin
$ timeguard leap-check 2016-12-31T23:59:60Z
$ timeguard smear 2016-12-31 --method google
$ timeguard now --zone Australia/Adelaide
$ timeguard validate-logs ./server.log
```

## Leap Second Data

`internal/timeutil/leapdata.json` contains known leap seconds. Update periodically from IERS.

## Packaging / Release

Goreleaser config (`.goreleaser.yml`) produces:
- Binaries (linux/darwin/windows; amd64/arm64)
- Homebrew formula (tap: `example/homebrew-timeguard`)
- Scoop manifest (bucket: `example/scoop-timeguard`)
- Debian/RPM packages via nfpm

Dry run:
```
goreleaser release --clean --snapshot
```

## Contributing

Pull requests welcome. Put reusable logic in `internal/timeutil` and keep CLI wiring in `cmd/`.

Test:
```
go test ./...
```

## License

MIT
