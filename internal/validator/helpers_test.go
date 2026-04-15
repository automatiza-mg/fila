package validator

import "testing"

func TestNotBlank(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "not blank",
			value: "foo",
			want:  true,
		},
		{
			name:  "not blank (with space)",
			value: " foo ",
			want:  true,
		},
		{
			name:  "blank",
			value: "",
		},
		{
			name:  "blank (tab)",
			value: "\t",
		},
		{
			name:  "blank (new line)",
			value: "\r\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NotBlank(tt.value)
			if tt.want != got {
				t.Fatalf("expected NotBlank(%q) to return %t", tt.value, tt.want)
			}
		})
	}
}

func TestEmailRX(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "valid email",
			value: "foo@bar.com",
			want:  true,
		},
		{
			name:  "blank",
			value: "",
		},
		{
			name:  "invalid email",
			value: "foo@bar@.com",
		},
		{
			name:  "no domain",
			value: "foo@",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Matches(tt.value, EmailRX)
			if tt.want != got {
				t.Fatalf("expected IsEmail(%q) to return %t", tt.value, tt.want)
			}
		})
	}
}

func TestCpfRX(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "formatted",
			value: "123.456.789-00",
			want:  true,
		},
		{
			name:  "unformatted",
			value: "12345678900",
			want:  true,
		},
		{
			name:  "blank",
			value: "",
		},
		{
			name:  "too short",
			value: "1234567890",
		},
		{
			name:  "too long",
			value: "123456789001",
		},
		{
			name:  "letters",
			value: "1234567890a",
		},
		{
			name:  "partial formatting",
			value: "123.456.78900",
		},
		{
			name:  "missing dash",
			value: "123.456.789.00",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Matches(tt.value, CpfRX)
			if tt.want != got {
				t.Fatalf("expected CpfRX.Match(%q) to return %t", tt.value, tt.want)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value string
		n     int
		want  bool
	}{
		{
			name:  "valid length",
			value: "foo",
			n:     2,
			want:  true,
		},
		{
			name:  "valid length (exact)",
			value: "foo",
			n:     3,
			want:  true,
		},
		{
			name:  "invalid length",
			value: "foo",
			n:     4,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := MinLength(tt.value, tt.n)
			if tt.want != got {
				t.Fatalf("expected MinLength(%q, %d) to return %t", tt.value, tt.n, tt.want)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value string
		n     int
		want  bool
	}{
		{
			name:  "valid length",
			value: "foo",
			n:     4,
			want:  true,
		},
		{
			name:  "valid length (exact)",
			value: "foo",
			n:     3,
			want:  true,
		},
		{
			name:  "invalid length",
			value: "foo",
			n:     2,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := MaxLength(tt.value, tt.n)
			if tt.want != got {
				t.Fatalf("expected MaxLength(%q, %d) to return %t", tt.value, tt.n, tt.want)
			}
		})
	}
}
