package goteller

import (
	"../ipaddr"
	"../messages"
	"bufio"
	"fmt"
	"net"
	"strings"
)

const HEADER_LEN int = 23
const CONNECTOR string = "GNUTELLA CONNECT/0.4\n\n"
const REPLY string = "GNUTELLA OK\n\n"

func (teller *GoTeller) startServant() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", teller.addr.Port))
	if err != nil { // If error, just panic. Node will not work if Listen fails
		panic(err.Error())
	}

	go teller.startPinger() // Will periodically send pings
	go func() {             // Wait for incoming connections
		for teller.alive {
			conn, err := listener.Accept()
			if err != nil && teller.debugFile != nil {
				fmt.Fprintln(teller.debugFile, err) // no worries. just print the error
			} else {
				go teller.handleConnection(conn)
			}
		}
	}()
}

func (teller *GoTeller) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		if r := recover(); r != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(teller.debugFile, "Recovered a Panic in handleConnection: ", r)
			}
		}
	}()

	connIO := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)) // ReaderWriter has a bufio.Reader which allows peeking of io.Reader
	peeked, err := connIO.Reader.Peek(3)                                        // Peek first 3 chars to see if it starts with GET .... indicates a GET request
	if l := len(peeked); l != 3 {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, "Couldn't peek 3 bytes properly...error:", err)
		}
		return
	} else if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
		return
	}
	// Peek worked fine
	if strings.HasPrefix(string(peeked), "GET") {
		// Its a http request! Send connIO to request handler
		teller.handleRequest(connIO)
		return
	}

	connected, err := gnutellaReplyToConnect(connIO)
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
		return
	} else if !connected {
		return // Probably failed to get correct CONNECTOR string
	}

	headerBuffer := make([]byte, HEADER_LEN)
	n, err := connIO.Reader.Read(headerBuffer)
	if n != HEADER_LEN {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, "Wasn't able to read HEADER_LEN bytes")
		}
	} else if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
	} else {
		header, err := messages.ParseHeaderBytes(headerBuffer)
		if err != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(teller.debugFile, err)
			}
			return
		}

		from, err := ipaddr.ParseAddrString(conn.LocalAddr().String()) // May need to switch to conn.LocalAddr()
		if err != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(teller.debugFile, err)
			}
			return
		}

		payloadBuffer := make([]byte, header.PayloadLen)
		if header.PayloadLen > 0 {
			n, err := connIO.Reader.Read(payloadBuffer)
			if uint32(n) != header.PayloadLen {
				if teller.debugFile != nil {
					fmt.Fprintln(teller.debugFile, "Couldn't read payloadLen bytes for Pong")
				}
				return
			} else if err != nil {
				if teller.debugFile != nil {
					fmt.Fprintln(teller.debugFile, err)
				}
				return
			}
		}
		switch header.PayloadDesc {
		case messages.PING:
			{
				teller.onPing(*header, *from)
			}
		case messages.PONG:
			{
				pong, err := messages.ParsePongBytes(payloadBuffer)
				if err != nil {
					if teller.debugFile != nil {
						fmt.Fprintln(teller.debugFile, err)
					}
				} else {
					teller.onPong(*header, *pong)
				}
			}
		case messages.PUSH:
			{
				/*push, err := messages.ParsePushBytes(payloadBuffer)
				if err != nil {
					if teller.debugFile != nil {
						fmt.Fprintln(teller.debugFile, err)
					}
				} else {
					// TODO: Create push handler
				}*/
			}
		case messages.QUERY:
			{
				query, err := messages.ParseQueryBytes(payloadBuffer)
				if err != nil {
					if teller.debugFile != nil {
						fmt.Fprintln(teller.debugFile, err)
					}
				} else {
					teller.onQuery(*header, *query, *from)
				}
			}
		case messages.QUERYHIT:
			{
				queryhit, err := messages.ParseQueryHitBytes(payloadBuffer)
				if err != nil {
					if teller.debugFile != nil {
						fmt.Fprintln(teller.debugFile, err)
					}
				} else {
					teller.onQueryHit(*header, *queryhit)
				}
			}
		}
	}
}

func gnutellaReplyToConnect(connIO *bufio.ReadWriter) (bool, error) {
	connectBuffer := make([]byte, len(CONNECTOR))
	_, err := connIO.Reader.Read(connectBuffer)
	if err != nil {
		return false, err
	}

	connectStr := messages.ReadStringLE(connectBuffer)

	if connectStr != CONNECTOR {
		return false, nil
	}

	replyBuffer := make([]byte, len(REPLY))
	messages.WriteStringLE(replyBuffer, REPLY)

	err = sendBytes(connIO, replyBuffer) // from requesthandler.go file
	if err != nil {
		return false, nil
	}
	return true, nil
}

func gnutellaConnect(connIO *bufio.ReadWriter) (bool, error) {
	connectBuffer := make([]byte, len(CONNECTOR))
	messages.WriteStringLE(connectBuffer, CONNECTOR)
	err := sendBytes(connIO, connectBuffer)
	if err != nil {
		return false, err
	}

	replyBuffer := make([]byte, len(REPLY))
	_, err = connIO.Reader.Read(replyBuffer)
	if err != nil {
		return false, err
	}

	replyStr := messages.ReadStringLE(replyBuffer)
	if replyStr != REPLY {
		return false, nil
	}

	return true, nil
}
