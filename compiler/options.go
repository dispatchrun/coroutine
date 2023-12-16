package compiler

// Option configures the compiler.
type Option func(*compiler)

func CallgraphType(callgraphType string) Option {
	return func(c *compiler) {
		c.callgraphType = callgraphType
	}
}

func OnlyListFiles(enabled bool) Option {
	return func(c *compiler) {
		c.onlyListFiles = enabled
	}
}

func DebugColors(enabled bool) Option {
	return func(c *compiler) {
		c.debugColors = enabled
	}
}
