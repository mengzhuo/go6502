L1	= $1000
L2	= $C000
LOGO	= L1+L2
	ORG LOGO
	HEX DEADBEEF
	ASC 'ASC World'
	STR 'STRING HELLO'