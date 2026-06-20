package nand16

// FlipFlop is a positive-edge-triggered D flip-flop.
// In cycle-based simulation, we sample D at clock edge and update Q.
type FlipFlop struct {
	D  *Wire
	Q  *Wire
	Qn *Wire
}

// NewDFF creates a D flip-flop with fresh Q/Qn wires.
func NewDFF(d *Wire) (*Wire, *Wire) {
	ff := &FlipFlop{
		D:  d,
		Q:  NewWire(),
		Qn: NewWire(),
	}
	ff.Qn.Val = true // initial: Q=0, Qn=1
	// Register globally via Module; store ff for later
	_ = ff
	return ff.Q, ff.Qn
}

// Tick captures D into Q (called once per clock cycle).
func (ff *FlipFlop) Tick() {
	ff.Q.Val = ff.D.Val
	ff.Qn.Val = !ff.D.Val
}
