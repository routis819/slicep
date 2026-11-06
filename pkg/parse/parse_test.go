package parse_test

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/parse"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []parse.Datum
		wantErr bool
	}{
		{
			name:  "Empty input",
			input: "   ; only comment\n",
			want:  []parse.Datum{},
		},
		{
			name:  "Integer literals",
			input: "123 -45 +67",
			want: []parse.Datum{
				&parse.IntDatum{Value: 123},
				&parse.IntDatum{Value: -45},
				&parse.IntDatum{Value: 67},
			},
		},
		{
			name:  "Float literals",
			input: "1.25 -0.5 1e3 +inf.0 -inf.0",
			want: []parse.Datum{
				&parse.FloatDatum{Value: 1.25},
				&parse.FloatDatum{Value: -0.5},
				&parse.FloatDatum{Value: 1000},
				&parse.FloatDatum{Value: math.Inf(1)},
				&parse.FloatDatum{Value: math.Inf(-1)},
			},
		},
		{
			name:  "Boolean literals",
			input: "#t #f #true #false",
			want: []parse.Datum{
				&parse.BoolDatum{Value: true},
				&parse.BoolDatum{Value: false},
				&parse.BoolDatum{Value: true},
				&parse.BoolDatum{Value: false},
			},
		},
		{
			name:  "String literals",
			input: `"hello" "a\nb\t\"\\" "line1\  ` + "\n" + `  line2"`,
			want: []parse.Datum{
				&parse.StringDatum{Value: "hello"},
				&parse.StringDatum{Value: "a\nb\t\"\\"},
				&parse.StringDatum{Value: "line1line2"},
			},
		},
		{
			name:  "Symbol literal",
			input: "foo",
			want: []parse.Datum{
				&parse.SymbolDatum{Value: "foo"},
			},
		},
		{
			name:  "Simple list",
			input: "(foo bar)",
			want: []parse.Datum{
				&parse.ListDatum{
					Elements: []parse.Datum{
						&parse.SymbolDatum{Value: "foo"},
						&parse.SymbolDatum{Value: "bar"},
					},
				},
			},
		},
		{
			name:  "Nested list and multiple top-level datums",
			input: "(define x (+ 1 2.5)) (set! x (* x 3))",
			want: []parse.Datum{
				&parse.ListDatum{
					Elements: []parse.Datum{
						&parse.SymbolDatum{Value: "define"},
						&parse.SymbolDatum{Value: "x"},
						&parse.ListDatum{
							Elements: []parse.Datum{
								&parse.SymbolDatum{Value: "+"},
								&parse.IntDatum{Value: 1},
								&parse.FloatDatum{Value: 2.5},
							},
						},
					},
				},
				&parse.ListDatum{
					Elements: []parse.Datum{
						&parse.SymbolDatum{Value: "set!"},
						&parse.SymbolDatum{Value: "x"},
						&parse.ListDatum{
							Elements: []parse.Datum{
								&parse.SymbolDatum{Value: "*"},
								&parse.SymbolDatum{Value: "x"},
								&parse.IntDatum{Value: 3},
							},
						},
					},
				},
			},
		},
		{
			name:    "Error: Unclosed parenthesis",
			input:   "(+ 1 2",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error: Unexpected closing parenthesis",
			input:   ")",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error: Lexical error propagated",
			input:   "123abc",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := parse.Parse(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
