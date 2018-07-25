package envman

import (
	"testing"

	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

func countOfEnvInEnvSlice(env models.EnvironmentItemModel, envSlice []models.EnvironmentItemModel) (cnt int, err error) {
	for _, e := range envSlice {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return 0, err
		}

		k, v, err := e.GetKeyValuePair()
		if err != nil {
			return 0, err
		}

		if key == k && value == v {
			cnt++
		}
	}
	return
}

func countOfEnvKeyInEnvSlice(env models.EnvironmentItemModel, envSlice []models.EnvironmentItemModel) (cnt int, err error) {
	for _, e := range envSlice {
		key, _, err := env.GetKeyValuePair()
		if err != nil {
			return 0, err
		}

		k, _, err := e.GetKeyValuePair()
		if err != nil {
			return 0, err
		}

		if key == k {
			cnt++
		}
	}
	return
}

func TestUpdateOrAddToEnvlist(t *testing.T) {
	env1 := models.EnvironmentItemModel{
		"test_key1": "test_value1",
	}
	require.Equal(t, nil, env1.FillMissingDefaults())

	env2 := models.EnvironmentItemModel{
		"test_key2": "test_value2",
	}
	require.Equal(t, nil, env2.FillMissingDefaults())

	// Should add to list, but not override
	oldEnvSlice := []models.EnvironmentItemModel{env1, env2}
	newEnvSlice, err := UpdateOrAddToEnvlist(oldEnvSlice, env1, false)
	require.Equal(t, nil, err)

	env1Cnt, err := countOfEnvKeyInEnvSlice(env1, newEnvSlice)
	require.Equal(t, nil, err)
	require.Equal(t, 2, env1Cnt)

	env2Cnt, err := countOfEnvKeyInEnvSlice(env2, newEnvSlice)
	require.Equal(t, nil, err)
	require.Equal(t, 1, env2Cnt)

	// Should update list
	oldEnvSlice = []models.EnvironmentItemModel{env1, env2}
	newEnvSlice, err = UpdateOrAddToEnvlist(oldEnvSlice, env1, true)
	require.Equal(t, nil, err)

	env1Cnt, err = countOfEnvKeyInEnvSlice(env1, newEnvSlice)
	require.Equal(t, nil, err)
	require.Equal(t, 1, env1Cnt)

	env2Cnt, err = countOfEnvKeyInEnvSlice(env2, newEnvSlice)
	require.Equal(t, nil, err)
	require.Equal(t, 1, env2Cnt)
}

func TestRemoveDefaults(t *testing.T) {
	// Filled env
	env := models.EnvironmentItemModel{
		"test_key": "test_value",
		models.OptionsKey: models.EnvironmentItemOptionsModel{
			Title:      pointers.NewStringPtr("test_title"),
			IsTemplate: pointers.NewBoolPtr(!models.DefaultIsTemplate),

			Description:       pointers.NewStringPtr(""),
			Summary:           pointers.NewStringPtr(""),
			ValueOptions:      []string{},
			IsRequired:        pointers.NewBoolPtr(models.DefaultIsRequired),
			IsDontChangeValue: pointers.NewBoolPtr(models.DefaultIsDontChangeValue),
			IsExpand:          pointers.NewBoolPtr(models.DefaultIsExpand),
			IsSensitive:       pointers.NewBoolPtr(models.DefaultIsSensitive),
			SkipIfEmpty:       pointers.NewBoolPtr(models.DefaultSkipIfEmpty),
		},
	}

	require.Equal(t, nil, removeDefaults(&env))

	opts, err := env.GetOptions()
	require.Equal(t, nil, err)

	require.NotEqual(t, (*string)(nil), opts.Title)
	require.Equal(t, "test_title", *opts.Title)
	require.NotEqual(t, (*bool)(nil), opts.IsTemplate)
	require.Equal(t, !models.DefaultIsTemplate, *opts.IsTemplate)

	require.Equal(t, (*string)(nil), opts.Description)
	require.Equal(t, (*string)(nil), opts.Summary)
	require.Equal(t, 0, len(opts.ValueOptions))
	require.Equal(t, (*bool)(nil), opts.IsRequired)
	require.Equal(t, (*bool)(nil), opts.IsDontChangeValue)
	require.Equal(t, (*bool)(nil), opts.IsExpand)
	require.Equal(t, (*bool)(nil), opts.IsSensitive)
	require.Equal(t, (*bool)(nil), opts.SkipIfEmpty)

}
