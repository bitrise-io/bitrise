package integration

import (
	"testing"

	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestStepInfo(t *testing.T) {
	out, err := command.New(binPath(), "setup", "-c", defaultLibraryURI).RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)

	t.Log("library step")
	{
		out, err = command.New(binPath(), "step-info", "--collection", defaultLibraryURI, "--id", "apk-info", "--version", "1.0.4").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, apkInfo104Defintiion, out)
	}

	t.Log("library step --format json")
	{
		out, err = command.New(binPath(), "step-info", "--collection", defaultLibraryURI, "--id", "apk-info", "--version", "1.0.4", "--format", "json").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, true, strings.Contains(out, apkInfo104DefintiionJSON), out)
	}

	t.Log("local step")
	{
		out, err := command.New(binPath(), "step-info", "--collection", "path", "--id", "./test-step").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, localTestStepDefintion, out)
	}

	t.Log("local step - deprecated --step-yml flag")
	{
		out, err := command.New(binPath(), "step-info", "--step-yml", "./test-step").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, localTestStepDefintion, out)
	}

	t.Log("local step --format json")
	{
		out, err := command.New(binPath(), "step-info", "--collection", "path", "--id", "./test-step", "--format", "json").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, localTestStepDefintionJSON, out)
	}

	t.Log("git step")
	{
		out, err := command.New(binPath(), "step-info", "--collection", "git", "--id", "https://github.com/bitrise-steplib/steps-xamarin-user-management.git", "--version", "1.0.3").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, gitTestStepDefinition, out)
	}

	t.Log("git step --format json")
	{
		out, err := command.New(binPath(), "step-info", "--collection", "git", "--id", "https://github.com/bitrise-steplib/steps-xamarin-user-management.git", "--version", "1.0.3", "--format", "json").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
		require.Equal(t, true, strings.Contains(out, gitTestStepDefintionJSON), out)
	}
}

