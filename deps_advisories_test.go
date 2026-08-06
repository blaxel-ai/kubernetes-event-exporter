package main

// Guards the four modules bumped to clear this fork's open critical/high advisories.
//
// Why a version assertion rather than a behavioural test: every one of the nine
// golang.org/x/crypto criticals is in golang.org/x/crypto/ssh or .../ssh/agent, and
// neither package appears in this binary's dependency graph (`go list -deps ./...`
// returns nothing for them). There is no reachable code path to exercise, so a test
// pretending to reproduce the weakness would prove nothing. What can regress is the
// version, and that is what this checks.
//
// It reads go.mod rather than using x/mod/semver so the check adds no dependency of
// its own.

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Minimum version per module, from each advisory's first patched release.
// otel is listed at 1.43.0 rather than its advisory floor of 1.41.0 because
// google.golang.org/grpc v1.82.1 requires 1.43.0 -- pinning 1.41.0 makes the module
// graph unresolvable, so the two cannot be chosen independently.
var minimums = map[string]string{
	"golang.org/x/crypto":             "v0.52.0", // 9 criticals, x/crypto/ssh + ssh/agent
	"google.golang.org/grpc":          "v1.82.1", // 2 high
	"github.com/sirupsen/logrus":      "v1.8.3",  // 1 high
	"go.opentelemetry.io/otel":        "v1.43.0", // 1 high (floor 1.41.0, raised by grpc)
	"go.opentelemetry.io/otel/metric": "v1.43.0", // kept in lockstep with otel
	"go.opentelemetry.io/otel/trace":  "v1.43.0", // kept in lockstep with otel
}

// requireLine matches a go.mod require entry, with or without the // indirect marker.
var requireLine = regexp.MustCompile(`(?m)^\s*(\S+)\s+(v\S+?)(?:\s+//.*)?\s*$`)

// compareVersions returns -1, 0 or 1 comparing two vMAJOR.MINOR.PATCH strings.
// Pre-release suffixes are ignored: none of the pinned modules use them, and treating
// "v1.2.3-rc1" as equal to "v1.2.3" would be the wrong way to be wrong -- so any
// suffix is reported so the test fails loudly instead of guessing.
func compareVersions(t *testing.T, a, b string) int {
	t.Helper()
	parse := func(v string) [3]int {
		v = strings.TrimPrefix(v, "v")
		if i := strings.IndexAny(v, "-+"); i >= 0 {
			t.Fatalf("version %q carries a pre-release or build suffix; this comparison does not handle those, tighten it before relying on it", v)
		}
		var out [3]int
		for i, part := range strings.SplitN(v, ".", 3) {
			n, err := strconv.Atoi(part)
			if err != nil {
				t.Fatalf("cannot parse version %q: %v", v, err)
			}
			out[i] = n
		}
		return out
	}
	x, y := parse(a), parse(b)
	for i := range x {
		switch {
		case x[i] < y[i]:
			return -1
		case x[i] > y[i]:
			return 1
		}
	}
	return 0
}

func TestDependencyAdvisoryMinimums(t *testing.T) {
	raw, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("cannot read go.mod: %v", err)
	}

	found := make(map[string]string)
	for _, m := range requireLine.FindAllStringSubmatch(string(raw), -1) {
		if _, watched := minimums[m[1]]; watched {
			found[m[1]] = m[2]
		}
	}

	for module, min := range minimums {
		got, ok := found[module]
		if !ok {
			// Not "pass because it is gone": a module dropping out of go.mod silently
			// would otherwise disable this assertion.
			t.Errorf("%s is not in go.mod; if it was intentionally dropped, remove it from this test in the same commit", module)
			continue
		}
		if compareVersions(t, got, min) < 0 {
			t.Errorf("%s is %s, below the advisory floor %s -- a patched version was reverted", module, got, min)
		}
	}
}
