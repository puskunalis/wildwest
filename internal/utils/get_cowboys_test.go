package utils

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAreCowboyNamesUnique(t *testing.T) {
	tests := []struct {
		name    string
		cowboys []Cowboy
		err     bool
	}{
		{
			name: "unique names",
			cowboys: []Cowboy{
				{Name: "John"},
				{Name: "Bill"},
				{Name: "Sam"},
				{Name: "Peter"},
				{Name: "Philip"},
			},
			err: false,
		},
		{
			name: "duplicate names",
			cowboys: []Cowboy{
				{Name: "John"},
				{Name: "Bill"},
				{Name: "Sam"},
				{Name: "Peter"},
				{Name: "Peter"},
			},
			err: true,
		},
		{
			name: "only duplicate names",
			cowboys: []Cowboy{
				{Name: "John"},
				{Name: "John"},
			},
			err: true,
		},
		{
			name: "three duplicate names",
			cowboys: []Cowboy{
				{Name: "John"},
				{Name: "John"},
				{Name: "John"},
			},
			err: true,
		},
		{
			name: "single name",
			cowboys: []Cowboy{
				{Name: "John"},
			},
			err: false,
		},
		{
			name:    "empty list",
			cowboys: []Cowboy{},
			err:     false,
		},
		{
			name:    "nil list",
			cowboys: nil,
			err:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := areCowboyNamesUnique(tc.cowboys)
			gotErr := got != nil

			if diff := cmp.Diff(tc.err, gotErr); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestLoadCowboys(t *testing.T) {
	validJSON := `[
	{
		"name": "John",
		"health": 10,
		"damage": 1
	},
	{
		"name": "Bill",
		"health": 8,
		"damage": 2
	},
	{
		"name": "Sam",
		"health": 10,
		"damage": 1
	},
	{
		"name": "Peter",
		"health": 5,
		"damage": 3
	},
	{
		"name": "Philip",
		"health": 15,
		"damage": 1
	}
]`

	invalidJSON := `[
	{
		"name": "John",
		"health": 10,
		"damage": 1
	},
	{
		"name": "Bill",
		"health": 8,
		"damage": 2
	}
	{
		"name": "Sam",
		"health": 10,
		"damage": 1
	}
]`

	// create temporary files for testing
	validFile, err := os.CreateTemp("", "valid_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(validFile.Name())

	invalidFile, err := os.CreateTemp("", "invalid_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(invalidFile.Name())

	// write json data to temporary files
	if err := os.WriteFile(validFile.Name(), []byte(validJSON), 0o644); err != nil {
		panic(err)
	}
	if err := os.WriteFile(invalidFile.Name(), []byte(invalidJSON), 0o644); err != nil {
		panic(err)
	}

	tests := []struct {
		name           string
		filename       string
		want           []Cowboy
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:     "valid json",
			filename: validFile.Name(),
			want: []Cowboy{
				{Name: "John", Health: 10, Damage: 1},
				{Name: "Bill", Health: 8, Damage: 2},
				{Name: "Sam", Health: 10, Damage: 1},
				{Name: "Peter", Health: 5, Damage: 3},
				{Name: "Philip", Health: 15, Damage: 1},
			},
			wantErr:        false,
			wantErrMessage: "",
		},
		{
			name:           "invalid json",
			filename:       invalidFile.Name(),
			want:           nil,
			wantErr:        true,
			wantErrMessage: "unmarshal JSON:",
		},
		{
			name:           "file not found",
			filename:       "non_existent_file.json",
			want:           nil,
			wantErr:        true,
			wantErrMessage: "open file:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := loadCowboys(tc.filename)

			if !tc.wantErr {
				if diff := cmp.Diff(err, nil); diff != "" {
					t.Fatalf(diff)
				}
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Fatalf(diff)
				}
			} else {
				if diff := cmp.Diff(err, nil); diff == "" {
					t.Fatalf(diff)
				}

				// TODO properly compare errors
				if !strings.Contains(err.Error(), tc.wantErrMessage) {
					t.Fatalf("expected error message to contain '%s', but got '%s'", tc.wantErrMessage, err.Error())
				}
			}
		})
	}
}
