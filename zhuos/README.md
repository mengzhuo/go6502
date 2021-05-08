ZHU OS
=============

ZHU OS (Zhuo's Hardly Usable Operating System) is a simple OS
for Go 6502 AppleII (emulator, 64 KByte) and fun only
Some modification for easy implement of modern os.

Modifications
===============




Memory map
==============

0x0000 - 0x00ff Program
0x0100 - 0x01ff System stack
0x0400 - 0x07ff 40 cols text output page 1 (RW) (40x24)
0x0800 - 0x0bff 40 cols text output page 2 (RW)

0xc000 - 0xc010 Keyboard input

0xfffa - 0xfffb NMI handler (0x1000)
0xfffc - 0xfffd OS start point (0x2000)
0xfffe - 0xffff IRQ handler (0x3000)
