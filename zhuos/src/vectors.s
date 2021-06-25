	ICL "zhuos/src/symbols.s"

	ORG $FFFA
	ADR NMIVECT ; NMI handler(0xfffa)
	ADR RESET  ; Reset handler(0xfffc)
	ADR IRQENTRY  ; IRQ handler(0xfffe)
