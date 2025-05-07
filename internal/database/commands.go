package database

type Command int

const (
	CommandUNK Command = iota // Unknown command
	CommandGET                // GET key
	CommandSET                // SET key value
	CommandDEL                // DEL key
)

type Query struct {
	Command Command
	Args    []string
}
