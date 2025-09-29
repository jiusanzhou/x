package x

import (
	"reflect"
)

// func DeepCopy[T any](v T) T {

// }

// DeepCopy 使用反射实现深度拷贝
func DeepCopy[T any](src T) T {
	srcValue := reflect.ValueOf(src)
	dstValue := deepCopyValue(srcValue)
	return dstValue.Interface().(T)
}

// deepCopyValue 递归拷贝值
func deepCopyValue(src reflect.Value) reflect.Value {
	// 处理零值
	if !src.IsValid() {
		return src
	}

	// 如果是指针，需要解引用
	if src.Kind() == reflect.Ptr {
		if src.IsNil() {
			return src
		}
		// 递归拷贝指针指向的值
		elem := deepCopyValue(src.Elem())
		// 创建新的指针
		newPtr := reflect.New(elem.Type())
		newPtr.Elem().Set(elem)
		return newPtr
	}

	// 根据不同类型进行处理
	switch src.Kind() {
	case reflect.Slice:
		return deepCopySlice(src)
	case reflect.Map:
		return deepCopyMap(src)
	case reflect.Struct:
		return deepCopyStruct(src)
	case reflect.Array:
		return deepCopyArray(src)
	case reflect.Interface:
		return deepCopyInterface(src)
	default:
		// 基本类型直接拷贝
		if src.CanInterface() {
			return reflect.ValueOf(src.Interface())
		}
		return src
	}
}

// 拷贝切片
func deepCopySlice(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return src
	}

	dst := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
	for i := 0; i < src.Len(); i++ {
		elem := deepCopyValue(src.Index(i))
		dst.Index(i).Set(elem)
	}
	return dst
}

// 拷贝数组
func deepCopyArray(src reflect.Value) reflect.Value {
	dst := reflect.New(src.Type()).Elem()
	for i := 0; i < src.Len(); i++ {
		elem := deepCopyValue(src.Index(i))
		dst.Index(i).Set(elem)
	}
	return dst
}

// 拷贝Map
func deepCopyMap(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return src
	}

	dst := reflect.MakeMapWithSize(src.Type(), src.Len())
	iter := src.MapRange()
	for iter.Next() {
		key := deepCopyValue(iter.Key())
		value := deepCopyValue(iter.Value())
		dst.SetMapIndex(key, value)
	}
	return dst
}

// 拷贝结构体
func deepCopyStruct(src reflect.Value) reflect.Value {
	dst := reflect.New(src.Type()).Elem()

	for i := 0; i < src.NumField(); i++ {
		// 检查字段是否可导出
		field := src.Type().Field(i)
		if field.PkgPath != "" && !field.Anonymous { // 非导出字段
			continue
		}

		srcField := src.Field(i)
		if srcField.CanInterface() {
			dstField := dst.Field(i)
			if dstField.CanSet() {
				copied := deepCopyValue(srcField)
				dstField.Set(copied)
			}
		}
	}
	return dst
}

// 拷贝接口
func deepCopyInterface(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return src
	}

	elem := src.Elem()
	copied := deepCopyValue(elem)
	return copied.Convert(src.Type())
}
