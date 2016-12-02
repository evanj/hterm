/**
@fileoverview Externs for jest tests https://facebook.github.io/jest/docs/api.html
@externs
*/

/**
@param {string} name
@param {function()} testFunc
*/
function test(name, testFunc) {};

// Much of this stolen from:
// From https://github.com/google/closure-compiler/blob/master/contrib/externs/jasmine-2.0.js

/**
 * @param {*} expectedValue
 * @return {jasmine.Matchers} matcher
 */
function expect(expectedValue) {}
