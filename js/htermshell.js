/** @const */
var terminalDivId = "terminal";

/** called on document load */
var loaded = function() {
  var terminalElement = document.getElementById(terminalDivId);
  if (terminalElement == null) {
    throw new Error("Terminal element id " + terminalDivId + " does not exist");
  }
  if (terminalElement.nodeName != "DIV") {
    throw new Error("Element with id " + terminalDivId + " must be a div; not " + terminalElement.nodeName);
  }

  // from https://chromium.googlesource.com/apps/libapps/+/master/hterm/doc/embed.md
  hterm.defaultStorage = new lib.Storage.Memory();

  /** @type{!hterm.Terminal} required to detect errors like terminal.foo */
  var terminal = new hterm.Terminal("default");

  terminal.onTerminalReady = function() {
    // Create a new terminal IO object and give it the foreground.
    var io = terminal.io.push();
    var channel = new consolechannel.Channel(new consolechannel.BrowserEnvironment(), "/", {});

    function send(str) {
      console.log("key/send from terminal:", str);
      channel.write(str, function(success) { console.log("send onComplete", success); });
    }
    io.onVTKeystroke = send;
    io.sendString = send;

    io.onTerminalResize = function(columns, rows) {
      // React to size changes here.
      // Secure Shell pokes at NaCl, which eventually results in
      // some ioctls on the host.
      console.log("onTerminalResize", columns, rows);
      channel.setSize(columns, rows);
    };

    console.log("hello terminal.onTerminalReady()");
    terminal.setCursorPosition(0, 0);
    terminal.setCursorVisible(true);
    terminal.installKeyboard();

    channel.startRead(io);
  };

  console.log("decorating", terminalElement);
  terminal.decorate(/** @type{!HTMLDivElement} */ (terminalElement));
};

document.addEventListener('DOMContentLoaded', loaded);
