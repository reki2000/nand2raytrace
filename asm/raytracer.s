\ =====================================================
\ Real Raytracer for NAND-16
\ Quadratic ray-sphere intersection, Half-Lambert + Phong
\ Two spheres + shadow rays + checkerboard ground
\ 64x32 RGB555, 8.8 fixed-point
\ =====================================================

: isqrt dup 2 < if else
  dup 1 rshift over over / + 1 rshift
  over over / + 1 rshift
  over over / + 1 rshift nip then ;
: fsqrt isqrt 4 lshift ;
: clamp0 dup 0< if drop 0 then ;

\ ground-t ( |ry| -- t )
: ground-t 128 swap f/ ;

\ sphere-hit ( cx cy cz r -- t or 0 )
\ Stores normal at 8222/8224/8226 on hit
: sphere-hit
  8210 ! 8208 ! 8206 ! 8204 !
  8210 @ dup f* 8212 !
  8192 @ dup f* 8194 @ dup f* + 256 + 8214 !
  0 8204 @ - 8192 @ f*
  0 8206 @ - 8194 @ f* +
  0 8208 @ - -256 f* +  8216 !
  8204 @ dup f* 8206 @ dup f* + 8208 @ dup f* + 8212 @ - 8218 !
  8216 @ dup f* 8214 @ 8218 @ f* -
  dup 0< if drop 0 else
    fsqrt 8216 @ negate swap - 8214 @ f/
    dup 1 < if drop 0 else
      8220 !
      8220 @ 8192 @ f* 8204 @ - 8210 @ f/ 8222 !
      8220 @ 8194 @ f* 8206 @ - 8210 @ f/ 8224 !
      8220 @ -256  f* 8208 @ - 8210 @ f/ 8226 !
      8220 @
    then
  then ;

\ shade ( guard base_r base_g base_b -- )
\ Half-Lambert diffuse + Phong specular (colored, no white additive)
: shade
  8236 ! 8234 ! 8232 ! drop
  8222 @ -148 f* 8224 @ 148 f* + 8226 @ 148 f* +
  256 + 1 rshift 160 f* 96 + 256 min 8228 !
  8222 @ -83 f* 8224 @ 83 f* + 8226 @ 227 f* + clamp0
  dup f* 77 f* 8228 @ + 26 + 256 min 8228 !
  8232 @ 8228 @ f* 31 min 8198 !
  8234 @ 8228 @ f* 31 min 8200 !
  8236 @ 8228 @ f* 31 min 8202 !
;

\ shadow? ( guard cx cy cz r -- flag )
\ disc = dot(oc,L)^2 - (dot(oc,oc) - r^2)
: shadow?
  8260 ! 8258 ! 8256 ! 8254 ! drop
  8238 @ 8254 @ - 8264 !
  -128 8256 @ - 8266 !
  8240 @ 8258 @ - 8268 !
  8264 @ dup f* 8266 @ dup f* + 8268 @ dup f* +
  8260 @ dup f* - 8270 !
  8264 @ -148 f* 8266 @ 148 f* + 8268 @ 148 f* +
  dup f* 8270 @ -
  0< if 0 else -1 then
;

\ =================== Main render loop ===================

32 0 do 64 0 do
  i 32 - 3 lshift 8192 !
  16 j - 3 lshift 8194 !
  j 2 rshift 8198 !  j 3 rshift 8200 !  31 j - 31 min 8202 !
  9999 8196 !

  \ Sphere 1: warm (80,0,-512) r=128
  80 0 -512 128 sphere-hit
  dup 0 > if dup 8196 @ < if
    8196 !  0 31 10 4 shade
  else drop then else drop then

  \ Sphere 2: cool (-80,-32,-384) r=96
  -80 -32 -384 96 sphere-hit
  dup 0 > if dup 8196 @ < if
    8196 !  0 6 18 31 shade
  else drop then else drop then

  \ Ground plane y=-128
  8194 @ dup 0< if
    negate ground-t
    dup 0 > if dup 8196 @ < if
      dup 8196 !
      dup 8192 @ f* 8238 !
      dup -256 f* 8240 !
      dup 8192 @ f* 256 and
      swap -256 f* 256 and
      xor if  26 8198 ! 25 8200 ! 22 8202 !
      else    10 8198 ! 10 8200 ! 8 8202 !
      then
      \ Shadow from sphere 1
      0 80 0 -512 128 shadow?
      if  8198 @ 1 rshift 8198 !
          8200 @ 1 rshift 8200 !
          8202 @ 1 rshift 8202 !
      then
      \ Shadow from sphere 2
      0 -80 -32 -384 96 shadow?
      if  8198 @ 1 rshift 8198 !
          8200 @ 1 rshift 8200 !
          8202 @ 1 rshift 8202 !
      then
    else drop then else drop then
  else drop then

  8198 @ 10 lshift 8200 @ 5 lshift or 8202 @ or
  j i pixel16
loop loop halt
