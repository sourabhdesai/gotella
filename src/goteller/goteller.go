package goteller

import (
	"../ipaddr"
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sync"
)

type HitResult messages.HitResult

const SEED = 7187

type GoTeller struct {
	alive          bool
	debugFile      *io.Writer
	addr           IPAddr
	Neighbors      []IPAddr
	NumShared      uint32
	NumKB          uint32
	Port           uint16
	NetworkSpeed   uint32
	hashCount      uint32
	servantID      string
	randGen        *rand.Rand
	savedPings     map[[16]byte]IPAddr
	savedQueries   map[[16]byte]IPAddr
	neighborsMutex sync.RWMutex
	pingMapMutex   sync.RWMutex
	queryMapMutex  sync.RWMutex
	queryFunc      func(string) []HitResult
	resultFunc     func([]QueryResult, uint32, string) []QueryResult
	dataFunc       func(error, uint32, string, *net.Response)
	requestFunc    func(uint32, string) []byte
}

func (teller *GoTeller) StartAtPort(port uint16) error {
	teller.alive = true
	teller.randGen = rand.New(SEED)
	if teller.servantID == nil {
		teller.alive = false
		return fmt.Errorf("Must set Servant ID (use SetServantID)")
	}
	if teller.queryFunc == nil {
		teller.alive = false
		return fmt.Errorf("Must set Query callback function (use OnQuery)")
	}
	if teller.resultFunc == nil {
		teller.alive = false
		return fmt.Errorf("Must set Hit callback function (use OnHit)")
	}
	if teller.dataFunc == nil {
		teller.alive = false
		return fmt.Errorf("Must set Data callback function (use OnData)")
	}
	teller.Port = port
	teller.addr.Port = port
	err := teller.SetToLocalIP()
	if err != nil {
		return err
	}
	// TODO: Initialize other things
	return nil
}

func (teller *GoTeller) SetDebugFile(file *io.Writer) {
	teller.debugFile = file
}

func (teller *GoTeller) SetServantID(id string) {
	if l := len(id); l <= 16 {
		teller.servantID = id
	} else {
		teller.servantID = id[:16]
	}
}

func (teller *GoTeller) OnQuery(qFunc func(string) []HitResult) {
	teller.queryFunc = qFunc
}

func (teller *GoTeller) OnHit(rFunc func([]QueryResult, uint32, string) []QueryResult) {
	teller.resultFunc = rFunc
}

func (teller *GoTeller) OnData(dFunc func(error, uint32, string, *net.Response)) {
	teller.dataFunc = dFunc
}

// send msg to all neighbors except for from
func (teller *GoTeller) floodToNeighbors(msg []byte, from IPAddr) {
	teller.neighborsMutex.RLock()
	defer teller.neighborsMutex.RUnclock()
	for _, addr := range teller.Neighbors {
		if from != addr {
			teller.sendToNeighbor(msg, addr)
		}
	}
}

func (teller *GoTeller) sendToNeighbor(msg []byte, to IPAddr) {
	conn, err := net.Dial("tcp", to.String())
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
		return
	}

	defer conn.Close()

	connIO := bufio.NewReadWriter(&conn, &conn)
	connected, err := gnutellaConnect(connIO)
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
		return
	}
	if !connected {
		return
	}

	err = sendBytes(connIO, msg) // in requesthandler.go
	if err != nil {
		if teller.debugFile != nil {
			fmt.Fprintln(teller.debugFile, err)
		}
	}
}

func (teller *GoTeller) isNeighbor(from IPAddr) bool {
	teller.neighborsMutex.RLock()
	defer teller.neighborsMutex.RUnlock()
	for _, addr := range teller.Neighbors {
		if from == addr {
			return true
		}
	}
	return false
}

func (teller *GoTeller) addNeighbor(newNode IPAddr) {
	teller.neighborsMutex.Lock()
	defer teller.neighborsMutex.Unlock()
	teller.Neighbors = append(teller.Neighbors, newNode)
}

func (teller *GoTeller) newID() [16]byte {
	var id [16]byte
	addrBuffer := teller.addr.ToBytes()
	copy(id[:6], addrBuffer)
	var numNeighbors uint16 = len(teller.Neighbors)
	var randomNum uint32 = teller.randGen.Int31n(int32(numNeighbors)) + teller.randGen.Int31()
	binary.LittleEndian.PutUint16(id[6:8], numNeighbors)
	binary.LittleEndian.PutUint32(id[8:12], teller.hashCount)
	binary.LittleEndian.PutUint32(id[12:], randomNum)
	teller.hashCount++
	return id
}

func (teller *GoTeller) SendQuery(query string, ttl byte, minSpeed uint32) {
	teller.sendQuery(query, ttl, minSpeed, teller.addr)
}
