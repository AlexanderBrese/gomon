# gomon [![build and test](https://github.com/AlexanderBrese/gomon/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/AlexanderBrese/gomon/actions/workflows/go.yml) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/217fd7aa6f224b8d8094c833d4c5b07a)](https://www.codacy.com/gh/AlexanderBrese/gomon/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=AlexanderBrese/gomon&amp;utm_campaign=Badge_Grade) [![Go Report Card](https://goreportcard.com/badge/github.com/AlexanderBrese/gomon)](https://goreportcard.com/report/github.com/AlexanderBrese/gomon) [![codecov](https://codecov.io/gh/AlexanderBrese/gomon/branch/master/graph/badge.svg)](https://codecov.io/gh/AlexanderBrese/gomon)

![how-does-it-work](https://github.com/AlexanderBrese/gomon/blob/master/docs/gomon.gif)

# Why is it necessary?

Because it also refreshes the `browser` upon change.

# How does it work?

Plain: Watches for code changes, reloads the specified binary e.g. a webserver and calls the client through a websocket to reload.

# Usage
## setup the client

Include the script below to your client e.g. a static main.js & change `YOUR_PORT` with a free port of your choice. The default port is `3000`.<br>
```js
function tryConnectToReload(address) {
  var conn;
  // This is a statically defined port on which the app is hosting the reload service.
  conn = new WebSocket("ws://localhost:3000/sync");

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

## install gomon

```
go get -u github.com/AlexanderBrese/gomon
```

## run gomon 

If you want to configure the reload behavior or set change paths then just provide a `configuration` to the process.

```
gomon [-c PATH_TO_YOUR_CONFIG]
```

## configure gomon

`Default` configuration:
```toml
# The port used for the browser syncing server
port = 3000
[build]
# What should the build be named?
build_name = "main"
# How should the build be done?
build_command = "go build -o"
# How should the build be run?
execution_command = ""
# What should we built from?
relative_source_dir = ""
# Where should the build be stored?
relative_build_dir = "tmp/build"
[filter]
# Watch these extensions for changes
include_exts = ["go", "tpl", "tmpl", "html", "css", "js", "env", "yaml"]
# Watch these directories for changes
include_relative_dirs = []
# Ignore these files
exclude_relative_files = []
# Ignore these directories
exclude_relative_dirs = ["assets", "tmp", "vendor", "node_modules", "build"]
[log]
# What should the Build log be named?
build_log_name = "gomon.log"
# Where should the Build log be stored?
relative_build_log_dir = "tmp"
# Should the Main log be enabled?
main = true
# Should the Detection log be enabled?
detection = false
# Should the Build log be enabled?
build = true
# Should the Run log be enabled?
run = false
# Should the Sync log be enabled?
sync = false
# Should the App log be enabled?
app = true
# Should a timestamp be appended to the log?
time = true
[color]
# The Main log color
main = "red"
# The Detection log color
detection = "magenta"
# The Build log color
build = "yellow"
# The Run log color
run = "green"
# The Sync log color
sync = "cyan"
# The App log color
app = "blue"
```

# What features is it going to provide?

The goals for version `1.0.0` are:
-  ~~Linux/MacOS Support (currently windows only)~~
-  ~~Colorful log messages~~
- Customize binary execution with environmental flags
