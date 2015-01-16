package goteller

import (
	"../ipaddr"
	"../messages"
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

const SEED int64 = 7187 // A large prime number
const DEFAULT_PING_INTERVAL time.Duration = 3 * time.Second

type HitResult messages.HitResult

type GoTeller struct {
	alive           bool
	debugFile       io.Writer
	addr            ipaddr.IPAddr
	Neighbors       []ipaddr.IPAddr
	NumShared       uint32
	NumKB           uint32
	Port            uint16
	NetworkSpeed    uint32
	PingInterval    time.Duration
	hashCount       uint32
	servantID       string
	randGen         *rand.Rand
	savedPings      map[[16]byte]ipaddr.IPAddr
	savedQueries    map[[16]byte]ipaddr.IPAddr
	myQueries       map[[16]byte]Query
	neighborsMutex  sync.RWMutex
	pingMapMutex    sync.RWMutex
	queryMapMutex   sync.RWMutex
	myQueryMapMutex sync.RWMutex
	queryFunc       func(string) []messages.HitResult
	requestFunc     func(uint32, string) (io.ReadCloser, int64)
}

func (teller *GoTeller) StartAtPort(port uint16) error {
	teller.alive = true
	teller.randGen = rand.New(rand.NewSource(SEED))
	if teller.servantID == "" {
		teller.alive = false
		return fmt.Errorf("Must set Servant ID (use SetServantID)")
	}
	if teller.queryFunc == nil {
		teller.alive = false
		return fmt.Errorf("Must set Query callback function (use OnQuery)")
	}
	if teller.requestFunc == nil {
		teller.alive = false
		return fmt.Errorf("Must set Request callback function (use OnRequest)")
	}
	if len(teller.Neighbors) == 0 {
		teller.alive = false
		return fmt.Errorf("Must set initial neighbors (use SetInitNeighbors)")
	}
	teller.Port = port
	teller.addr.Port = port
	err := teller.addr.SetToLocalIP()
	if err != nil {
		return err
	}
	if teller.PingInterval == 0 {
		teller.PingInterval = DEFAULT_PING_INTERVAL
	}
	teller.savedPings = make(map[[16]byte]ipaddr.IPAddr)
	teller.savedQueries = make(map[[16]byte]ipaddr.IPAddr)
	teller.myQueries = make(map[[16]byte]Query)
	err = teller.startServant()
	if err != nil {
		return err
	}
	return nil
}

func (teller *GoTeller) Stop() {
	teller.alive = false
}

func (teller *GoTeller) IsRunning() bool {
	return teller.alive
}

func (teller *GoTeller) SetInitNeighbors(addrs []string) error {
	for _, address := range addrs {
		addr, err := ipaddr.ParseAddrString(address)
		if err != nil {
			return err
		}
		teller.Neighbors = append(teller.Neighbors, *addr)
	}
	return nil
}

func (teller *GoTeller) SetDebugFile(file io.Writer) {
	teller.debugFile = file
}

func (teller *GoTeller) SetServantID(id string) {
	if l := len(id); l <= 16 {
		teller.servantID = id
	} else {
		teller.servantID = id[:16]
	}
}

func (teller *GoTeller) OnQuery(qFunc func(string) []messages.HitResult) {
	teller.queryFunc = qFunc
}

func (teller *GoTeller) OnRequest(reqFunc func(uint32, string) (io.ReadCloser, int64)) {
	teller.requestFunc = reqFunc
}

// send msg to all neighbors except for from
func (teller *GoTeller) floodToNeighbors(msg []byte, from ipaddr.IPAddr) {
	teller.neighborsMutex.RLock()
	defer teller.neighborsMutex.RUnlock()
	for _, addr := range teller.Neighbors {
		if from != addr {
			teller.sendToNeighbor(msg, addr)
		}
	}
}

func (teller *GoTeller) sendToNeighbor(msg []byte, to ipaddr.IPAddr) bool {
	conn, err := net.Dial("tcp", to.String())
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
		return false
	}

	defer conn.Close()

	connIO := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	connected, err := gnutellaConnect(connIO)
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
		return false
	}
	if !connected {
		err = fmt.Errorf("Didn't receive a valid connect reply")
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
		return false
	}

	err = sendBytes(connIO, msg) // in requesthandler.go
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
		return false
	}

	return true
}

func (teller *GoTeller) isNeighbor(from ipaddr.IPAddr) bool {
	teller.neighborsMutex.RLock()
	defer teller.neighborsMutex.RUnlock()
	for _, addr := range teller.Neighbors {
		if from == addr {
			return true
		}
	}
	return false
}

func (teller *GoTeller) neighborWithSameIP(addr ipaddr.IPAddr) (ipaddr.IPAddr, bool) {
	teller.neighborsMutex.RLock()
	defer teller.neighborsMutex.RUnlock()
	for _, neighbor := range teller.Neighbors {
		if neighbor.IP == addr.IP {
			return neighbor, true
		}
	}
	return ipaddr.IPAddr{}, false
}

func (teller *GoTeller) addNeighbor(newNode ipaddr.IPAddr) {
	teller.neighborsMutex.Lock()
	defer teller.neighborsMutex.Unlock()
	teller.Neighbors = append(teller.Neighbors, newNode)
}

func (teller *GoTeller) removeNeighbor(deadNeighbor ipaddr.IPAddr) {
	teller.neighborsMutex.Lock()
	defer teller.neighborsMutex.Unlock()
	for i, addr := range teller.Neighbors {
		if addr == deadNeighbor {
			teller.Neighbors = append(teller.Neighbors[:i], teller.Neighbors[i+1:]...)
			break
		}
	}
}

func (teller *GoTeller) newID() [16]byte {
	var id [16]byte
	addrBuffer := teller.addr.ToBytes()
	copy(id[:6], addrBuffer)
	var numNeighbors uint16 = uint16(len(teller.Neighbors))
	var randomNum uint32 = uint32(teller.randGen.Int31n(int32(numNeighbors)+1) + teller.randGen.Int31())
	binary.LittleEndian.PutUint16(id[6:8], numNeighbors)
	binary.LittleEndian.PutUint32(id[8:12], teller.hashCount)
	binary.LittleEndian.PutUint32(id[12:], randomNum)
	teller.hashCount++
	return id
}

func (teller *GoTeller) SendQuery(query Query) error {
	defer func() {
		if r := recover(); r != nil {
			if teller.debugFile != nil {
				fmt.Fprintln(teller.debugFile, r)
			}
		}
	}()
	// First check if both callbacks have been set
	if query.onHit == nil && query.onResponse != nil {
		return fmt.Errorf("Must set OnHit callback for query (Use OnHit(callback))")
	} else if query.onResponse == nil && query.onHit != nil {
		return fmt.Errorf("Must set OnResponse callback for query (Use OnResponse(callback))")
	} else if query.onResponse == nil && query.onHit == nil {
		return fmt.Errorf("Must set OnHit and OnResponse callbacks for query (Use OnHit(callback) & OnResponse(callback))")
	}
	if query.TTL == 0 {
		return fmt.Errorf("TTL on query for \"%s\" was 0. Query TTL must be greater than 0.", query.SearchQuery)
	}
	// Save query into myQueries map
	descID := teller.sendQuery(query.SearchQuery, query.TTL, query.MinSpeed, teller.addr)
	teller.myQueryMapMutex.Lock()
	defer teller.myQueryMapMutex.Unlock()
	teller.myQueries[descID] = query
	return nil
}
