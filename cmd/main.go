package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const host = "halcyon.dal.net:6667"
const username = "mayav"
const realname = "Maya V."
const daemonname = "tome"

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const (
	cmdPing           = "PING"
	cmdMOTDStart      = "375"
	cmdMOTDItem       = "372"
	cmdMOTDEnd        = "376"
	cmdReplyListStart = "321"
	cmdReplyListItem  = "322"
	cmdReplyListEnd   = "323"
)

func main() {
	fmt.Println("connecting to", host)
	conn, err := net.Dial("tcp", host)
	check(err)
	defer conn.Close()
	fmt.Println("connected to", host)

	go func() {
		conn.SetReadDeadline(time.Time{}) // Disable read timeout

		reader := bufio.NewReader(conn)

		var motdFrom string
		var motdBuffer bytes.Buffer

		for {
			raw, err := reader.ReadString('\n')
			check(err)

			args := strings.Split(raw, " ")

			//var prefix string
			var command string

			if len(args) > 0 {
				if strings.HasPrefix(args[0], ":") {
					//prefix = args[0][:len(args[0])]
				}

				args = args[1:len(args)]
			}

			if len(args) > 0 {
				command = args[0]

				args = args[1:len(args)]
			}

			switch command {
			case cmdPing:
				fmt.Fprintf(conn, "PONG %s\n", daemonname)
			case cmdMOTDStart:
				for i, arg := range args {
					if arg == ":-" {
						if len(args) >= i+1 {
							motdFrom = args[i+1]
						} else {
							break
						}
					}
				}
			case cmdMOTDItem:
				parts := strings.Split(raw, ":- ")
				if len(parts) > 1 {
					motdBuffer.WriteString(parts[1])
				}
			case cmdMOTDEnd:
				fmt.Printf("MOTD from %s:\n", motdFrom)
				motdFrom = ""

				fmt.Printf(motdBuffer.String())
				motdBuffer.Reset()
			default:
				fmt.Print(raw)
			}
		}
	}()

	fmt.Printf("logging in as %s(%s)\n", realname, username)

	_, err = fmt.Fprintf(conn, "USER %s * *: %s\n", username, realname)
	check(err)

	_, err = fmt.Fprintf(conn, "NICK %s\n", username)
	check(err)

	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		check(err)

		_, err = io.WriteString(conn, input)
		check(err)
	}
}
