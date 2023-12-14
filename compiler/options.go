package compiler

// Option configures the compiler.
type Option func(*compiler)

func OnlyListFiles(enabled bool) Option {
	return func(c *compiler) {
		c.onlyListFiles = enabled
	}
}
