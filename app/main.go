package main

import (
   "fmt"
   glfw "github.com/go-gl/glfw3"
   gtk "github.com/GlenKelley/go-glutil"
   gl "github.com/GlenKelley/go-gl/gl32"
   glm "github.com/Jragonmiris/mathgl"
   sim "github.com/GlenKelley/carsim"
)

func main() {
   fmt.Println("Start")
   receiver := &Receiver{}
   gtk.CreateWindow(640, 480, "carsim", false, receiver)
}

type Receiver struct {
   Window   *glfw.Window
   Data     DataBindings
   Shaders  gtk.ShaderLibrary
   SceneLoc SceneBindings
   Car      sim.Car
   Controls  gtk.ControlBindings
   UIState  UIState
}

type DataBindings struct {
   Vao  gl.VertexArrayObject
   Scene *gtk.Model
   Car *gtk.Model
   Projection glm.Mat4d
   Cameraview glm.Mat4d
}

type SceneBindings struct {
   Projection gl.UniformLocation `gl:"projection"`
   Cameraview gl.UniformLocation `gl:"cameraview"`
   Worldview  gl.UniformLocation `gl:"worldview"`
   Position gl.AttributeLocation `gl:"position"`
}

type UIState struct {
   IsRotating bool
   Theta      float64
   Controls   sim.Controls
}

const (
   ProgramScene = "scene"
   Fov = 60
   Near = 0.1
   Far = 100
)

func (r *Receiver) Init(window *glfw.Window) {
   r.Window = window
   gtk.Bind(&r.Data)
   r.Shaders = gtk.NewShaderLibrary()
   r.Shaders.LoadProgram(ProgramScene, "scene.v.glsl", "scene.f.glsl")
   r.Shaders.BindProgramLocations(ProgramScene, &r.SceneLoc)
   gtk.PanicOnError()
   
   r.Data.Projection = glm.Ident4d()
   r.Data.Cameraview = glm.Ident4d()
   
   r.Data.Scene = gtk.EmptyModel("root")
   model, err := gtk.LoadSceneAsModel("drivetrain.dae")
   if err != nil { panic(err) }
   r.Data.Scene.AddChild(model)
   r.Data.Car = model.Children[0]
   r.Data.Scene.AddGeometry(gtk.Grid(10))
   
   r.Car = sim.NewCar()
   r.ResetKeyBindingDefaults()
}

func (r *Receiver) ResetKeyBindingDefaults() {
   c := &r.Controls
   c.ResetBindings()
   c.BindKeyPress(glfw.KeyW, r.PushFuelPedal, r.ReleaseFuelPedal)
   c.BindKeyPress(glfw.KeyS, r.PushReversePedal, r.ReleaseReversePedal)
   c.BindKeyPress(glfw.KeySpace, r.PushBreakPedal, r.ReleaseBreakPedal)
   c.BindKeyPress(glfw.KeyR, r.ToggleRotate, nil)
   c.BindKeyPress(glfw.KeyEscape, r.Quit, nil)
}

func (r *Receiver) Draw(window *glfw.Window) {
   bg := gtk.SoftBlack
   gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
   gl.Enable(gl.DEPTH_TEST)
   gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
   
   r.Shaders.UseProgram(ProgramScene)
   gl.UniformMatrix4fv(r.SceneLoc.Projection, 1, gl.FALSE, gtk.MatArray(r.Data.Projection))
   gl.UniformMatrix4fv(r.SceneLoc.Cameraview, 1, gl.FALSE, gtk.MatArray(r.Data.Cameraview))
   gtk.PanicOnError()
   gtk.DrawModel(glm.Ident4d(), r.Data.Scene, r.SceneLoc.Worldview, r.SceneLoc.Position, r.Data.Vao)
}

func (r *Receiver) Reshape(window *glfw.Window, width, height int) {
   aspectRatio := gtk.WindowAspectRatio(window)
   r.Data.Projection = glm.Perspectived(Fov, aspectRatio, Near, Far)
}

func (r *Receiver) MouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
}

func (r *Receiver) MouseMove(window *glfw.Window, xpos float64, ypos float64) {
}

func (r *Receiver) KeyPress(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey) {
   r.Controls.DoKeyAction(k, action)
}

func (r *Receiver) Scroll(window *glfw.Window, xoff float64, yoff float64) {
}

func (r *Receiver) Simulate(time gtk.GameTime) {
   dt := time.Delta.Seconds()
   r.Car.Simulate(r.UIState.Controls, dt)
   r.SetCarTransform()
   
   //camera
   if r.UIState.IsRotating {
      period := 5.0
      dtheta := dt * 360 / period
      r.UIState.Theta += dtheta
   }
   
   p := glm.Vec4d{0,5,12,2}
   rotation := glm.HomogRotate3DYd(r.UIState.Theta)
   t := glm.Translate3Dd(-p[0], -p[1], -p[2])
   r.Data.Cameraview = t.Mul4(rotation)
}

func (r *Receiver) SetCarTransform() {
   p := r.Car.Center
   m := glm.Translate3Dd(p[0], p[1], p[2])
   r.Data.Car.Transform = m
}

func (r *Receiver) OnClose(window *glfw.Window) {
}

func (r *Receiver) IsIdle() bool {
   return false
}

func (r *Receiver) NeedsRender() bool {
   return true
}

func (r *Receiver) Quit() {
   r.Window.SetShouldClose(true)
}

func (r *Receiver) PushFuelPedal() {
   r.UIState.Controls.FuelPedal++
}

func (r *Receiver) ReleaseFuelPedal() {
   r.UIState.Controls.FuelPedal--
}

func (r *Receiver) PushReversePedal() {
   r.UIState.Controls.FuelPedal--
}

func (r *Receiver) ReleaseReversePedal() {
   r.UIState.Controls.FuelPedal++
}

func (r *Receiver) PushBreakPedal() {
   r.UIState.Controls.BreakPedal++
}

func (r *Receiver) ReleaseBreakPedal() {
   r.UIState.Controls.BreakPedal--
}

func (r *Receiver) ToggleRotate() {
   r.UIState.IsRotating = !r.UIState.IsRotating
}
