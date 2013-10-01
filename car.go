package car

import (
   // "fmt"
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
   RearWheelAngularDeviation float64
   FrontWheelAngularDeviation float64
   FrontWheelO float64
   
   //Transient properties for numerical integration
   TransientAcceleration glm.Vec4d
}

type Controls struct {
   FuelPedal float64
   BreakPedal float64
}

type EngineProfile interface {
   Torque(rpm float64) float64
   StallingRPM() float64
}

type SimpleEngine struct {
   torque float64
   stallingRPM float64
}

func (engine *SimpleEngine) Torque(rpm float64) float64 {
   return engine.torque
}

func (engine *SimpleEngine) StallingRPM() float64 {
   return engine.stallingRPM
}

func NewCar() Car {
   return Car {
      Profile {
         1500,                      //Mass Kg
         &SimpleEngine{448, 1000},  //Engine N, rpm
         0.4257,  //Drag
         120.8,    //rolling friction
         10000,   //breaking power N
         1,       //center of gravity height m
         3,       //rear axel 
         3,       //front axel
         10000,   //type friction mu
         1,   //tyre traction constant
         2.66,    //gear ratio
         3.42,    //diff ratio
         0.7,     //efficiency
         1,//0.34,    //wheel radius
         150,      //whell mass

         // Mass              float64 //Kg
         // Engine            EngineProfile
         // Drag              float64 
         // RollingResistance float64 
         // BreakingPower     float64 //N
         //    
         // CenterOfGravityHeight float64 //m
         // RearAxelDisplacement  float64 //m
         // FrontAxelDisplacement float64 //m
         // TyreFrictionMu        float64 
         // TyreTractionConstant    float64
         //    
         // GearRatio               float64
         // DifferentialRatio       float64
         // TransmissionEfficiency  float64
         // WheelRadius             float64
         // WheelMass               float64
      },
      glm.Vec4d{0,0,1,1},
      glm.Vec4d{0,0,0,0},
      glm.Vec4d{0,1,0,0},
      0,
      0,
      0,
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
   
   // maxFrontTyreTraction := rl / axelDisplacement * weight - (h-wr)/axelLength * m * at.Dot(u)
   dynamicWeight := (h-wr) / axelDisplacement * m * at.Dot(u)
   weight := fl / axelDisplacement * staticWeight + dynamicWeight
   maxRearTyreTraction := mu * weight

   freeRollingAV := v.Dot(u) / wr
   rwav = freeRollingAV
   slipRatio := 0.0
   if !IsZero(v) {
      slipRatio = (rwav * wr - v.Dot(u)) / vmag
   }
   //  else {
   //    // slipRatio = rwav * wr //-rwav * wr
   // }
   tractionForce := math.Min(tc * slipRatio, maxRearTyreTraction)
   tractionTorque := tractionForce * wr
   // fmt.Println("tyre traction force", tractionForce)

   vn := glm.Vec4d{}
   if !IsZero(v) {
      vn = v.Normalize()
   }
   
   brakeTorque := bmaxToruqe * vn.Dot(u) * controls.BreakPedal

   rpm := freeRollingAV * xg * xd * 60 / (2 * math.Pi)
   engineTorque := car.Profile.Engine.Torque(rpm)
   
   appliedTorque := engineTorque * controls.FuelPedal
   driveTorque := appliedTorque * xg * xd * te
   driveForce := driveTorque / wr
   
   totalTorque := driveTorque - 2 * tractionTorque - brakeTorque
   
   // fmt.Println("wrav", rwav)
   // fmt.Println("driveTorque", driveTorque)
   // fmt.Println("tractionTorque", tractionTorque)
   // fmt.Println("brakeTorque", brakeTorque)
   // fmt.Println("totalTorque", totalTorque)
   
   b := bmax * controls.BreakPedal
   
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
   
   // fmt.Println(rwav, dv)
   
   p = p.Add(dp)
   v = v.Add(dv)
   rwav = rwav + drwav
   
   rwav = v.Dot(u) / wr
   
   car.FrontWheelO = 45 * math.Pi / 180
   if car.FrontWheelO > 0.0001 {
      turningRadius := axelDisplacement / math.Sin(car.FrontWheelO)
      angularVelocity := math.Copysign(v.Len() / turningRadius, v.Dot(u))
      rotation := dt * angularVelocity 
      // fmt.Println("turning radius", turningRadius, rotation)
      m := glm.HomogRotate3DZd(-rotation * 180 / math.Pi)
      u = m.Mul4x1(u)
      v = m.Mul4x1(v)
   }
   
   car.Center = p
   car.Velocity = v
   car.TransientAcceleration = a
   car.RearWheelAngularVelocity = rwav
   car.Direction = u
   car.FrontWheelAngularDeviation += freeRollingAV * dt
   car.RearWheelAngularDeviation += car.RearWheelAngularVelocity * dt 
}
