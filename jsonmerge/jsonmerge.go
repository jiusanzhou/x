package jsonmerge

import (
	"errors"
	"fmt"
	"reflect"
)

// Merge merge json value
func Merge(dst interface{}, src ...interface{}) error {
	var err error
	for _, v := range src {
		err = merge(dst, v)
		// TODO: continue with param
		if err != nil {
			return err
		}
	}

	return err
}

func merge(dst interface{}, src interface{}) error {
	if dst != nil && reflect.ValueOf(dst).Kind() != reflect.Ptr {
		return ErrNonPointerArgument
	}
	var (
		vDst, vSrc reflect.Value
		err        error
	)

	if vDst, vSrc, err = resolveValues(dst, src); err != nil {
		return err
	}
	if !vDst.CanSet() {
		return fmt.Errorf("cannot set dst, needs reference")
	}
	if vDst.Type() != vSrc.Type() {
		return ErrDifferentArgumentsTypes
	}
	_, err = deepMerge(vDst, vSrc, 0)
	return err
}

// Traverses recursively both values, assigning src's fields values to dst.
// The map argument tracks comparisons that have already been seen, which allows
// short circuiting on recursive types.
//
// if the type of dstIn or src is slice and the items type is map
// and the length or dst and src is the same
// merge each other
func deepMerge(dstIn, src reflect.Value, depth int) (dst reflect.Value, err error) {
	dst = dstIn

	overwrite := false
	overwriteWithEmptySrc := true
	overwriteSliceWithEmptySrc := false
	typeCheck := true
	appendSlice := true
	mergeArray := false

	// TODO:
	switch dst.Kind() {
	case reflect.Struct:
	case reflect.Map:
		if dst.IsNil() && !src.IsNil() {
			if dst.CanSet() {
				dst.Set(reflect.MakeMap(dst.Type()))
			} else {
				dst = src
				return
			}
		}

		for _, key := range src.MapKeys() {
			srcElement := src.MapIndex(key)
			if !srcElement.IsValid() {
				continue
			}

			dstElement := dst.MapIndex(key)
			if dstElement.IsValid() {
				k := dstElement.Interface()
				dstElement = reflect.ValueOf(k)
			}

			if isReflectNil(srcElement) {
				// if overwrite || isReflectNil(dstElement) {
				// 	dst.SetMapIndex(key, srcElement)
				// }

				continue
			}

			if !srcElement.CanInterface() {
				continue
			}

			if srcElement.CanInterface() {
				srcElement = reflect.ValueOf(srcElement.Interface())

				if dstElement.IsValid() {
					dstElement = reflect.ValueOf(dstElement.Interface())
				}
			}

			dstElement, err = deepMerge(dstElement, srcElement, depth+1)
			if err != nil {
				return
			}

			dst.SetMapIndex(key, dstElement)
		}
	case reflect.Slice:
		// merge slice items
		newSlice := dst

		// if the slice items type is not simple and same lenght
		// length < x
		if (!isEmptyValue(src) ||
			overwriteWithEmptySrc ||
			overwriteSliceWithEmptySrc) &&
			(overwrite || isEmptyValue(dst)) && !appendSlice {
			if typeCheck && src.Type() != dst.Type() {
				err = fmt.Errorf("cannot override two slices with different type (%s, %s)", src.Type(), dst.Type())
				return
			}

			newSlice = src
		} else if mergeArray && depth <= 2 && dst.Len() == src.Len() {
			// merge item
			// TODO: hardcode the maxDepth
			var nislice = reflect.MakeSlice(reflect.TypeOf([]interface{}{}), dst.Len(), dst.Len())
			for i := 0; i < dst.Len(); i++ {
				var ni reflect.Value
				ni, err = deepMerge(dst.Index(i), src.Index(i), depth+1)
				nislice.Index(i).Set(ni)
			}

			newSlice = nislice
		} else if appendSlice {
			if typeCheck && src.Type() != dst.Type() {
				err = fmt.Errorf("cannot append two slice with different type (%s, %s)", src.Type(), dst.Type())
				return
			}
			newSlice = reflect.AppendSlice(dst, src)
		}

		if dst.CanSet() {
			dst.Set(newSlice)
		} else {
			dst = newSlice
		}
	case reflect.Ptr, reflect.Interface:
		if isReflectNil(src) {
			break
		}

		if dst.Kind() != reflect.Ptr && src.Type().AssignableTo(dst.Type()) {
			if dst.IsNil() || overwrite {
				if overwrite || isEmptyValue(dst) {
					if dst.CanSet() {
						dst.Set(src)
					} else {
						dst = src
					}
				}
			}
		}

		if src.Kind() != reflect.Interface {
			if dst.IsNil() || (src.Kind() != reflect.Ptr && overwrite) {
				if dst.CanSet() && (overwrite || isEmptyValue(dst)) {
					dst.Set(src)
				}
			} else if src.Kind() == reflect.Ptr {
				if dst, err = deepMerge(dst.Elem(), src.Elem(), depth+1); err != nil {
					return
				}
				dst = dst.Addr()
			} else if dst.Elem().Type() == src.Type() {
				if dst, err = deepMerge(dst.Elem(), src, depth+1); err != nil {
					return
				}
			} else {
				err = ErrDifferentArgumentsTypes
				return
			}
			break
		}

		if dst.IsNil() || overwrite {
			if (overwrite || isEmptyValue(dst)) && (overwriteWithEmptySrc || !isEmptyValue(src)) {
				if dst.CanSet() {
					dst.Set(src)
				} else {
					dst = src
				}
			}
		} else if _, err = deepMerge(dst.Elem(), src.Elem(), depth+1); err != nil {
			return
		}

	default:
		mustSet := (!isEmptyValue(src) || overwriteWithEmptySrc) || (isEmptyValue(dst))
		if mustSet {
			if dst.CanSet() {
				dst.Set(src)
			} else {
				dst = src
			}
		}
	}

	return dst, err
}

// MergeJSONWithHardCode ...
func MergeJSONWithHardCode(dst interface{}, src interface{}) error {

	return nil
}

// IsReflectNil is the reflect value provided nil
func isReflectNil(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Interface, reflect.Slice, reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr:
		// Both interface and slice are nil if first word is 0.
		// Both are always bigger than a word; assume flagIndir.
		return v.IsNil()
	default:
		return false
	}
}

// From src/pkg/encoding/json/encode.go.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isEmptyValue(v.Elem())
	case reflect.Func:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}

func resolveValues(dst, src interface{}) (vDst, vSrc reflect.Value, err error) {
	if dst == nil || src == nil {
		err = ErrNilArguments
		return
	}
	vDst = reflect.ValueOf(dst).Elem()
	if vDst.Kind() != reflect.Struct && vDst.Kind() != reflect.Map {
		err = ErrNotSupported
		return
	}
	vSrc = reflect.ValueOf(src)
	// We check if vSrc is a pointer to dereference it.
	if vSrc.Kind() == reflect.Ptr {
		vSrc = vSrc.Elem()
	}
	return
}

// Errors reported by Mergo when it finds invalid arguments.
var (
	ErrNilArguments                = errors.New("src and dst must not be nil")
	ErrDifferentArgumentsTypes     = errors.New("src and dst must be of same type")
	ErrNotSupported                = errors.New("only structs and maps are supported")
	ErrExpectedMapAsDestination    = errors.New("dst was expected to be a map")
	ErrExpectedStructAsDestination = errors.New("dst was expected to be a struct")
	ErrNonPointerArgument          = errors.New("dst must be a pointer")
)
