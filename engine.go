package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	//"os"
)

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}


type Board struct {
	//Message string
	World [][]byte
	P     Params
}
type BoardResponse struct {
	//Message  string
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

type Engine struct{}

func (e *Engine) NewBoard(req Board, res *BoardResponse) (err error) {
	var boardRequest *Board
	if boardRequest == nil {
		err = errors.New("???")
		return
	}

	//fmt.Println("engine:" + req.Message)

	newWorld := req.World
	turn := 0
	for ; turn < req.P.Turns; turn++ {
		newWorld = calculateNextState(req.P, newWorld, turn)
	}
	res.NewWorld = newWorld
	res.NewTurn = turn
	//*res = reply
	//res.NewTurn = 100
	return
	
}

func main() {
	//var api = new(Engine)
	rpc.Register(&Engine{})

	//tcpAddr, err := net.ResolveTCPAddr("tcp", ":8033")
	//checkError(err)

	pAddr := flag.String("port", "8033", "Port to listen on")
	//flag.StringVar(&nextAddr, "next", "localhost:8040", "IP:Port string for the next member")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	/*rr := rpc.Register(api)
	if err != nil {
		log.Fatal("err API", err)
	}*/
	
	listener, err := net.Listen("tcp", ":"+*pAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Println("Engine Start")
	defer listener.Close()
	
	//rpc.Accept(listener)
	_, err = listener.Accept()
		if err != nil {
			log.Print("rpc.Serve: accept:", err.Error())
			return
		}
}

/*func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}*/

