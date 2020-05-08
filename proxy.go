package reflectplus

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var proxyFactories = make(map[string]ProxyFactory)

func AddProxyFactory(importPath string, name string, fac ProxyFactory) {
	proxyFactories[importPath+"#"+name] = fac
}

type InvocationHandler func(method string, args ...interface{}) []interface{}

type ProxyFactory func(px InvocationHandler) interface{}

func NewProxy(importPath string, name string, handler InvocationHandler) (interface{}, error) {
	fac := proxyFactories[importPath+"#"+name]
	if fac == nil {
		return nil, fmt.Errorf("no proxy factory for %s#%s", importPath, name)
	}
	return fac(handler), nil
}

func generateSrcProxy(w *goGenFile, i Interface) {
	tName := typesafeName(i.ImportPath) + i.Name + "Proxy"

	// force the compiler to validate interface implementation
	w.Printf("var _ = (%s)(%s{})\n", w.ImportName(i.ImportPath, i.Name), tName)

	w.Printf("type %s struct {\n", tName)
	w.ShiftRight()
	w.Printf("Handler %s\n", w.ImportName("github.com/worldiety/reflectplus", "InvocationHandler"))
	w.ShiftLeft()
	w.Printf("}\n")
	for _, m := range i.Methods {
		w.Printf("func (_self %s) %s(", tName, m.Name)
		var paramNames []string
		for i, p := range m.Params {
			pName := p.Name
			if len(pName) == 0 {
				pName = "p" + strconv.Itoa(i)
			}
			paramNames = append(paramNames, pName)
			w.Printf("%s %s", pName, typeDeclToGo(w, p.Type))
			if i < len(m.Params)-1 {
				w.Printf(",")
			}
		}
		w.Printf(")")
		if len(m.Returns) > 0 {
			w.Printf("(")
			for i, p := range m.Returns {
				w.Printf(typeDeclToGo(w, p.Type))
				if i < len(m.Returns)-1 {
					w.Printf(",")
				}
			}
			w.Printf(")")
		}
		w.Printf("{\n")

		if len(m.Returns) > 0 {
			w.Printf("res := ")
		}

		w.Printf("_self.Handler(\"%s\",", m.Name)
		for i, p := range paramNames {
			w.Printf(p)
			if i < len(paramNames)-1 {
				w.Printf(",")
			}
		}
		w.Printf(")\n")
		if len(m.Returns) > 0 {
			w.Printf("return ")
			for i, r := range m.Returns {
				w.Printf("res[%d].(%s)", i, typeDeclToGo(w, r.Type))
				if i < len(m.Returns)-1 {
					w.Printf(",")
				}
			}
		}

		w.Printf("\n")
		w.Printf("}\n\n")
	}
}

func typesafeName(importPath string) string {
	tokens := strings.Split(importPath, "/")
	sb := &strings.Builder{}
	for _, token := range tokens {
		for i, r := range token {
			if i == 0 {
				r = unicode.ToUpper(r)
			}
			if r >= '0' && r <= '1' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
				sb.WriteRune(r)
			}
		}
	}
	return sb.String()
}

func typeDeclToGo(w *goGenFile, t TypeDecl) string {
	sb := &strings.Builder{}

	if t.Var{
		sb.WriteString("...")
	}

	for i := 0; i < t.Stars; i++ {
		sb.WriteByte('*')
	}

	switch t.Identifier {
	case "map":
		sb.WriteString("map[")
		sb.WriteString(typeDeclToGo(w, t.Params[0]))
		sb.WriteString("]")
		sb.WriteString(typeDeclToGo(w, t.Params[1]))
	case "[]":
		if t.Length == -1 {
			sb.WriteString("[]")
		} else {
			sb.WriteString("[")
			sb.WriteString(strconv.Itoa(t.Length))
			sb.WriteString("]")
		}
		sb.WriteString(typeDeclToGo(w, t.Params[0]))
	case "chan":
		sb.WriteString("chan ")
		sb.WriteString(typeDeclToGo(w, t.Params[0]))
	case "func":
		sb.WriteString("func(")
		for i, p := range t.Func.Params {
			sb.WriteString(p.Name)
			sb.WriteString(" ")
			sb.WriteString(typeDeclToGo(w, p.Type))
			if i < len(t.Func.Params)-1 {
				sb.WriteString(",")
			}
		}
		sb.WriteString(")")
		if len(t.Func.Returns) > 0 {
			sb.WriteString("(")
			for i, p := range t.Func.Returns {
				sb.WriteString(p.Name)
				sb.WriteString(" ")
				sb.WriteString(typeDeclToGo(w, p.Type))
				if i < len(t.Func.Params)-1 {
					sb.WriteString(",")
				}
			}
			sb.WriteString(")")
		}
	default:
		sb.WriteString(w.ImportName(t.ImportPath, t.Identifier))
	}
	return sb.String()
}