func TestStepInfoExitCode(t *testing.T) {
	t.Log("default setup - desired exit code: 0")
	{
		out, err := command.New(binPath(), "setup", "--collection", defaultLibraryURI).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("latest version - desired exit code: 0")
	{
		out, err := command.New(binPath(), "step-info", "--collection", defaultLibraryURI, "--id", "script").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("latest version, json format - desired exit code: 0")
	{
		out, err := command.New(binPath(), "step-info", "--collection", defaultLibraryURI, "--id", "script", "--format", "json").RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("invalid version -1 - desired exit code: NOT 0")
	{
		out, err := command.New(binPath(), "step-info", "--collection", defaultLibraryURI, "--id", "script", "--version", "-1").RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}

	t.Log("invalid version -1 - desired exit code: NOT 0")
	{
		out, err := command.New(binPath(), "step-info", "--collection", defaultLibraryURI, "--id", "script", "--version", "-1", "--format", "json").RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)
	}
}

const gitTestStepDefintionJSON = `{"library":"git","id":"https://github.com/bitrise-steplib/steps-xamarin-user-management.git","version":"1.0.3","info":{},"step":{"title":"Xamarin User Management","summary":"This step helps you authenticate your user with Xamarin and to download your Xamarin liceses.","description":"This step helps you authenticate your user with Xamarin and to download your Xamarin licenses.","website":"https://github.com/bitrise-steplib/steps-xamarin-user-management","source_code_url":"https://github.com/bitrise-steplib/steps-xamarin-user-management","support_url":"https://github.com/bitrise-steplib/steps-xamarin-user-management/issues","host_os_tags":["osx-10.10"],"project_type_tags":["xamarin"],"is_requires_admin_user":false,"is_always_run":true,"is_skippable":false,"run_if":".IsCI","timeout":0,"inputs":[{"build_slug":"$BITRISE_BUILD_SLUG","opts":{"is_expand":true,"skip_if_empty":false,"title":"Bitrise build slug","description":"Bitrise build slug\n","summary":"","category":"","is_required":true,"is_dont_change_value":false,"is_template":false}},{"opts":{"is_expand":true,"skip_if_empty":false,"title":"Xamarin.iOS License","description":"Set to yes if you want to download the Xamarin.iOS license file\n","summary":"","category":"","value_options":["yes","no"],"is_required":true,"is_dont_change_value":false,"is_template":false},"xamarin_ios_license":"yes"},{"opts":{"is_expand":true,"skip_if_empty":false,"title":"Xamarin.Android License","description":"Set to yes if you want to download the Xamarin.Android license file\n","summary":"","category":"","value_options":["yes","no"],"is_required":true,"is_dont_change_value":false,"is_template":false},"xamarin_android_license":"yes"},{"opts":{"is_expand":true,"skip_if_empty":false,"title":"Xamarin.Mac License","description":"Set to yes if you want to download the Xamarin.Mac license file\n","summary":"","category":"","value_options":["yes","no"],"is_required":true,"is_dont_change_value":false,"is_template":false},"xamarin_mac_license":"no"}]}`

const gitTestStepDefinition = "\x1b[34;1m" + `Library:` + "\x1b[0m" + ` git
` + "\x1b[34;1m" + `ID:` + "\x1b[0m" + ` https://github.com/bitrise-steplib/steps-xamarin-user-management.git
` + "\x1b[34;1m" + `Version:` + "\x1b[0m" + ` 1.0.3
` + "\x1b[34;1m" + `LatestVersion:` + "\x1b[0m" + ` 
` + "\x1b[34;1m" + `Definition:` + "\x1b[0m" + `

title: "Xamarin User Management"
summary: This step helps you authenticate your user with Xamarin and to download your Xamarin liceses.
description: |-
  This step helps you authenticate your user with Xamarin and to download your Xamarin licenses.
website: https://github.com/bitrise-steplib/steps-xamarin-user-management
source_code_url: https://github.com/bitrise-steplib/steps-xamarin-user-management
support_url: https://github.com/bitrise-steplib/steps-xamarin-user-management/issues
host_os_tags:
  - osx-10.10
project_type_tags:
  - xamarin
type_tags:
is_requires_admin_user: false
is_always_run: true
is_skippable: false
run_if: .IsCI
inputs:
  - build_slug: $BITRISE_BUILD_SLUG
    opts:
      title: Bitrise build slug
      description: |
        Bitrise build slug
      is_required: true
      is_expand: true
  - xamarin_ios_license: "yes"
    opts:
      title: Xamarin.iOS License
      description: |
        Set to yes if you want to download the Xamarin.iOS license file
      value_options:
          - "yes"
          - "no"
      is_required: true
      is_expand: true
  - xamarin_android_license: "yes"
    opts:
      title: Xamarin.Android License
      description: |
        Set to yes if you want to download the Xamarin.Android license file
      value_options:
          - "yes"
          - "no"
      is_required: true
      is_expand: true
  - xamarin_mac_license: "no"
    opts:
      title: Xamarin.Mac License
      description: |
        Set to yes if you want to download the Xamarin.Mac license file
      value_options:
          - "yes"
          - "no"
      is_required: true
      is_expand: true`

const localTestStepDefintionJSON = "{\"library\":\"path\",\"id\":\"./test-step\",\"info\":{},\"step\":{\"title\":\"STEP TEMPLATE\",\"summary\":\"A short summary of the step. Don't make it too long ;)\",\"description\":\"This is a Step template.\\nContains everything what's required for a valid Stepman managed step.\\n\\nA Step's description (and generally any description property)\\ncan be a [Markdown](https://en.wikipedia.org/wiki/Markdown) formatted text.\\n\\nTo create your own Step:\\n\\n1. Create a new repository on GitHub\\n2. Copy the files from this folder into your repository\\n3. That's all, you can use it on your own machine\\n4. Once you're happy with it you can share it with others.\",\"website\":\"https://github.com/...\",\"source_code_url\":\"https://github.com/...\",\"support_url\":\"https://github.com/.../issues\",\"host_os_tags\":[\"osx-10.10\"],\"project_type_tags\":[\"ios\",\"android\",\"xamarin\"],\"type_tags\":[\"script\"],\"deps\":{\"brew\":[{\"name\":\"git\"},{\"name\":\"wget\"}],\"apt_get\":[{\"name\":\"git\"},{\"name\":\"wget\"}]},\"is_requires_admin_user\":true,\"is_always_run\":false,\"is_skippable\":false,\"run_if\":\"\",\"timeout\":0,\"inputs\":[{\"example_step_input\":\"Default Value - you can leave this empty if you want to\",\"opts\":{\"is_expand\":true,\"skip_if_empty\":false,\"title\":\"Example Step Input\",\"description\":\"Description of this input.\\n\\nCan be Markdown formatted text.\\n\",\"summary\":\"Summary. No more than 2-3 sentences.\",\"category\":\"\",\"is_required\":true,\"is_dont_change_value\":false,\"is_template\":false}}],\"outputs\":[{\"EXAMPLE_STEP_OUTPUT\":null,\"opts\":{\"is_expand\":true,\"skip_if_empty\":false,\"title\":\"Example Step Output\",\"description\":\"Description of this output.\\n\\nCan be Markdown formatted text.\\n\",\"summary\":\"Summary. No more than 2-3 sentences.\",\"category\":\"\",\"is_required\":false,\"is_dont_change_value\":false,\"is_template\":false}}]},\"definition_pth\":\"test-step/step.yml\"}"

const localTestStepDefintion = "\x1b[34;1m" + `Library:` + "\x1b[0m" + ` path
` + "\x1b[34;1m" + `ID:` + "\x1b[0m" + ` ./test-step
` + "\x1b[34;1m" + `Version:` + "\x1b[0m" + ` 
` + "\x1b[34;1m" + `LatestVersion:` + "\x1b[0m" + ` 
` + "\x1b[34;1m" + `Definition:` + "\x1b[0m" + `

title: "STEP TEMPLATE"
summary: A short summary of the step. Don't make it too long ;)
description: |-
  This is a Step template.
  Contains everything what's required for a valid Stepman managed step.

  A Step's description (and generally any description property)
  can be a [Markdown](https://en.wikipedia.org/wiki/Markdown) formatted text.

  To create your own Step:

  1. Create a new repository on GitHub
  2. Copy the files from this folder into your repository
  3. That's all, you can use it on your own machine
  4. Once you're happy with it you can share it with others.
website: https://github.com/...
source_code_url: https://github.com/...
support_url: https://github.com/.../issues
host_os_tags:
  - osx-10.10
project_type_tags:
  - ios
  - android
  - xamarin
type_tags:
  - script
is_requires_admin_user: true
is_always_run: false
is_skippable: false
deps:
  brew:
  - name: git
  - name: wget
  apt_get:
  - name: git
  - name: wget
run_if: ""
inputs:
  - example_step_input: Default Value - you can leave this empty if you want to
    opts:
      title: "Example Step Input"
      summary: Summary. No more than 2-3 sentences.
      description: |
        Description of this input.

        Can be Markdown formatted text.
      is_expand: true
      is_required: true
      value_options: []
outputs:
  - EXAMPLE_STEP_OUTPUT:
    opts:
      title: "Example Step Output"
      summary: Summary. No more than 2-3 sentences.
      description: |
        Description of this output.

        Can be Markdown formatted text.`

const apkInfo104DefintiionJSON = "{\"library\":\"https://github.com/bitrise-io/bitrise-steplib.git\",\"id\":\"apk-info\",\"version\":\"1.0.4\",\"latest_version\":\"1.3.0\",\"info\":{},\"step\":{\"title\":\"APK info\",\"summary\":\"APK Android info provider\",\"description\":\"Provides all possible Android APK information as package name, version name or version code.\",\"website\":\"https://github.com/thefuntasty/bitrise-step-apk-info\",\"source_code_url\":\"https://github.com/thefuntasty/bitrise-step-apk-info\",\"support_url\":\"https://github.com/thefuntasty/bitrise-step-apk-info/issues\",\"published_at\":\"2016-10-19T15:35:00.882498804+02:00\",\"source\":{\"git\":\"https://github.com/thefuntasty/bitrise-step-apk-info.git\",\"commit\":\"104e26a8800fc9363658b5837cf4747e5f26b032\"},\"asset_urls\":{\"icon.svg\":\"https://bitrise-steplib-collection.s3.amazonaws.com/steps/apk-info/assets/icon.svg\"},\"project_type_tags\":[\"android\"],\"type_tags\":[\"android\",\"apk\"],\"is_requires_admin_user\":false,\"is_always_run\":false,\"is_skippable\":false,\"run_if\":\"\",\"timeout\":0,\"inputs\":[{\"apk_path\":\"$BITRISE_APK_PATH\",\"opts\":{\"category\":\"\",\"description\":\"File path to APK file to get info from.\\n\",\"is_dont_change_value\":false,\"is_expand\":true,\"is_required\":true,\"is_template\":false,\"skip_if_empty\":false,\"summary\":\"\",\"title\":\"APK file path\"}}],\"outputs\":[{\"ANDROID_APP_PACKAGE_NAME\":null,\"opts\":{\"category\":\"\",\"description\":\"Android application package name, ex. com.package.my\",\"is_dont_change_value\":false,\"is_expand\":true,\"is_required\":false,\"is_template\":false,\"skip_if_empty\":false,\"summary\":\"\",\"title\":\"Android application package name\"}},{\"ANDROID_APK_FILE_SIZE\":null,\"opts\":{\"category\":\"\",\"description\":\"Android APK file size, in bytes\",\"is_dont_change_value\":false,\"is_expand\":true,\"is_required\":false,\"is_template\":false,\"skip_if_empty\":false,\"summary\":\"\",\"title\":\"Android APK file size\"}},{\"ANDROID_APP_NAME\":null,\"opts\":{\"category\":\"\",\"description\":\"Android application name from APK\",\"is_dont_change_value\":false,\"is_expand\":true,\"is_required\":false,\"is_template\":false,\"skip_if_empty\":false,\"summary\":\"\",\"title\":\"Android application name\"}},{\"ANDROID_APP_VERSION_NAME\":null,\"opts\":{\"category\":\"\",\"description\":\"Android application version name from APK, ex. 1.0.0\",\"is_dont_change_value\":false,\"is_expand\":true,\"is_required\":false,\"is_template\":false,\"skip_if_empty\":false,\"summary\":\"\",\"title\":\"Android application version name\"}},{\"ANDROID_APP_VERSION_CODE\":null,\"opts\":{\"category\":\"\",\"description\":\"Android application version code from APK, ex. 10\",\"is_dont_change_value\":false,\"is_expand\":true,\"is_required\":false,\"is_template\":false,\"skip_if_empty\":false,\"summary\":\"\",\"title\":\"Android application version code\"}},{\"ANDROID_ICON_PATH\":null,\"opts\":{\"category\":\"\",\"description\":\"File path to android application icon\",\"is_dont_change_value\":false,\"is_expand\":true,\"is_required\":false,\"is_template\":false,\"skip_if_empty\":false,\"summary\":\"\",\"title\":\"File path to icon\"}}]}"

const apkInfo104Defintiion = "\x1b[34;1m" + `Library:` + "\x1b[0m" + ` https://github.com/bitrise-io/bitrise-steplib.git
` + "\x1b[34;1m" + `ID:` + "\x1b[0m" + ` apk-info
` + "\x1b[34;1m" + `Version:` + "\x1b[0m" + ` 1.0.4
` + "\x1b[34;1m" + `LatestVersion:` + "\x1b[0m" + ` 1.3.0
` + "\x1b[34;1m" + `Definition:` + "\x1b[0m" + `

title: APK info
summary: APK Android info provider
description: Provides all possible Android APK information as package name, version
  name or version code.
website: https://github.com/thefuntasty/bitrise-step-apk-info
source_code_url: https://github.com/thefuntasty/bitrise-step-apk-info
support_url: https://github.com/thefuntasty/bitrise-step-apk-info/issues
published_at: 2016-10-19T15:35:00.882498804+02:00
source:
  git: https://github.com/thefuntasty/bitrise-step-apk-info.git
  commit: 104e26a8800fc9363658b5837cf4747e5f26b032
project_type_tags:
- android
type_tags:
- android
- apk
is_requires_admin_user: false
is_always_run: false
is_skippable: false
inputs:
- apk_path: $BITRISE_APK_PATH
  opts:
    description: |
      File path to APK file to get info from.
    is_required: true
    title: APK file path
outputs:
- ANDROID_APP_PACKAGE_NAME: null
  opts:
    description: Android application package name, ex. com.package.my
    title: Android application package name
- ANDROID_APK_FILE_SIZE: null
  opts:
    description: Android APK file size, in bytes
    title: Android APK file size
- ANDROID_APP_NAME: null
  opts:
    description: Android application name from APK
    title: Android application name
- ANDROID_APP_VERSION_NAME: null
  opts:
    description: Android application version name from APK, ex. 1.0.0
    title: Android application version name
- ANDROID_APP_VERSION_CODE: null
  opts:
    description: Android application version code from APK, ex. 10
    title: Android application version code
- ANDROID_ICON_PATH: null
  opts:
    description: File path to android application icon
    title: File path to icon`
