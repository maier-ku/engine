package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	//"uk.ac.bris.cs/gameoflife/gol"
	//"uk.ac.bris.cs/gameoflife/gol"
)

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

//PGM image
type Board struct {
	World [][]byte
	Turn  int
	P     Params
}
type BoardResponse struct {
	NewWorld [][]byte
	NewTurn  int
}

const alive = 0xFF
const dead = 0x00

func mod(x, m int) int {
	return (x + m) % m
}

func countNeighbours(p Params, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

func calculateNextState(p Params, world [][]byte, turn int) [][]byte {
	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := countNeighbours(p, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

type Engine struct {
}

func (e *Engine) NewBoard(req Board, res *BoardResponse) (err error) {
	fmt.Println("-----> Engine req:", req.P)

	var reply BoardResponse
	turn := 0
	reply.NewWorld = req.World
	for ; turn < req.P.Turns; turn++ {
		reply.NewWorld = calculateNextState(req.P, req.World, turn)
		reply.NewTurn = turn + 1
	}
	fmt.Println("-----> Engine req.P:", req.P)
	fmt.Println("-----> Engine Turn:", turn)

	*res = reply
	return
}

func main() {
	var api = new(Engine)
	err := rpc.Register(api)
	if err != nil {
		log.Fatal("err API", err)
	}
	pAddr := flag.String("port", "8033", "Port to listen on")
	//flag.StringVar(&nextAddr, "next", "localhost:8040", "IP:Port string for the next member")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	//rpc.Register(&Engine{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)

	fmt.Println("Engine Start")
	defer listener.Close()
	rpc.Accept(listener)
}
