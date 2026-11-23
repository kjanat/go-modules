package cm

import (
	"unsafe"

	upstream "go.bytecodealliance.org/cm"
)

type (
	HostLayout   = upstream.HostLayout
	AnyInteger   = upstream.AnyInteger
	Discriminant = upstream.Discriminant
	BoolResult   = upstream.BoolResult
	Resource     = upstream.Resource
	Rep          = upstream.Rep
	ErrorContext = upstream.ErrorContext
)

type (
	List[T any]                                    = upstream.List[T]
	AnyList[T any]                                 = upstream.AnyList[T]
	Option[T any]                                  = upstream.Option[T]
	Result[Shape, OK, Err any]                     = upstream.Result[Shape, OK, Err]
	AnyResult[Shape, OK, Err any]                  = upstream.AnyResult[Shape, OK, Err]
	Variant[Tag Discriminant, Shape, Align any]    = upstream.Variant[Tag, Shape, Align]
	AnyVariant[Tag Discriminant, Shape, Align any] = upstream.AnyVariant[Tag, Shape, Align]
	Future[T any]                                  = upstream.Future[T]
	Stream[T any]                                  = upstream.Stream[T]
)

type (
	Tuple[T0, T1 any]                                                                 = upstream.Tuple[T0, T1]
	Tuple3[T0, T1, T2 any]                                                            = upstream.Tuple3[T0, T1, T2]
	Tuple4[T0, T1, T2, T3 any]                                                        = upstream.Tuple4[T0, T1, T2, T3]
	Tuple5[T0, T1, T2, T3, T4 any]                                                    = upstream.Tuple5[T0, T1, T2, T3, T4]
	Tuple6[T0, T1, T2, T3, T4, T5 any]                                                = upstream.Tuple6[T0, T1, T2, T3, T4, T5]
	Tuple7[T0, T1, T2, T3, T4, T5, T6 any]                                            = upstream.Tuple7[T0, T1, T2, T3, T4, T5, T6]
	Tuple8[T0, T1, T2, T3, T4, T5, T6, T7 any]                                        = upstream.Tuple8[T0, T1, T2, T3, T4, T5, T6, T7]
	Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8 any]                                    = upstream.Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8]
	Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 any]                               = upstream.Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]
	Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any]                          = upstream.Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]
	Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any]                     = upstream.Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]
	Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any]                = upstream.Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]
	Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any]           = upstream.Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]
	Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any]      = upstream.Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]
	Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any] = upstream.Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]
)

const (
	MaxTuple     = upstream.MaxTuple
	ResourceNone = upstream.ResourceNone
	ResultOK     = upstream.ResultOK
	ResultErr    = upstream.ResultErr
)

func NewList[T any, Len AnyInteger](data *T, len Len) List[T] {
	return upstream.NewList(data, len)
}

func ToList[S ~[]T, T any](s S) List[T] {
	return upstream.ToList(s)
}

func None[T any]() Option[T] {
	return upstream.None[T]()
}

func Some[T any](v T) Option[T] {
	return upstream.Some(v)
}

func OK[R AnyResult[Shape, OK, Err], Shape, OK, Err any](ok OK) R {
	return upstream.OK[R](ok)
}

func Err[R AnyResult[Shape, OK, Err], Shape, OK, Err any](err Err) R {
	return upstream.Err[R](err)
}

func NewVariant[Tag Discriminant, Shape, Align any, T any](tag Tag, data T) Variant[Tag, Shape, Align] {
	return upstream.NewVariant[Tag, Shape, Align](tag, data)
}

func New[V AnyVariant[Tag, Shape, Align], Tag Discriminant, Shape, Align any, T any](tag Tag, data T) V {
	return upstream.New[V](tag, data)
}

func Case[T any, V AnyVariant[Tag, Shape, Align], Tag Discriminant, Shape, Align any](v *V, tag Tag) *T {
	return upstream.Case[T](v, tag)
}

func CaseUnmarshaler[T ~uint8 | ~uint16 | ~uint32](cases []string) func(v *T, text []byte) error {
	return upstream.CaseUnmarshaler[T](cases)
}

func Reinterpret[T, From any](from From) (to T) {
	return upstream.Reinterpret[T](from)
}

func LowerString[S ~string](s S) (*byte, uint32) {
	return upstream.LowerString(s)
}

func LiftString[T ~string, Data unsafe.Pointer | uintptr | *uint8, Len AnyInteger](data Data, len Len) T {
	return upstream.LiftString[T](data, len)
}

func LowerList[L AnyList[T], T any](list L) (*T, uint32) {
	return upstream.LowerList(list)
}

func LiftList[L AnyList[T], T any, Data unsafe.Pointer | uintptr | *T, Len AnyInteger](data Data, len Len) L {
	return upstream.LiftList[L](data, len)
}

func BoolToU32[B ~bool](v B) uint32 { return upstream.BoolToU32(v) }

func U32ToBool(v uint32) bool { return upstream.U32ToBool(v) }

func F32ToU32(v float32) uint32 { return upstream.F32ToU32(v) }

func U32ToF32(v uint32) float32 { return upstream.U32ToF32(v) }

func F64ToU64(v float64) uint64 { return upstream.F64ToU64(v) }

func U64ToF64(v uint64) float64 { return upstream.U64ToF64(v) }

func F32ToU64(v float32) uint64 { return upstream.F32ToU64(v) }

func U64ToF32(v uint64) float32 { return upstream.U64ToF32(v) }

func PointerToU32[T any](v *T) uint32 { return upstream.PointerToU32(v) }

func U32ToPointer[T any](v uint32) *T { return upstream.U32ToPointer[T](v) }

func PointerToU64[T any](v *T) uint64 { return upstream.PointerToU64(v) }

func U64ToPointer[T any](v uint64) *T { return upstream.U64ToPointer[T](v) }
