package stats

import (
	"testing"
	"time"
)

// TestLineTimeHonoursTheZuluMarker is the guard for changing the device's
// timezone.
//
// kixdns stamps every line in UTC ("…T09:19:05.199374387Z"). Parsing that as
// local time is only correct while the device itself runs on UTC; after a switch
// to UTC+8 the same line would land eight hours later, moving every hourly
// bucket, every daily total and the session start. The zone in the stamp must
// win over the machine's zone.
func TestLineTimeHonoursTheZuluMarker(t *testing.T) {
	const line = `2026-09-16T09:19:05.199374387Z  INFO config loaded target="config" version=2.0`

	zones := []struct {
		name string
		loc  *time.Location
	}{
		{"UTC", time.UTC},
		{"Asia/Shanghai (+8)", time.FixedZone("CST", 8*3600)},
		{"America/New_York (-4)", time.FixedZone("EDT", -4*3600)},
	}

	var first int64
	for i, z := range zones {
		got, ok := lineTime(line, z.loc)
		if !ok {
			t.Fatalf("%s: failed to parse the line", z.name)
		}
		if i == 0 {
			first = got.Unix()
			// 09:19:05Z is 09:19:05 UTC by definition.
			if h := got.UTC().Hour(); h != 9 {
				t.Errorf("UTC: hour = %d, want 9", h)
			}
			continue
		}
		if got.Unix() != first {
			t.Errorf("%s: instant = %d, want %d (the instant must not depend on the machine's zone)",
				z.name, got.Unix(), first)
		}
	}

	// The local hour legitimately differs per zone — that is the bucketing.
	sh, _ := lineTime(line, time.FixedZone("CST", 8*3600))
	if h := sh.Hour(); h != 17 {
		t.Errorf("Asia/Shanghai: hour = %d, want 17 (09:19Z + 8h)", h)
	}
	t.Logf("09:19:05Z -> UTC hour 9, UTC+8 hour 17, same instant")
}

// TestLineTimeWithoutZoneFallsBackToLocal covers a stamp that carries no zone,
// and a line with no timestamp at all.
func TestLineTimeWithoutZoneFallsBackToLocal(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	got, ok := lineTime("2026-09-16T09:19:05 some event", loc)
	if !ok {
		t.Fatal("failed to parse a zoneless stamp")
	}
	if h := got.Hour(); h != 9 {
		t.Errorf("zoneless: hour = %d, want 9 (read as the log's own zone)", h)
	}

	if _, ok := lineTime("no timestamp here", loc); ok {
		t.Error("a line without a timestamp should not parse")
	}
}

// TestLineTimeWithNumericOffset covers an explicit +08:00 style stamp.
func TestLineTimeWithNumericOffset(t *testing.T) {
	got, ok := lineTime("2026-09-16T17:19:05+08:00 event", time.UTC)
	if !ok {
		t.Fatal("failed to parse an offset stamp")
	}
	if got.UTC().Hour() != 9 {
		t.Errorf("UTC hour = %d, want 9", got.UTC().Hour())
	}
}
