package envman

import (
	"testing"

	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/utils"
)

var (
	testKey          = "test_key"
	testValue        = "test_value"
	testKey1         = "test_key1"
	testValue1       = "test_value1"
	testKey2         = "test_key2"
	testValue2       = "test_value2"
	testTitle        = "test_title"
	testDescription  = "test_description"
	testValueOptions = []string{testKey2, testValue2}
	testTrue         = true
	testFalse        = false
	testEmptyString  = ""
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
	err := env1.FillMissingDefaults()
	if err != nil {
		t.Fatal(err)
	}

	env2 := models.EnvironmentItemModel{
		"test_key2": "test_value2",
	}
	err = env2.FillMissingDefaults()
	if err != nil {
		t.Fatal(err)
	}

	// Should add to list, but not override
	oldEnvSlice := []models.EnvironmentItemModel{env1, env2}
	newEnvSlice, err := UpdateOrAddToEnvlist(oldEnvSlice, env1, false)
	if err != nil {
		t.Fatal(err)
	}

	env1Cnt, err := countOfEnvKeyInEnvSlice(env1, newEnvSlice)
	if err != nil {
		t.Fatal(err)
	}
	if env1Cnt != 2 {
		t.Fatalf("Failed to proper add env, %d x (test_key1)", env1Cnt)
	}

	env2Cnt, err := countOfEnvKeyInEnvSlice(env2, newEnvSlice)
	if err != nil {
		t.Fatal(err)
	}
	if env2Cnt != 1 {
		t.Fatalf("Failed to proper add env, %d x (test_key2)", env2Cnt)
	}

	// Should update list
	oldEnvSlice = []models.EnvironmentItemModel{env1, env2}
	newEnvSlice, err = UpdateOrAddToEnvlist(oldEnvSlice, env1, true)
	if err != nil {
		t.Fatal(err)
	}

	env1Cnt, err = countOfEnvKeyInEnvSlice(env1, newEnvSlice)
	if err != nil {
		t.Fatal(err)
	}
	if env1Cnt != 1 {
		t.Fatalf("Failed to proper add env, %d x (test_key1)", env1Cnt)
	}

	env2Cnt, err = countOfEnvKeyInEnvSlice(env2, newEnvSlice)
	if err != nil {
		t.Fatal(err)
	}
	if env2Cnt != 1 {
		t.Fatalf("Failed to proper add env, %d x (test_key2)", env2Cnt)
	}
}

func TestRemoveDefaults(t *testing.T) {
	defaultIsRequired := models.DefaultIsRequired
	defaultIsExpand := models.DefaultIsExpand
	defaultIsDontChangeValue := models.DefaultIsDontChangeValue

	// Filled env
	env := models.EnvironmentItemModel{
		testKey: testValue,
		models.OptionsKey: models.EnvironmentItemOptionsModel{
			Title:             utils.NewStringPtr(testTitle),
			Description:       utils.NewStringPtr(testEmptyString),
			ValueOptions:      []string{},
			IsRequired:        utils.NewBoolPtr(defaultIsRequired),
			IsExpand:          utils.NewBoolPtr(defaultIsExpand),
			IsDontChangeValue: utils.NewBoolPtr(defaultIsDontChangeValue),
		},
	}

	err := removeDefaults(&env)
	if err != nil {
		t.Fatal(err)
	}

	opts, err := env.GetOptions()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Title == nil {
		t.Fatal("Removed Title")
	}
	if opts.Description != nil {
		t.Fatal("Failed to remove default Description")
	}
	if opts.IsRequired != nil {
		t.Fatal("Failed to remove default IsRequired")
	}
	if opts.IsExpand != nil {
		t.Fatal("Failed to remove default IsExpand")
	}
	if opts.IsDontChangeValue != nil {
		t.Fatal("Failed to remove default IsDontChangeValue")
	}
}

// func TestGenerateFormattedYMLForEnvModels(t *testing.T) {
// 	t.Log("TestGenerateFormattedYMLForEnvModels")
// }
//
// func TestClearPathIfExist(t *testing.T) {
// 	t.Log("TestClearPathIfExist")
// }
//
// func TestInitAtPath(t *testing.T) {
// 	t.Log("TestInitAtPath")
// }
//
// func TestReadEnvs(t *testing.T) {
// 	t.Log("TestReadEnvs")
// }
//
// func TestReadEnvsOrCreateEmptyList(t *testing.T) {
// 	t.Log("TestReadEnvsOrCreateEmptyList")
// }
//
// func TestWriteEnvMapToFile(t *testing.T) {
// 	t.Log("TestWriteEnvMapToFile")
// }
