# Travelling Drone Competition

A hypothetical drone medical delivery system delivers many types of products to many hospitals using a fleet of drones. In order to provide the best experience to customers, we want to minimize time-to-delivery. However, hospitals may have different needs for different orders, so they can classify high-urgency orders as "Emergency", while low-urgency orders are classified as "Resupply".

The system has a limited number of drones, and each drone can carry multiple packages. This enables a series of packages to be delivered in sequence to a series of hospitals, or multiple packages to the same hospital in delivery.

With the noisy, random nature of these hospital orders, scheduling which drones carry which packages to which hospitals becomes a (maybe) interesting scheduling challenge. The Travelling Drone Competition is an invitation to write a drone delivery scheduling algorithm and have it compete against other schedulers. This code repository contains instructions and example starting code in multiple languages to facilitate building a scheduler entry.

The competition engine runs a scenario every N minutes. Each submitted scheduler gets its own instance of the scenario, executed in lockstep with all others. The engine provides the current simulation state to each scheduler entry. The entry then returns the list of drone launches it wants to make. The engine applies the moves, generates the next tick of the simulation in a consistent and fair way, and the game proceeds until the simulation is concluded.

Throughout the simulation, the engine tracks various metrics on how each scheduler performs. The results are displayed in a dashboard, which shows the history of all scenarios.

Travelling Drone Competition makes no judgments about which metrics are good or bad, or whether higher values are better or worse. It simply collects and displays the metrics data. What you choose to optimize for is entirely up to you!

# Technical Overview

To play, an entrant must write a program that can connect to the server via websocket. The program and engine communicate over the websocket using a simple RPC-like protocol.

