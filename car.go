package car

import (
   glm "github.com/Jragonmiris/mathgl"
)

const (
   Gravity = 9.8
)

type Profile struct {
   Mass              float64 //Kg
   EngineForce       float64 //N
   Drag              float64 
   RollingResistance float64
   BreakingPower     float64
}

type Car struct {
   Profile   Profile
   Center    glm.Vec4d
   Velocity  glm.Vec4d
   Direction glm.Vec4d
}

type Controls struct {
   FuelPedal float64
   BreakPedal float64
}

func NewCar() Car {
   return Car {
      Profile {
         1000,
         1000,
         0.4257,
         12.8,
         1000,
      },
      glm.Vec4d{0,0,0,1},
      glm.Vec4d{0,0,0,0},
      glm.Vec4d{0,1,0,0},
   }
}

func IsZero(v glm.Vec4d) bool {
   return v.ApproxEqual(glm.Vec4d{})
}

func (car *Car) Simulate(controls Controls, timestep float64) {
   p := car.Center
   v := car.Velocity
   u := car.Direction
   m := car.Profile.Mass
   d := car.Profile.Drag
   rr := car.Profile.RollingResistance
   bmax := car.Profile.BreakingPower
   emax := car.Profile.EngineForce
   vmag := v.Len()
   dt := timestep
   
   e := emax * controls.FuelPedal
   b := bmax * controls.BreakPedal
   
   vn := glm.Vec4d{}
   if !IsZero(v) {
      vn = v.Normalize()
   }
   
   forceTraction := u.Mul(e)
   forceDrag := v.Mul(-d * vmag)
   forceRollingResistance := v.Mul(-rr)
   forceBreaking := u.Mul(-b * vn.Dot(u))
   
   force := forceTraction.Add(forceDrag).Add(forceRollingResistance).Add(forceBreaking)
   
   a := force.Mul(1.0/m)
   p = p.Add(v.Mul(dt)).Add(a.Mul(dt*dt*0.5))
   v = v.Add(a.Mul(dt))
   
   car.Center = p
   car.Velocity = v
}
