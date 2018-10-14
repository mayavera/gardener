package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
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
	cmdWhoIsUser      = "311"
	cmdWhoIsServer    = "312"
	cmdWhoIsIdle      = "317"
	cmdEndOfWhoIs     = "318"
	cmdListStart      = "321"
	cmdListItem       = "322"
	cmdListEnd        = "323"
)

const (
	errUnknownCommand = "421"
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

type WhoisResponse struct {
	nick         string
	user         string
	name         string
	host         string
	server       string
	serverInfo   string
	idleDuration time.Duration
	joinedAt     time.Time
}

func main() {
	fmt.Println("connecting to", host)
	conn, err := net.Dial("tcp", host)
	check(err)
	defer conn.Close()
	fmt.Println("connected to", host)

	go func() {
		conn.SetReadDeadline(time.Time{}) // Disable read timeout

		reader := bufio.NewReader(conn)

		whoisResponses := map[string]WhoisResponse{}

		for {
			raw, err := reader.ReadString('\n')
			check(err)
			raw = strings.TrimSuffix(raw, "\n")

			args := strings.Split(raw, " ")

			trailerParts := strings.Split(raw, ":")
			var trailer string
			if len(trailerParts) > 2 {
				trailer = trailerParts[2]
			} else if len(trailerParts) == 2 && !strings.HasPrefix(raw, ":") {
				trailer = trailerParts[1]
			}

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
			case cmdWhoIsUser:
				fmt.Println(raw)
				res := WhoisResponse{
					nick: args[1],
					user: args[2],
					host: args[3],
					name: trailer,
				}

				whoisResponses[res.nick] = res
			case cmdWhoIsServer:
				if len(args) < 3 {
					break
				}

				nick := args[1]
				res, found := whoisResponses[nick]
				if found {
					res.server = args[2]
					res.serverInfo = trailer
					whoisResponses[nick] = res
				}
			case cmdWhoIsIdle:
				nick := args[1]
				res, found := whoisResponses[nick]
				if found {
					numIdleSeconds, err := strconv.Atoi(args[2])
					if err != nil {
						break
					}

					res.idleDuration = time.Duration(numIdleSeconds) * time.Second

					joinedAtUnixTimestamp, err := strconv.ParseInt(args[3], 10, 64)
					if err != nil {
						break
					}

					res.joinedAt = time.Unix(joinedAtUnixTimestamp, 0)

					whoisResponses[nick] = res
				}
			case cmdEndOfWhoIs:
				nick := args[1]
				res, found := whoisResponses[nick]
				if found {
					delete(whoisResponses, nick)
					fmt.Printf("WHOIS for %s:\n", nick)
					fmt.Printf("  user: %s\n", res.user)
					fmt.Printf("  name: %s\n", res.name)
					fmt.Printf("  host: %s\n", res.host)
					fmt.Printf("  server: %s\n", res.server)
					fmt.Printf("  server info: %s\n", res.serverInfo)
					fmt.Printf("  idle: %s\n", res.idleDuration)
					fmt.Printf("  joined: %s\n", res.joinedAt)
				} else {
					fmt.Println(raw)
				}
			case cmdListStart:
				fmt.Println("channels:")
				fmt.Println("----------")
			case cmdListItem:
				var channel string
				if len(args) > 1 {
					channel = args[1]
				} else {
					break
				}

				var userCount int
				if len(args) > 2 {
					userCount, _ = strconv.Atoi(args[2])
				}

				info := trailer

				fmt.Printf("%s (%d)\n", channel, userCount)
				fmt.Printf("  %s\n", info)
			case cmdListEnd:
				fmt.Println("----------")
			case errUnknownCommand:
				if len(args) >= 2 {
					fmt.Printf("unknown command: %s\n", args[1])
				}
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
