package torque

import (
	"fmt"
	"reflect"
)

type ReflectedFieldNode struct {
	Value       reflect.Value
	StructField reflect.StructField

	Parent   *ReflectedFieldNode
	Children []*ReflectedFieldNode
}

func (node *ReflectedFieldNode) IsKind(kind reflect.Kind) (reflect.Kind, bool) {
	if node.StructField.Type.Kind() == reflect.Interface && node.Value.Kind() != kind {
		return node.Value.Kind(), false
	} else if node.StructField.Type.Kind() != kind {
		return node.StructField.Type.Kind(), false
	}
	return kind, true
}

func (node *ReflectedFieldNode) GetKind() reflect.Kind {
	if node.StructField.Type.Kind() == reflect.Interface {
		return node.Value.Kind()
	}
	return node.StructField.Type.Kind()
}

func (node *ReflectedFieldNode) FindPath(path []string) *ReflectedFieldNode {
	if len(path) == 0 {
		return node
	}

	for _, child := range node.Children {
		if child.StructField.Name == path[0] {
			return child.FindPath(path[1:])
		}
	}

	return nil
}

// createReflectedFieldTree can be used to create a tree structure of the fields in a struct
func createReflectedFieldTree(structOrPtr interface{}) (*ReflectedFieldNode, error) {
	return reflectedFieldTreeRecursive(structOrPtr, make(map[reflect.Type]struct{}))
}

func reflectedFieldTreeRecursive(structOrPtr interface{}, visited map[reflect.Type]struct{}) (root *ReflectedFieldNode, err error) {
	typ := reflect.TypeOf(structOrPtr)
	if _, seen := visited[typ]; seen {
		return &ReflectedFieldNode{
			Value:    reflect.ValueOf(structOrPtr),
			Children: make([]*ReflectedFieldNode, 0),
		}, nil
	}
	visited[typ] = struct{}{}

	root = &ReflectedFieldNode{
		Value: reflect.ValueOf(structOrPtr),
		StructField: reflect.StructField{
			Name: fmt.Sprintf("%T", structOrPtr),
		},

		Parent:   nil,
		Children: make([]*ReflectedFieldNode, 0),
	}

	if root.Value.Kind() == reflect.Ptr {
		// detect all methods on this pointer
		val := root.Value
		for i := 0; i < val.NumMethod(); i++ {
			methodVal := val.Method(i)
			methodTyp := val.Type().Method(i)

			node := &ReflectedFieldNode{
				Value: methodVal,
				StructField: reflect.StructField{
					Name: methodTyp.Name,
					Type: methodTyp.Type,
				},
				Parent:   root,
				Children: make([]*ReflectedFieldNode, 0),
			}
			root.Children = append(root.Children, node)

			// for each of the values returned by this method,
			// create a field tree and append it as a child
			for j := 0; j < methodVal.Type().NumOut(); j++ {
				retTyp := methodVal.Type().Out(j)
				var retVal reflect.Value
				if retTyp.Kind() == reflect.Ptr {
					retVal = reflect.New(retTyp.Elem())
				} else {
					retVal = reflect.New(retTyp).Elem()
				}

			retTypSwitch:
				switch retTyp.Kind() {
				case reflect.Ptr:
					fallthrough
				case reflect.Struct:
					// check for circular dependencies, if found, append the parent node
					// as a child of the returned tree instead of recurring again
					for temp := node.Parent; temp != nil; temp = temp.Parent {
						if temp.Value.Type() == retTyp {
							node.Children = append(node.Children, temp.Children...)
							break retTypSwitch
						}
					}

					tree, err := reflectedFieldTreeRecursive(retVal.Interface(), visited)
					if err != nil {
						return root, err
					}
					// the children of the returned tree should be children of this node
					node.Children = append(node.Children, tree.Children...)
				}
			}

		}

		// convert this pointer to a value
		root.Value = val.Elem()
	}

	if root.Value.Kind() != reflect.Struct {
		return
	}

	val := root.Value
	for i := 0; i < val.NumField(); i++ {
		iface := zeroValueInterfaceFromField(val.Field(i))
		if iface != nil {
			node, err := reflectedFieldTreeRecursive(iface, visited)
			if err != nil {
				return nil, err
			}
			node.StructField = val.Type().Field(i)
			node.Parent = root
			root.Children = append(root.Children, node)

			//support embedded struct fields
			if node.StructField.Anonymous {
				for _, child := range node.Children {
					child.Parent = root
					root.Children = append(root.Children, child)
				}
			}
		} else if val.Field(i).Kind() == reflect.Struct {
			node := &ReflectedFieldNode{
				Value:       val.Field(i),
				StructField: val.Type().Field(i),
				Parent:      root,
				Children:    make([]*ReflectedFieldNode, 0),
			}
			root.Children = append(root.Children, node)
		} else {
			node := &ReflectedFieldNode{
				Value: val.Field(i),
				StructField: reflect.StructField{
					Name: val.Type().Field(i).Name,
					Type: val.Type().Field(i).Type,
				},
				Parent:   root,
				Children: make([]*ReflectedFieldNode, 0),
			}
			root.Children = append(root.Children, node)
		}
	}

	return root, nil
}

