package forms

import "sync"

type InputCollection struct {
	groups map[string]*InputGroup

	icMx sync.Mutex
}

func NewInputCollection() *InputCollection {
	return &InputCollection{
		groups: make(map[string]*InputGroup),
	}
}

func (ic *InputCollection) NewGroup(
	name string,
	description string,
) *InputGroup {

	ic.icMx.Lock()
	defer ic.icMx.Unlock()

	ig := &InputGroup{
		name:        name,
		description: description,
		inputs:      []Input{},

		containers:   make(map[int]*InputGroup),
		fieldNameSet: make(map[string]Input),

		fieldValueLookupHints: make(map[string][]string),
	}
	ic.groups[name] = ig
	return ig
}

func (ic *InputCollection) HasGroup(name string) bool {
	ic.icMx.Lock()
	defer ic.icMx.Unlock()

	_, exists := ic.groups[name]
	return exists
}

func (ic *InputCollection) Group(name string) *InputGroup {
	ic.icMx.Lock()
	defer ic.icMx.Unlock()

	return ic.groups[name]
}

func (ic *InputCollection) Groups() []*InputGroup {
	ic.icMx.Lock()
	defer ic.icMx.Unlock()

	groupList := []*InputGroup{}
	for _, g := range ic.groups {
		groupList = append(groupList, g)
	}
	return groupList
}
