package main

import (
	"bufio"
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
	cmdMode           = "MODE"
	cmdMOTDStart      = "375"
	cmdMOTDItem       = "372"
	cmdMOTDEnd        = "376"
	cmdReplyListStart = "321"
	cmdReplyListItem  = "322"
	cmdReplyListEnd   = "323"
	cmdNowAway        = "306"
	cmdUnaway         = "305"
)

const (
	modeAway          = 'a'
	modeInvisible     = 'i'
	modeWallops       = 'w'
	modeRestricted    = 'r'
	modeOperator      = 'o'
	modeLocalOperator = 'O'
	modeNotifyee      = 's'
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

		for {
			raw, err := reader.ReadString('\n')
			check(err)
			raw = strings.TrimSuffix(raw, "\n")

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
							fmt.Printf("MOTD from %s\n", args[i+1])
							fmt.Println("----------")
						} else {
							break
						}
					}
				}
			case cmdMOTDItem:
				parts := strings.Split(raw, ":- ")
				if len(parts) > 1 {
					fmt.Println(parts[1])
				}
			case cmdMOTDEnd:
				fmt.Println("----------")
			case cmdMode:
				if len(args) > 2 || args[0] != username {
					break
				}

				modes := args[1:]
				if len(modes[0]) > 1 {
					modes[0] = modes[0][1:]
				}

				for _, mode := range modes {
					if len(mode) < 2 {
						break
					}

					var enabled bool
					if mode[0] == '+' {
						enabled = true
					} else if mode[0] == '-' {
						enabled = false
					} else {
						break
					}

					switch mode[1] {
					case modeInvisible:
						if enabled {
							fmt.Println("you are invisible")
						} else {
							fmt.Println("you are visible")
						}
					case modeAway:
						if enabled {
							fmt.Println("you are away")
						} else {
							fmt.Println("you are active")
						}
					}
				}
			case cmdNowAway:
				fmt.Println("you are away")
			case cmdUnaway:
				fmt.Println("you are active")
			default:
				fmt.Println(raw)
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
