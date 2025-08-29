package chioas

import "golang.org/x/exp/maps"

// WalkPaths walks all the paths in the definition calling the supplied func with each
//
// If the pathDef is altered, it is also altered in the definition (unless an error is returned)
//
// Walking continues until an error is encountered or the func returns a false contd
func (d *Definition) WalkPaths(fn func(path string, pathDef *Path) (cont bool, err error)) error {
	_, err := d.Paths.walk("", fn)
	return err
}

// WalkMethods walks all the methods in the definition calling the supplied func with each
//
// If the methodDef is altered, it is also altered in the definition (unless an error is returned)
//
// Walking continues until an error is encountered or the func returns a false contd
func (d *Definition) WalkMethods(fn func(path string, method string, methodDef *Method) (contd bool, err error)) error {
	contd := true
	var err error
	keys := maps.Keys(d.Methods)
	for i := 0; contd && err == nil && i < len(keys); i++ {
		method := keys[i]
		methodDef := d.Methods[method]
		contd, err = fn(root, method, &methodDef)
		if err == nil {
			d.Methods[method] = methodDef
		}
	}
	if contd && err == nil {
		_, err = d.Paths.walk("", func(path string, pathDef *Path) (contd bool, err error) {
			contd = true
			k := maps.Keys(pathDef.Methods)
			for i := 0; contd && err == nil && i < len(k); i++ {
				method := k[i]
				methodDef := pathDef.Methods[method]
				contd, err = fn(path, method, &methodDef)
				if err == nil {
					pathDef.Methods[method] = methodDef
				}
			}
			return contd, err
		})
	}
	return err
}

func (ps Paths) walk(currPath string, fn func(path string, pathDef *Path) (bool, error)) (contd bool, err error) {
	contd = true
	keys := maps.Keys(ps)
	for i := 0; contd && err == nil && i < len(keys); i++ {
		path := keys[i]
		pathDef := ps[path]
		contd, err = fn(currPath+path, &pathDef)
		if err == nil {
			ps[path] = pathDef
		}
		if contd && err == nil {
			contd, err = pathDef.Paths.walk(currPath+path, fn)
		}
	}
	return contd, err
}
