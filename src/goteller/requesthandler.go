package goteller

import (
	"../ipaddr"
	"bufio"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
)

func (teller *GoTeller) sendRequest(fileIndex uint32, filename string, to ipaddr.IPAddr) {
	endpoint := to.String()
	path := fmt.Sprintf("/get/%d/%s", fileIndex, filename)
	req, err := http.NewRequest("GET", endpoint+path, nil)
	if err != nil {
		teller.dataFunc(err, fileIndex, filename, nil)
		return
	}
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		teller.dataFunc(err, fileIndex, filename, nil)
		return
	}
	err = req.Write(conn)
	if err != nil {
		teller.dataFunc(err, fileIndex, filename, nil)
		return
	}

	connReader := bufio.NewReader(conn)
	res, err := http.ReadResponse(connReader, req)
	if err != nil {
		teller.dataFunc(err, fileIndex, filename, nil)
		return
	}

	teller.dataFunc(nil, fileIndex, filename, res)

}

func (teller *GoTeller) handleRequest(connIO *bufio.ReadWriter) {
	req, err := http.ReadRequest(connIO.Reader)
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(*teller.debugFile, err)
		}
		return
	}

	path := req.URL.Path[1:] // drop the leading '/'
	match, err := regexp.MatchString("get/[0-9]+/..*", path)
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(*teller.debugFile, err)
		}
	}
	if !match {
		// Respond with 404
		err := respondNotFound(connIO)
		if err != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(*teller.debugFile, err)
			}
		}
	} else {
		var fileIdx uint32
		var filename string
		n, err := fmt.Sscanf(path, "get/%d/%s", &fileIdx, &filename)
		if err != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(*teller.debugFile, err)
			}
		} else if n != 2 {
			if teller.debugFile != nil {
				fmt.Fprintf(*teller.debugFile, "Scanned %d out of 2 values in path \"%s\"", n, path)
			}
		} else {
			// Valid request
			bodyBuffer := teller.requestFunc(fileIdx, filename)
			httpHeader := "HTTP 200 OK\r\nServer: Gnutella\r\nContent-type: application/binary\r\nContent-length: " + strconv.Itoa(len(bodyBuffer)) + "\r\n\r\n"
			response := append([]byte(httpHeader), bodyBuffer...)
			err := sendBytes(connIO, response)
			if err != nil {
				if teller.debugFile != nil {
					fmt.Fprintln(*teller.debugFile, err)
				}
			}
		}
	}
}

func respondNotFound(connIO *bufio.ReadWriter) error {
	responseString := "HTTP/1.0 404 Not Found"
	err := sendBytes(connIO, []byte(responseString))
	return err
}

func sendBytes(connIO *bufio.ReadWriter, buffer []byte) error {
	_, err := connIO.Writer.Write(buffer)
	if err != nil {
		return err
	} else {
		err := connIO.Writer.Flush() // Flush the buffer
		return err                   // might be nil
	}
}
