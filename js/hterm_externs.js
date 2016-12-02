/**
@fileoverview Externs for hterm
@externs
*/

/** @const */
var hterm = {};
/** @const */
hterm.Terminal = {};

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
