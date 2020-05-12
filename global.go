package reflectplus

import (
	"encoding/json"
	"reflect"
)

var packages []Package
var typesByName map[string]reflect.Type = make(map[string]reflect.Type)

func AddPackage(pkg Package) {
	packages = append(packages, pkg)
}

func ImportMetaData(jsn []byte) (Package, error) {
	pkg := Package{}
	err := json.Unmarshal(jsn, &pkg)
	if err != nil {
		return pkg, err
	}
	AddPackage(pkg)
	return pkg, nil
}

func Packages() []Package {
	return packages
}

// FindType returns the type or nil
func FindType(importPath string, name string) reflect.Type {
	return typesByName[importPath+"#"+name]
}

func AddType(importPath string, name string, p reflect.Type) {
	typesByName[importPath+"#"+name] = p
}

func Interfaces() []Interface {
	var res []Interface
	for _, p := range packages {
		for _, iface := range p.AllInterfaces() {
			res = append(res, iface)
		}
	}
	return res
}

func FindInterface(importPath string, name string) *Interface {
	for _, p := range packages {
		for _, iface := range p.AllInterfaces() {
			if iface.ImportPath == importPath && iface.Name == name {
				return &iface
			}
		}
	}
	return nil
}

func FindStruct(importPath string, name string) *Struct {
	for _, p := range packages {
		for _, iface := range p.AllStructs() {
			if iface.ImportPath == importPath && iface.Name == name {
				return &iface
			}
		}
	}
	return nil
}

func FindPackage(importPath string) *Package {
	for _, p := range packages {
		if p.ImportPath == importPath {
			return &p
		}
		var r *Package
		p.VisitPackages(func(pkg Package) bool {
			if pkg.ImportPath == importPath {
				r = &pkg
				return false
			}
			return true
		})
		if r != nil {
			return r
		}
	}
	return nil
}
