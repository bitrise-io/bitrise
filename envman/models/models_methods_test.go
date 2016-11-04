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
			Category:          pointers.NewStringPtr("category"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(false),
			IsDontChangeValue: pointers.NewBoolPtr(true),
			IsTemplate:        pointers.NewBoolPtr(false),
			SkipIfEmpty:       pointers.NewBoolPtr(false),
		},
	}

	key, value, err := env.GetKeyValuePair()
	require.NoError(t, err)

	require.Equal(t, "test_key", key)
	require.Equal(t, "test_value", value)

	// More then 2 fields
	env = EnvironmentItemModel{
		"test_key":  "test_value",
		"test_key1": "test_value1",
		OptionsKey:  EnvironmentItemOptionsModel{Title: pointers.NewStringPtr("test_title")},
	}

	key, value, err = env.GetKeyValuePair()
	require.EqualError(t, err, `more than 2 keys specified: [opts test_key test_key1]`)

	// 2 key-value fields
	env = EnvironmentItemModel{
		"test_key":  "test_value",
		"test_key1": "test_value1",
	}

	key, value, err = env.GetKeyValuePair()
	require.EqualError(t, err, `more than 1 environment key specified: [test_key test_key1]`)

	// Not string value
	env = EnvironmentItemModel{"test_key": true}

	key, value, err = env.GetKeyValuePair()
	require.NoError(t, err)

	require.Equal(t, "test_key", key)
	require.Equal(t, "true", value)

	// Empty key
	env = EnvironmentItemModel{"": "test_value"}

	key, value, err = env.GetKeyValuePair()
	require.EqualError(t, err, "no environment key found, keys: []")

	// Missing key-value
	env = EnvironmentItemModel{OptionsKey: EnvironmentItemOptionsModel{Title: pointers.NewStringPtr("test_title")}}

	key, value, err = env.GetKeyValuePair()
	require.EqualError(t, err, "no environment key found, keys: [opts]")
}

