package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// This is the core scheduler. Everything in main() is just managing the
// websocket message protocol.
func DecideLaunches(_ Scenario, state State) []Launch {
	pendingOrders := state.PendingOrders
	availableDroneIds := state.AvailableDroneIds

	launches := []Launch{}

	nextDroneIndex := 0
	for _, o := range pendingOrders {
		if nextDroneIndex >= len(availableDroneIds) {
			break
		}

		droneId := availableDroneIds[nextDroneIndex]
		nextDroneIndex++

		launches = append(launches, Launch{droneId, []int{o.OrderId}})
	}

	return launches
}

func main() {
	serverUrlBase := flag.String("server_url_base", "", "a wss:// or ws:// URL")
	entryName := flag.String("entry_name", "", "Name of this scheduler entry.")
	authToken := flag.String("auth_token", "", "Auth token for the human submitting this scheduler.")
	compMode := flag.Bool("comp_mode", false, "If true, connect to the server in competition mode.")

	flag.Parse()

	wsEndpoint := WsEndpointTesting
	if *compMode {
		wsEndpoint = WsEndpointComp
	}

	dialer := websocket.Dialer{}
	serverUrl := *serverUrlBase + "/" + wsEndpoint
	fmt.Println("Connecting to", serverUrl)
	conn, _, err := dialer.Dial(
		serverUrl,
		http.Header{},
	)
	if err != nil {
		fmt.Printf("cannot connect: %s", err.Error())
		return
	}

	err = conn.WriteJSON(MessageToServer{
		Handshake: &Handshake{
			EntryName: *entryName,
			AuthToken: *authToken,
		},
	})
	if err != nil {
		fmt.Printf("Error sending handshake: %s", err.Error())
		return
	}

	var message MessageToClient
	err = conn.ReadJSON(&message)
	if err != nil {
		fmt.Printf("Error reading HandshakeResult: %s", err.Error())
		return
	}

	handshakeResult := message.HandshakeResult
	if handshakeResult == nil {
		fmt.Println("Error: malformed handshake result, missing 'HandshakeResult' key in JSON.")
		return
	}

	if !handshakeResult.IsOk {
		fmt.Printf("Error: Failed handshake: %s\n", handshakeResult.Message)
		return
	}

	fmt.Printf("Handshake success! %s\n", handshakeResult.Message)

SCENARIOS_LOOP:
	for {
		message = MessageToClient{}
		err = conn.ReadJSON(&message)
		if err != nil {
			fmt.Printf("Error reading StartScenarioRun: %s", err.Error())
			break SCENARIOS_LOOP
		}

		if message.StartScenarioRun == nil {
			fmt.Println("Error: malformed StartScenarioRun, missing 'StartScenarioRun key in JSON.")
			return
		}
		scenario := message.StartScenarioRun.Scenario

	ONE_LOOP:
		for {
			// Messages are received in a JSON format. There's an object with
			// a single key representing the message type. The value for that
			// key is the message payload. See types.go for the overall
			// structure details.
			var message MessageToClient
			err := conn.ReadJSON(&message)
			if err != nil {
				fmt.Printf("Error reading: %s\n", err.Error())
				break SCENARIOS_LOOP
			}

			switch {
			case message.GetMoves != nil:
				state := message.GetMoves.State
				fmt.Printf(
					"Time %d: Pending orders %d, available drones %d.\n",
					state.TimeOfDay,
					len(state.PendingOrders),
					len(state.AvailableDroneIds),
				)

				launches := DecideLaunches(scenario, state)

				result := Moves{
					Launches: launches,
				}
				err = conn.WriteJSON(MessageToServer{
					Moves: &result,
				})
				if err != nil {
					fmt.Printf("Error writing Moves response: %s\n", err.Error())
					break SCENARIOS_LOOP
				}

			case message.EndScenarioRun != nil:
				stats := message.EndScenarioRun.Stats

				fmt.Printf("\nSTATS\n")
				for _, keySpec := range stats.KeySpecs {
					value := stats.Values[keySpec.FullName]
					fmt.Printf("%s: %v\n", keySpec.FullName, value)
				}
				break ONE_LOOP

			case message.Close != nil:
				closeMessage := message.Close
				if closeMessage.IsOk {
					fmt.Println(closeMessage.Message)
				} else {
					fmt.Printf("Error: %s\nConnection close is imminent!\n", closeMessage.Message)
				}
				break SCENARIOS_LOOP
			}
		}

		if !*compMode {
			break SCENARIOS_LOOP
		}
	}

	fmt.Println("Closed.")
}
