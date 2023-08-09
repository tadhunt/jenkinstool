package jenkinstool

import(
	"testing"
)

func TestString(t *testing.T) {
	s1 := String("foo")
	if s1 != "foo" {
		t.Fatalf("1: expected foo")
	}

	s2 := String(&s1)
	if s2 != "foo" {
		t.Fatalf("2: expected foo got '%s'", s2)
	}

	var fv = float64(12345.6)
	f1 := String(fv)
	if f1 != "12345.6" {
		t.Fatalf("3: expected 12345.6 got '%s'", f1)
	}

	f2 := String(&fv)
	if f2 != "12345.6" {
		t.Fatalf("4: expected 12345.6 got '%s'", f2)
	}
}

