package gol

import (
	"fmt"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events    chan<- Event
	ioCommand chan<- ioCommand
	ioIdle    <-chan bool

	ioFileName chan<- string //send-only
	output     chan<- uint8
	input      chan uint8
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

//This function takes the current state of the world and completes one evolution of the world. It then returns the result.
func calculateNextState(p Params, world [][]byte, c distributorChannels, turn int) [][]byte {
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
					cell := util.Cell{X: x, Y: y}
					c.events <- CellFlipped{turn, cell}
				}
			}

			if world[y][x] == dead {
				if neighbours == 3 {
					newWorld[y][x] = alive
					cell := util.Cell{X: x, Y: y}
					c.events <- CellFlipped{turn, cell}
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

//This function takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	aliveCell := []util.Cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				aliveCell = append(aliveCell, util.Cell{X: x, Y: y})
			}
		}
	}
	return aliveCell
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	command := ioInput
	c.ioCommand <- command

	filename := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioFileName <- filename

	// output := ioOutput
	// c.output <- output

	// TODO: Create a 2D slice to store the world.
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			val := <-c.input
			world[y][x] = val
		}
	}

	// TODO: For all initially alive cells send a CellFlipped Event.

	turn := 0
	c.events <- AliveCellsCount{0, 0}
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			<-ticker.C
			c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world))}
		}
	}()

	//Implement logic here
	// TODO: Execute all turns of the Game of Life.
	for ; turn < p.Turns; turn++ {
		world = calculateNextState(p, world, c, turn)
		c.events <- TurnComplete{turn}
	}

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.
	//finalrurncomplete
	aliveCell := calculateAliveCells(p, world)
	c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: aliveCell}

	outputCommand := ioOutput
	c.ioCommand <- outputCommand

	outputFilename := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, p.Turns)
	c.ioFileName <- outputFilename

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.output <- world[y][x]
		}
	}
	c.events <- ImageOutputComplete{turn, outputFilename}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
	ticker.Stop()

}
