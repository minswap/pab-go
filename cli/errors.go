package cli

// CLIError is error when running cardano-cli
type CLIError struct {
	message string
	Args    []string
}

func NewCLIError(msg string, args []string) *CLIError {
	return &CLIError{msg, args}
}

func (e *CLIError) Error() string {
	return e.message
}
