package main

import (
   "fmt"
   glfw "github.com/go-gl/glfw3"
   gtk "github.com/GlenKelley/go-glutil"
)

type Receiver struct {
}

func main() {
   fmt.Println("Start")
   receiver := &Receiver{}
   gtk.CreateWindow(640, 480, "carsim", false, receiver)
}

func (r *Receiver) Init(window *glfw.Window) {
}

func (r *Receiver) Draw(window *glfw.Window) {
}

func (r *Receiver) Reshape(window *glfw.Window, width, height int) {
}

func (r *Receiver) MouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
}

func (r *Receiver) MouseMove(window *glfw.Window, xpos float64, ypos float64) {
}

func (r *Receiver) KeyPress(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey) {
}

func (r *Receiver) Scroll(window *glfw.Window, xoff float64, yoff float64) {
}

func (r *Receiver) Simulate(time gtk.GameTime) {
}

func (r *Receiver) OnClose(window *glfw.Window) {
   
}

func (r *Receiver) IsIdle() bool {
   return true
}

func (r *Receiver) NeedsRender() bool {
   return false
}