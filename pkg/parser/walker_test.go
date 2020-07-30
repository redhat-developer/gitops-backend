package parser

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func get(k string, v interface{}) (interface{}, error) {
	switch v := v.(type) {
	case *[]interface{}:
		v = *v
	case *map[string]interface{}:
		v = *v
	}
	if v, ok := v.(*[]interface{}); ok {
		v = *v
	}
	if v, ok := v.(*map[string]interface{}); ok {
		v = *v
	}

	switch v := v.(type) {
	case []interface{}:
		i, err := strconv.Atoi(k)
		if err != nil {
			return nil, fmt.Errorf("unable to convert index %q", k)
		}
		if i < 0 || i >= len(v) {
			return nil, errors.New("index out of range")
		}

		return v[i], nil

	case map[string]interface{}:
		value, ok := v[k]
		if !ok {
			return nil, nil
		}
		return value, nil

	default:
		return nil, fmt.Errorf("cannot get key '%s' on type %T", k, v)
	}
}

func TestGet(t *testing.T) {
	data := map[string]interface{}{
		"test": map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"test": "value1",
				},
				{
					"test": "value2",
				},
			},
		},
	}

	res, err := get("test", data)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"value1", "value2"}
	if diff := cmp.Diff(want, res); diff != "" {
		t.Fatal("failed:\n%s", diff)
	}
}
