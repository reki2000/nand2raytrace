	li r6, 61184
	li r5, 57088
	addi r4, r0, 0
	jal .Lmain_1
.Ludiv_2:
	addi r5, r5, -2
	sw r7, 0(r5)
	addi r3, r0, 16
.Ludiv_loop_3:
	add r2, r2, r2
	bge r4, r0, .Ludiv_skip1_4
	addi r2, r2, 1
.Ludiv_skip1_4:
	add r4, r4, r4
	blt r2, r1, .Ludiv_skip2_5
	sub r2, r2, r1
	addi r4, r4, 1
.Ludiv_skip2_5:
	addi r3, r3, -1
	blt r0, r3, .Ludiv_loop_3
	lw r7, 0(r5)
	addi r5, r5, 2
	ret
.Lmain_1:
	jal .Lword_end_6
.Lword_7:
	addi r5, r5, -2
	sw r7, 0(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 2
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Ltrue_8
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_9
.Ltrue_8:
	addi r4, r0, -1
.Lcmp_end_9:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_10
	jal .Lif_false_11
.Lif_body_10:
	jal .Lif_end_12
.Lif_false_11:
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 2(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 2(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r2, r0, 0
	call .Ludiv_2
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 2(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 2(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r2, r0, 0
	call .Ludiv_2
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 2(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 2(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r2, r0, 0
	call .Ludiv_2
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, 2
.Lif_end_12:
	lw r7, 0(r5)
	addi r5, r5, 2
	ret
.Lword_end_6:
	jal .Lword_end_13
.Lword_14:
	addi r5, r5, -2
	sw r7, 0(r5)
	call .Lword_7
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 4
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shl r4, r4, r1
	lw r7, 0(r5)
	addi r5, r5, 2
	ret
.Lword_end_13:
	jal .Lword_end_15
.Lword_16:
	addi r5, r5, -2
	sw r7, 0(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	blt r4, r0, .Ltrue_17
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_18
.Ltrue_17:
	addi r4, r0, -1
.Lcmp_end_18:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_19
	jal .Lif_false_20
.Lif_body_19:
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
.Lif_false_20:
	lw r7, 0(r5)
	addi r5, r5, 2
	ret
.Lword_end_15:
	jal .Lword_end_21
.Lword_22:
	addi r5, r5, -2
	sw r7, 0(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 128
	lw r1, 0(r6)
	sw r4, 0(r6)
	add r4, r1, r0
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r3, r0, r0
	bge r4, r0, .Lfdiv_pos_23
	sub r4, r0, r4
	addi r3, r0, -1
.Lfdiv_pos_23:
	addi r5, r5, -2
	sw r3, 0(r5)
	addi r3, r0, 8
	shr r2, r4, r3
	shl r4, r4, r3
	call .Ludiv_2
	lw r3, 0(r5)
	addi r5, r5, 2
	beq r3, r0, .Lfdiv_neg_24
	sub r4, r0, r4
.Lfdiv_neg_24:
	lw r7, 0(r5)
	addi r5, r5, 2
	ret
.Lword_end_21:
	jal .Lword_end_25
.Lword_26:
	addi r5, r5, -2
	sw r7, 0(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8210
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8208
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8206
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8204
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8210
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8212
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8192
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8194
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 256
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8214
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8204
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8192
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8206
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8194
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8208
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65280
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8216
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8204
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8206
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8208
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8212
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8218
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8216
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8214
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8218
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	blt r4, r0, .Ltrue_27
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_28
.Ltrue_27:
	addi r4, r0, -1
.Lcmp_end_28:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_29
	jal .Lif_false_30
.Lif_body_29:
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	jal .Lif_end_31
.Lif_false_30:
	call .Lword_14
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8216
	lw r4, 0(r4)
	sub r4, r0, r4
	lw r1, 0(r6)
	sw r4, 0(r6)
	add r4, r1, r0
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8214
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r3, r0, r0
	bge r4, r0, .Lfdiv_pos_32
	sub r4, r0, r4
	addi r3, r0, -1
.Lfdiv_pos_32:
	addi r5, r5, -2
	sw r3, 0(r5)
	addi r3, r0, 8
	shr r2, r4, r3
	shl r4, r4, r3
	call .Ludiv_2
	lw r3, 0(r5)
	addi r5, r5, 2
	beq r3, r0, .Lfdiv_neg_33
	sub r4, r0, r4
.Lfdiv_neg_33:
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Ltrue_34
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_35
.Ltrue_34:
	addi r4, r0, -1
.Lcmp_end_35:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_36
	jal .Lif_false_37
.Lif_body_36:
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	jal .Lif_end_38
.Lif_false_37:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8220
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8220
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8192
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8204
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8210
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r3, r0, r0
	bge r4, r0, .Lfdiv_pos_39
	sub r4, r0, r4
	addi r3, r0, -1
.Lfdiv_pos_39:
	addi r5, r5, -2
	sw r3, 0(r5)
	addi r3, r0, 8
	shr r2, r4, r3
	shl r4, r4, r3
	call .Ludiv_2
	lw r3, 0(r5)
	addi r5, r5, 2
	beq r3, r0, .Lfdiv_neg_40
	sub r4, r0, r4
.Lfdiv_neg_40:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8222
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8220
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8194
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8206
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8210
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r3, r0, r0
	bge r4, r0, .Lfdiv_pos_41
	sub r4, r0, r4
	addi r3, r0, -1
.Lfdiv_pos_41:
	addi r5, r5, -2
	sw r3, 0(r5)
	addi r3, r0, 8
	shr r2, r4, r3
	shl r4, r4, r3
	call .Ludiv_2
	lw r3, 0(r5)
	addi r5, r5, 2
	beq r3, r0, .Lfdiv_neg_42
	sub r4, r0, r4
.Lfdiv_neg_42:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8224
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8220
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65280
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8208
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8210
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r3, r0, r0
	bge r4, r0, .Lfdiv_pos_43
	sub r4, r0, r4
	addi r3, r0, -1
.Lfdiv_pos_43:
	addi r5, r5, -2
	sw r3, 0(r5)
	addi r3, r0, 8
	shr r2, r4, r3
	shl r4, r4, r3
	call .Ludiv_2
	lw r3, 0(r5)
	addi r5, r5, 2
	beq r3, r0, .Lfdiv_neg_44
	sub r4, r0, r4
.Lfdiv_neg_44:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8226
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8220
	lw r4, 0(r4)
.Lif_end_38:
.Lif_end_31:
	lw r7, 0(r5)
	addi r5, r5, 2
	ret
.Lword_end_25:
	jal .Lword_end_45
.Lword_46:
	addi r5, r5, -2
	sw r7, 0(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8236
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8234
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8232
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8222
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65388
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8224
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 148
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8226
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 148
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 256
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 160
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 96
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 256
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Lmin_47
	add r4, r1, r0
.Lmin_47:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8228
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8222
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65453
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8224
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 83
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8226
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 227
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	call .Lword_16
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 77
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8228
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 26
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 256
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Lmin_48
	add r4, r1, r0
.Lmin_48:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8228
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8232
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8228
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 31
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Lmin_49
	add r4, r1, r0
.Lmin_49:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8234
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8228
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 31
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Lmin_50
	add r4, r1, r0
.Lmin_50:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8236
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8228
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 31
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Lmin_51
	add r4, r1, r0
.Lmin_51:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	lw r7, 0(r5)
	addi r5, r5, 2
	ret
.Lword_end_45:
	jal .Lword_end_52
.Lword_53:
	addi r5, r5, -2
	sw r7, 0(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8260
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8258
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8256
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8254
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8238
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8254
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8264
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65408
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8256
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8266
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8240
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8258
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8268
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8264
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8266
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8268
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8260
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8270
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8264
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65388
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8266
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 148
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8268
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 148
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8270
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	blt r4, r0, .Ltrue_54
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_55
.Ltrue_54:
	addi r4, r0, -1
.Lcmp_end_55:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_56
	jal .Lif_false_57
.Lif_body_56:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	jal .Lif_end_58
.Lif_false_57:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65535
.Lif_end_58:
	lw r7, 0(r5)
	addi r5, r5, 2
	ret
.Lword_end_52:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 32
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r5, r5, -2
	sw r4, 0(r5)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r5, r5, -2
	sw r4, 0(r5)
	lw r4, 0(r6)
	addi r6, r6, 2
.Ldo_59:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 64
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r5, r5, -2
	sw r4, 0(r5)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r5, r5, -2
	sw r4, 0(r5)
	lw r4, 0(r6)
	addi r6, r6, 2
.Ldo_60:
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 2(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 32
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shl r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8192
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 16
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 6(r5)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shl r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8194
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 6(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 2
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 6(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 3
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 31
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 6(r5)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sub r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 31
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Lmin_61
	add r4, r1, r0
.Lmin_61:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 9999
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8196
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 80
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65024
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 128
	call .Lword_26
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r1, r4, .Ltrue_62
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_63
.Ltrue_62:
	addi r4, r0, -1
.Lcmp_end_63:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_64
	jal .Lif_false_65
.Lif_body_64:
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8196
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Ltrue_66
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_67
.Ltrue_66:
	addi r4, r0, -1
.Lcmp_end_67:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_68
	jal .Lif_false_69
.Lif_body_68:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8196
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 31
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 10
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 4
	call .Lword_46
	jal .Lif_end_70
.Lif_false_69:
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_end_70:
	jal .Lif_end_71
.Lif_false_65:
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_end_71:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65456
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65504
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65152
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 96
	call .Lword_26
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r1, r4, .Ltrue_72
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_73
.Ltrue_72:
	addi r4, r0, -1
.Lcmp_end_73:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_74
	jal .Lif_false_75
.Lif_body_74:
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8196
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Ltrue_76
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_77
.Ltrue_76:
	addi r4, r0, -1
.Lcmp_end_77:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_78
	jal .Lif_false_79
.Lif_body_78:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8196
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 6
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 18
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 31
	call .Lword_46
	jal .Lif_end_80
.Lif_false_79:
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_end_80:
	jal .Lif_end_81
.Lif_false_75:
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_end_81:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8194
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	blt r4, r0, .Ltrue_82
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_83
.Ltrue_82:
	addi r4, r0, -1
.Lcmp_end_83:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_84
	jal .Lif_false_85
.Lif_body_84:
	sub r4, r0, r4
	call .Lword_22
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r1, r4, .Ltrue_86
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_87
.Ltrue_86:
	addi r4, r0, -1
.Lcmp_end_87:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_88
	jal .Lif_false_89
.Lif_body_88:
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8196
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	blt r4, r1, .Ltrue_90
	addi r4, r0, 0
	beq r0, r0, .Lcmp_end_91
.Ltrue_90:
	addi r4, r0, -1
.Lcmp_end_91:
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_92
	jal .Lif_false_93
.Lif_body_92:
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8196
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8192
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8238
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65280
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8240
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8192
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 256
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	and r4, r4, r1
	lw r1, 0(r6)
	sw r4, 0(r6)
	add r4, r1, r0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65280
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	mul r2, r1, r4
	mulh r3, r1, r4
	addi r1, r0, 8
	shr r2, r2, r1
	shl r3, r3, r1
	or r4, r2, r3
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 256
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	and r4, r4, r1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	xor r4, r4, r1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_94
	jal .Lif_false_95
.Lif_body_94:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 26
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 25
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 22
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	jal .Lif_end_96
.Lif_false_95:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 10
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 10
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_end_96:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 80
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65024
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 128
	call .Lword_53
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_97
	jal .Lif_false_98
.Lif_body_97:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_false_98:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 0
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65456
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65504
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 65152
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 96
	call .Lword_53
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	bne r1, r0, .Lif_body_99
	jal .Lif_false_100
.Lif_body_99:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shr r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_false_100:
	jal .Lif_end_101
.Lif_false_93:
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_end_101:
	jal .Lif_end_102
.Lif_false_89:
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_end_102:
	jal .Lif_end_103
.Lif_false_85:
	lw r4, 0(r6)
	addi r6, r6, 2
.Lif_end_103:
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8198
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 10
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shl r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8200
	lw r4, 0(r4)
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 5
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	shl r4, r4, r1
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	or r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	li r4, 8202
	lw r4, 0(r4)
	add r1, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	or r4, r4, r1
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 6(r5)
	addi r6, r6, -2
	sw r4, 0(r6)
	lw r4, 2(r5)
	add r2, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	add r3, r4, r0
	lw r4, 0(r6)
	addi r6, r6, 2
	li r1, 7
	shl r3, r3, r1
	add r2, r2, r2
	li r1, 61440
	add r1, r1, r3
	add r1, r1, r2
	sw r4, 0(r1)
	lw r4, 0(r6)
	addi r6, r6, 2
	lw r1, 2(r5)
	addi r1, r1, 1
	sw r1, 2(r5)
	lw r2, 0(r5)
	bge r1, r2, .Lloop_exit_104
	jal .Ldo_60
.Lloop_exit_104:
	addi r5, r5, 4
	lw r1, 2(r5)
	addi r1, r1, 1
	sw r1, 2(r5)
	lw r2, 0(r5)
	bge r1, r2, .Lloop_exit_105
	jal .Ldo_59
.Lloop_exit_105:
	addi r5, r5, 4
	halt
	halt
