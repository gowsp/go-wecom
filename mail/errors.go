package mail

import "fmt"

func errPair(op, fields string) error {
	return fmt.Errorf("wecom: %s: one of %s is required", op, fields)
}

func errValue(op, field string) error {
	return fmt.Errorf("wecom: %s: %s has an unsupported value", op, field)
}
