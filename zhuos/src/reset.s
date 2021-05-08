* Reset Routine
* Clear screen
* Clear keyboard
* Show Logo
* JMP to main loop

CLRKB = $C010
MAIN  = $B000
LOGO  = $D000
SCRSTR = $400
SCREND = $08


	ORG $a000
	CLV
	LDA #!0 ; clear 40 cols screen
	STA SCRSTR
	LDA $40
	STA SCRSTR+!1 ; use $400 itself as indirect
	LDA $1

SCRLOOP	STA (SCRSTR),Y
	ADC SCRSTR
	BVC SCRNOF
	INC SCRSTR+!1
SCRNOF	LDA #SCREND
	CMP SCREND+!1
	BNE SCRLOOP


	LDA #!1 ; clear keyboard
	STA CLRKB


*	load Logo
	LDA LOGO,X
