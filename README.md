# Why is it necessary?

It also reloads the `browser` upon change.

# How does it work?

Plain: Watches for changes on the server, reloads the server and reloads the browser.<br><br>
![how-does-it-work](https://github.com/AlexanderBrese/go-server-browser-reload/blob/main/go-server-browser-reload.png)

# Usage

## setup the client

Include the script below to your client e.g. static main.js & change `YOUR_PORT` with a free port of your choice. The default port is `3000`.<br>
```js
function tryConnectToReload(address) {
  var conn;
  // This is a statically defined port on which the app is hosting the reload service.
  conn = new WebSocket("ws://localhost:YOUR_PORT");

  conn.onclose = function(evt) {
    // The reload endpoint hasn't been started yet, we are retrying in 2 seconds.
    setTimeout(() => tryConnectToReload(), 2000);
  };

  conn.onmessage = function(evt) {
    console.log("Refresh received!");

    // the page will refresh every time a message is received.
    location.reload()
  };
}

try {
  if (window["WebSocket"]) {
    tryConnectToReload();
  } else {
    console.log("Your browser does not support WebSocket, cannot connect to the reload service.");
  }
} catch (ex) {
  console.log('Exception during connecting to reload:', ex);
}
```

## install go-server-browser-reload

```
go get -u github.com/AlexanderBrese/go-server-browser-reload
```

## run go-server-browser-reload 

If you want to configure the reload behavior or set change paths then just provide a `configuration` to the process.

```
go-server-browser-reload -c PATH_TO_YOUR_CONFIG
```

## configure go-server-browser-reload

Default configuration:
```toml
# Use go to build the binary as usual
cmd = "go build -o ./tmp/main ."
# The port used for the browser syncing server
port = 3000
# Watch these extensions
include_ext = ["go", "tpl", "tmpl", "html", "css", "js", "env", "yaml"]
# Watch these directories
include_dir = []
# Ignore these files
exclude_file = []
# Ignore these directories
exclude_dir = ["assets", "tmp", "vendor", "node_modules", "build"]
# Buffer changes before rebuilding for a certain amount of time (ms)
delay = 1000
```

# What features is it going to provide?

The goals for version `1.0.0` are:
- Linux/MacOS Support (currently windows only)
- Colorful log messages
- Customize binary execution with environmental flags