func recurseFieldsImplementing[T interface{}](structOrPtr interface{}, fn func(val T, field reflect.StructField) error) error {
	val := reflect.ValueOf(structOrPtr)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Prefer the original value over a zero-value reflection for the root check,
	// so that TemplateProvider implementations whose Template() returns instance
	// data (not a constant) are handled correctly.
	if t, ok := structOrPtr.(T); ok {
		err := fn(t, reflect.StructField{
			Name: fmt.Sprintf("%T", structOrPtr),
		})
		if err != nil {
			return err
		}
	} else {
		iface := zeroValueInterfaceFromField(val)
		if t, ok := iface.(T); ok {
			err := fn(t, reflect.StructField{
				Name: fmt.Sprintf("%T", structOrPtr),
			})
			if err != nil {
				return err
			}
		}
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanInterface() {
			continue // skip unexported fields
		}

		if t, ok := tryAssertInterface[T](field); ok {
			err := fn(t, fieldType)
			if err != nil {
				return err
			}
		}

		if field.Kind() != reflect.Ptr &&
			field.Kind() != reflect.Slice &&
			field.Kind() != reflect.Struct {
			continue
		}

		iface := zeroValueInterfaceFromField(field)
		if t, ok := iface.(T); ok {
			err := fn(t, val.Type().Field(i))
			if err != nil {
				return err
			}
		}

		if field.Kind() == reflect.Slice {
			// Get the underlying type of this slice
			underlyingType := field.Type().Elem()
			if underlyingType.Kind() != reflect.Ptr &&
				underlyingType.Kind() != reflect.Struct {
				continue
			}

			iface = zeroValueInterfaceFromField(field)
		} else if field.Kind() != reflect.Struct {
			// If this is not a struct or pointer, we can't recurse
			continue
		}

		// Even if this field is not the interface we're looking for, its
		// child fields might be... So recurse on
		err := recurseFieldsImplementing[T](iface, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func tryAssertInterface[T any](val reflect.Value) (T, bool) {
	var zero T

	// Unwrap pointers and interfaces
	for val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		if val.IsNil() {
			elemType := val.Type().Elem()
			if !elemType.Implements(reflect.TypeOf((*T)(nil)).Elem()) &&
				elemType.Kind() != reflect.Struct && elemType.Kind() != reflect.Ptr {
				return zero, false
			}

			val = reflect.New(elemType)
		}
		val = val.Elem()
	}

	if !val.IsValid() || !val.CanInterface() {
		return zero, false
	}

	iface := val.Interface()
	t, ok := iface.(T)
	return t, ok
}

// zeroValueInterfaceFromField converts a reflected field to a zero'd version of itself as an interface type.
// this makes it easier to perform type assertions on reflected struct fields
func zeroValueInterfaceFromField(field reflect.Value) interface{} {
	switch field.Kind() {
	case reflect.Struct:
		if field.Type().Kind() == reflect.Ptr {
			return reflect.New(field.Type().Elem()).Interface()
		}
		return reflect.New(field.Type()).Interface()
	case reflect.Ptr:
		fallthrough
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Ptr {
			return reflect.New(field.Type().Elem().Elem()).Interface()
		}
		return reflect.New(field.Type().Elem()).Interface()
	}
	return nil
}
