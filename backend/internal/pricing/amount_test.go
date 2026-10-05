package pricing

import "testing"

func TestParseMajorAmount(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  int64
	}{
		{"0", 0},
		{"1", 100},
		{"1.5", 150},
		{"123.45", 12345},
	} {
		got, err := ParseMajorAmount(tc.input)
		if err != nil {
			t.Fatalf("%q: %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("%q: got %d want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseMajorAmountRejectsInvalid(t *testing.T) {
	for _, value := range []string{"", "-1", "1.234", "abc", ".50"} {
		if _, err := ParseMajorAmount(value); err == nil {
			t.Fatalf("expected %q to fail", value)
		}
	}
}
