package nand16

// Gate represents a 2-input NAND gate.
type Gate struct {
	A, B *Wire
	Out  *Wire
}

// Eval computes NAND: Out = !(A && B)
func (g *Gate) Eval() {
	g.Out.Val = !(g.A.Val && g.B.Val)
}
