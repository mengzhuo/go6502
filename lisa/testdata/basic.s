********
* Test *
********

; This is a test comment

	LDA LABEL-SYMBOLIC
	LDA  LABEL+$1
	LDA   $1
	LDA  $800
	LDA LBL,X
	LDA LBL+$1,X
	LDA $10,X
	LDA $1010,X
	LDA LBL,Y
	STA LBL+$80,Y
	LDX $0,Y
	BNE LBL
	BCS LBL+$3
	BVC *+$5
	BMI $900
	BEQ LBL-$3
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
^0 LDA #0 
^9 STA LBL
^7 BIT $C010
