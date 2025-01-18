# Example Traveller Entry - Javascript

## Disclaimer

This example code probably doesn't work as-is. The server-client protocol has changed a couple of times since this was last updated. If you're interested in a JS solution, you have some options:
1. Contact the game manager (that's a human) and ask for help.
1. Study the Go (recommended) or Python protocols, and update the JS example code to match. It's somewhat close, just not all the way there.

## Actual Documentation

This is an example Traveller Competition scheduler, implemented in Javascript. Per the technical details, it:
1. Opens a websocket
1. Sends the handshake message
1. Listens for the handshake response
1. Waits for the next scenario run to start by listening for the scenario initialization
1. Responds to GetMoves requests with a list of drone launches. It uses a very naive strategy of immediately launching a drone with a single package, whenever both an order is pending and a drone is present.

## To run

This JS example is a bit janky. To minimize dependencies, we implemented a simple web page that connects the websocket and runs the scheduler, but you still need some way to serve the static HTML and JS file. We use a simple webserver available at https://github.com/treaster/gohttp, but any static files server will do.

With that in mind:
1. Serve the files however you can.
1. Open the hosted web page. We go to http://localhost:8080, but your address will vary.
1. Click the "Test" button to run a test, or the "Competition" button to enter the competition.
1. Watch the browser's debug console for output.

## Code notes
* The core scheduler logic is implemented in `decideLaunches()`. If you want to improve the scheduler's performance, this is the place to start. It takes the current simulation State as an argument, and returns the list of Launches that it decides on. It would be reasonable to pass the Scenario definition as an additional argument.
* This can surely be improved in all manner of ways, like making it executable with nodejs or similar instead of running in a web page. Such configuration, especially idiomatically, is beyond the knowledge of this author.
* AuthToken and EntryName are hardcoded. Replace these with your own values!
* Unlike the Go and Python examples, the JS example is written in a more loosely-typed manner. The Javascript JSON parser is able to convert directly to JS objects, which while not typesafe, are at least nicer to use than Go or Python dictionaries. Additionally, we didn't want to wire up Typescript machinery, and without that then JS typing is less useful. As a result, there is no native JS definition of the websocket message types in this example.
* Error handling is pretty weak, and this example doesn't display as many informational and progress messages as the others.
