package main

// This file is a copy of the simulator's types.go, except the package name is
// changed and this comment is added. Copying this file instead of importing
// keeps the example_go implementation fully standalone, and thus more
// equivalent to the Python and Javascript examples.

const WsEndpointComp = "ws-competition"
const WsEndpointTesting = "ws-testing"

// ** WEBSOCKET PROTOCOL - MESSAGES SENT FROM SERVER TO CLIENT
//    Exactly one field should be non-nil.

type MessageToClient struct {
	HandshakeResult  *HandshakeResult
	StartScenarioRun *StartScenarioRun
	GetMoves         *GetMoves
	EndScenarioRun   *EndScenarioRun
	Close            *Close
}

// Sent from server to client on Handshake.
type HandshakeResult struct {
	IsOk              bool
	Message           string
	TimeoutMs         int
	ScenarioFreqSecs  int
	NextStartDatetime string // RFC3339 format
}

// Sent from server to client when scenario is about to start.
type StartScenarioRun struct {
	Scenario Scenario
}

// Sent from server to client on each simulation tick/turn.
// NOTE: If there's nothing for the client to do (no pending orders OR no
// available drones) then this message will not be sent and the client's turn
// will be skipped.
type GetMoves struct {
	State State
}

type EndScenarioRun struct {
	Stats Stats
}

// If IsOk is true, then the close is expected, healthy behavior.
// If IsOk is false, then the client demonstrated some kind of erroneous behavior.
// Message contains a human-readable explanation.
// After sending this message, the server will close the client's connection.
type Close struct {
	IsOk    bool
	Message string
}

// ** WEBSOCKET PROTOCOL - MESSAGES SENT FROM CLIENT TO SERVER
//    Exactly one field should be non-nil.

type MessageToServer struct {
	Handshake *Handshake
	Moves     *Moves
}

// Sent from client to server on each GetMoves.
type Moves struct {
	Launches []Launch
}

// Sent from client to server on connection.
type Handshake struct {
	AuthToken string
	EntryName string
}

// ** COMMON MESSAGE TYPES **
type State struct {
	TimeOfDay       int
	PendingOrders   []Order
	AvailableDrones []int
	BusyDrones      []DroneStatus
}

type KeySpec struct {
	FullName      string
	StatName      string
	Priority      Priority
	Type          string
	ShouldDisplay bool
	DisplayName   string
}

type DroneConfig struct {
	DroneId     int
	MaxCapacity int
	MaxSpeed    float64
	MaxRange    float64
}

type DroneStatus struct {
	DroneId            int
	TimeToAvailability int
}

type Coord struct {
	X float64
	Y float64
}

type Scenario struct {
	WarehousePosition Coord
	Hospitals         map[string]Coord
	Drones            []DroneConfig
	MaxTime           int
	SlaSecs           map[Priority]int
}

type Priority string

const (
	EMERGENCY Priority = "Emergency"
	RESUPPLY  Priority = "Resupply"
)

type Order struct {
	OrderId  int
	Time     int
	Hospital string
	Priority Priority
}

type Launch struct {
	DroneId  int
	OrderIds []int
}

type Stats struct {
	KeySpecs []KeySpec
	Values   map[string]any
}
