********
* Test *
********

; This is a test comment
	ORG LABEL
SYMBOLIC EQU $1000
LABEL	= $fffe
LBL	= LABEL-SYMBOLIC
OK	= LABEL-SYMBOLIC
	LDA  LABEL+$1
	LDA   $1
LBL1	LDA  $800
	LDA LBL,X
	LDA LBL+$1,X
	LDA $10,X
	LDA $1010,X
	LDA LBL,Y
START	EQU $8000
	STA $80,Y
	LDX $0,Y
	BNE >6
	BCS LBL+$3
^6	BVC *+$5
	BMI <6
	BEQ LBL-$3
	BEQ >7-$3
	LDA (LBL),Y
	LDA (LBL+$2),Y
	LDA ($2),Y
	LDA (!10+%101),Y
	LDA (LBL,X)
	ADC (LBL+$3,X)
	STA (LABEL-!2,X)
	STA ($00,X)
	JMP (LBL)
	JMP (LBL+$3)
	JMP ($800)
^0 LDA #666
^9 STA <0
^7 BIT $C010
	BRK
