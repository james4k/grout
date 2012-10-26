package grout

type Generator func(sitecfg, cfg M, info ContentInfo) (Content, error)

var generators = make(map[string]Generator)

func RegisterGenerator(name string, gen Generator) {
	if _, ok := generators[name]; ok {
		halt("Generator '%s' already exists!\n", name)
	}
	generators[name] = gen
}
