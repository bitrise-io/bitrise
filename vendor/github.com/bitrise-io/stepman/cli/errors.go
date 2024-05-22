package cli

import "fmt"

var errStepNotAvailableOfflineMode error = fmt.Errorf("step not available in offline mode")
