package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/go-cmp/cmp"
)

const (
	validJSON = `[
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
	invalidJSON = `[
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
	negativeHealthCowboys = `[
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
		"health": -1,
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
	zeroHealthCowboys = `[
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
		"health": 0,
		"damage": 3
	},
	{
		"name": "Philip",
		"health": 15,
		"damage": 1
	}
]`
	negativeDamageCowboys = `[
	{
		"name": "John",
		"health": 10,
		"damage": -2
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
	zeroDamageCowboys = `[
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
		"damage": 0
	},
	{
		"name": "Philip",
		"health": 15,
		"damage": 1
	}
]`
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

func TestAreCowboyHealthValuesPositive(t *testing.T) {
	tests := []struct {
		name    string
		want    error
		cowboys []Cowboy
	}{
		{"no cowboys", nil, nil},
		{"1 cowboy with positive health", nil, []Cowboy{
			{Health: 1},
		}},
		{"2 cowboys with positive health", nil, []Cowboy{
			{Health: 1},
			{Health: 2},
		}},
		{"1 cowboy with 0 health", ErrCowboyHealthNotPositive, []Cowboy{
			{Health: 0},
		}},
		{"2 cowboys with 1 cowboy with 0 health", ErrCowboyHealthNotPositive, []Cowboy{
			{Health: 1},
			{Health: 0},
		}},
		{"2 cowboys with 0 health", ErrCowboyHealthNotPositive, []Cowboy{
			{Health: 0},
			{Health: 0},
		}},
		{"1 cowboy with negative health", ErrCowboyHealthNotPositive, []Cowboy{
			{Health: -1},
		}},
		{"2 cowboy with 1 cowboy with negative health", ErrCowboyHealthNotPositive, []Cowboy{
			{Health: 1},
			{Health: -1},
		}},
		{"3 cowboys with positive, 0, and negative health", ErrCowboyHealthNotPositive, []Cowboy{
			{Health: 1},
			{Health: 0},
			{Health: -1},
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := areCowboyHealthValuesPositive(tc.cowboys)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAreCowboyDamageValuesPositive(t *testing.T) {
	tests := []struct {
		name    string
		want    error
		cowboys []Cowboy
	}{
		{"no cowboys", nil, nil},
		{"1 cowboy with positive damage", nil, []Cowboy{
			{Damage: 1},
		}},
		{"2 cowboys with positive damage", nil, []Cowboy{
			{Damage: 1},
			{Damage: 2},
		}},
		{"1 cowboy with 0 damage", ErrCowboyDamageNotPositive, []Cowboy{
			{Damage: 0},
		}},
		{"2 cowboys with 1 cowboy with 0 damage", ErrCowboyDamageNotPositive, []Cowboy{
			{Damage: 1},
			{Damage: 0},
		}},
		{"2 cowboys with 0 damage", ErrCowboyDamageNotPositive, []Cowboy{
			{Damage: 0},
			{Damage: 0},
		}},
		{"1 cowboy with negative damage", ErrCowboyDamageNotPositive, []Cowboy{
			{Damage: -1},
		}},
		{"2 cowboy with 1 cowboy with negative damage", ErrCowboyDamageNotPositive, []Cowboy{
			{Damage: 1},
			{Damage: -1},
		}},
		{"3 cowboys with positive, 0, and negative damage", ErrCowboyDamageNotPositive, []Cowboy{
			{Damage: 1},
			{Damage: 0},
			{Damage: -1},
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := areCowboyDamageValuesPositive(tc.cowboys)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLoadCowboys(t *testing.T) {
	tests := []struct {
		name           string
		jsonCowboys    string
		want           []Cowboy
		wantErrMessage string
	}{
		{
			name:        "valid json",
			jsonCowboys: validJSON,
			want: []Cowboy{
				{Name: "John", Health: 10, Damage: 1},
				{Name: "Bill", Health: 8, Damage: 2},
				{Name: "Sam", Health: 10, Damage: 1},
				{Name: "Peter", Health: 5, Damage: 3},
				{Name: "Philip", Health: 15, Damage: 1},
			},
			wantErrMessage: "",
		},
		{
			name:           "invalid json",
			jsonCowboys:    invalidJSON,
			want:           nil,
			wantErrMessage: "unmarshal JSON:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			file, err := os.CreateTemp("", "*.json")
			assert.NoError(t, err)
			defer os.Remove(file.Name())

			err = os.WriteFile(file.Name(), []byte(tc.jsonCowboys), 0o644)
			assert.NoError(t, err)

			// execute
			got, err := loadCowboys(file.Name())

			// verify
			if err != nil {
				assert.Contains(t, err.Error(), tc.wantErrMessage)
			}

			if err == nil {
				assert.Empty(t, tc.wantErrMessage)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLoadCowboysNoFile(t *testing.T) {
	got, err := loadCowboys("file_that_does_not_exist.json")
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestGetCowboys(t *testing.T) {
	tests := []struct {
		name        string
		jsonCowboys string
		want        []Cowboy
		err         error
	}{
		{"valid cowboy list", validJSON, []Cowboy{
			{Name: "John", Health: 10, Damage: 1},
			{Name: "Bill", Health: 8, Damage: 2},
			{Name: "Sam", Health: 10, Damage: 1},
			{Name: "Peter", Health: 5, Damage: 3},
			{Name: "Philip", Health: 15, Damage: 1},
		}, nil},
		{"invalid cowboy list", invalidJSON, nil, ErrCowboyListInvalidJSONFormat},
		{"negative health cowboys", negativeHealthCowboys, nil, ErrCowboyHealthNotPositive},
		{"0 health cowboys", zeroHealthCowboys, nil, ErrCowboyHealthNotPositive},
		{"negative damage cowboys", negativeDamageCowboys, nil, ErrCowboyDamageNotPositive},
		{"zero damage cowboys", zeroDamageCowboys, nil, ErrCowboyDamageNotPositive},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			file, err := os.CreateTemp("", "*.json")
			assert.NoError(t, err)
			defer os.Remove(file.Name())

			err = os.WriteFile(file.Name(), []byte(tc.jsonCowboys), 0o644)
			assert.NoError(t, err)

			// execute
			got, err := GetCowboys(file.Name(), 1)

			// verify
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetCowboysListTooShort(t *testing.T) {
	tests := []struct {
		name        string
		jsonCowboys string
		replicas    int
		err         error
	}{
		{"5 cowboys, 3 replicas", validJSON, 3, nil},
		{"5 cowboys, 4 replicas", validJSON, 4, nil},
		{"5 cowboys, 5 replicas", validJSON, 5, nil},
		{"5 cowboys, 6 replicas", validJSON, 6, ErrCowboyListTooShort},
		{"5 cowboys, 6 replicas", validJSON, 7, ErrCowboyListTooShort},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			file, err := os.CreateTemp("", "*.json")
			assert.NoError(t, err)
			defer os.Remove(file.Name())

			err = os.WriteFile(file.Name(), []byte(tc.jsonCowboys), 0o644)
			assert.NoError(t, err)

			// execute
			_, err = GetCowboys(file.Name(), tc.replicas)

			// verify
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
