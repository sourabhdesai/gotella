package goteller

import (
	"../ipaddr"
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
)

func (teller *GoTeller) sendRequest(fileIndex uint32, filename string, to ipaddr.IPAddr) {
	endpoint := to.String()
	path := fmt.Sprintf("/get/%d/%s", fileIndex, filename)
	req, err := http.NewRequest("GET", "http://"+endpoint+path, nil)
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
			fmt.Fprintln(teller.debugFile, err)
		}
		return
	}

	path := req.URL.Path[1:] // drop the leading '/'
	var fileIdx uint32
	var filename string
	n, err := fmt.Sscanf(path, "get/%d/%s", &fileIdx, &filename)
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
	} else if n != 2 {
		if teller.debugFile != nil {
			fmt.Fprintf(teller.debugFile, "Scanned %d out of 2 values in path \"%s\"", n, path)
		}
	}
	if err != nil || n != 2 {
		res := buildNotFoundResponse(req)
		res.Write(connIO.Writer)
		err = connIO.Writer.Flush()
		if err != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(teller.debugFile, err)
			}
		}
	} else {
		// Valid request
		bodyReader, length := teller.requestFunc(fileIdx, filename)
		res := buildResponse("200 OK", 200, bodyReader, length, req)
		res.Write(connIO.Writer)
		err = connIO.Writer.Flush()
		if err != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(teller.debugFile, err)
			}
		}
	}
}

func buildNotFoundResponse(req *http.Request) http.Response {
	return buildResponse("404 Not Found", 404, nil, 0, req)
}

func buildResponse(status string, statuscode int, body io.ReadCloser, bodyLen int64, req *http.Request) http.Response {
	res := http.Response{
		Status:        status,
		StatusCode:    statuscode,
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		ContentLength: bodyLen,
		Close:         true,
		Request:       req,
	}
	if bodyLen > int64(0) {
		res.Body = body
	}
	return res
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
