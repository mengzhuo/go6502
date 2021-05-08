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

func isNonAddress(m Mnemonic) bool {
	switch m {
	case OBJ, ORG, EPZ, EQU:
		return true
	}
	return false
}

func isRawData(m Mnemonic) bool {
	switch m {
	case ASC, STR, HEX, ADR:
		return true
	}
	return false
}

func isRelative(m Mnemonic) bool {
	switch m {
	case BCC, BCS, BEQ, BNE,
		BNC, BPL, BTR, BFL,
		BGE, BLT, BMI, BVC, BVS:
		return true
	}
	return false
}
