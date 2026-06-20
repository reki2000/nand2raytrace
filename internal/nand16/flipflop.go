package nand16

// FlipFlop is a positive-edge-triggered D flip-flop.
// In cycle-based simulation, we sample D at clock edge and update Q.
type FlipFlop struct {
	D  *Wire
	Q  *Wire
	Qn *Wire
}

// Tick captures D into Q (called once per clock cycle).
func (ff *FlipFlop) Tick() {
	ff.Q.Val = ff.D.Val
	ff.Qn.Val = !ff.D.Val
}
