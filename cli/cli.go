package cli

type CliCommandType = uint8

const (
	CliCommandGetBalance CliCommandType = iota
	CliCommandSend
)

type CliCommand struct {
	Type CliCommandType
}
