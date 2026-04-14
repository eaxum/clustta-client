package system_icon

import "fmt"

// GetExtensionIcon returns a PNG image of the system icon for the given file extension.
// Linux does not have a native API for this, so we return an error.
func GetExtensionIcon(extension string) ([]byte, error) {
	return nil, fmt.Errorf("system icon retrieval is not supported on Linux for extension %s", extension)
}
