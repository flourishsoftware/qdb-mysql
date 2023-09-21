package utils

type BashCmd struct {
	BashCommand string
	BashOptions []string
	Command     string
	Options     []string
}

func NewBashCmd(programName string) BashCmd {
	return BashCmd{
		BashCommand: "bash",
		BashOptions: []string{"-c"},
		Command:     programName,
	}
}

func (mc BashCmd) Append(cmdOptions ...string) BashCmd {
	mc.Options = append(mc.Options, cmdOptions...)
	return mc
}

func (mc BashCmd) AsBashCmd(cmdOptions ...string) (string, []string) {
	result1, result2 := mc.BashCommand, mc.BashOptions
	result2 = append(result2, mc.Command)
	result2 = append(result2, mc.Options...)
	return result1, result2
}

func (mc BashCmd) AsCmd(cmdOptions ...string) (string, []string) {
	result1, result2 := mc.Command, mc.Options
	return result1, result2
}
