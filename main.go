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
   Data     DataBindings
   Shaders  gtk.ShaderLibrary
   SceneLoc SceneBindings
   Car      Car
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
}

const (
   PROGRAM_SCENE = "scene"
   FOV = 60
   NEAR = 0.1
   FAR = 100
)

func (r *Receiver) Init(window *glfw.Window) {
   gtk.Bind(&r.Data)
   r.Shaders = gtk.NewShaderLibrary()
   r.Shaders.LoadProgram(PROGRAM_SCENE, "scene.v.glsl", "scene.f.glsl")
   r.Shaders.BindProgramLocations(PROGRAM_SCENE, &r.SceneLoc)
   gtk.PanicOnError()
   
   r.Data.Projection = glm.Ident4d()
   r.Data.Cameraview = glm.Ident4d()
   
   r.Data.Scene = gtk.EmptyModel()
   var err error
   r.Data.Scene, err = gtk.LoadSceneAsModel("car.dae")
   if err != nil { panic(err) }
   
   r.Car = Car {
      glm.Vec4d{0,0,0,1},
      glm.QuatIdentd(),
   }
}

func (r *Receiver) Draw(window *glfw.Window) {
   bg := gtk.SoftBlack
   gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
   gl.Enable(gl.DEPTH_TEST)
   gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
   
   r.Shaders.UseProgram(PROGRAM_SCENE)
   gl.UniformMatrix4fv(r.SceneLoc.Projection, 1, gl.FALSE, gtk.MatArray(r.Data.Projection))
   gl.UniformMatrix4fv(r.SceneLoc.Cameraview, 1, gl.FALSE, gtk.MatArray(r.Data.Cameraview))
   mv := glm.Ident4d()
   gl.UniformMatrix4fv(r.SceneLoc.Worldview, 1, gl.FALSE, gtk.MatArray(mv))
   gtk.PanicOnError()
   gtk.DrawModel(mv, r.Data.Scene, r.SceneLoc.Worldview, r.SceneLoc.Position, r.Data.Vao)
}

func (r *Receiver) Reshape(window *glfw.Window, width, height int) {
   aspectRatio := gtk.WindowAspectRatio(window)
   r.Data.Projection = glm.Perspectived(FOV, aspectRatio, NEAR, FAR)
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
   period := 5.0
   elapsed := time.Elapsed.Seconds()
   theta := elapsed * 360 / period
   p := glm.Vec4d{0,2,6,1}
   rotation := glm.HomogRotate3DYd(theta)
   r.Car.Position = p
   t := glm.Translate3Dd(-p[0], -p[1], -p[2])
   r.Data.Cameraview = t.Mul4(rotation)
}

func (r *Receiver) OnClose(window *glfw.Window) {
   
}

func (r *Receiver) IsIdle() bool {
   return false
}

func (r *Receiver) NeedsRender() bool {
   return true
}