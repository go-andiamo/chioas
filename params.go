package chioas

type PathParams map[string]PathParam

type PathParam struct {
	Description string
	Example     any
}
