"use strict";

/** @suppress{duplicate} */
var consolechannel = consolechannel || require("../js/consolechannel.js");

/**
@constructor
@struct
@param {string} url
@param {string} body
@param {function(string)} onSuccess
@param {function()} onError
*/
var PostArgs = function(url, body, onSuccess, onError) {
  this.url = url;
  this.struct = JSON.parse(body);
  /** @type {function(string)} */
  this.onSuccess = onSuccess;
  /** @type {function()} */
  this.onError = onError;
};

/**
@constructor
@implements {consolechannel.Environment}
*/
var FakeEnvironment = function() {
  /** @type {!Array<!PostArgs>} */
  this.posts = [];
};
/** @override */
FakeEnvironment.prototype.getRandomValues = function(typedArray) {
  typedArray[0] = 42;
  return typedArray;
};
/** @override */
FakeEnvironment.prototype.post = function(url, body, onSuccess, onError) {
  this.posts.push(new PostArgs(url, body, onSuccess, onError));
};

it("consolechannel multiple keystrokes are batched", () => {
  var env = new FakeEnvironment();
  var extraParams = {param: 'something'};
  var channel = new consolechannel.Channel(env, "http://localhost:8080", extraParams);

  // write: sends the post
  channel.write("helloworld");
  // must be accessed with [] to avoid closure compiler renaming
  expect(env.posts[0].struct["session_id"]).toBe("KgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=");
  expect(env.posts[0].struct["data"]).toBe("helloworld");
  expect(env.posts[0].struct["extra"]).toEqual(extraParams);

  // more writes: batched
  channel.write("one");
  expect(env.posts.length).toBe(1);
  channel.write("two");
  expect(env.posts.length).toBe(1);

  // on success: the batch is flushed
  env.posts[0].onSuccess('{}');
  expect(env.posts[1].struct["data"]).toBe("onetwo");
});