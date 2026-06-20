package nand16

// Module is a hierarchical collection of gates, flip-flops, and sub-modules.
type Module struct {
	Name  string
	Gates []*Gate
	FFs   []*FlipFlop
	Sub   []*Module
}

// NewModule creates an empty module.
func NewModule(name string) *Module {
	return &Module{Name: name}
}

// NAND adds a NAND gate to this module, returns the output wire.
func (m *Module) NAND(a, b *Wire) *Wire {
	out := NewWire()
	m.Gates = append(m.Gates, &Gate{A: a, B: b, Out: out})
	return out
}

// DFF adds a D flip-flop, returns (Q, Qn).
func (m *Module) DFF(d *Wire) (*Wire, *Wire) {
	ff := &FlipFlop{
		D:  d,
		Q:  NewWire(),
		Qn: NewWire(),
	}
	ff.Qn.Val = true
	m.FFs = append(m.FFs, ff)
	return ff.Q, ff.Qn
}

// DFFTo adds a D flip-flop that writes to a pre-existing Q wire.
// Essential for feedback loops (register hold).
func (m *Module) DFFTo(d, q *Wire) {
	ff := &FlipFlop{
		D:  d,
		Q:  q,
		Qn: NewWire(),
	}
	m.FFs = append(m.FFs, ff)
}

// Add includes a sub-module.
func (m *Module) Add(sub *Module) {
	m.Sub = append(m.Sub, sub)
}

// AllGates recursively collects all gates.
func (m *Module) AllGates() []*Gate {
	gates := make([]*Gate, 0, len(m.Gates))
	gates = append(gates, m.Gates...)
	for _, s := range m.Sub {
		gates = append(gates, s.AllGates()...)
	}
	return gates
}

// AllFFs recursively collects all flip-flops.
func (m *Module) AllFFs() []*FlipFlop {
	ffs := make([]*FlipFlop, 0, len(m.FFs))
	ffs = append(ffs, m.FFs...)
	for _, s := range m.Sub {
		ffs = append(ffs, s.AllFFs()...)
	}
	return ffs
}
