package main

import (
	"bufio"
	"io"
	"regexp"
)

var (
	memberPattern = regexp.MustCompile("\\s{4}([0-9:]+)?([\\w.]+)\\s([a-zA-Z][\\w]+)(\\([\\w,.]*\\))?\\s->\\s([\\w]+)")
	classPattern  = regexp.MustCompile("([\\w.]+)\\s->\\s([\\w.]+):")
)

type MemberType int

const (
	Method MemberType = iota
	Field
)

type GoldenField struct {
	obfName, goldenName string
	FieldType           string // fields type
}

func (f *GoldenField) Type() MemberType {
	return Field
}

func (f *GoldenField) ObfName() string {
	return f.obfName
}

func (f *GoldenField) GoldenName() string {
	return f.goldenName
}

type GoldenMethod struct {
	obfName, goldenName string
	ReturnType          string
	Parameters          string
}

func (m *GoldenMethod) Type() MemberType {
	return Method
}

func (m *GoldenMethod) ObfName() string {
	return m.obfName
}

func (m *GoldenMethod) GoldenName() string {
	return m.goldenName
}

type GoldenMember interface {
	Type() MemberType

	ObfName() string
	GoldenName() string
}

type GoldenMap struct {
	obfToGold map[string]GoldenMember
	goldToObf map[string]GoldenMember
}

func (g *GoldenMap) ObfName(gold string) GoldenMember {
	return g.goldToObf[gold]
}

func (g *GoldenMap) GoldName(obf string) GoldenMember {
	return g.obfToGold[obf]
}

func (g *GoldenMap) Put(obf, gold string, member GoldenMember) {
	g.obfToGold[obf] = member
	g.goldToObf[gold] = member
}

func NewGoldenMap() *GoldenMap {
	return &GoldenMap{
		obfToGold: map[string]GoldenMember{},
		goldToObf: map[string]GoldenMember{},
	}
}

type GoldenClass struct {
	ObfName    string
	GoldenName string

	Methods *GoldenMap
	Fields  *GoldenMap

	AllMembers []GoldenMember
}

func (gc *GoldenClass) PutField(field *GoldenField) {
	gc.Fields.Put(field.ObfName(), field.GoldenName(), field)
	gc.AllMembers = append(gc.AllMembers, field)
}

func (gc *GoldenClass) PutMethod(method *GoldenMethod) {
	gc.Methods.Put(method.obfName, method.goldenName, method)
	gc.AllMembers = append(gc.AllMembers, method)
}

func NewClass() *GoldenClass {
	return &GoldenClass{
		Methods: NewGoldenMap(),
		Fields:  NewGoldenMap(),
	}
}

type GoldenMappingsReader struct {
	ObfClassMap    map[string]*GoldenClass
	GoldenClassMap map[string]*GoldenClass
}

func NewGoldenMappingsReader() *GoldenMappingsReader {
	return &GoldenMappingsReader{
		ObfClassMap:    map[string]*GoldenClass{},
		GoldenClassMap: map[string]*GoldenClass{},
	}
}

func (r *GoldenMappingsReader) Read(reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	var currentClass *GoldenClass

	for scanner.Scan() {
		text := scanner.Text()
		if classPattern.MatchString(text) {

			if currentClass != nil {
				r.ObfClassMap[currentClass.ObfName] = currentClass
				r.GoldenClassMap[currentClass.GoldenName] = currentClass
			}

			currentClass = NewClass()
			matches := classPattern.FindAllStringSubmatch(text, -1)

			currentClass.GoldenName = matches[0][1]
			currentClass.ObfName = matches[0][2]
		} else if memberPattern.MatchString(text) {
			if currentClass != nil {
				matches := memberPattern.FindAllStringSubmatch(text, -1)

				if matches[0][1] == "" {
					// field

					f := &GoldenField{
						obfName:    matches[0][5],
						goldenName: matches[0][3],
						FieldType:  matches[0][2],
					}

					currentClass.PutField(f)
				} else {
					m := &GoldenMethod{
						obfName:    matches[0][5],
						goldenName: matches[0][3],
						ReturnType: matches[0][2],
						Parameters: matches[0][4],
					}
					currentClass.PutMethod(m)
				}
			}
		}
	}

}
