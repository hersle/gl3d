package console

type Console struct {
	lines []string
	head int
	capacity int
	length int

	prompt string
}

func NewConsole() *Console {
	var c Console
	c.capacity = 10
	c.length = 0
	c.lines = make([]string, c.capacity)
	c.head = -1 // initially empty
	return &c
}

func (c *Console) AddLine(line string) {
	c.head = (c.head + 1) % c.capacity
	c.lines[c.head] = line
	if c.length < c.capacity {
		c.length++
	}
}

func (c *Console) String() string {
	str := ""
	if c.head != -1 {
		if c.full() {
			for d := 0; d < c.capacity; d++ {
				i := (c.head + 1 + d) % c.capacity
				str += c.lines[i] + "\n"
			}
		} else {
			for i := 0; i < c.length; i++ {
				str += c.lines[i] + "\n"
			}
		}
	}
	str += "> " + c.prompt
	return str
}

func (c *Console) Prompt() string {
	return c.prompt
}

func (c *Console) SetPrompt(prompt string) {
	c.prompt = prompt
}

func (c *Console) Execute() {
	c.AddLine(c.Prompt())
	// TODO: do command
	c.SetPrompt("")
}

func (c *Console) AddToPrompt(char rune) {
	c.prompt = c.prompt + string(char)
}

func (c *Console) DeleteFromPrompt() {
	if len(c.prompt) > 0 {
		c.prompt = c.prompt[:len(c.prompt)-1]
	}
}

func (c *Console) full() bool {
	return c.length == c.capacity
}
