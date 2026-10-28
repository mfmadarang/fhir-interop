package validate

import "fmt"

// represents a single validation problem found on a resource
type Issue struct {
	ResourceType string
	ResourceID   string
	Field        string
	Message      string
	Warning      bool // true when we couldn't check it, false means a real problem
}

func (i Issue) String() string {
	id := i.ResourceID
	if id == "" {
		id = "(no id)"
	}
	return fmt.Sprintf("%s/%s: %s: %s", i.ResourceType, id, i.Field, i.Message)
}
