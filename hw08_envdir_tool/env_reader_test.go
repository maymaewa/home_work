package main

import (
	"testing"
)

func TestReadDir(t *testing.T) {
	env, err := ReadDir("testdata/env")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("BAR", func(t *testing.T) {
		v, ok := env["BAR"]
		if !ok {
			t.Fatal("BAR not found")
		}

		if v.NeedRemove {
			t.Fatal("BAR should not be removed")
		}

		if v.Value != "bar" {
			t.Fatalf("expected bar, got %q", v.Value)
		}
	})

	t.Run("EMPTY", func(t *testing.T) {
    	v, ok := env["EMPTY"]
    	if !ok {
        	t.Fatal("EMPTY not found")
    	}

    	if v.NeedRemove {
        	t.Fatal("EMPTY should not be marked for removal")
    	}

    	if v.Value != "" {
        	t.Fatalf("expected empty string, got %q", v.Value)
    	}
	})

	t.Run("FOO", func(t *testing.T) {
	v, ok := env["FOO"]
	if !ok {
		t.Fatal("FOO not found")
	}

	if v.NeedRemove {
		t.Fatal("FOO should not be removed")
	}

	expected := "   foo\nwith new line"

	if v.Value != expected {
		t.Fatalf("expected %q, got %q", expected, v.Value)
	}
})

	t.Run("UNSET", func(t *testing.T) {
		v, ok := env["UNSET"]
		if !ok {
			t.Fatal("UNSET not found")
		}

		if !v.NeedRemove {
			t.Fatal("UNSET should be marked for removal")
		}
	})

	t.Run("HELLO", func(t *testing.T) {
		v, ok := env["HELLO"]
		if !ok {
			t.Fatal("HELLO not found")
		}

		if v.NeedRemove {
			t.Fatal("HELLO should not be removed")
		}

		if v.Value != "\"hello\"" && v.Value != "hello" {
			t.Fatalf("unexpected HELLO value: %q", v.Value)
		}
	})
}

func TestReadDir_NotExists(t *testing.T) {
	_, err := ReadDir("testdata/not_exists")

	if err == nil {
		t.Fatal("expected error")
	}
}