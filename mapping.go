package chioas

import (
	"net/http"
	"regexp"
	"strings"
)

func (d *Definition) MapRequest(r *http.Request) (path *Path, method *Method) {
	if path = d.MapPath(r.URL.Path); path != nil {
		if m, ok := d.Methods[r.Method]; ok {
			method = &m
		}
	}
	return
}

func (d *Definition) MapPath(url string) (path *Path) {
	parts, err := pathSplitter.Split(url)
	if err != nil {
		return nil
	}
	l := len(parts)
	currPaths := d.Paths
	for i, part := range parts {
		if p, ok := currPaths["/"+part]; ok {
			if i == l {
				path = &p
				break
			} else {
				currPaths = p.Paths
			}
		} else if varPath := findPathVarMatch(part, currPaths); varPath != nil {
			if i == l {
				path = varPath
				break
			} else {
				currPaths = varPath.Paths
			}
		} else {
			break
		}
	}
	return path
}

func findPathVarMatch(path string, curr Paths) (found *Path) {
	for k, rp := range curr {
		if strings.HasPrefix(k, "/{") && strings.HasSuffix(k, "}") {
			pth := strings.TrimSuffix(strings.TrimPrefix(k, `/{`), `}`)
			matches := !strings.Contains(pth, ":")
			if !matches {
				paramRx := strings.SplitN(pth, ":", 2)
				if rx, err := regexp.Compile(paramRx[1]); err == nil {
					matches = rx.MatchString(path)
				}
			}
			if matches {
				found = &rp
				break
			}
		}
	}
	return
}
