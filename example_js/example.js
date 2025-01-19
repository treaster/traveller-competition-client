"use strict";


function decideLaunches(scenarioSpec, state) {
    const pendingOrders = state.PendingOrders
    const availableDroneIds = state.AvailableDroneIds;

    const launches = []

    let nextDroneIndex = 0
    for (const orderId in pendingOrders) {
        if (nextDroneIndex >= availableDroneIds.length) {
            break
        }

        const droneId = availableDroneIds[nextDroneIndex]
        nextDroneIndex += 1

        launches.push(
            {
                DroneId: droneId,
                OrderIds: [orderId],
            },
        )
    }

    return launches
}


function initPageInputs() {
    const testButton = document.getElementById("testButton");
    testButton.onclick = (evt) => {
        const serverUrlBase = document.getElementById("serverUrlBase").value;
        run(serverUrlBase, "ws-testing");
    };

    const compButton = document.getElementById("compButton");
    compButton.onclick = (evt) => {
        const serverUrlBase = document.getElementById("serverUrlBase").value;
        run(serverUrlBase, "ws-competition");
    };
}


function disablePageInputs() {
    document.getElementById("testButton").disabled = "disabled";
    document.getElementById("compButton").disabled = "disabled";
    document.getElementById("serverUrlBase").disabled = "disabled";
}


export function run(serverUrlBase, wsEndpoint) {
    disablePageInputs();

    const serverUrl = serverUrlBase + "/" + wsEndpoint;

    const socket = new WebSocket(serverUrl);
    socket.addEventListener("open", (evt) => {
        send(socket, {
            Handshake: {
                AuthToken: "[auth token]",
                EntryName: "I didn't read the instructions",
            },
        });
    });

    let scenarioSpec = null;

    socket.addEventListener("message", (evt) => {
        console.log("RECV MESSAGE", evt.data);
        const messageObj = JSON.parse(evt.data);

        if (messageObj.HandshakeResult) {
            const handshakeResult = messageObj.HandshakeResult;
            console.log("Got handshake", handshakeResult);
            if (!handshakeResult.IsOk) {
                socket.close();
            }
            return;
        }

        if (messageObj.StartScenarioRun) {
            const initScenario = messageObj.StartScenarioRun;
            scenarioSpec = initScenario.Scenario;
            console.log("Got scenario", initScenario);
            return;
        }

        if (messageObj.GetMoves) {
            const getMovesReq = messageObj.GetMoves;
            console.log(
                `Time ${getMovesReq.State.TimeOfDay}: Pending orders ${Object.keys(getMovesReq.State.PendingOrders).length}, available drones ${getMovesReq.State.AvailableDroneIds.length}.`,
            )

            const launches = decideLaunches(scenarioSpec, getMovesReq.State);
    
            send(socket, {
                Moves: {
                    Launches: launches,
                },
            });

            return;
        }

        if (messageObj.EndScenarioRun) {
            const endScenario = messageObj.EndScenarioRun;
            console.log(JSON.stringify(endScenario.Stats.Values, "", 4));
            return;
        }

        if (messageObj.Close) {
            const closeMsg = messageObj.Close;
            console.log(`Websocket closing. ${closeMsg.Message}`);
            return;
        }
    });
}

function send(ws, message) {
    const messageJson = JSON.stringify(message);
    console.log("SEND MESSAGE", messageJson);
    ws.send(messageJson);
}

initPageInputs();
