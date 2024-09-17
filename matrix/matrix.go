package matrix

import (
	"github.com/bonoboris/satisfied/math32"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Matrix represents a 3x3 matrix (OpenGL style 3x3 - right handed, column major)
type Matrix struct {
	M0, M3, M6 float32
	M1, M4, M7 float32
	M2, M5, M8 float32
}

// NewIndentity returns an identity matrix.
func NewIndentity() Matrix {
	return Matrix{
		M0: 1, M3: 0, M6: 0,
		M1: 0, M4: 1, M7: 0,
		M2: 0, M5: 0, M8: 1,
	}
}

// NewTranslate returns a translation transformation matrix.
func NewTranslate(x, y float32) Matrix {
	return Matrix{
		M0: 1, M3: 0, M6: 0,
		M1: 0, M4: 1, M7: 0,
		M2: x, M5: y, M8: 1,
	}
}

// NewTranslateV returns a translation transformation matrix.
func NewTranslateV(v rl.Vector2) Matrix {
	return Matrix{
		M0: 1, M3: 0, M6: 0,
		M1: 0, M4: 1, M7: 0,
		M2: v.X, M5: v.Y, M8: 1,
	}
}

// NewRotate returns a rotation transformation matrix (angle in degrees).
func NewRotate(angle int32) Matrix {
	// Special case for (nearly) multiples of 90 degrees
	switch angle % 360 {
	case 0:
		return Matrix{
			M0: 1, M3: 0, M6: 0,
			M1: 0, M4: 1, M7: 0,
			M2: 0, M5: 0, M8: 1,
		}
	case 90, -270:
		return Matrix{
			M0: 0, M3: 1, M6: 0,
			M1: -1, M4: 0, M7: 0,
			M2: 0, M5: 0, M8: 1,
		}
	case 180, -180:
		return Matrix{
			M0: -1, M3: 0, M6: 0,
			M1: 0, M4: -1, M7: 0,
			M2: 0, M5: 0, M8: 1,
		}
	case 270, -90:
		return Matrix{
			M0: 0, M3: -1, M6: 0,
			M1: 1, M4: 0, M7: 0,
			M2: 0, M5: 0, M8: 1,
		}
	default:
		return NewRotateRad(float32(angle) * rl.Deg2rad)
	}
}

func NewRotateRad(angle float32) Matrix {
	cos := math32.Cos(angle)
	sin := math32.Sin(angle)
	return Matrix{
		M0: cos, M3: sin, M6: 0,
		M1: -sin, M4: cos, M7: 0,
		M2: 0, M5: 0, M8: 1,
	}
}

// NewRotateAround returns a rotation transformation matrix by angle around point (x, y).
func NewRotateAround(angle int32, x, y float32) Matrix {
	return NewTranslate(-x, -y).Rotate(angle).Translate(x, y)
}

// NewRotateAroundV returns a rotation transformation matrix by angle around point v.
func NewRotateAroundV(angle int32, v rl.Vector2) Matrix {
	return NewTranslateV(v).Rotate(angle).TranslateV(v.Negate())
}

// NewScale returns a scale transformation matrix.
func NewScale(s float32) Matrix {
	return Matrix{
		M0: s, M3: 0, M6: 0,
		M1: 0, M4: s, M7: 0,
		M2: 0, M5: 0, M8: 1,
	}
}

func (mat Matrix) IsIdentity() bool {
	return mat.M0 == 1 && mat.M1 == 0 && mat.M2 == 0 &&
		mat.M3 == 0 && mat.M4 == 1 && mat.M5 == 0 &&
		mat.M6 == 0 && mat.M7 == 0 && mat.M8 == 1
}

// Mult returns the matrix product mat * right.
func (mat Matrix) Mult(right Matrix) Matrix {
	var result Matrix

	result.M0 = mat.M0*right.M0 + mat.M1*right.M3 + mat.M2*right.M6
	result.M1 = mat.M0*right.M1 + mat.M1*right.M4 + mat.M2*right.M7
	result.M2 = mat.M0*right.M2 + mat.M1*right.M5 + mat.M2*right.M8
	result.M3 = mat.M3*right.M0 + mat.M4*right.M3 + mat.M5*right.M6
	result.M4 = mat.M3*right.M1 + mat.M4*right.M4 + mat.M5*right.M7
	result.M5 = mat.M3*right.M2 + mat.M4*right.M5 + mat.M5*right.M8
	result.M6 = mat.M6*right.M0 + mat.M7*right.M3 + mat.M8*right.M6
	result.M7 = mat.M6*right.M1 + mat.M7*right.M4 + mat.M8*right.M7
	result.M8 = mat.M6*right.M2 + mat.M7*right.M5 + mat.M8*right.M8

	return result
}

// Translate applies a translation transformation to the matrix.
func (mat Matrix) Translate(x, y float32) Matrix { return mat.Mult(NewTranslate(x, y)) }

// TranslateV applies a translation transformation to the matrix.
func (mat Matrix) TranslateV(v rl.Vector2) Matrix { return mat.Mult(NewTranslateV(v)) }

// Rotate applies a rotation transformation to the matrix (angle in degrees).
func (mat Matrix) Rotate(angle int32) Matrix { return mat.Mult(NewRotate(angle)) }

// Rotate applies a rotation transformation to the matrix (angle in radians).
func (mat Matrix) RotateRad(angle float32) Matrix { return mat.Mult(NewRotateRad(angle)) }

// RotateAround applies a rotation transformation to the matrix around point (x, y) (angle in degrees).
func (mat Matrix) RotateAround(angle int32, x, y float32) Matrix {
	return mat.Mult(NewRotateAround(angle, x, y))
}

// RotateAroundV applies a rotation transformation to the matrix around point v.
func (mat Matrix) RotateAroundV(angle int32, v rl.Vector2) Matrix {
	return mat.Mult(NewRotateAroundV(angle, v))
}

// Scale applies a scale transformation to the matrix.
func (mat Matrix) Scale(s float32) Matrix { return mat.Mult(NewScale(s)) }

// Apply applies transformation matrix to the vector (x, y).
func (mat Matrix) Apply(x, y float32) rl.Vector2 {
	return rl.Vector2{
		X: mat.M0*x + mat.M1*y + mat.M2,
		Y: mat.M3*x + mat.M4*y + mat.M5,
	}
}

// ApplyV applies transformation matrix to the vector v.
func (mat Matrix) ApplyV(v rl.Vector2) rl.Vector2 {
	return rl.Vector2{
		X: mat.M0*v.X + mat.M1*v.Y + mat.M2,
		Y: mat.M3*v.X + mat.M4*v.Y + mat.M5,
	}
}

// ApplyRec applies transformation matrix to the rect defined by top-left corner (x, y) and size (w, h).
//
// NOTE: The rect is properly rotated only for 0, 90, 180 and 270 degrees rotations.
func (mat Matrix) ApplyRec(x, y, w, h float32) rl.Rectangle {
	return rl.NewRectangleCorners(mat.Apply(x, y), mat.Apply(x+w, y+h))
}

// ApplyRecV applies transformation matrix to the rect defined by top-left corner v and size s.
//
// NOTE: The rect is properly rotated only for 0, 90, 180 and 270 degrees rotations.
func (mat Matrix) ApplyRecV(v rl.Vector2, s rl.Vector2) rl.Rectangle {
	return mat.ApplyRec(v.X, v.Y, s.X, s.Y)
}

// ApplyRecRec applies transformation matrix to the rect r.
//
// NOTE: The rect is properly rotated only for 0, 90, 180 and 270 degrees rotations.
func (mat Matrix) ApplyRecRec(r rl.Rectangle) rl.Rectangle {
	return mat.ApplyRec(r.X, r.Y, r.Width, r.Height)
}
