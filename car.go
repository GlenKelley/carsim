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
   Engine            EngineProfile
   Drag              float64 
   RollingResistance float64 
   BreakingPower     float64 //N
   
   CenterOfGravityHeight float64 //m
   RearAxelDisplacement  float64 //m
   FrontAxelDisplacement float64 //m
   TyreFrictionMu        float64 
   TyreTractionConstant    float64
   
   GearRatio               float64
   DifferentialRatio       float64
   TransmissionEfficiency  float64
   WheelRadius             float64
   WheelMass               float64

}

type Car struct {
   Profile   Profile
   Center    glm.Vec4d
   Velocity  glm.Vec4d
   Direction glm.Vec4d
   RearWheelAngularVelocity float64
   //Transient properties for numerical integration
   TransientAcceleration glm.Vec4d
}

type Controls struct {
   FuelPedal float64
   BreakPedal float64
}

type EngineProfile interface {
   Torque(rpm float64) float64
}

type SimpleEngine struct {
   torque float64
}

func (engine *SimpleEngine) Torque(rpm float64) float64 {
   return engine.torque
}

func NewCar() Car {
   return Car {
      Profile {
         1500,
         &SimpleEngine{448},
         0.4257,
         12.8,
         10000,
         1,
         1,
         1,
         1,
         1,
         2.66,
         3.42,
         0.7,
         0.34,
         1,
      },
      glm.Vec4d{0,0,0,1},
      glm.Vec4d{0,0,0,0},
      glm.Vec4d{0,1,0,0},
      0,
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
   rwav := car.RearWheelAngularVelocity
   m := car.Profile.Mass
   d := car.Profile.Drag
   rr := car.Profile.RollingResistance
   h := car.Profile.CenterOfGravityHeight
   fl := car.Profile.FrontAxelDisplacement
   rl := car.Profile.RearAxelDisplacement
   xg := car.Profile.GearRatio
   xd := car.Profile.DifferentialRatio
   te := car.Profile.TransmissionEfficiency
   wr := car.Profile.WheelRadius
   wm := car.Profile.WheelMass
   bmax := car.Profile.BreakingPower
   bmaxToruqe := bmax / wr
   mu := car.Profile.TyreFrictionMu
   tc := car.Profile.TyreTractionConstant
   g := Gravity
   dt := timestep
   vmag := v.Len()
   
   staticWeight := g * m
   axelDisplacement := fl + rl
   
   // maxFrontTyreTraction := rl / axelDisplacement * weight - h/axelLength * m * at.Dot(u)
   dynamicWeight := h / axelDisplacement * m * at.Dot(u)
   weight := fl / axelDisplacement * staticWeight + dynamicWeight
   maxRearTyreTraction := mu * weight
   
   slipRatio := (rwav * wr - v.Dot(u)) / vmag
   tractionForce := math.Min(tc * slipRatio, maxRearTyreTraction)
   tractionTorque := tractionForce * wr
   
   brakeTorque := bmaxToruqe * controls.BreakPedal

   rpm := rwav * xg * xd * 60 / (2 * math.Pi)
   engineTorque := car.Profile.Engine.Torque(rpm)
   appliedTorque := engineTorque * controls.FuelPedal
   driveTorque := appliedTorque * xg * xd * te
   driveForce := driveTorque / wr

   totalTorque := driveTorque + 2 * tractionTorque + brakeTorque
   
   b := bmax * controls.BreakPedal

   vn := glm.Vec4d{}
   if !IsZero(v) {
      vn = v.Normalize()
   }
   
   rearForceTraction := math.Copysign(math.Min(math.Abs(driveForce), maxRearTyreTraction), driveForce)
   forceTraction := u.Mul(rearForceTraction)
   
   forceBreaking := u.Mul(-b * vn.Dot(u))
   forceDrag := v.Mul(-d * vmag)
   forceRollingResistance := v.Mul(-rr)
   
   force := forceTraction.Add(forceDrag).Add(forceRollingResistance).Add(forceBreaking)

   rearWheelInertia := wm * wr * wr / 2
   wheelAcceleration := totalTorque / rearWheelInertia
   drwav := wheelAcceleration * dt
   
   a := force.Mul(1.0/m)
   dv := a.Mul(dt)
   dp := v.Mul(dt).Add(a.Mul(dt*dt*0.5))
   if v.Len() < forceBreaking.Len()*dt/m {
      dv = v.Mul(-1)
   }
   
   p = p.Add(dp)
   v = v.Add(dv)
   rwav = rwav + drwav
   
   car.Center = p
   car.Velocity = v
   car.TransientAcceleration = a
   car.RearWheelAngularVelocity = rwav
}
