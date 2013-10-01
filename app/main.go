package main

import (
   "os"
   "fmt"
   "math"
   "encoding/json"
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

func panicOnErr(err error) {
   if err != nil {
      panic(err)
   }
}

type Receiver struct {
   Window   *glfw.Window
   Data     DataBindings
   Shaders  gtk.ShaderLibrary
   SceneLoc SceneBindings
   Car      sim.Car
   Controls  gtk.ControlBindings
   Constants Constants
   UIState  UIState
}

type Constants struct {
   PanSensitivity float64
   Fov float64
   Near float64
   Far float64
}
var DefaultConstants = Constants{10, 60, 0.1, 100}

const (
   ProgramScene = "scene"
)

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
   Controls   sim.Controls
   PanAxis   glm.Vec4d
   TiltAxis  glm.Vec4d
   Orientation glm.Quatd
   CameraDistance float64
}

func (r *Receiver) Init(window *glfw.Window) {
   r.Window = window
   r.LoadConfiguration("gameconf.json")
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
   
   r.UIState = UIState{
      sim.Controls{},
      glm.Vec4d{0,1,0,0},
      glm.Vec4d{1,0,0,0},
      glm.QuatIdentd(),
      10,
   }
}

func (r *Receiver) LoadConfiguration(confFile string) {
   r.Constants = DefaultConstants
   r.ResetKeyBindingDefaults()
   file, err := os.Open(confFile)
   if err == nil {
      defer file.Close()
      decoder := json.NewDecoder(file)
      root := map[string]interface{}{}
      err = decoder.Decode(&root)
      panicOnErr(err)
      if constants, ok := root["constants"]; ok {
         bytes, err := json.Marshal(constants)
         panicOnErr(err)
         err = json.Unmarshal(bytes, &r.Constants)
         panicOnErr(err)
      }
      if controls, ok := root["controls"]; ok {
         sc := make(map[string]string)
         for k, v := range controls.(map[string]interface{}) {
            sc[k] = v.(string)
         }
         r.Controls.Apply(r, sc)
      }
   }
}

func (r *Receiver) ResetKeyBindingDefaults() {
   c := &r.Controls
   c.ResetBindings()
   c.BindKeyPress(glfw.KeyW, r.PushFuelPedal, r.ReleaseFuelPedal)
   c.BindKeyPress(glfw.KeyS, r.PushReversePedal, r.ReleaseReversePedal)
   c.BindKeyPress(glfw.KeySpace, r.PushBreakPedal, r.ReleaseBreakPedal)
   c.BindKeyPress(glfw.KeyEscape, r.Quit, nil)
   c.BindMouseMovement(r.PanView)
   c.BindScroll(r.Zoom)
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
   r.Data.Projection = glm.Perspectived(r.Constants.Fov, aspectRatio, r.Constants.Near, r.Constants.Far)
}

func (r *Receiver) MouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
   r.Controls.DoMouseButtonAction(button, action)
}

func (r *Receiver) MouseMove(window *glfw.Window, xpos float64, ypos float64) {
   r.Controls.DoMouseMoveAction(window, xpos, ypos)
}

func (r *Receiver) KeyPress(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey) {
   r.Controls.DoKeyAction(k, action)
}

func (r *Receiver) Scroll(window *glfw.Window, xoff float64, yoff float64) {
   r.Controls.DoScrollAction(xoff, yoff)
}

func (r *Receiver) Simulate(time gtk.GameTime) {
   dt := time.Delta.Seconds()
   r.Car.Simulate(r.UIState.Controls, dt)
   r.SetCarTransform()
   
   
   c := r.Data.Car.WorldTransform().Mul4x1(glm.Vec4d{0,0,0,1})
   p := glm.Vec4d{0,0,1,0}.Mul(r.UIState.CameraDistance)
   rotation := r.UIState.Orientation.Conjugate().Mat4()
   t := glm.Translate3Dd(-p[0], -p[1], -p[2])
   r.Data.Cameraview = t.Mul4(rotation).Mul4(glm.Translate3Dd(-c[0], -c[1], -c[2]))
}

func (r *Receiver) SetCarTransform() {
   p := r.Car.Center
   m := glm.Translate3Dd(p[0], p[1], p[2])
   r.Data.Car.Transform = m
}


func (r *Receiver) PanView(pos, delta glm.Vec2d) {
   theta := delta.Mul(r.Constants.PanSensitivity * r.Constants.Fov)
   
   turnV := glm.QuatRotated(theta[1], gtk.ToVec3D(r.UIState.TiltAxis))
   turnH := glm.QuatRotated(-theta[0], gtk.ToVec3D(r.UIState.PanAxis))
   
   r.UIState.Orientation = turnH.Mul(r.UIState.Orientation).Mul(turnV)
}

func (r *Receiver) Zoom(xoff, yoff float64) {
   f := math.Max(0.1, math.Exp2(-yoff/16))
   r.UIState.CameraDistance *= f
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
