package gitlab

import (
	"reflect"
	"strings"
	"testing"
)

// TestServiceMethodsInInterface ensures that all exported methods on Service structs
// are part of their corresponding interfaces
func TestServiceMethodsInInterface(t *testing.T) {
	t.Parallel()
	for concreteService, interfaceType := range serviceMap {
		concreteType := reflect.TypeOf(concreteService)
		interfaceTyp := reflect.TypeOf(interfaceType).Elem()
		serviceName := concreteType.Elem().Name()

		// Check that the concrete service implements the interface
		if !reflect.TypeOf(concreteService).Implements(interfaceTyp) {
			t.Errorf("%s doesn't implement %s", serviceName, interfaceTyp.Name())
			continue
		}

		// Get all exported methods from the concrete service struct
		concreteMethodMap := make(map[string]struct{})
		for i := range concreteType.NumMethod() {
			method := concreteType.Method(i)
			if method.IsExported() {
				concreteMethodMap[method.Name] = struct{}{}
			}
		}

		// Get all methods from the interface
		interfaceMethodMap := make(map[string]struct{})
		for i := range interfaceTyp.NumMethod() {
			method := interfaceTyp.Method(i)
			interfaceMethodMap[method.Name] = struct{}{}
		}

		// Find methods in concrete type that are not in the interface
		var notInInterface []string
		for methodName := range concreteMethodMap {
			if isStandardMethod(methodName) {
				continue
			}

			if _, ok := interfaceMethodMap[methodName]; !ok {
				notInInterface = append(notInInterface, methodName)
			}
		}

		if len(notInInterface) > 0 {
			t.Errorf(
				"%s has exported methods not in %s: %s",
				serviceName,
				interfaceTyp.Name(),
				strings.Join(notInInterface, ", "),
			)
		}
	}
}

// isStandardMethod checks if the method name is likely from a standard Go interface
// You might need to customize this list based on your project
func isStandardMethod(name string) bool {
	standardMethods := map[string]bool{
		"String":        true, // fmt.Stringer
		"Error":         true, // error
		"MarshalJSON":   true, // json.Marshaler
		"UnmarshalJSON": true, // json.Unmarshaler
	}
	return standardMethods[name]
}
