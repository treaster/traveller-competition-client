# Example Traveller Entry - Python

This is an example Traveller Competition scheduler, implemented in Python. Per the technical details, it:
1. Opens a websocket
1. Sends the handshake message
1. Listens for the handshake response
1. Waits for the next scenario run to start by listening for the scenario initialization
1. Responds to GetMoves requests with a list of drone launches. It uses a very naive strategy of immediately launching a drone with a single package, whenever both an order is pending and a drone is present.

## To run

From the repository root directory, execute:
```
pip install -r example_py/requirements.txt

# Testing
go run ./example_py/ \
    --server_url_base=wss://treaster.net/tsc

# Competition
go run ./example_py/ \
    --server_url_base=wss://treaster.net/tsc \
    --auth_token=[your auth token] \
    --entry_name=[publicly visible name of this specific entry] \
    --comp_mode

```


## Code notes
* The core scheduler logic is implemented in `decide_launches()`. If you want to improve the scheduler's performance, this is the place to start. It takes the scenario specifation (unused) and the current simulation State as an argument, and returns the list of Launches that it decides on.
* The websocket messages are decoded into untyped objects with arbitrary fields. This makes it easy to access the fields, but it makes the code less self-documenting. Check the top-level READMME.md or example_go/types.go for websocket API documentation.
* Running the main program vs. running the unit test with pytest seems to require different import styles, which we haven't had time to resolve yet. As such, running pytest on example_py/ will fail with obtuse module errors.
