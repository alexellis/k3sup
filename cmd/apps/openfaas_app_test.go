package apps

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func Test_getValuesSuffix_arm64(t *testing.T) {
	want := "-arm64"
	got := getValuesSuffix("arm64")
	if want != got {
		t.Errorf("suffix, want: %s, got: %s", want, got)
	}
}

func Test_getValuesSuffix_aarch64(t *testing.T) {
	want := "-arm64"
	got := getValuesSuffix("aarch64")
	if want != got {
		t.Errorf("suffix, want: %s, got: %s", want, got)
	}
}

func Test_mergeFlags(t *testing.T) {
	var (
		unexpectedErr         = "failed to merge err: %v, existing flags: %v, set overrides: %v"
		unexpectedMergeErr    = "error merging flags, want: %v, got: %v"
		inconsistentErrString = "inconsistent error return want: %v, got: %v"
	)
	tests := []struct {
		title          string
		flags          map[string]string
		overrides      []string
		resultExpected map[string]string
		errExpected    error
	}{
		// positive cases:
		{"Single key with numeric value and no flags",
			map[string]string{},
			[]string{"b=1"},
			map[string]string{"b": "1"},
			nil,
		},
		{"Empty set and no flags, should not fail",
			map[string]string{},
			[]string{},
			map[string]string{},
			nil,
		},
		{"No set key and an existing flag",
			map[string]string{"a": "1"},
			[]string{},
			map[string]string{"a": "1"},
			nil,
		},
		{"Single key with numeric value, override the flag with numeric value",
			map[string]string{"a": "1"},
			[]string{"a=2"},
			map[string]string{"a": "2"},
			nil,
		},
		{"Single key with numeric value and single flag key with numeric value",
			map[string]string{"a": "1"},
			[]string{"b=1"},
			map[string]string{"a": "1", "b": "1"},
			nil,
		},
		{"Single key with numeric value, update existing key and a new key",
			map[string]string{"a": "1"},
			[]string{"a=2", "b=1"},
			map[string]string{"a": "2", "b": "1"},
			nil,
		},
		{"Update all existing flags in the map",
			map[string]string{"a": "1", "b": "2"},
			[]string{"a=2", "b=3"},
			map[string]string{"a": "2", "b": "3"},
			nil,
		},

		// check errors
		{"Incorrect flag format, providing : as a delimiter",
			map[string]string{"a": "1"},
			[]string{"a:2"},
			nil,
			fmt.Errorf("incorrect format for custom flag `a:2`"),
		},
		{"Incorrect flag format, providing space as a delimiter",
			map[string]string{"a": "1"}, []string{"a 2"},
			nil,
			fmt.Errorf("incorrect format for custom flag `a 2`"),
		},
		{"Incorrect flag format, providing - as a delimiter",
			map[string]string{"a": "1"},
			[]string{"a-2"},
			nil,
			fmt.Errorf("incorrect format for custom flag `a-2`"),
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			err := mergeFlags(test.flags, test.overrides)
			if err != nil {
				if test.errExpected == nil {
					t.Errorf(unexpectedErr, err, test.flags, test.overrides)
				} else if !strings.EqualFold(err.Error(), test.errExpected.Error()) {
					t.Errorf(inconsistentErrString, test.errExpected, err)
				}
				return
			}
			if !reflect.DeepEqual(test.resultExpected, test.flags) {
				t.Errorf(unexpectedMergeErr, test.resultExpected, test.flags)
			}
		})
	}
}
