/**
@fileoverview Externs for hterm
@externs
*/

/** const */
var lib = {};
/** @const */
lib.Storage = {};

/**
@constructor
*/
lib.Storage.Memory = function() {};

/** @const */
var hterm = {};
/**
@constructor
@param {string=} opt_profileId
*/
hterm.Terminal = function(opt_profileId) {};

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
@param {hterm.Terminal} terminal
*/
hterm.Terminal.IO = function(terminal) {};
/**
 * Write a UTF-16 JavaScript string to the terminal.
 *
 * @param {string} string The string to print.
 */
hterm.Terminal.IO.prototype.writeUTF16 = function(string) {};
