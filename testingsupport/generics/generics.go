package generics

type SingleAny[T any] struct{}
type SingleString[T string] struct{}
type SingleFloat32[T float32] struct{}
type SingleFloat64[T float64] struct{}
type SingleInt[T int] struct{}
type SingleInt16[T int16] struct{}
type SingleInt32[T int32] struct{}
type SingleInt64[T int64] struct{}
type SingleBool[T bool] struct{}

type OrAny[T any | any] struct{}
type OrMixed[T string | bool] struct{}
type ManyOrAny[T any | any | any | any] struct{}
type ManyOrMixed[T string | bool | int | int16] struct{}

type MultipleAny[T any, K any] struct{}
type MultipleAnyOneType[T, K any] struct{}
type MultipleString[T string, K string] struct{}
type MultipleStringOneType[T, K string] struct{}

type AnonIface[T interface{}] struct{}
type named interface{}
type NamedIface[T named] struct{}
