#!/usr/bin/env python3

from typing import (
    get_args,
    get_origin,
    Type,
    Union,
)
from websockets.sync.client import (
    ClientConnection,
    connect,
)
from argparse import ArgumentParser
import json


class Struct:
    def __init__(self, **entries):
        results = {}
        for k, v in entries.items():
            results[k] = deserialize(v)
        self.__dict__.update(results)

    def __getattr__(self, name: str) -> any:
        return self.__dict__.get(name, None)


def decide_launches(scenario: Struct, state: Struct) -> list[Struct]:
    pendingOrders = state.PendingOrders
    availableDrones = state.AvailableDrones

    launches: list[Struct] = []

    nextDroneIndex = 0
    for o in pendingOrders:
        if nextDroneIndex >= len(availableDrones):
            break

        droneId = availableDrones[nextDroneIndex]
        nextDroneIndex += 1

        launches.append(
            Struct(
                DroneId=droneId,
                OrderIds=[o.OrderId],
            )
        )

    return launches


def main() -> None:
    parser = ArgumentParser()
    parser.add_argument(
        "--server_url_base",
        default="",
        help="a wss:// or ws:// URL",
    )
    parser.add_argument(
        "--entry_name",
        default="",
        help="Name of this scheduler entry.",
    )
    parser.add_argument(
        "--auth_token",
        default="",
        help="Auth token for the human submitting this scheduler.",
    )
    parser.add_argument(
        "--comp_mode",
        default=False,
        action="store_true",
        help="If true, connect to the server in competition mode.",
    )

    args = parser.parse_args()

    ws_endpoint = "ws-testing"
    if args.comp_mode:
        ws_endpoint = "ws-competition"

    server_url = args.server_url_base + "/" + ws_endpoint

    print(f"Connecting to {server_url}")
    with connect(server_url) as websocket:
        send(
            websocket,
            Struct(
                Handshake=Struct(
                    AuthToken=args.auth_token,
                    EntryName=args.entry_name,
                ),
            ),
        )

        handshake_res = recv(websocket, "HandshakeResult")
        if not handshake_res.IsOk:
            print("Bad handshake")
            return

        print("Handshake success!")

        scenario = recv(websocket, "StartScenarioRun")
        while True:
            message = recv(websocket, None)

            if message.GetMoves:
                submessage = message.GetMoves
                print(
                    f"Time {submessage.State.TimeOfDay}: Pending orders {len(submessage.State.PendingOrders)}, available drones {len(submessage.State.AvailableDrones)}.",
                )

                launches = decide_launches(scenario.Scenario, submessage.State)
                send(
                    websocket,
                    Struct(
                        Moves=Struct(
                            Launches=launches,
                        ),
                    ),
                )

            if message.Error:
                print(message.Error.Message)
                return

            if message.EndScenarioRun:
                print("Done!")
                stats = serialize(message.EndScenarioRun.Stats)
                print(json.dumps(stats, indent=4))
                if not args.comp_mode:
                    return


def send(ws: ClientConnection, obj: Struct) -> str:
    obj_dict = serialize(obj)
    obj_json = json.dumps(obj_dict)
    ws.send(obj_json)


def recv(ws: ClientConnection, expected_submessage: str | None) -> any:
    obj_json = ws.recv()
    obj_dict = json.loads(obj_json)

    obj = deserialize(obj_dict)
    if expected_submessage:
        obj = getattr(obj, expected_submessage)
    return obj


def deserialize(data: any) -> any:
    if isinstance(data, list):
        results = []
        for item in data:
            results.append(deserialize(item))
        return results

    if isinstance(data, dict):
        return Struct(**data)

    return data


def serialize(obj: any) -> dict[str, any]:
    if isinstance(obj, Struct):
        obj = obj.__dict__

    if isinstance(obj, dict):
        result = {}
        for k, v in obj.items():
            result[k] = serialize(v)
        return result

    if isinstance(obj, list):
        result = []
        for item in obj:
            result.append(serialize(item))
        return result

    return obj


if __name__ == "__main__":
    main()
