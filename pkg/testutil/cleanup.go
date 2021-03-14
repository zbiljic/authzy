package testutil

// Cleanup contains a list of function that are called to cleanup a fixture.
type Cleanup struct {
	funcs []func()
}

// Add adds function to funcs list.
func (c *Cleanup) Add(f ...func()) {
	c.funcs = append(c.funcs, f...)
}

// AppendFront append funcs from another cleanup in front of the funcs list.
func (c *Cleanup) AppendFront(c1 *Cleanup) {
	c.funcs = append(c1.funcs, c.funcs...)
}

// Recover runs cleanup functions after test exit with exception.
func (c *Cleanup) Recover() {
	if err := recover(); err != nil {
		c.run()
		panic(err)
	}
}

// Run runs cleanup functions when a test finishes running.
func (c *Cleanup) Run() {
	c.run()
}

func (c *Cleanup) run() {
	for _, f := range c.funcs {
		f()
	}
}
