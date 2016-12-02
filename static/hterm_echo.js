'use strict';

var ejhack = {};

ejhack.Relay = function(url) {
  if (!url.endsWith("/")) {
    url += "/"
  }
  this.url_ = url;

  // generate a 32-byte unique random id as a base64-encoded string
  var array = new Uint8Array(32);
  window.crypto.getRandomValues(array);
  var s = '';
  for(var i = 0; i < array.byteLength; i++) {
    s += String.fromCharCode(array[i]);
  }
  this.session_id_ = btoa(s);

  // HACK: Pass any query parameters to the server
  var query = '';
  var questionIndex = window.location.href.indexOf('?');
  if (questionIndex >= 0) {
    query = window.location.href.substring(questionIndex + 1);
  }
  this.query_ = query;
}

/**
 * Write data to the terminal program/server. Stolen from
 * nassh.Stream.GoogleRelay.prototype.asyncOpen_.
 */
ejhack.Relay.prototype.write = function(data, onComplete) {
  var writeRequest = new XMLHttpRequest();

  function onError() {
    console.error('write onError');
    onComplete(false);
  }

  function onReady() {
    if (writeRequest.readyState != XMLHttpRequest.DONE) {
      return;
    }
    if (writeRequest.status != 200) {
      return onError();
    }
    console.log('write success');
    onComplete(true);
  }
  writeRequest.open('POST', this.url_ + "write", true);
  writeRequest.withCredentials = true;  // allow cookies for when we eventually get there?
  writeRequest.onabort = writeRequest.ontimeout = writeRequest.onerror = onError;
  writeRequest.onloadend = onReady;

  var request = JSON.stringify({session_id: this.session_id_, data: data});
  writeRequest.send(request);
};

/**
 * Sets the terminal size to  rows, cols.
 */
ejhack.Relay.prototype.setSize = function(columns, rows) {
  var sizeRequest = new XMLHttpRequest();

  function onError() {
    console.error('setSize onError');
  }

  function onReady() {
    if (sizeRequest.readyState != XMLHttpRequest.DONE) {
      return;
    }
    if (sizeRequest.status != 200) {
      return onError();
    }
    console.log('setSize success');
  }
  sizeRequest.open('POST', this.url_ + "setSize", true);
  sizeRequest.withCredentials = true;  // allow cookies for when we eventually get there?
  sizeRequest.onabort = sizeRequest.ontimeout = sizeRequest.onerror = onError;
  sizeRequest.onloadend = onReady;

  var request = JSON.stringify({session_id: this.session_id_, columns: columns, rows: rows});
  sizeRequest.send(request);
}

/**
 * Start reading data that should be written to io.
 */
ejhack.Relay.prototype.startRead = function(io) {
  if (typeof io === 'undefined' || typeof io.writeUTF16 === undefined) {
    throw "Relay.startRead: io.writeUTF16 must be defined";
  }
  var self = this;
  var readRequest = new XMLHttpRequest();

  function onError() {
    console.error('read onError');
  }

  function onReady() {
    if (readRequest.readyState != XMLHttpRequest.DONE) {
      return;
    }
    if (readRequest.status != 200) {
      return onError();
    }

    var response = JSON.parse(readRequest.responseText);
    console.log('read success; length:', response.data.length);
    io.writeUTF16(response.data);
    // read again!
    self.startRead(io);
  }
  readRequest.open('POST', this.url_ + "read", true);
  readRequest.withCredentials = true;  // allow cookies for when we eventually get there?
  readRequest.onabort = readRequest.ontimeout = readRequest.onerror = onError;
  readRequest.onloadend = onReady;
  var request = JSON.stringify({
    session_id: this.session_id_,
    query: this.query_
  });
  readRequest.send(request);
};
