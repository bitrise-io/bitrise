package tools

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func TestEnvListToMap(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	m, err := envListToMap([]string{"TEST=test"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"TEST": "test"}, m)
}

func TestExpandEnvItems(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("expands using the external environment")
	{
		toExpand := []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"KEY": "${EXTERNAL_KEY} value", "opts": map[string]interface{}{"is_expand": true}},
		}
		external := []string{"EXTERNAL_KEY=some"}
		m, err := ExpandEnvItems(toExpand, external)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"KEY": "some value"}, m)
	}

	t.Log("external environment is extended by the expand")
	{
		toExpand := []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"HOME": "some/home/path", "opts": map[string]interface{}{"is_expand": true}},
			envmanModels.EnvironmentItemModel{"KEY": "${HOME}", "opts": map[string]interface{}{"is_expand": true}},
		}
		external := []string{}
		m, err := ExpandEnvItems(toExpand, external)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"KEY": "some/home/path", "HOME": "some/home/path"}, m)
	}

	t.Log("last value used")
	{
		toExpand := []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"KEY": "value1", "opts": map[string]interface{}{"is_expand": true}},
			envmanModels.EnvironmentItemModel{"KEY": "value2", "opts": map[string]interface{}{"is_expand": true}},
		}
		external := []string{}
		m, err := ExpandEnvItems(toExpand, external)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"KEY": "value2"}, m)
	}

	{
		toExpand := []envmanModels.EnvironmentItemModel{
			envmanModels.EnvironmentItemModel{"KEY": "some", "opts": map[string]interface{}{"is_expand": true}},
			envmanModels.EnvironmentItemModel{"KEY": "$KEY value", "opts": map[string]interface{}{"is_expand": true}},
			envmanModels.EnvironmentItemModel{"TEST": "$KEY", "opts": map[string]interface{}{"is_expand": true}},
		}
		external := []string{}
		m, err := ExpandEnvItems(toExpand, external)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"KEY": "some value", "TEST": "some value"}, m)
	}
}
