package main

import "C"

// // These functions will be called from the Objective-C code
// //

//export goFullscreenWillEnter
func goFullscreenWillEnter() {
	println("[Go] Fullscreen will enter")
	if app != nil {
		app.EmitEvent("fullscreen-enter")
	}
}

//export goFullscreenDidExit
func goFullscreenDidExit() {
	println("[Go] Fullscreen did exit")
	if app != nil {
		app.EmitEvent("fullscreen-exit")
	}
}
