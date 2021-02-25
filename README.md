# why is it necessary

# how does it work

Plain: Watches for changes on the server, reloads the server and reloads the browser.<br><br>
![how-does-it-work](https://github.com/AlexanderBrese/go-server-browser-reload/blob/main/go-server-browser-reload.png)

# how is it used

## setup the client

Include the script below to your client e.g. static main.js<br>
Change YOUR_PORT with a free port of your choice e.g. 3000.<br>
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
