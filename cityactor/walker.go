package main

import (
	"fmt"
	"os"
	"time"

	"github.com/klekih/city/common"
)

var ticker *time.Ticker
var myRoute *Route
var currentInstructionIndex = -1
var currentPosInInstruction float64
var averageSpeed float64 = 10 // meter/second

// WalkerStatus is the answer for reporting walker status
type WalkerStatus struct {
	arrivedAtDestination bool
}

// StartWalker is the main entry point for starting a walker
func StartWalker(config *Config, city *common.CityInterface) chan WalkerStatus {

	// first, get a route to go on
	myRoute = getRoute(config)

	if len(myRoute.Paths) == 0 {
		panic("Route has no path to go on")
	}

	fmt.Println("Got route, going on it")

	reportChan, junctionChan := Connect()

	// create the chan to report back the status
	c := make(chan WalkerStatus)

	duration := time.Duration(config.SimulationStep)
	ticker = time.NewTicker(duration * time.Millisecond)

	go func() {
		for {
			select {
			case <-ticker.C:
				advance(city, reportChan, junctionChan)
			case j := <-junctionChan:
				fmt.Println("Received answer with junction", j)
			}
		}
	}()

	return c
}

func getRoute(config *Config) *Route {

	route, err := generateRandomRoute(config)

	if err != nil {
		panic(err)
	}

	return route
}

func advance(city *common.CityInterface, chanReport chan common.Report, junctionChan chan common.Junction) {

	if myRoute == nil {
		panic("No route to go on")
	}

	// Find out which step from instruction are we on.
	// Begin with first if needed
	if currentInstructionIndex == -1 {
		currentInstructionIndex = 1
	}

	// Check if all instructions have been passed.
	if len(myRoute.Paths[0].Instructions) <= currentInstructionIndex {
		fmt.Println("Route finished")
		os.Exit(0)
	}

	// Get the current distance to cover
	currentInstruction := myRoute.Paths[0].Instructions[currentInstructionIndex]
	distance := currentInstruction.Distance

	currentPosInInstruction += averageSpeed

	if currentPosInInstruction >= distance {
		currentInstructionIndex++
		currentPosInInstruction = 0
	} else {
		fmt.Println("On", currentInstruction.StreetName, ":",
			currentPosInInstruction, "of", distance)
	}

	// Now get the corresponding line string to report to the city
	// the path we're going on

	firstPointIndex := currentInstruction.Interval[0]
	secondPointIndex := currentInstruction.Interval[1]

	line := myRoute.Paths[0].Points.Coordinates[firstPointIndex:secondPointIndex]

	fmt.Println("Report to the city")

	report := common.Report{}.WithCurrentLine(line)

	chanReport <- report

	fmt.Println("Ask for next junction")
	junctionChan <- common.Junction{}
}
