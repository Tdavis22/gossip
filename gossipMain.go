package gossip

import "time"

type heartbeat int

type serverHeartStats struct {
	id int
	heartbeatCounter int
	heartBeatTime time.Time
}

type neighborhood struct { //The neighbors of one node to all of it's friends nodes
	neighborCommunication []neighborCommunication
}

type neighborCommunication struct {
	neighborId int
	outgoing *chan heartbeat //from myId to neighborId
	incoming *chan heartbeat //from neighborId to myId
}


func runGossipSimulation(numServers int) {

	allNeighborhoods := initializeNeighbors(numServers)

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
	chan1to2 := make(chan heartbeat)
	chan2to1 := make(chan heartbeat)

	neighbor1To2 := neighborCommunication{id2, &chan1to2, &chan2to1}
	neighbor2To1 := neighborCommunication{id1, &chan2to1, &chan1to2}

	return neighbor1To2, neighbor2To1
}



func serverSimulation(id int, neighbors neighborhood) {
  heartBeatTable := []*serverHeartStats{}

  serverStatus := new(serverHeartStats)

  println(id)


  heartBeatTable = append(heartBeatTable, serverStatus)
}


func main() {
	runGossipSimulation(4)
}