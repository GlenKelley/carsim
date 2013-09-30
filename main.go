package main

import (
   "fmt"
   glfw "github.com/go-gl/glfw3"
   gtk "github.com/GlenKelley/go-glutil"
   gl "github.com/GlenKelley/go-gl/gl32"
   glm "github.com/Jragonmiris/mathgl"
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
   Car      Car
   Controls  gtk.ControlBindings
   UIState  UIState
}

type DataBindings struct {
   Vao  gl.VertexArrayObject
   Scene *gtk.Model
   Projection glm.Mat4d
   Cameraview glm.Mat4d
}

type SceneBindings struct {
   Projection gl.UniformLocation `gl:"projection"`
   Cameraview gl.UniformLocation `gl:"cameraview"`
   Worldview  gl.UniformLocation `gl:"worldview"`
   Position gl.AttributeLocation `gl:"position"`
}

type Car struct {
   Position  glm.Vec4d
   Orientation glm.Quatd
   Velocity glm.Vec4d
   EngineForce float64
   BreakForce float64
}

type UIState struct {
   IsRotating bool
   Theta      float64
}

const (
   ProgramScene = "scene"
   Fov = 60
   Near = 0.1
   Far = 100
   CarMass = 1.0
   MaximumEngineForce = 10
   MaximumBreakForce = 10
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
   
   r.Data.Scene = gtk.EmptyModel()
   var err error
   r.Data.Scene, err = gtk.LoadSceneAsModel("car.dae")
   if err != nil { panic(err) }
   
   r.Car = Car{
      glm.Vec4d{0,0,0,1},
      glm.QuatIdentd(),
      glm.Vec4d{},
      0.0,
      0.0,
   }
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
   mv := glm.Ident4d()
   gl.UniformMatrix4fv(r.SceneLoc.Worldview, 1, gl.FALSE, gtk.MatArray(mv))
   gtk.PanicOnError()
   gtk.DrawModel(mv, r.Data.Scene, r.SceneLoc.Worldview, r.SceneLoc.Position, r.Data.Vao)
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
   
   //update car movement
   position := r.Car.Position
   velocity := r.Car.Velocity
   
   carFront := glm.Vec4d{0,1,0,0}
   engineForce := carFront.Mul(r.Car.EngineForce)
   var breakForce glm.Vec4d
   if !velocity.ApproxEqual(glm.Vec4d{}) {
      breakForce = velocity.Normalize().Mul(-r.Car.BreakForce)
   } else {
      velocity = glm.Vec4d{}
   }
   
   mass := CarMass
   force := engineForce.Add(breakForce)
   acceleration := force.Mul(1.0 / mass)
   
   position = position.Add(velocity.Mul(dt)).Add(acceleration.Mul(dt*dt/2))
   velocity = velocity.Add(acceleration.Mul(dt))
   
   r.Car.Position = position
   r.Car.Velocity = velocity
   r.SetCarTransform(glm.Translate3Dd(position[0], position[1], position[2]))
   
   //camera
   if r.UIState.IsRotating {
      period := 5.0
      dtheta := dt * 360 / period
      r.UIState.Theta += dtheta
   }
   
   p := glm.Vec4d{0,2,6,1}
   rotation := glm.HomogRotate3DYd(r.UIState.Theta)
   t := glm.Translate3Dd(-p[0], -p[1], -p[2])
   r.Data.Cameraview = t.Mul4(rotation)
}

func (r *Receiver) SetCarTransform(m glm.Mat4d) {
   r.Data.Scene.Children[0].Transform = m
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
   r.Car.EngineForce += MaximumEngineForce
}

func (r *Receiver) ReleaseFuelPedal() {
   r.Car.EngineForce -= MaximumEngineForce
}

func (r *Receiver) PushReversePedal() {
   r.Car.EngineForce -= MaximumEngineForce
}

func (r *Receiver) ReleaseReversePedal() {
   r.Car.EngineForce += MaximumEngineForce
}

func (r *Receiver) PushBreakPedal() {
   r.Car.BreakForce += MaximumBreakForce
}

func (r *Receiver) ReleaseBreakPedal() {
   r.Car.BreakForce -= MaximumBreakForce
}

func (r *Receiver) ToggleRotate() {
   r.UIState.IsRotating = !r.UIState.IsRotating
}
