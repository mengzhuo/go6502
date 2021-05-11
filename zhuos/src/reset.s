* Reset Routine
* Clear screen
* Clear keyboard
* Show Logo
* JMP to main loop

CLRKB = $C010
MAIN  = $B000
LOGO  = $D000

	ORG $a500
	CLV
	LDA #$04
	STA $41
	LDA #$D0 ; setup logo position
	STA $43
	LDY #!0 ; y = data pnt
	LDX #!0 ; x = src pnt

SCRLOOP	LDA ($42),Y
	STA ($40),Y
	INY
	CPY #!27
	BNE SCRLOOP
	TYA
	LDY #!0
	CLC
	ADC $42 ; store y
	STA $42

	CLC
	LDA #$80
	ADC $40
	STA $40
	BNE SCRLOOP
	INC $41
	LDA #!6
	CMP $41
	BNE SCRLOOP


	LDA #!1 ; clear keyboard
	STA CLRKB

	JMP (MAIN)
