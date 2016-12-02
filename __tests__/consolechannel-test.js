"use strict";

/** @suppress{duplicate} */
var consolechannel = consolechannel || require("../js/consolechannel.js");

/**
@constructor
@implements {consolechannel.Environment}
*/
var FakeEnvironment = function() {
  this.posts = [];
};
/** @override */
FakeEnvironment.prototype.getRandomValues = function(typedArray) {
  typedArray[0] = 42;
  return typedArray;
};
/** @override */
FakeEnvironment.prototype.post = function(url, struct, onSuccess, onError) {
  this.posts.push([url, struct, onSuccess, onError]);
};

it("consolechannel multiple keystrokes are batched", () => {
  var env = new FakeEnvironment();
  var extraParams = {param: 'something'};
  var channel = new consolechannel.Channel(env, "http://localhost:8080", extraParams);

  // write: sends the post
  channel.write("helloworld");
  expect(env.posts[0][1].session_id).toBe("KgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=");
  expect(env.posts[0][1].data).toBe("helloworld");
  expect(env.posts[0][1].extra).toBe(extraParams);

  // more writes: batched
  channel.write("one");
  expect(env.posts.length).toBe(1);
  channel.write("two");
  expect(env.posts.length).toBe(1);

  // on success: the batch is flushed
  env.posts[0][2]();
  expect(env.posts[1][1].data).toBe("onetwo");

  
});
