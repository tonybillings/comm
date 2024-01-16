/*******************************************************************************
DESCRIPTION

  Simple messaging app based on the Node API.

INSTALLATION

  1. Run the following:
     go clean
     go mod tidy
     go install

  2. Ensure $GOPATH/bin is in your PATH.

USAGE

  gochat <your_name> <port>

  your_name - This is your display name that others will see.
  port - The port number your node will bind to for incoming messages.

  Example: gochat John 8111

  This will launch th interactive prompt and bind to the socket 0.0.0.0:8111.
  At this point, other users of gochat can send you messages via that address.

INTERACTIVE PROMPT

  To send a message, enter the destination address or alias for the recipient,
  followed by a comma, then your message.  Padding around the comma is optional.

  Format: [<dest_addr>|<alias>], <message>

  The destination address is in the format <host>:<port> (<host> and colon are
  optional).

  Equivalent examples:
    8222, Hello Jane!
    :8222, Hello Jane!
    127.0.0.1:8222, Hello Jane!

  Sending using an alias:
    jane, Hello Jane!

ALIASES

  To create an alias, enter the alias name, followed by the assignment operator,
  and then the destination address for the alias (<host> and colon are optional).

  Format: <alias>=<dest_addr>

  Equivalent examples:
    jane=8222
    jane=:8222
    jane=127.0.0.1:8222

  Example alias using a remote address:
    jill=10.0.0.107:8333

  Note that whenever you receive a message from someone, an alias is automatically
  added for that person.  If they specified their name as being "John" when they
  started their instance of gochat, then the alias created for you will have that
  same name.  Also note aliases are case-sensitive!

*******************************************************************************/

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"tonysoft.com/comm/pkg/comerr"
	"tonysoft.com/comm/pkg/node"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func handleSendErr(err error, input string) {
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, net.ErrClosed) {
			fmt.Println("recipient is unavailable")
		} else if errors.Is(err, comerr.ErrAddressFormatUnknown) {
			fmt.Println("unknown recipient (add via <alias>=<address>)")
		} else {
			fmt.Println("failed to send message")
		}
	} else {
		fmt.Printf("\033[38;5;245m\033[1F\033[2K\r%s\n\033[m", input)
	}
}

func main() {
	if len(os.Args) != 3 {
		panic("invalid number of args (expected 2)\n\nUSAGE: gochat <your_name> <port>")
	}

	name := os.Args[1]
	port := os.Args[2]

	fmt.Println("Starting GoChat, enter CTRL+C to quit...")

	cfg := node.NewConfig(":" + port)
	n, err := node.New[string](cfg)
	checkErr(err)

	err = n.Start()
	checkErr(err)
	defer n.Stop()

	aliases := make(map[string]string)
	ctx, cancel := context.WithCancel(context.Background())
	var targetRecipient atomic.Value

	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	go func() {
		<-intChan
		fmt.Println("Interrupt received, shutting down...")
		signal.Stop(intChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Print("> ")
				if target := targetRecipient.Load(); target != nil && target.(string) != "" {
					fmt.Printf("%s > ", target.(string))
				}

				reader := bufio.NewReader(os.Stdin)
				input, e := reader.ReadString('\n')
				checkErr(e)

				input = strings.TrimSpace(input)

				if strings.HasPrefix(input, "<<") {
					targetRecipient.Store("")
				} else if strings.HasPrefix(input, ">>") {
					targetRecipient.Store(strings.TrimSpace(strings.Split(input, ">>")[1]))
				} else {
					if target := targetRecipient.Load(); target != nil && target.(string) != "" {
						message := fmt.Sprintf("%s: %s", name, input)
						recipient := target.(string)
						if alias, ok := aliases[recipient]; ok {
							recipient = alias
						}
						_, se := n.Send(recipient, &message)
						handleSendErr(se, input)
					} else {
						if strings.Contains(input, ",") {
							inputArr := strings.Split(input, ",")
							recipient := strings.TrimSpace(inputArr[0])
							if alias, ok := aliases[recipient]; ok {
								recipient = alias
							}

							p, pe := strconv.ParseInt(recipient, 10, 16)
							if pe == nil {
								recipient = ":" + strconv.Itoa(int(p))
							}

							text := strings.TrimSpace(strings.Join(inputArr[1:], ","))
							message := fmt.Sprintf("%s: %s", name, text)
							_, se := n.Send(recipient, &message)
							handleSendErr(se, input)
						} else {
							inputArr := strings.Split(input, "=")
							if len(inputArr) == 2 {
								alias := strings.TrimSpace(inputArr[0])
								address := strings.TrimSpace(inputArr[1])
								p, pe := strconv.ParseInt(address, 10, 16)
								if pe == nil {
									address = ":" + strconv.Itoa(int(p))
								}
								aliases[alias] = address
							}
						}
					}
				}
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-n.Recv():
				if !ok {
					return
				}
				text := *msg.Data
				textArr := strings.Split(text, ":")
				aliases[textArr[0]] = msg.FromNode()
				fmt.Printf("\r%s\n> ", text)

				if target := targetRecipient.Load(); target != nil && target.(string) != "" {
					fmt.Printf("%s > ", target.(string))
				}
			}
		}
	}()
	wg.Wait()
}
