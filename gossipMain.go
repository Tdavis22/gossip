package main

import (
    "fmt"
    "sync"
    "time"
)

type heartbeat int

type serverHeartStats struct {
	id int
	heartBeatCounter int
	heartBeatTime time.Time
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

func runGossipSimulation(numServers int) {

	allNeighborhoods := initializeNeighbors(numServers)
    wg.Add(numServers)

	for i:= 0; i < numServers; i++ {
		go serverSimulation(i, allNeighborhoods[i])
	}

}

func initializeNeighbors(numServers int) []neighborhood {
	allNeighborhoods := make([]neighborhood, numServers)

	for i := 0; i < numServers; i++ {
		neighborId := i + 1
		if neighborId >= numServers  {
			neighborId = 0
		}
		neighbor1To2, neighbor2To1 := connectNeighbors(i, neighborId)
		allNeighborhoods[i].neighborCommunication = append(allNeighborhoods[i].neighborCommunication, neighbor1To2)
		allNeighborhoods[neighborId].neighborCommunication = append(allNeighborhoods[neighborId].neighborCommunication, neighbor2To1)
	}

	return allNeighborhoods
}

func connectNeighbors(id1 int, id2 int) (neighborCommunication, neighborCommunication) {
	chan1to2 := make(chan []serverHeartStats)
	chan2to1 := make(chan []serverHeartStats)

	neighbor1To2 := neighborCommunication{id2, &chan1to2, &chan2to1}
	neighbor2To1 := neighborCommunication{id1, &chan2to1, &chan1to2}

	return neighbor1To2, neighbor2To1
}



func serverSimulation(id int, neighbors neighborhood) {
  defer wg.Done()

  heartBeatTable := []serverHeartStats{}
   
  serverStatus := serverHeartStats{id, 0, time.Now()}
  heartBeatTable = append(heartBeatTable, serverStatus)

  sendClock := time.Now()

  fmt.Println(id)
  for {
      if time.Now().Sub(sendClock) > time.Duration(sendout)*time.Second {
          sendClock = time.Now()
          (*neighbors.neighborCommunication[0].outgoing) <- heartBeatTable;
          (*neighbors.neighborCommunication[1].outgoing) <- heartBeatTable;   
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
          default:
              if time.Now().Sub(serverStatus.heartBeatTime) > time.Duration(heartRate)*time.Second {
                  for idx, _ := range heartBeatTable {
                      if heartBeatTable[idx].id == id {
                          heartBeatTable[idx].heartBeatCounter += 1
                          heartBeatTable[idx].heartBeatTime = time.Now()
                          serverStatus.heartBeatTime = time.Now()
                      }
                  }
              }                  
      }
  }
}

func updateHeartBeatTable(heartBeatTable *[]serverHeartStats, tableLength int, m serverHeartStats) {
    tableIndex := 0
    var found bool = false
    for idx, _ := range (*heartBeatTable) {
        if (*heartBeatTable)[idx].id == m.id {
            if (*heartBeatTable)[idx].heartBeatCounter < m.heartBeatCounter {
                (*heartBeatTable)[idx].heartBeatCounter = m.heartBeatCounter
                (*heartBeatTable)[idx].heartBeatTime = time.Now()
            }
            found = true
            break
        }
    }
    if found == false {
        (*heartBeatTable) = append((*heartBeatTable), serverHeartStats{m.id, m.heartBeatCounter, time.Now()})
    }
}

func checkForTimeouts(heartBeatTable *[]serverHeartStats, tableLength int) {
    for idx, _ := range (*heartBeatTable) {
        if time.Now().Sub((*heartBeatTable)[idx].heartBeatTime) > time.Duration(2*timeout)*time.Second {
            (*heartBeatTable) = append((*heartBeatTable)[:idx], (*heartBeatTable)[idx+1:]...)
        }
    }
}


func main() {
	runGossipSimulation(4)
    wg.Wait()
}
