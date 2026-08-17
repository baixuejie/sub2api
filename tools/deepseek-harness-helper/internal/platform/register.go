package platform

import "fmt"

type Registration struct {
	Executable string
}

func RegisterProtocol(executable string) error {
	if executable == "" {
		return fmt.Errorf("executable path is required")
	}
	return registerProtocol(Registration{Executable: executable})
}
