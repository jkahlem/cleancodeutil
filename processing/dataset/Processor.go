package dataset

type Filter interface {
	Include(TargetType) bool
}

type TargetType struct {
	Filter  Filter
	Subsets []TargetType
}

type Processor struct {
	TargetSet     TargetType
	SubProcessors []Processor
}

func NewProcessor(set TargetType) Processor {
	processor := Processor{
		SubProcessors: make([]Processor, len(set.Subsets)),
	}

	// {Perform any initialization here}
	for i, subset := range set.Subsets {
		processor.SubProcessors[i] = NewProcessor(subset)
	}
	return processor
}

func (p *Processor) Process(t TargetType) {
	if !p.TargetSet.Filter.Include(t) {
		return
	}

	// {Do actual activity here}

	for i := range p.SubProcessors {
		p.SubProcessors[i].Process(t)
	}
}
