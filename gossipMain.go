package gossip
import (
	"sync"
	"time"
)

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
	writeMutex := sync.Mutex{}
	allNeighborhoods := initializeNeighbors(numServers)

	for i:= 0; i < numServers; i++ {
		println(i, " neighbors [", allNeighborhoods[i].neighborCommunication[0].neighborId, ",",
			allNeighborhoods[i].neighborCommunication[1].neighborId, ",",
			allNeighborhoods[i].neighborCommunication[2].neighborId, "]")
	}

	for i:= 0; i < numServers; i++ {
		go serverSimulation(i, allNeighborhoods[i], writeMutex)
	}

	time.Sleep(20)

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
	chan1to2 := make(chan heartbeat)
	chan2to1 := make(chan heartbeat)

	neighbor1To2 := neighborCommunication{id2, &chan1to2, &chan2to1}
	neighbor2To1 := neighborCommunication{id1, &chan2to1, &chan1to2}

	return neighbor1To2, neighbor2To1
}



func serverSimulation(id int, neighbors neighborhood, writeMutex sync.Mutex) {
	//isSick := false
	heartbeatTable := initializeHeartBeatTable(neighbors)

	printHeartBeatTable(id, heartbeatTable, writeMutex)

}


func initializeHeartBeatTable(neighbors neighborhood) []serverHeartStats{

	numNeighbors := len(neighbors.neighborCommunication)
	heartbeatTable := make([]serverHeartStats, numNeighbors, numNeighbors )

	for i := 0; i < numNeighbors; i++ {
		neighborComm := neighbors.neighborCommunication[i]
		heartbeatTable[i] = serverHeartStats{id: neighborComm.neighborId, heartbeatCounter: -1, heartBeatTime: time.Now()}
	}

	return heartbeatTable
}

func printHeartBeatTable(id int, heartBeatTable []serverHeartStats, writeMutex sync.Mutex) {
	println(id, "heartbeats: " )
	for i := 0; i < len(heartBeatTable); i++ {
		neighborStats := heartBeatTable[i]
		println("id:", neighborStats.id, "heartCounter:", neighborStats.heartbeatCounter, "timer:", neighborStats.heartBeatTime.String())
	}
}

func main() {
	runGossipSimulation(8)
}