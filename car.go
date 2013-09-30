package car

import (
   "math"
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
   BreakingPower     float64 //N
   
   CenterOfGravityHeight float64 //m
   RearAxelDisplacement  float64 //m
   FrontAxelDisplacement float64 //m
   TyreFrictionMu        float64 
}

type Car struct {
   Profile   Profile
   Center    glm.Vec4d
   Velocity  glm.Vec4d
   Direction glm.Vec4d
   //Transient properties for numerical integration
   TransientAcceleration glm.Vec4d
}

type Controls struct {
   FuelPedal float64
   BreakPedal float64
}

func NewCar() Car {
   return Car {
      Profile {
         100,
         500,
         0.4257,
         12.8,
         1000,
         1,
         1,
         1,
         1,
      },
      glm.Vec4d{0,0,0,1},
      glm.Vec4d{0,0,0,0},
      glm.Vec4d{0,1,0,0},
      glm.Vec4d{0,0,0,0},
   }
}

func IsZero(v glm.Vec4d) bool {
   return v.ApproxEqual(glm.Vec4d{})
}

func (car *Car) Simulate(controls Controls, timestep float64) {
   p := car.Center
   v := car.Velocity
   u := car.Direction
   at := car.TransientAcceleration
   m := car.Profile.Mass
   d := car.Profile.Drag
   rr := car.Profile.RollingResistance
   bmax := car.Profile.BreakingPower
   emax := car.Profile.EngineForce
   mu := car.Profile.TyreFrictionMu
   g := Gravity
   dt := timestep
   staticWeight := g * m
   vmag := v.Len()
   h := car.Profile.CenterOfGravityHeight
   fl := car.Profile.FrontAxelDisplacement
   rl := car.Profile.RearAxelDisplacement
   axelDisplacement := fl + rl
   
   e := emax * controls.FuelPedal
   b := bmax * controls.BreakPedal

   vn := glm.Vec4d{}
   if !IsZero(v) {
      vn = v.Normalize()
   }
   
   // maxFrontTyreTraction := rl / axelDisplacement * weight - h/axelLength * m * at.Dot(u)
   dynamicWeight := h / axelDisplacement * m * at.Dot(u)
   weight := fl / axelDisplacement * staticWeight + dynamicWeight
   maxRearTyreTraction := mu * weight
   rearForceTraction := math.Copysign(math.Min(math.Abs(e), maxRearTyreTraction), e)
   forceTraction := u.Mul(rearForceTraction)
   
   forceBreaking := u.Mul(-b * vn.Dot(u))
   forceDrag := v.Mul(-d * vmag)
   forceRollingResistance := v.Mul(-rr)
   
   force := forceTraction.Add(forceDrag).Add(forceRollingResistance).Add(forceBreaking)
   
   a := force.Mul(1.0/m)
   dv := a.Mul(dt)
   dp := v.Mul(dt).Add(a.Mul(dt*dt*0.5))

   if v.Len() < forceBreaking.Len()*dt/m {
      dv = v.Mul(-1)
   }
   
   p = p.Add(dp)
   v = v.Add(dv)
   
   car.Center = p
   car.Velocity = v
   car.TransientAcceleration = a
}
