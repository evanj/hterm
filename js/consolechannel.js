"use strict";

/** @const */
var consolechannel = {};

/** @typedef {{data: (string|undefined), columns: (number|undefined), rows: (number|undefined)}} */
consolechannel.PartialRequest;
/** @typedef {{data: string}} */
consolechannel.ResponseUnion;

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
@param {string} requestSerialized
@param {function(string)} onSuccess
@param {function()} onError
*/
consolechannel.Environment.prototype.post = function(url, requestSerialized, onSuccess, onError) {};

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
consolechannel.BrowserEnvironment.prototype.post = function(url, requestSerialized, onSuccess, onError) {
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
    onSuccess(request.responseText);
  }
  request.onreadystatechange = onReadyStateChange;
  request.open("POST", url, true);
  // allow cookies for when we eventually get there
  request.withCredentials = true;

  request.send(requestSerialized);
};


/**
@constructor
@struct
@param {!consolechannel.Environment} env
@param {string} url
@param {!Object<string, string>} extra
*/
consolechannel.Channel = function(env, url, extra) {
  /** @type {!consolechannel.Environment} */
  this.env_ = env;
  /** @type {string} */
  this.url_ = url;
  /** @type {!Object<string, string>} */
  this.extra_ = extra;

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
@private
@param {string} path
@param {!consolechannel.PartialRequest} struct
@param {function(!consolechannel.ResponseUnion)} onSuccess
@param {function()} onError
*/
consolechannel.Channel.prototype.postStruct_ = function(path, struct, onSuccess, onError) {
  var jsonDict = {};
  // common
  jsonDict["session_id"] = this.session_id_
  jsonDict["extra"] = this.extra_;
  // write
  jsonDict["data"] = struct.data;
  // setSize
  jsonDict["columns"] = struct.columns;
  jsonDict["rows"] = struct.rows;
  var serialized = JSON.stringify(jsonDict)

  /** @param {string} responseSerialized */
  var rawOnSucccess = function(responseSerialized) {
    // convert a raw JSON message to a Closure compiler friendly struct
    var raw = JSON.parse(responseSerialized);
    if (typeof raw !== "object") {
      console.error("unexpected type from server response: " + typeof raw);
      onError();
      return
    }
    var struct = {data: raw["data"]};
    onSuccess(struct);
  }

  this.env_.post(this.url_ + path, serialized, rawOnSucccess, onError);
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

  var self = this;
  var onSuccess = function() {
    self.onWriteComplete_(true);
  };
  var onError = function() {
    self.onWriteComplete_(false);
  };

  var request = {data: data};
  this.postStruct_("write", request, onSuccess, onError);
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

/**
Sets the terminal size to rows, cols.
@param {number} columns
@param {number} rows
*/
consolechannel.Channel.prototype.setSize = function(columns, rows) {
  function onError() {
    console.error("setSize onError");
  }

  function onSuccess() {
    console.log("setSize success");
  }

  var request = {
    columns: columns,
    rows: rows
  };
  this.postStruct_("setSize", request, onSuccess, onError);
};

/**
Start reading data that should be written to io.
@param {!hterm.Terminal.IO} io
*/
consolechannel.Channel.prototype.startRead = function(io) {
  // if (typeof io === "undefined" || typeof io.writeUTF16 === undefined) {
  //   throw "Channel.startRead: io.writeUTF16 must be defined";
  // }
  var self = this;

  function onError() {
    console.error("read onError");
  }

  /** @param {!consolechannel.ResponseUnion} struct */
  function onSuccess(struct) {
    console.log("read success; length:", struct.data.length);
    io.writeUTF16(struct.data);
    // read again!
    self.startRead(io);
  }

  this.postStruct_("read", {}, onSuccess, onError)
};

// export in order to be required by node
if (typeof module !== "undefined" && module.exports) {
  // can't just assign consolechannel due to Closure namespace aliasing rules
  module.exports = {
    Channel: consolechannel.Channel,
  };
}
