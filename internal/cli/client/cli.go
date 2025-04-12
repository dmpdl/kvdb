package cli

import (
	"context"
	"fmt"
	"strings"
)

type Client interface {
	Send(ctx context.Context, request []byte) ([]byte, error)
	Close() error
}

type Reader interface {
	ReadString(sep byte) (string, error)
}

func Run(ctx context.Context, input Reader, client Client) {
	defer client.Close()
	fmt.Println("Key Value DB Client")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("exit")
			return
		default:
		}

		fmt.Print("> ")

		request, err := input.ReadString('\n')
		if err != nil {
			fmt.Println("read input error:", err)
			continue
		}

		request = strings.TrimSpace(request)

		if strings.ToLower(request) == "exit" {
			fmt.Println("bye")
			break
		}

		output, err := client.Send(ctx, []byte(request))
		if err != nil {
			fmt.Println("failed send")
		}

		fmt.Println(">", string(output))
	}
}
