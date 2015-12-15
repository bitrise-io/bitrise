package models

import (
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

func TestGetKeyValuePair(t *testing.T) {
	// Filled env
	env := EnvironmentItemModel{
		"test_key": "test_value",
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			Summary:           pointers.NewStringPtr("test_summary"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(false),
			IsDontChangeValue: pointers.NewBoolPtr(true),
			IsTemplate:        pointers.NewBoolPtr(false),
			SkipIfEmpty:       pointers.NewBoolPtr(false),
		},
	}

	key, value, err := env.GetKeyValuePair()
	require.Equal(t, nil, err)

	require.Equal(t, "test_key", key)
	require.Equal(t, "test_value", value)

	// More then 2 fields
	env = EnvironmentItemModel{
		"test_key":  "test_value",
		"test_key1": "test_value1",
		OptionsKey:  EnvironmentItemOptionsModel{Title: pointers.NewStringPtr("test_title")},
	}

	key, value, err = env.GetKeyValuePair()
	require.NotEqual(t, nil, err)

	// 2 key-value fields
	env = EnvironmentItemModel{
		"test_key":  "test_value",
		"test_key1": "test_value1",
	}

	key, value, err = env.GetKeyValuePair()
	require.NotEqual(t, nil, err)

	// Not string value
	env = EnvironmentItemModel{"test_key": true}

	key, value, err = env.GetKeyValuePair()
	require.Equal(t, nil, err)
	require.Equal(t, "test_key", key)
	require.Equal(t, "true", value)

	// Empty key
	env = EnvironmentItemModel{"": "test_value"}

	key, value, err = env.GetKeyValuePair()
	require.NotEqual(t, nil, err)

	// Missing key-value
	env = EnvironmentItemModel{OptionsKey: EnvironmentItemOptionsModel{Title: pointers.NewStringPtr("test_title")}}

	key, value, err = env.GetKeyValuePair()
	require.NotEqual(t, nil, err)
}

func TestParseFromInterfaceMap(t *testing.T) {
	envOptions := EnvironmentItemOptionsModel{}
	model := map[string]interface{}{}

	// Normal
	model["title"] = "test_title"
	model["value_options"] = []string{"test_key2", "test_value2"}
	model["is_expand"] = true
	require.Equal(t, nil, envOptions.ParseFromInterfaceMap(model))

	// title is not a string
	model = map[string]interface{}{}
	model["title"] = true
	require.Equal(t, nil, envOptions.ParseFromInterfaceMap(model))

	// value_options is not a string slice
	model = map[string]interface{}{}
	model["value_options"] = []interface{}{true, false}
	require.Equal(t, nil, envOptions.ParseFromInterfaceMap(model))

	// is_required is not a bool
	model = map[string]interface{}{}
	model["is_required"] = pointers.NewBoolPtr(true)
	require.NotEqual(t, nil, envOptions.ParseFromInterfaceMap(model))

	model = map[string]interface{}{}
	model["is_required"] = "YeS"
	require.Equal(t, nil, envOptions.ParseFromInterfaceMap(model))

	model = map[string]interface{}{}
	model["is_required"] = "NO"
	require.Equal(t, nil, envOptions.ParseFromInterfaceMap(model))

	model = map[string]interface{}{}
	model["is_required"] = "y"
	require.Equal(t, nil, envOptions.ParseFromInterfaceMap(model))

	model = map[string]interface{}{}
	model["skip_if_empty"] = "true"
	require.Equal(t, nil, envOptions.ParseFromInterfaceMap(model))

	// other_key is not supported key
	model = map[string]interface{}{}
	model["other_key"] = true
	require.NotEqual(t, nil, envOptions.ParseFromInterfaceMap(model))
}

func TestGetOptions(t *testing.T) {
	// Filled env
	env := EnvironmentItemModel{
		"test_key": "test_value",
		OptionsKey: EnvironmentItemOptionsModel{
			Title:    pointers.NewStringPtr("test_title"),
			IsExpand: pointers.NewBoolPtr(false),
		},
	}
	opts, err := env.GetOptions()
	require.Equal(t, nil, err)

	require.NotEqual(t, nil, opts.Title)
	require.Equal(t, "test_title", *opts.Title)

	require.NotEqual(t, nil, opts.IsExpand)
	require.Equal(t, false, *opts.IsExpand)

	// Missing opts
	env = EnvironmentItemModel{
		"test_key": "test_value",
	}
	_, err = env.GetOptions()
	require.Equal(t, nil, err)

	// Wrong opts
	env = EnvironmentItemModel{
		"test_key": "test_value",
		OptionsKey: map[interface{}]interface{}{
			"title": "test_title",
			"test":  "test_description",
		},
	}
	_, err = env.GetOptions()
	require.NotEqual(t, nil, err)
}

func TestNormalize(t *testing.T) {
	// Filled with map[string]interface{} options
	env := EnvironmentItemModel{
		"test_key": "test_value",
		OptionsKey: map[interface{}]interface{}{
			"title":         "test_title",
			"description":   "test_description",
			"summary":       "test_summary",
			"value_options": []string{"test_key2", "test_value2"},
			"is_required":   true,
			"skip_if_empty": false,
		},
	}

	require.Equal(t, nil, env.Normalize())

	opts, err := env.GetOptions()
	require.Equal(t, nil, err)

	require.NotEqual(t, nil, opts.Title)
	require.Equal(t, "test_title", *opts.Title)

	require.NotEqual(t, nil, opts.Description)
	require.Equal(t, "test_description", *opts.Description)

	require.NotEqual(t, nil, opts.Summary)
	require.Equal(t, "test_summary", *opts.Summary)

	require.Equal(t, 2, len(opts.ValueOptions))

	require.NotEqual(t, nil, opts.IsRequired)
	require.Equal(t, true, *opts.IsRequired)

	require.Equal(t, false, *opts.SkipIfEmpty)

	// Filled with EnvironmentItemOptionsModel options
	env = EnvironmentItemModel{
		"test_key": "test_value",
		OptionsKey: EnvironmentItemOptionsModel{
			Title:        pointers.NewStringPtr("test_title"),
			Description:  pointers.NewStringPtr("test_description"),
			Summary:      pointers.NewStringPtr("test_summary"),
			ValueOptions: []string{"test_key2", "test_value2"},
			IsRequired:   pointers.NewBoolPtr(true),
		},
	}

	require.Equal(t, nil, env.Normalize())

	opts, err = env.GetOptions()
	require.Equal(t, nil, err)

	require.NotEqual(t, nil, opts.Title)
	require.Equal(t, "test_title", *opts.Title)

	require.NotEqual(t, nil, opts.Description)
	require.Equal(t, "test_description", *opts.Description)

	require.NotEqual(t, nil, opts.Summary)
	require.Equal(t, "test_summary", *opts.Summary)

	require.Equal(t, 2, len(opts.ValueOptions))

	require.NotEqual(t, nil, opts.IsRequired)
	require.Equal(t, true, *opts.IsRequired)

	// Empty options
	env = EnvironmentItemModel{
		"test_key": "test_value",
	}

	require.Equal(t, nil, env.Normalize())

	opts, err = env.GetOptions()
	require.Equal(t, nil, err)

	require.Equal(t, (*string)(nil), opts.Title)
	require.Equal(t, (*string)(nil), opts.Description)
	require.Equal(t, (*string)(nil), opts.Summary)
	require.Equal(t, 0, len(opts.ValueOptions))
	require.Equal(t, (*bool)(nil), opts.IsRequired)
	require.Equal(t, (*bool)(nil), opts.IsDontChangeValue)
	require.Equal(t, (*bool)(nil), opts.IsExpand)
	require.Equal(t, (*bool)(nil), opts.IsTemplate)
	require.Equal(t, (*bool)(nil), opts.SkipIfEmpty)
}

func TestFillMissingDefaults(t *testing.T) {
	// Empty env
	env := EnvironmentItemModel{
		"test_key": "test_value",
	}

	require.Equal(t, nil, env.FillMissingDefaults())

	opts, err := env.GetOptions()
	require.Equal(t, nil, err)

	require.NotEqual(t, nil, opts.Description)
	require.Equal(t, "", *opts.Description)

	require.NotEqual(t, nil, opts.Summary)
	require.Equal(t, "", *opts.Summary)

	require.NotEqual(t, nil, opts.IsRequired)
	require.Equal(t, DefaultIsRequired, *opts.IsRequired)

	require.NotEqual(t, nil, opts.IsExpand)
	require.Equal(t, DefaultIsExpand, *opts.IsExpand)

	require.NotEqual(t, nil, opts.IsDontChangeValue)
	require.Equal(t, DefaultIsDontChangeValue, *opts.IsDontChangeValue)

	require.NotEqual(t, nil, opts.IsTemplate)
	require.Equal(t, DefaultIsDontChangeValue, *opts.IsTemplate)

	require.NotEqual(t, nil, opts.SkipIfEmpty)
	require.Equal(t, DefaultSkipIfEmpty, *opts.SkipIfEmpty)

	// Filled env
	env = EnvironmentItemModel{
		"test_key": "test_value",
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			Summary:           pointers.NewStringPtr("test_summary"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(true),
			IsDontChangeValue: pointers.NewBoolPtr(false),
			IsTemplate:        pointers.NewBoolPtr(false),
			SkipIfEmpty:       pointers.NewBoolPtr(false),
		},
	}

	require.Equal(t, nil, env.FillMissingDefaults())

	opts, err = env.GetOptions()
	require.Equal(t, nil, err)

	require.NotEqual(t, nil, opts.Title)
	require.Equal(t, "test_title", *opts.Title)

	require.NotEqual(t, nil, opts.Description)
	require.Equal(t, "test_description", *opts.Description)

	require.NotEqual(t, nil, opts.Summary)
	require.Equal(t, "test_summary", *opts.Summary)

	require.Equal(t, 2, len(opts.ValueOptions))

	require.NotEqual(t, nil, opts.IsRequired)
	require.Equal(t, true, *opts.IsRequired)

	require.NotEqual(t, nil, opts.IsExpand)
	require.Equal(t, true, *opts.IsExpand)

	require.NotEqual(t, nil, opts.IsDontChangeValue)
	require.Equal(t, false, *opts.IsDontChangeValue)

	require.NotEqual(t, nil, opts.IsTemplate)
	require.Equal(t, false, *opts.IsTemplate)

	require.NotEqual(t, nil, opts.SkipIfEmpty)
	require.Equal(t, false, *opts.SkipIfEmpty)
}

func TestValidate(t *testing.T) {
	// No key-value
	env := EnvironmentItemModel{
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			Summary:           pointers.NewStringPtr("test_summary"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(true),
			IsDontChangeValue: pointers.NewBoolPtr(false),
		},
	}
	require.NotEqual(t, nil, env.Validate())

	// Empty key
	env = EnvironmentItemModel{
		"": "test_value",
	}
	require.NotEqual(t, nil, env.Validate())

	// Valid env
	env = EnvironmentItemModel{
		"test_key": "test_value",
	}
	require.Equal(t, nil, env.Validate())
}
