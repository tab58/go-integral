package utils

func Map[T any, R any](slice []T, fn func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func MapErr[T any, R any](slice []T, fn func(T) (R, error)) ([]R, error) {
	result := make([]R, len(slice))
	var err error
	for i, v := range slice {
		result[i], err = fn(v)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func Reduce[Value any, Result any](slice []Value, fn func(Result, Value) Result, initialValue Result) Result {
	result := initialValue
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

func ReduceErr[Value any, Result any](slice []Value, fn func(Result, Value) (Result, error), initialValue Result) (Result, error) {
	result := initialValue
	for _, v := range slice {
		r, err := fn(result, v)
		if err != nil {
			return r, err
		}
		result = r
	}
	return result, nil
}

func Filter[Value any](slice []Value, fn func(Value) bool) []Value {
	result := make([]Value, 0)
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}
