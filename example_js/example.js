"use strict";


function decideLaunches(state) {
    const pendingOrders = state.PendingOrders
    const availableDrones = state.AvailableDrones

    const launches = []

    let nextDroneIndex = 0
    for (let o of pendingOrders) {
        if (nextDroneIndex >= availableDrones.length) {
            break
        }

        const droneId = availableDrones[nextDroneIndex]
        nextDroneIndex += 1

        launches.push(
            {
                DroneId: droneId,
                OrderIds: [o.OrderId],
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
        console.log("SEND");
        send(socket, {
            AuthToken: "[auth token]",
            EntryName: "I didn't read the instructions",
        });
    });

    let handshakeResult = null;
    let initScenario = null;

    socket.addEventListener("message", (evt) => {
        console.log("MESSAGE", typeof(evt.data), evt.data);
        const messageObj = JSON.parse(evt.data);

        if (!handshakeResult) {
            handshakeResult = messageObj;
            console.log("Got handshake", handshakeResult);
            if (!handshakeResult.IsOk) {
                socket.close();
            }
            return;
        }

        if (!initScenario) {
            initScenario = messageObj;
            console.log("Got scenario", initScenario);
            return;
        }

        console.log("Got state");
        const getMovesReq = messageObj;
        if (getMovesReq.Error) {
            console.log(getMovesReq.Error.Message);
            socket.close();
            return;
        }

        if (getMovesReq.Stats) {
            console.log(JSON.stringify(getMovesReq.Stats, "", 4));
            socket.close();
            return;
        }

        console.log(
            `Time ${getMovesReq.State.TimeOfDay}: Pending orders ${getMovesReq.State.PendingOrders.length}, available drones ${getMovesReq.State.AvailableDrones.length}.`,
        )

        const launches = decideLaunches(getMovesReq.State);

        send(socket, {
            Launches: launches,
        });
    });
}

function send(ws, message) {
    const messageJson = JSON.stringify(message);
    ws.send(messageJson);
}

initPageInputs();
