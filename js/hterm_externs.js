/**
@fileoverview Externs for hterm
@externs
*/

/** @const */
var lib = {};

/**
 * Namespace for implementations of persistent, possibly cloud-backed
 * storage.
 * @const
 */
lib.Storage = {};

/**
@interface
*/
lib.Storage.Store = function() {};
// /**
//  * Register a function to observe storage changes.
//  *
//  * @param {function()} callback The function to invoke when the storage
//  *     changes.
//  */
// lib.Storage.Store.prototype.addObserver = function(callback) {};

// /**
//  * Unregister a change observer.
//  *
//  * @param {function()} callback A previously registered callback.
//  */
// lib.Storage.Store.prototype.removeObserver = function(callback) {};

// /**
//  * Delete everything in this storage.
//  *
//  * @param {function()=} opt_callback The function to invoke when the delete
//  *     has completed.
//  */
// lib.Storage.Store.prototype.clear = function(opt_callback) {};

// /**
//  * Return the current value of a storage item.
//  *
//  * @param {string} key The key to look up.
//  * @param {function(*)} callback The function to invoke when the value has
//  *     been retrieved.
//  */
// lib.Storage.Store.prototype.getItem = function(key, callback) {};

// /**
//  * Fetch the values of multiple storage items.
//  *
//  * @param {!Array<string>} keys The keys to look up.
//  * @param {function(!Object)} callback The function to invoke when the values have
//  *     been retrieved.
//  */
// lib.Storage.Store.prototype.getItems = function(keys, callback) {};

// *
//  * Set a value in storage.
//  *
//  * @param {string} key The key for the value to be stored.
//  * @param {*} value The value to be stored.  Anything that can be serialized
//  *     with JSON is acceptable.
//  * @param {function()=} opt_callback Optional function to invoke when the
//  *     set is complete.  You don't have to wait for the set to complete in order
//  *     to read the value, since the local cache is updated synchronously.
 
// lib.Storage.Store.prototype.setItem = function(key, value, opt_callback) {};

// /**
//  * Set multiple values in storage.
//  *
//  * @param {!Object} obj A map of key/values to set in storage.
//  * @param {function()=} opt_callback Optional function to invoke when the
//  *     set is complete.  You don't have to wait for the set to complete in order
//  *     to read the value, since the local cache is updated synchronously.
//  */
// lib.Storage.Store.prototype.setItems = function(obj, opt_callback) {};

// /**
//  * Remove an item from storage.
//  *
//  * @param {string} key The key to be removed.
//  * @param {function()=} opt_callback Optional function to invoke when the
//  *     remove is complete.  You don't have to wait for the set to complete in
//  *     order to read the value, since the local cache is updated synchronously.
//  */
// lib.Storage.Store.prototype.removeItem = function(key, opt_callback) {};

// /**
//  * Remove multiple items from storage.
//  *
//  * @param {!Array<string>} ary The keys to be removed.
//  * @param {function()=} opt_callback Optional function to invoke when the
//  *     remove is complete.  You don't have to wait for the set to complete in
//  *     order to read the value, since the local cache is updated synchronously.œ
//  */
// lib.Storage.Store.prototype.removeItems = function(ary, opt_callback) {};

/**
@constructor
@implements {lib.Storage.Store}
*/
lib.Storage.Memory = function() {};

/** @const */
var hterm = {};

/** @type {!lib.Storage.Store} */
hterm.defaultStorage;

/**
@constructor
@struct
@param {string=} opt_profileId
*/
hterm.Terminal = function(opt_profileId) {};

/**
 * Clients should override this to be notified when the terminal is ready
 * for use.
 *
 * The terminal initialization is asynchronous, and shouldn't be used before
 * this method is called.
 */
hterm.Terminal.prototype.onTerminalReady = function() {};

/** @type {!hterm.Terminal.IO} */
hterm.Terminal.prototype.io;

/**
 * Set the cursor position.
 *
 * The cursor row is relative to the scroll region if the terminal has
 * 'origin mode' enabled, or relative to the addressable screen otherwise.
 *
 * @param {number} row The new zero-based cursor row.
 * @param {number} column The new zero-based cursor column.
 */
hterm.Terminal.prototype.setCursorPosition = function(row, column) {};

/**
 * Set the cursor-visible mode bit.
 *
 * If cursor-visible is on, the cursor will be visible.  Otherwise it will not.
 *
 * Defaults to on.
 *
 * @param {boolean} state True to set cursor-visible mode, false to unset.
 */
hterm.Terminal.prototype.setCursorVisible = function(state) {};

/**
 * Take over the given DIV for use as the terminal display.
 *
 * @param {!HTMLDivElement} div The div to use as the terminal display.
 */
hterm.Terminal.prototype.decorate = function(div) {};

/**
 * Install the keyboard handler for this terminal.
 *
 * This will prevent the browser from seeing any keystrokes sent to the
 * terminal.
 */
hterm.Terminal.prototype.installKeyboard = function() {};

/**
@constructor
@struct
@param {hterm.Terminal} terminal
*/
hterm.Terminal.IO = function(terminal) {};
/**
 * Write a UTF-16 JavaScript string to the terminal.
 *
 * @param {string} string The string to print.
 */
hterm.Terminal.IO.prototype.writeUTF16 = function(string) {};

/**
 * Create a new hterm.Terminal.IO instance and make it active on the Terminal
 * object associated with this instance.
 *
 * This is used to pass control of the terminal IO off to a subcommand.  The
 * IO.pop() method can be used to restore control when the subcommand completes.
 * @return {!hterm.Terminal.IO}
 */
hterm.Terminal.IO.prototype.push = function() {}

/**
 * Called when a terminal keystroke is detected.
 *
 * Clients should override this to receive notification of keystrokes.
 *
 * The keystroke data will be encoded according to the 'send-encoding'
 * preference.
 *
 * @param {string} string The VT key sequence.
 */
 hterm.Terminal.IO.prototype.onVTKeystroke = function(string) {}

 /**
 * Called when data needs to be sent to the current command.
 *
 * Clients should override this to receive notification of pending data.
 *
 * @param {string} string The data to send.
 */
hterm.Terminal.IO.prototype.sendString = function(string) {}

/**
 * Called when terminal size is changed.
 *
 * Clients should override this to receive notification of resize.
 *
 * @param {number} width terminal width.
 * @param {number} height terminal height.
 */
hterm.Terminal.IO.prototype.onTerminalResize = function(width, height) {}
