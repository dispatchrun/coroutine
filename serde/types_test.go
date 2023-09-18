package serde_test

import (
	"testing"
	"time"

	serdeinternal "github.com/stealthrocket/coroutine/internal/serde"
)

func TestSerdeTime(t *testing.T) {
	t.Run("time zero", func(t *testing.T) {
		testSerdeTime(t, time.Time{})
	})

	t.Run("time.Now", func(t *testing.T) {
		testSerdeTime(t, time.Now())
	})

	t.Run("fixed zone", func(t *testing.T) {
		parsed, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		if err != nil {
			t.Error(err)
		}
		loc, err := time.LoadLocation("US/Eastern")
		if err != nil {
			t.Error("failed to load location", err)
		}
		t2 := parsed.In(loc)

		testSerdeTime(t, t2)
	})
}

func testSerdeTime(t *testing.T, x time.Time) {
	b := serdeinternal.Serialize(x)
	out, b := serdeinternal.Deserialize(b)

	if !x.Equal(out.(time.Time)) {
		t.Errorf("expected %v, got %v", x, out)
	}
}
