# Example Traveller Entry - Go

This is an example Traveller Competition entry, implemented in Go. Per the technical details, it:
1. Opens a websocket
1. Sends the handshake message
1. Listens for the handshake response
1. Waits for the next scenario run to start by listening for the scenario initialization
1. Responds to GetMoves requests with a list of drone launches. It uses a very naive strategy of immediately launching a drone with a single package, whenever both an order is pending and a drone is present.

## To run

From the repository root directory, execute:
```
# Testing
go run ./example_go/ \
    --server_url_base=wss://treaster.net/tzc

# Competition
go run ./example_go/ \
    --server_url_base=wss://treaster.net/tzc \
    --auth_token=[your auth token] \
    --entry_name=[publicly visible name of this specific entry] \
    --comp_mode
```

## Code notes
* The core scheduler logic is implemented in `DecideLaunches()`. If you want to improve the scheduler's performance, this is the place to start. It takes the scenario specification (unused) and the current simulation State as an argument, and returns the list of Launches that it decides on.
* The websocket message types are represented as structs, defined in types.go. The example code reads JSON off the wire, then converts to Go structures as early as possible for more typesafe code. The types.go file is a *copy* of the types.go file used by the simulator engine. The copy keeps this example fully independent of the engine implementation.
