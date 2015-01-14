package main

import (
	"../goteller"
	"../messages"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		fmt.Println("Need 3 Arguments: <Port> <InitAddress> <ServantID>")
		return
	}
	port, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	initAddr := args[1]
	servantID := args[2]

	teller := goteller.GoTeller{}
	teller.SetServantID(servantID)
	err = teller.SetInitNeighbors([]string{initAddr})
	if err != nil {
		fmt.Println(err)
		return
	}
	teller.SetDebugFile(os.Stderr)
	teller.OnQuery(OnQuery)
	teller.OnHit(OnHit)
	teller.OnData(OnData)
	teller.OnRequest(OnRequest)

	err = teller.StartAtPort(uint16(port))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Started Servant at port", teller.Port)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Type Query:")
		query, _ := reader.ReadString('\n')
		teller.SendQuery(query[:len(query)-1], 1, 0) // ttl = 2; minspeed = 0 KB/s
	}
}

func OnQuery(query string) []messages.HitResult {
	fmt.Printf("Received Query: \"%s\"\n", query)
	return []messages.HitResult{
		messages.HitResult{
			FileIndex: 0,
			FileSize:  uint32(len(getFile())),
			Filename:  "file.txt",
		},
	}
}

func OnHit(hits []goteller.QueryResult, fileidx uint32, servantID string) []goteller.QueryResult {
	fmt.Println("Received Query Hits")
	for i, result := range hits {
		fmt.Printf("%d: %+v\n", i, result)
	}
	return hits[0:1]
}

func OnData(err error, fileindex uint32, filename string, res *http.Response) {
	fmt.Println("Received Response from Data Request")
	fmt.Printf("fileindex: %d; filename: %s\n", fileindex, filename)
	if err != nil {
		fmt.Println("Received Error in OnData:", err)
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Response Body:\n\t" + string(data))
}

func OnRequest(fileIndex uint32, filename string) (io.ReadCloser, int64) {
	fmt.Printf("Called On Request: fileIdx: %d, filename: %s\n", fileIndex, filename)
	readCloser := ioutil.NopCloser(bytes.NewReader(getFile()))
	return readCloser, int64(len(getFile()))
}

func getFile() []byte {
	return []byte("Hi Im Bob")
}
