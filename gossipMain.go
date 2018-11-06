package main
import (
	"fmt"
    "math/rand"
	"sync"
	"time"
)

type heartbeat int

type serverHeartStats struct {
	id int
	heartBeatCounter int
	heartBeatTime time.Time
	failing bool
}

type neighborhood struct { //The neighbors of one node to all of it's friends nodes
	neighborCommunication []neighborCommunication
}

type neighborCommunication struct {
	neighborId int
	outgoing *chan []serverHeartStats //from myId to neighborId
	incoming *chan []serverHeartStats //from neighborId to myId
}

var wg sync.WaitGroup

// time constants parameters
var timeout int = 5
var heartRate int = 2
var sendout int = 3
var failRate int = 2

func runGossipSimulation(numServers int) {
	allNeighborhoods := initializeNeighbors(numServers)
	wg.Add(numServers)

	for i:= 0; i < numServers; i++ {
		println(i, " neighbors [", allNeighborhoods[i].neighborCommunication[0].neighborId, ",",
			allNeighborhoods[i].neighborCommunication[1].neighborId, ",",
			allNeighborhoods[i].neighborCommunication[2].neighborId, "]")
	}

	for i:= 0; i < numServers; i++ {
            go serverSimulation(i, allNeighborhoods[i], numServers)
	}

	time.Sleep(1000000)

}

//Each node has a connection to the ones next to it in the circle and the ones directly across
func initializeNeighbors(numServers int) []neighborhood {
	allNeighborhoods := make([]neighborhood, numServers)

	for i := 0; i < numServers; i++ { //connect neighbors circularly
		neighborId := i + 1
		if neighborId >= numServers  {
			neighborId = 0
		}
		neighbor1To2, neighbor2To1 := connectNeighbors(i, neighborId)
		allNeighborhoods[i].neighborCommunication = append(allNeighborhoods[i].neighborCommunication, neighbor1To2)
		allNeighborhoods[neighborId].neighborCommunication = append(allNeighborhoods[neighborId].neighborCommunication, neighbor2To1)
	}

	for i := 0; i < numServers/2; i++ { //connect neighbor across
		neighborId := numServers/2 + i
		neighbor1To2, neighbor2To1 := connectNeighbors(i, neighborId)
		allNeighborhoods[i].neighborCommunication = append(allNeighborhoods[i].neighborCommunication, neighbor1To2)
		allNeighborhoods[neighborId].neighborCommunication = append(allNeighborhoods[neighborId].neighborCommunication, neighbor2To1)
	}

	return allNeighborhoods
}

func connectNeighbors(id1 int, id2 int) (neighborCommunication, neighborCommunication) {
	chan1to2 := make(chan []serverHeartStats, 100)
	chan2to1 := make(chan []serverHeartStats, 100)

	neighbor1To2 := neighborCommunication{id2, &chan1to2, &chan2to1}
	neighbor2To1 := neighborCommunication{id1, &chan2to1, &chan1to2}

	return neighbor1To2, neighbor2To1
}

func serverSimulation(id int, neighbors neighborhood, numServers int) {
	isFailing := false;
	defer wg.Done()

        printMutex := &sync.Mutex{}
	heartBeatTable := initializeHeartBeatTable(neighbors, numServers)

	heartBeatTable[id] = serverHeartStats{id: id, heartBeatCounter: 0, heartBeatTime: time.Now()}
	sendClock := time.Now()
        printMutex.Lock()
    failClock := time.Now()

	fmt.Println(id)
	for {
        if time.Now().Sub(failClock) > time.Duration(failRate)*time.Second {
			failClock = time.Now()
            if (rand.Intn(8) == id) {
                isFailing = !isFailing
                fmt.Println(id, "'s Failing State change to: ", isFailing)
            }
		}

		if !isFailing && time.Now().Sub(sendClock) > time.Duration(sendout)*time.Second {
			sendClock = time.Now()
			for idx, _ := range neighbors.neighborCommunication {
				*neighbors.neighborCommunication[idx].outgoing <- heartBeatTable
			}
			printHeartBeatTable(id, heartBeatTable)
		}
		select {
		    case m1 := <- (*neighbors.neighborCommunication[0].incoming):
			    for _, m := range m1 {
				    tableLength := len(heartBeatTable)
				    updateHeartBeatTable(&heartBeatTable, tableLength, m)
			    }
		    case m2 := <- (*neighbors.neighborCommunication[1].incoming):
			    for _, m := range m2 {
				    tableLength := len(heartBeatTable)
				    updateHeartBeatTable(&heartBeatTable, tableLength, m)
			    }
		    case m3 := <- (*neighbors.neighborCommunication[2].incoming):
			    for _, m := range m3 {
				    tableLength := len(heartBeatTable)
				    updateHeartBeatTable(&heartBeatTable, tableLength, m)
			    }
		    default:
			    if time.Now().Sub(heartBeatTable[id].heartBeatTime) > time.Duration(heartRate)*time.Second {
				    heartBeatTable[id].heartBeatCounter += 1
				    heartBeatTable[id].heartBeatTime = time.Now()
			    }
		}
	}
        printMutex.Unlock()
}

func initializeHeartBeatTable(neighbors neighborhood, numServers int) []serverHeartStats{

	numNeighbors := len(neighbors.neighborCommunication)
	heartBeatTable := make([]serverHeartStats, numServers, numServers)

	for i := 0; i < numNeighbors; i++ {
		neighborComm := neighbors.neighborCommunication[i]
		heartBeatTable[neighborComm.neighborId] = serverHeartStats{id: neighborComm.neighborId, heartBeatCounter: -1, heartBeatTime: time.Now()}
	}

	return heartBeatTable
}

func updateHeartBeatTable(heartBeatTable *[]serverHeartStats, tableLength int, m serverHeartStats) {
	if (*heartBeatTable)[m.id].heartBeatCounter < m.heartBeatCounter {
		(*heartBeatTable)[m.id].id = m.id
		(*heartBeatTable)[m.id].heartBeatCounter = m.heartBeatCounter
		(*heartBeatTable)[m.id].heartBeatTime = m.heartBeatTime
		(*heartBeatTable)[m.id].failing = m.failing
	}
}

func checkForTimeouts(heartBeatTable *[]serverHeartStats, tableLength int) {
	for idx, _ := range (*heartBeatTable) {
		if time.Now().Sub((*heartBeatTable)[idx].heartBeatTime) > time.Duration(2*timeout)*time.Second {
			(*heartBeatTable) = append((*heartBeatTable)[:idx], (*heartBeatTable)[idx+1:]...)
		}
	}
}

func printHeartBeatTable(id int, heartBeatTable []serverHeartStats) {
        printMutex := &sync.Mutex{}
        printMutex.Lock()
        println(id, "heartBeatTable: " )
	for i := 0; i < len(heartBeatTable); i++ {
		neighborStats := heartBeatTable[i]
		println("id:", neighborStats.id, "heartCounter:", neighborStats.heartBeatCounter, "timer:", neighborStats.heartBeatTime.String())
	}
        printMutex.Unlock()
}

func main() {
	runGossipSimulation(8)
	wg.Wait()
}
