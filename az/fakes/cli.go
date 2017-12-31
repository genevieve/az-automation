package fakes

type CLI struct {
	ExecuteCall struct {
		CallCount int
		Receives  struct {
			Args []string
		}
		Returns struct {
			Output string
			Error  error
		}
	}
}

func (c *CLI) Execute(args []string) (string, error) {
	c.ExecuteCall.CallCount++
	c.ExecuteCall.Receives.Args = args

	return c.ExecuteCall.Returns.Output, c.ExecuteCall.Returns.Error
}
