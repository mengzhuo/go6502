package lisa

type Mnemonic uint16

//go:generate stringer --type=Mnemonic
const (
	ADC Mnemonic = iota + 1
	AND
	ASL
	BCC
	BCS
	BEQ
	BIT
	BNE
	BPL
	BRK
	BVC
	BVS
	CLC
	CLD
	CLI
	CLV
	CMP
	CPX
	CPY
	DEC
	DEX
	DEY
	EOR
	INC
	INX
	INY
	JMP
	JSR
	LDA
	LDX
	LDY
	LSR
	NOP
	ORA
	PHA
	PHP
	PLA
	PLP
	ROL
	ROR
	RTI
	RTS
	SBC
	SEC
	SED
	SEI
	STA
	STX
	STY
	TAX
	TAY
	TSX
	TXA
	TXS
	TYA
	BTR
	BFL
	BGE
	BLT
	XOR
	SET
	LDR
	STO
	LDD
	STD
	POP
	STP
	ADD
	SUB
	PPD
	CPR
	INR
	DCR
	RTN
	BRA
	BNC
	BIC
	BIP
	BIM
	BMI
	BNM
	BKS
	RSB
	BSB
	BNZ
	OBJ
	ORG
	EPZ
	EQU
	ASC
	STR
	HEX
	LST
	NLS
	DCM
	ICL
	END
	ADR
	DCI
	INV
	BLK
	DFS
	PAG
	PAU
	BYT
	HBY
	DBY
	LET
	TTL
	NOG
	GEN
	PHS
	DPH
	DA // dots prefix
	IF
	EL
	FI
	USR
)

func isPseudo(m Mnemonic) bool {
	switch m {
	case OBJ, ORG, EPZ, EQU, ASC, STR, HEX, LST, NLS, DCM,
		ICL, END, ADR, DCI, INV, BLK, DFS, PAG, PAU, BYT,
		HBY, DBY, LET, TTL, NOG, GEN, PHS, DPH, DA, IF, EL, FI, USR:
		return true
	}
	return false
}

func isJump(m Mnemonic) bool {
	switch m {
	case BCC, BCS, BEQ, BNE, BNC, BPL, JMP, JSR, RTS, BTR, BFL, BGE, BLT, BMI,
		BRK, BVC, BVS:
		return true
	}
	return false
}
