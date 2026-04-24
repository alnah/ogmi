package cli_test

import "testing"

func TestVersionCommandDefaultsToDev(t *testing.T) {
	result := runOgmi(t, "version")
	requireSuccess(t, result)
	if got, want := result.stdout, "ogmi version dev\n"; got != want {
		t.Fatalf("ogmi version output = %q, want %q", got, want)
	}
}
