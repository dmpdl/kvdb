package cli

import (
	"context"
	"fmt"
	"strings"
)

type executor interface {
	RunCommand(ctx context.Context, rawQuery string) string
}

type reader interface {
	ReadString(sep byte) (string, error)
}

func Run(ctx context.Context, input reader, exec executor) {
	fmt.Println("Key Value DB Demo")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("exit")
			return
		default:
		}

		fmt.Print("> ")

		rawQuery, err := input.ReadString('\n')
		if err != nil {
			fmt.Println("read input error:", err)
			continue
		}

		rawQuery = strings.TrimSpace(rawQuery)

		if strings.ToLower(rawQuery) == "exit" {
			fmt.Println("Выход из программы.")
			break
		}

		output := exec.RunCommand(ctx, rawQuery)

		fmt.Println(">", output)
	}
}
