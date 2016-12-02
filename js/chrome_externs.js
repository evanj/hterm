// https://github.com/google/closure-compiler/blob/master/contrib/externs/chrome_extensions.js

/**
 * @fileoverview Definitions for the Chromium extensions API.
 * @externs
 */

/**
 * @const
 * @see https://developer.chrome.com/extensions/i18n.html
 */
chrome.i18n = {};

/**
 * @param {function(Array<string>): void} callback The callback function which
 *     accepts an array of the accept languages of the browser, such as
 *     'en-US','en','zh-CN'.
 * @return {undefined}
 */
chrome.i18n.getAcceptLanguages = function(callback) {};

/**
 * @param {string} path A path to a resource within an extension expressed
 *     relative to it's install directory.
 * @return {string} The fully-qualified URL to the resource.
 */
chrome.runtime.getURL = function(path) {};
