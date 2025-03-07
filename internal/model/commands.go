package model

type Command int

const (
	CommandUNK Command = iota // Unknown command
	CommandGET                // GET key
	CommandSET                // SET key value
	CommandDEL                // DEL key
)

const (
	CommandGETArgsLen = 1
	CommandSETArgsLen = 2
	CommandDELArgsLen = 1
)

type Query struct {
	Command Command
	Args    []string
}
