# GoTella

An implementation of the [Gnutella Protocol v0.4](http://courses.cs.washington.edu/courses/cse522/05au/gnutella_protocol_0.4.pdf) in [Go](http://golang.org/).

## Motivation
No practical reasons to make this. Gnutella isn't really used much anymore, much less this version of it, I know. Just wanted to have experience with creating a P2P system from scratch. Decided to use Go for the implementation so that I could gain more experience with using the language and also for its adeptness when working with Networking related projects.

## Design
This module is designed to be an API for the Gnutella protocol whereby you would only need to provide various callback functions for specific parts of the Gnutella protocol. The rest of the details such as the networking, periodic pinging, message serialization/deserialization, and general message bookkeeping are handled internally.

## Usage
*An example usage for this package is available in tests/simple_app.go*
To use this module, you must import the `goteller` & `messages` packages as well as the `net/http` package that comes with the Go standard library. All of the Gnutella communication is handled with an instance of a `goteller.Goteller` struct. The first step is to properly initialize this instance.

### Initialization
First steps after properly importing the required packages (listed above), are to instantiate a `goteller.Goteller` and set its various properties.

    teller := goteller.GoTeller{}
    teller.NetworkSpeed = 10 // Network speed for current process (Optional)
    teller.SetServantID("myusername123456") // ID for servant. No more than 16 characters. Will be truncated to 16 characters if exceeds. (Required)
    err := teller.SetInitNeighbors([]string{"localhost:4000", "10.11.12.13:4000"}) // Array of strings containing IP:Port addresses of other known servants. Can use "localhost:<Port>" if you know other servant is on the same machine. (Required)
    if err != nil {
	    … Handle Error // Usually because of bad formatting of the "IP:Port" strings
    }
    teller.SetDebugFile(os.Stdout) // Give an io.Writer to which internal error messages can be written to. Use os.Stdout to just write to stdout (Optional)
    teller.OnQuery(OnQueryCallback) // A callback function for any incoming queries to this servant (Required)
    teller.OnRequest(OnRequestCallback) // A callback function for any incoming HTTP Requests for resources at this node. Requests will be for resources returned as query hits on OnQuery callback. (Required) 

### Callbacks

More details on the `OnQuery` and `OnRequest` callback funcitons. Keep in mind that these callback functions are run on their own separate goroutines and can be called multiple times. Be careful about mutual exclusion and whatnot. IO done within these callback funcitons will be non-blocking by virtue of being on their own goroutines.

#### OnQuery Callback
Called whenever another servant sends out a query message which reaches your servant.
The onQuery callback funciton must have the following parameters and return type:

`func OnQueryCallback(query string) []messages.HitResult {…}`

The query paramater is simple the search term in string form. The callback function must return a list of query hits in the form of the `messages.HitResult` struct. This struct is of the following form:

    type HitResult struct {
	    FileIndex uint32 // Index of file. You can give it your own significance (Index in an array?). Otherwise can just leave as 0.
	    FileSize  uint32 // Size of data represented by this query hit
	    Filename  string // Name of data represented by this query hit
    }

#### OnRequest Callback
Called when, after you respond to a servant's Query message with a query hit(s), that servant sends a request to your servant for the file represented by that query hit.
The OnRequest callback function must have the following parameters and return type:

`func OnRequestCallback(fileIndex uint32, filename string) (io.ReadCloser, int64)`

fileIndex and filename correspond to the values from the HitResult struct. Returns an [io.ReadCloser](http://golang.org/pkg/io/#ReadCloser) and the length of the file in bytes. The io.ReadCloser is closed after the response is sent. If the length return is less than 0, a `404 Not Found` response will be sent.

### Starting the Servant
After initialization of the `goteller.GoTeller`, you can start servant with the following snippet:

    err := teller.StartAtPort(3000) // Starts the servants networking on multiple goroutines. Will not start for multiple reasons (Improper initialization, bad port number).
    if err != nil {
	    … Handle error // err != nil is a direct indication that the Servant cannot start
    }

### Sending a Query
Queries can be sent over the Gnutella network by constructing a `goteller.Query` struct and sending it with `teller.SendQuery(query)`. The query struct has the following form:

    type Query struct {
	    TTL         byte // TTL for the query message as it is flooded out over the Gnutella network
	    MinSpeed    uint16 // Minimum speed of a responding servant 
	    SearchQuery string // Search term for query
	    // Other private fields
    }

Each query object also has two callback functions that must be set before the query can be sent. One is the `OnHit` callback, and the other is the `OnResponse` callback.

#### OnHit callback
Called whenever a query hit is received for the corresponding query. It must have the following parameters and return type:

`func OnHitCallback(queryHits []goteller.QueryResult, servantSpeed uint32, servantID string) []goteller.QueryResult`

The `goteller.QueryResult` struct has the following form (All fields are hidden but have getter methods):

    type QueryResult struct {
	    fileIndex uint32 // Accessed with .GetFileIndex()
	    fileSize  uint32 // Accessed with .GetFileSize()
	    filename  string // Accessed with .GetFilename()
	    // Other hidden fields
    }

The function must return a slice of `QueryResult` structs. Each `QueryResult` in the returned slice must be from the slice `QueryResult`s given as a parameter to the callback function. In other words, the return value must be a subset of the `queryHits` input value. Requests are then sent for each of the `QueryResult` structs returned by the callback function.

#### OnResponse callback
Called when a response is received for a request sent for resource specified from return value of `OnHit` callback function. It must have the following parameters:

`func OnResponseCallback(err error, fileIndex uint32, filename string, res *http.Response)`

If the error parameter is not `nil`, the res pointer will be. The same is true vice-versa.

#### Example

    // Instantiate and start teller goteller.GoTeller at a port
    // First, construct a query:
    query := goteller.Query{
	    TTL: 1, // Will only go to immediate neighbors in Gnutella overlay. Must be greater than 0.
	    MinSpeed: teller.NetworkSpeed, // Only servants with as good network speed as me should respond
	    SearchQuery: "myfile.txt", // Search term of this query
    }
    query.OnHit(OnHitCallback) // Required
    query.OnResponse(OnResponseCallback) // Required
    err := teller.SendQuery(query)
    if err != nil {
	    … Handle Error // Occurrs upon improper initialization of query
    }

    // OnHit callback function just requests first QueryResult in input list of results
    func OnHitCallback(queryHits []goteller.QueryResult, servantSpeed uint32, servantID string) []goteller.QueryResult {
	    return queryHits[:1] // Just want the first one
    }
    // OnResponse callback function that prints response body to output
    func OnResponseCallback(err error, fileIndex uint32, filename string, res *http.Response) {
	    if err != nil {
		    fmt.Println(err)
		    return
	    }
	    if res.Status != 200 {
		    // Probably got a 404 response!
		    fmt.Printf("Got response status code %d for file \"%s\"\n", res.Status, filename)
		    return
	    }
	    body, err := ioutil.ReadAll(res.Body)
	    if err != nil {
		    fmt.Println(err)
		    return
	    }
	    fmt.Println("Got response for file \"%s\": %s\n", filename, string(body))
    }

### Limitations
As of now, the Push messages used in the Gnutella protocol haven't been implemented. No Push messages are sent, and all received push messages will be dropped with this implementation. I may add it later but it the rest basic ping/pong, query/queryhit, and HTTP requesting parts of the protocol work correctly.