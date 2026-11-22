// Package changes counts the changes widgets make to an interface's
// picture as they update, without a caller asking: a caret blinking, a
// tooltip showing or going, a GIF's frame, a window's transition, a list
// gliding on after a drag. ebitenui.UI.Changed reports them, so a game
// that draws only what changed knows to draw.
package changes

var count uint64

// Mark records a change of the picture.
func Mark() { count++ }

// Count returns how many changes were marked so far.
func Count() uint64 { return count }