func TestParseFromInterfaceMap(t *testing.T) {
	envOptions := EnvironmentItemOptionsModel{}
	model := map[string]interface{}{}

	// Normal
	model["title"] = "test_title"
	model["value_options"] = []string{"test_key2", "test_value2"}
	model["is_expand"] = true
	require.NoError(t, envOptions.ParseFromInterfaceMap(model))

	// title is not a string
	model = map[string]interface{}{}
	model["title"] = true
	require.NoError(t, envOptions.ParseFromInterfaceMap(model))

	// value_options is not a string slice
	model = map[string]interface{}{}
	model["value_options"] = []interface{}{true, false}
	require.NoError(t, envOptions.ParseFromInterfaceMap(model))

	// is_required is not a bool
	model = map[string]interface{}{}
	model["is_required"] = pointers.NewBoolPtr(true)
	require.NotEqual(t, nil, envOptions.ParseFromInterfaceMap(model))

	model = map[string]interface{}{}
	model["is_required"] = "YeS"
	require.NoError(t, envOptions.ParseFromInterfaceMap(model))

	model = map[string]interface{}{}
	model["is_required"] = "NO"
	require.NoError(t, envOptions.ParseFromInterfaceMap(model))

	model = map[string]interface{}{}
	model["is_required"] = "y"
	require.NoError(t, envOptions.ParseFromInterfaceMap(model))

	model = map[string]interface{}{}
	model["skip_if_empty"] = "true"
	require.NoError(t, envOptions.ParseFromInterfaceMap(model))

	// other_key is not supported key
	model = map[string]interface{}{}
	model["other_key"] = true
	require.EqualError(t, envOptions.ParseFromInterfaceMap(model), "not supported key found in options: other_key")
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
	require.NoError(t, err)

	require.NotNil(t, opts.Title)
	require.Equal(t, "test_title", *opts.Title)

	require.NotNil(t, opts.IsExpand)
	require.Equal(t, false, *opts.IsExpand)

	// Missing opts
	env = EnvironmentItemModel{
		"test_key": "test_value",
	}
	_, err = env.GetOptions()
	require.NoError(t, err)

	// Wrong opts
	env = EnvironmentItemModel{
		"test_key": "test_value",
		OptionsKey: map[interface{}]interface{}{
			"title": "test_title",
			"test":  "test_description",
		},
	}
	_, err = env.GetOptions()
	require.EqualError(t, err, "not supported key found in options: test")
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

	require.NoError(t, env.Normalize())

	opts, err := env.GetOptions()
	require.NoError(t, err)

	require.NotNil(t, opts.Title)
	require.Equal(t, "test_title", *opts.Title)

	require.NotNil(t, opts.Description)
	require.Equal(t, "test_description", *opts.Description)

	require.NotNil(t, opts.Summary)
	require.Equal(t, "test_summary", *opts.Summary)

	require.Equal(t, 2, len(opts.ValueOptions))

	require.NotNil(t, opts.IsRequired)
	require.Equal(t, true, *opts.IsRequired)

	require.NotNil(t, opts.SkipIfEmpty)
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

	require.NoError(t, env.Normalize())

	opts, err = env.GetOptions()
	require.NoError(t, err)

	require.NotNil(t, opts.Title)
	require.Equal(t, "test_title", *opts.Title)

	require.NotNil(t, opts.Description)
	require.Equal(t, "test_description", *opts.Description)

	require.NotNil(t, opts.Summary)
	require.Equal(t, "test_summary", *opts.Summary)

	require.Equal(t, 2, len(opts.ValueOptions))

	require.NotNil(t, opts.IsRequired)
	require.Equal(t, true, *opts.IsRequired)

	// Empty options
	env = EnvironmentItemModel{
		"test_key": "test_value",
	}

	require.NoError(t, env.Normalize())

	opts, err = env.GetOptions()
	require.NoError(t, err)

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

	require.NoError(t, env.FillMissingDefaults())

	opts, err := env.GetOptions()
	require.NoError(t, err)

	require.NotNil(t, opts.Description)
	require.Equal(t, "", *opts.Description)

	require.NotNil(t, opts.Summary)
	require.Equal(t, "", *opts.Summary)

	require.NotNil(t, opts.Category)
	require.Equal(t, "", *opts.Category)

	require.NotNil(t, opts.IsRequired)
	require.Equal(t, DefaultIsRequired, *opts.IsRequired)

	require.NotNil(t, opts.IsExpand)
	require.Equal(t, DefaultIsExpand, *opts.IsExpand)

	require.NotNil(t, opts.IsDontChangeValue)
	require.Equal(t, DefaultIsDontChangeValue, *opts.IsDontChangeValue)

	require.NotNil(t, opts.IsTemplate)
	require.Equal(t, DefaultIsDontChangeValue, *opts.IsTemplate)

	require.NotNil(t, opts.SkipIfEmpty)
	require.Equal(t, DefaultSkipIfEmpty, *opts.SkipIfEmpty)

	// Filled env
	env = EnvironmentItemModel{
		"test_key": "test_value",
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			Summary:           pointers.NewStringPtr("test_summary"),
			Category:          pointers.NewStringPtr("required"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(true),
			IsDontChangeValue: pointers.NewBoolPtr(false),
			IsTemplate:        pointers.NewBoolPtr(false),
			SkipIfEmpty:       pointers.NewBoolPtr(false),
		},
	}

	require.NoError(t, env.FillMissingDefaults())

	opts, err = env.GetOptions()
	require.NoError(t, err)

	require.NotNil(t, opts.Title)
	require.Equal(t, "test_title", *opts.Title)

	require.NotNil(t, opts.Description)
	require.Equal(t, "test_description", *opts.Description)

	require.NotNil(t, opts.Summary)
	require.Equal(t, "test_summary", *opts.Summary)

	require.NotNil(t, opts.Category)
	require.Equal(t, "required", *opts.Category)

	require.Equal(t, 2, len(opts.ValueOptions))

	require.NotNil(t, opts.IsRequired)
	require.Equal(t, true, *opts.IsRequired)

	require.NotNil(t, opts.IsExpand)
	require.Equal(t, true, *opts.IsExpand)

	require.NotNil(t, opts.IsDontChangeValue)
	require.Equal(t, false, *opts.IsDontChangeValue)

	require.NotNil(t, opts.IsTemplate)
	require.Equal(t, false, *opts.IsTemplate)

	require.NotNil(t, opts.SkipIfEmpty)
	require.Equal(t, false, *opts.SkipIfEmpty)
}

func TestValidate(t *testing.T) {
	// No key-value
	env := EnvironmentItemModel{
		OptionsKey: EnvironmentItemOptionsModel{
			Title:             pointers.NewStringPtr("test_title"),
			Description:       pointers.NewStringPtr("test_description"),
			Summary:           pointers.NewStringPtr("test_summary"),
			Category:          pointers.NewStringPtr("required"),
			ValueOptions:      []string{"test_key2", "test_value2"},
			IsRequired:        pointers.NewBoolPtr(true),
			IsExpand:          pointers.NewBoolPtr(true),
			IsDontChangeValue: pointers.NewBoolPtr(false),
		},
	}
	require.EqualError(t, env.Validate(), "no environment key found, keys: [opts]")

	// Empty key
	env = EnvironmentItemModel{
		"": "test_value",
	}
	require.EqualError(t, env.Validate(), "no environment key found, keys: []")

	// Valid env
	env = EnvironmentItemModel{
		"test_key": "test_value",
	}
	require.NoError(t, env.Validate())
}
