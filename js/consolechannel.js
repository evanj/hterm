"use strict";

/** @const */
var consolechannel = {};

/** @typedef {{session_id: string, data: string, extra: !Object}} */
consolechannel.WriteRequest;
/** @typedef {{data: string}} */
consolechannel.ReadResponse;

/** @record */
consolechannel.Environment = function() {};
/**
@param {!ArrayBufferView} typedArray
@return {!ArrayBufferView}
@throws {Error}
*/
consolechannel.Environment.prototype.getRandomValues = function(typedArray) {};
/**
@param {string} url
@param {!Object} struct
@param {function()} onSuccess
@param {function()} onError
*/
consolechannel.Environment.prototype.post = function(url, struct, onSuccess, onError) {};

/**
@constructor
@implements {consolechannel.Environment}
*/
consolechannel.BrowserEnvironment = function() {};
/** @override */
consolechannel.BrowserEnvironment.prototype.getRandomValues = function(typedArray) {
  return window.crypto.getRandomValues(typedArray);
};
/** @override */
consolechannel.BrowserEnvironment.prototype.post = function(url, struct, onSuccess, onError) {
  var request = new XMLHttpRequest();

  function onReadyStateChange() {
    if (request.readyState != XMLHttpRequest.DONE) {
      return;
    }
    if (request.status != 200) {
      onError();
      return;
    }
    console.log("write success");
    onSuccess();
  }
  request.onreadystatechange = onReadyStateChange;
  request.open("POST", url, true);
  // allow cookies for when we eventually get there
  request.withCredentials = true;

  var body = JSON.stringify(struct);
  request.send(body);
};


/**
@constructor
@struct
@param {!consolechannel.Environment} env
@param {string} url
@param {!Object} extraParams
*/
consolechannel.Channel = function(env, url, extraParams) {
  /** @type {!consolechannel.Environment} */
  this.env_ = env;
  /** @type {string} */
  this.url_ = url;
  /** @type {!Object} */
  this.extraParams_ = extraParams;

  // generate a 32-byte unique random id as a base64-encoded string
  var array = new Uint8Array(32);
  this.env_.getRandomValues(array);
  var s = "";
  for(var i = 0; i < array.byteLength; i++) {
    s += String.fromCharCode(array[i]);
  }
  /** @type {string} */
  this.session_id_ = btoa(s);

  /** @type {boolean} */
  this.writePending_ = false;
  /** @type {string} */
  this.writeBuffer_ = "";
};

/**
Write data to the terminal program/server. Stolen from
nassh.Stream.GoogleRelay.prototype.asyncOpen_.

@param {string} data
*/
consolechannel.Channel.prototype.write = function(data) {
  if (this.writePending_) {
    this.writeBuffer_ += data;
  } else {
    console.log("write calling doSend");
    this.doSend_(data);
  }
};

/**
@param {string} data
*/
consolechannel.Channel.prototype.doSend_ = function(data) {
  console.log("doSend");
  if (this.writePending_) {
    throw "writePending_ must be false";
  }
  if (this.writeBuffer_.length != 0) {
    throw "writeBuffer_ must be empty";
  }
  if (data.length == 0) {
    throw "data must not be empty";
  }
  this.writePending_ = true;
  var request = {session_id: this.session_id_, data: data, extra: this.extraParams_};

  var self = this;
  var onSuccess = function() {
    self.onWriteComplete_(true);
  };
  var onError = function() {
    self.onWriteComplete_(false);
  };
  this.env_.post(this.url_, request, onSuccess, onError);
};

/**
@private
@param {boolean} success
*/
consolechannel.Channel.prototype.onWriteComplete_ = function(success) {
  // TODO: distinguish between success and failure
  if (!this.writePending_) {
    throw "bug: write callback without writePending_";
  }
  if (!success) {
    console.log("write error occurred TODO: handle?");
  }

  this.writePending_ = false;
  if (this.writeBuffer_.length > 0) {
    var data = this.writeBuffer_;
    this.writeBuffer_ = "";
    this.doSend_(data);
  }
};

// /**
// Sets the terminal size to rows, cols.
// @param {number} columns
// @param {number} rows
// */
// consolechannel.Channel.prototype.setSize = function(columns, rows) {
//   var sizeRequest = new XMLHttpRequest();

//   function onError() {
//     console.error("setSize onError");
//   }

//   function onReady() {
//     if (sizeRequest.readyState != XMLHttpRequest.DONE) {
//       return;
//     }
//     if (sizeRequest.status != 200) {
//       return onError();
//     }
//     console.log("setSize success");
//   }
//   sizeRequest.open("POST", this.url_ + "setSize", true);
//   sizeRequest.withCredentials = true;  // allow cookies for when we eventually get there?
//   sizeRequest.onabort = sizeRequest.ontimeout = sizeRequest.onerror = onError;
//   sizeRequest.onloadend = onReady;

//   var request = JSON.stringify({session_id: this.session_id_, columns: columns, rows: rows});
//   sizeRequest.send(request);
// };

// /**
// Start reading data that should be written to io.
// @param {!hterm.Terminal.IO} io
// */
// consolechannel.Channel.prototype.startRead = function(io) {
//   // if (typeof io === "undefined" || typeof io.writeUTF16 === undefined) {
//   //   throw "Channel.startRead: io.writeUTF16 must be defined";
//   // }
//   var self = this;
//   var readRequest = new XMLHttpRequest();

//   function onError() {
//     console.error("read onError");
//   }

//   function onReady() {
//     if (readRequest.readyState != XMLHttpRequest.DONE) {
//       return;
//     }
//     if (readRequest.status != 200) {
//       return onError();
//     }

//     var response = /** @type {consolechannel.ReadResponse} */ (JSON.parse(readRequest.responseText));
//     console.log("read success; length:", response.data.length);
//     io.writeUTF16(response.data);
//     // read again!
//     self.startRead(io);
//   }
//   readRequest.open("POST", this.url_ + "read", true);
//   readRequest.withCredentials = true;  // allow cookies for when we eventually get there?
//   readRequest.onabort = readRequest.ontimeout = readRequest.onerror = onError;
//   readRequest.onloadend = onReady;
//   var request = JSON.stringify({
//     session_id: this.session_id_,
//     query: this.query_
//   });
//   readRequest.send(request);
// };

// export in order to be required by node
if (typeof module !== "undefined" && module.exports) {
  // can't just assign consolechannel due to Closure namespace aliasing rules
  module.exports = {
    Channel: consolechannel.Channel,
  };
}
