; ============================================
; BIOS for NAND-16
; Memory map:
;   0x0000-0x01FF  BIOS code
;   0x0200-0xCFFF  User code + data
;   0xD000-0xDFFF  Return stack (4KB, grows down)
;   0xE000-0xEFFF  Data stack (4KB, grows down)
;   0xF000-0xF7FF  Framebuffer (64x32)
;   0xF800-0xF80F  UART
;   0xF810-0xF81F  Timer
;   0xFE00-0xFFFF  Boot ROM
; ============================================

; === Boot entry point (loaded at 0x0000) ===
_start:
	addi r1, r0, -17  ; r1 = -17 = 0xFFEF
	addi r2, r0, 8    ; r2 = 8
	shl  sp, r1, r2   ; sp = (-17 << 8) & 0xFFFF = 0xEF00

	; Init return stack pointer (R5) = 0xDF00
	addi r3, r0, -32  ; r3 = 0xFFE0
	addi r3, r3, -1   ; r3 = 0xFFDF
	shl  r5, r3, r2   ; r5 = 0xDF00

	; Jump to user code at 0x0200
	addi r1, r0, 2    ; r1 = 2
	shl  r1, r1, r2   ; r1 = 0x0200 (r2 still = 8)
	jalr r1            ; jump to 0x0200

; === Utility subroutines ===

; putchar: R2 = character -> UART data port (MMIO)
os_putchar:
	addi r1, r0, -16
	addi r3, r0, 8
	shl  r1, r1, r3   ; r1 = 0xF000
	addi r3, r0, 8
	shl  r3, r3, r3   ; r3 = 0x0800
	add  r1, r1, r3   ; r1 = 0xF800 (UART data)
	sb   r2, 0(r1)
	ret

; putpixel: R2=x, R3=y, R4=color -> framebuffer
os_putpixel:
	addi r1, r0, -16
	addi r7, r0, 8
	shl  r1, r1, r7   ; r1 = 0xF000
	addi r7, r0, 6
	shl  r3, r3, r7   ; r3 = y * 64
	add  r1, r1, r3
	add  r1, r1, r2
	sb   r4, 0(r1)
	ret

; clear_fb: fill framebuffer with R2
os_clear_fb:
	addi r1, r0, -16
	addi r3, r0, 8
	shl  r1, r1, r3   ; r1 = 0xF000
	addi r3, r0, 8
	shl  r3, r3, r3   ; r3 = 0x800
	add  r3, r1, r3   ; r3 = 0xF800
_cf_loop:
	sb   r2, 0(r1)
	addi r1, r1, 1
	blt  r1, r3, _cf_loop
	ret
