L1	= $1000
L2	= $C000
LOGO	= L1+L2
	ORG LOGO
	HEX DEAD_BEEF
	ASC 'ASC World' + 'AFFF'
	STR 'STRING HELLO' + 'OK'
