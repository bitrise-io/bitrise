package debug

import (
	"fmt"
	"os"
	"path/filepath"
)

func W(str string) {
	f, err := os.OpenFile(filepath.Join(os.Getenv("BITRISE_DEPLOY_DIR"), "times.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	if _, err := f.WriteString(str); err != nil {
		fmt.Println(err)
	}
}
