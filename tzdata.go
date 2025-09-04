//go:build !no_tzdata
// +build !no_tzdata

package main

// Importing time/tzdata embeds the IANA timezone database into the binary.
// This is required for Windows or minimal container images where the zoneinfo
// files are not available at runtime. Without it, commands using time.LoadLocation
// (e.g. --zone America/New_York) will return "unknown time zone" on hosts
// lacking /usr/share/zoneinfo or GOROOT/lib/time/zoneinfo.zip.
//
// To produce a smaller binary (excluding tzdata) you can build with:
//   go build -tags no_tzdata .
// or configure goreleaser with build flags for a separate "slim" variant.
import _ "time/tzdata"