1. Connecting the websocket requires an authorization token. Contact the game manager (that's a human) to obtain a token.
1. The program must send the handshake message, including the previously mentioned token, in a particular format to initialize the connection. A malformed handshake will result in the websocket being closed immediately.
1. Every N minutes, the competition engine starts the next scenario. All open websockets will be included in the round. Any websocket connections that are opened after a round begins must wait for the following round.
1. The engine will send an initial specification of the scenario parameters. This includes things like hospital location, number and configuration of drones, etc.
1. The engine will send move request. This request contains the simulation state at that time, including currently pending orders and available drones. The scheduler program's response must contain the list of which drones to launch, with which orders loaded onto each.
1. When the scenario run is completed, the server will send an end message which instead contains the metrics gathered for that run. Once this message is sent, the scheduler program should prepare itself for a new scenario, per step 2.

# Technical Details

1. Open a websocket to `wss://treaster.net/tsc/ws-competition`
1. Send a message on the websocket with the following JSON-formatted message:

    ```
    {
        // The top-level key name indicates the message type and payload structure.
        // All websocket messages use this approach.
        "Handshake": {
            // An authorization token that allows access to play the game. The token
            // will be used to verify your participation in the competition, and to
            // look up your user initials for public display.
            //
            // Contact the game manager (that's a human) to obtain a token.
            "AuthToken": string

            // The name of your entry. This will be displayed in competition
            // leaderboards, alongside your user initials. The name could be descriptive
            // of your strategy with this entry, or could be something otherwise fun or
            // interesting. Ideally, you'd use different names for different strategies.
            // Incremental changes might result in small name changes -
            // "Greedy v1" -> "Greedy V2", while bigger implementation changes might
            // justify bigger name changes.
            "EntryName": string
        }
    }
    ```

1. Read from the websocket to receive a JSON-formatted handshake response:

    ```
    {
        "HandshakeResult": {
            // whether the handshake was successful
            "IsOk": bool

            // a human-readable string with more information
            "Message": string

            // A time limit in milliseconds of how long the client has to submit
            // moves back to the server. Note that this is from the server's
            // perspective, so message transport times are included. If a move
            // response ever exceeds this threshold, the server will close the
            // websocket. Yes, this is unfair to users with slower or less reliable
            // internet connections, but thems the rules!
            TimeoutMs int

            // Number of seconds between the start of each scenario run.
            ScenarioFreqSecs int

            // Start time of the next scenario run, in RFC3339 format.
            NextStartDatetime string
        }
    }
    ```

1. Read from the websocket to receive a JSON-formatted description of the scenario:

    ```
    {
        "StartScenarioRun": {
            "Scenario": {
                // The coordinate of the warehouse from which drones are launched
                "WarehousePosition": {
                    "X": float64
                    "Y": float64
                },

                // A mapping of hospital name -> hospital coordinate
                "Hospitals": {
                    "[hospital name]": {
                        "X": float64
                        "Y": float64
                    },
                    ...
                }

                // A list of drone descriptions.
                // Note that different drones may have different capabilities.
                "Drones": [
                     {
                         "DroneId": int,
                         "MaxCapacity": int
                         "MaxSpeed": float64
                         "MaxRange": float64
                     },
                     ...
                ],

                // the end time of the scenario
                "MaxTime": int,
            }
        }
    }
    ```

1. MAIN GAME LOOP:

    1. Read from the websocket to receive a JSON-formatted message. Three messages are possible: GetMoves, EndScenarioRun, and Close.

       GetMoves indicates the simulation needs a new set of moves from the client/entry. The client should make its decisions, then send a "Moves" response. (see further below)

    ```
    {
        "GetMoves": {
            "State": {
                // the tick number of the simulation. When this reaches MaxTime,
                // the scenario ends.
                "TimeOfDay": int
                "PendingOrders": [
                    {
                        "OrderId": int
                        "Time": int
                        "Hospital": string
                        "Priority" string ("Emergency" | "Resupply")
                    },
                    ...
                ]

                // List of DroneIds that are available for immediate launch.
                "AvailableDrones": []int

                // The current status of drones that are in flight.
                "BusyDrones": [
                    {
                        "DroneId": int

                        // How many simulation ticks until the drone is returned to
                        // the warehouse and is available for another launch.
                        "TimeToAvailability": int

                        "Position": {
                            "X": float64
                            "Y": float64
                        }
                    },
                    ...
                ]
            },
        }
    }
    ```

        If the simulation has completed, the server will send EndScenarioRun. This indicates that no more calls to GetMoves will be made until the next scenario begins. Additionally, it includes statistics on how this entry performed during the scenario. Although no more GetMoves messages will be sent, Close may still be sent after EndScenarioRun.
    ```
    {
        "EndScenarioRun": {
            // Stats are intended for display only, not for structured computation.
            "Stats": {
                "[metric name]": [metric value]
                ...
            }
        }
    }
    ```


        Close indicates the websocket will be closed by the server.
    ```
    {
        "Close": {
            // If IsOk is true, everything is healthy and this close is
            // expected and normal. If IsOk is false, some sort of error has
            // occurred. This may be an internal server error, or a violation
            // of this communication protocol by the client.
            "IsOk: bool,

            // A human-readable string indicating the reason for the close.
            "Message": string
        }
    }
    ```


    1. If "State" was received, write to the websocket a JSON-formatted message describing which drones should launch, carrying which orders. This message must be *received* by the server within [TimeoutMs] of when the request was sent, otherwise the server will terminate the websocket.

        ```
        {
            "Moves": {
                "Launches": [
                    {
                        "DroneId": int
                        "OrderIds": []int
                    },
                    ...
                ]
            }
        }
        ```

1. Run your scheduler wherever you like which can access the competition engine's URL. As long as the program runs and the websocket is connected, the scheduler will be automatically entered into each scenario. If the program terminates or the websocket otherwise disconnects, the scheduler will no longer participate.


# How To Test


Run the client entry as usual, but change the websocket URL to `wss://treaster.net/tsc/ws-testing. This will cause the engine to immediately kick off a single scenario with a fixed configuration. The scenario will run normally end-to-end, but with only one participating scheduler, and no stats are recorded. If no error is received, and stats are received, then your scheduler works! If an error is received ... you might have more work to do.


# Traveller Competition Design Considerations

The highest-priority design consideration for the implementation of this game was Security: How does one run arbitrary code submissions in a way that protects against a submission either attempting to root the host machine, or use the machine's resources to mine bitcoin?

The secondary consideration was User Experience: How can the system design facilitate a low-friction process for building and submitting a scheduler implementation?

Finally, the system design needed to be practical to build by a solo developer with full-time employment on the side and who's self-funding the infrastructure expenses. So, no big Kubernetes clusters!

How did the websocket approach stack up? What other solutions can *you* imagine?
