\ =====================================================
\ Real Raytracer for NAND-16 (animated orbit camera)
\ Quadratic ray-sphere intersection, Half-Lambert + Phong
\ Two spheres + shadow rays + checkerboard ground
\ 64x32 RGB555, 8.8 fixed-point
\
\ The camera orbits 360 degrees around the scene centre while
\ bobbing up and down, then returns to the start (one full
\ revolution = 24 frames). Each completed frame bumps the frame
\ counter at address 32974 so a host driver can snapshot it.
\
\ Memory map (8.8 fixed-point world units):
\   32750 cos(theta)   32752 sin(theta)
\   32756 ox  32758 oy  32760 oz        ray origin (camera position)
\   32764 dz                          ray direction z (32768 dx, 32770 dy)
\   32818 32820 32822                   sphere-hit oc scratch
\   32974 frame counter
\   32976.. sine table (24 entries, 256*sin(2*pi*k/24))
\ =====================================================

: isqrt dup 2 < if else
  dup 1 rshift over over / + 1 rshift
  over over / + 1 rshift
  over over / + 1 rshift nip then ;
: fsqrt isqrt 4 lshift ;
: clamp0 dup 0< if drop 0 then ;

\ sin@ ( idx -- 256*sin(2pi*idx/24) ) table lookup, idx in 0..23
: sin@ 2 * 32976 + @ ;

\ ground-t ( |dy| -- t ) hit distance to plane y=-128, t = (128 + oy) / |dy|
: ground-t 128 32758 @ + swap f/ ;

\ sphere-hit ( cx cy cz r -- t or 0 )
\ Ray O + t*d against sphere(C,r). Stores hit normal at 32798/32800/32802.
: sphere-hit
  32786 ! 32784 ! 32782 ! 32780 !
  \ a = dot(d,d)
  32768 @ dup f* 32770 @ dup f* + 32764 @ dup f* +  32790 !
  \ oc = O - C
  32756 @ 32780 @ -  32818 !
  32758 @ 32782 @ -  32820 !
  32760 @ 32784 @ -  32822 !
  \ b = dot(oc,d)
  32818 @ 32768 @ f* 32820 @ 32770 @ f* + 32822 @ 32764 @ f* +  32792 !
  \ c = dot(oc,oc) - r^2
  32818 @ dup f* 32820 @ dup f* + 32822 @ dup f* + 32786 @ dup f* -  32794 !
  \ disc = b^2 - a*c
  32792 @ dup f* 32790 @ 32794 @ f* -
  dup 0< if drop 0 else
    fsqrt 32792 @ negate swap - 32790 @ f/   \ t = (-b - sqrt(disc)) / a
    dup 1 < if drop 0 else
      32796 !
      \ normal = (O + t*d - C) / r
      32796 @ 32768 @ f* 32756 @ + 32780 @ - 32786 @ f/ 32798 !
      32796 @ 32770 @ f* 32758 @ + 32782 @ - 32786 @ f/ 32800 !
      32796 @ 32764 @ f* 32760 @ + 32784 @ - 32786 @ f/ 32802 !
      32796 @
    then
  then ;

\ shade ( guard base_r base_g base_b -- )
\ Half-Lambert diffuse + Phong specular (colored, no white additive)
: shade
  32812 ! 32810 ! 32808 ! drop
  32798 @ -148 f* 32800 @ 148 f* + 32802 @ 148 f* +
  256 + 1 rshift 160 f* 96 + 256 min 32804 !
  32798 @ -83 f* 32800 @ 83 f* + 32802 @ 227 f* + clamp0
  dup f* 77 f* 32804 @ + 26 + 256 min 32804 !
  32808 @ 32804 @ f* 31 min 32774 !
  32810 @ 32804 @ f* 31 min 32776 !
  32812 @ 32804 @ f* 31 min 32778 !
;

\ shadow? ( guard cx cy cz r -- flag )
\ Shadow ray from ground hit (32814, -128, 32816) toward light (1,1,1).
\ disc = dot(oc,L)^2 - (dot(oc,oc) - r^2)
: shadow?
  32836 ! 32834 ! 32832 ! 32830 ! drop
  32814 @ 32830 @ - 32840 !
  -128 32832 @ - 32842 !
  32816 @ 32834 @ - 32844 !
  32840 @ dup f* 32842 @ dup f* + 32844 @ dup f* +
  32836 @ dup f* - 32846 !
  32840 @ -148 f* 32842 @ 148 f* + 32844 @ 148 f* +
  dup f* 32846 @ -
  0< if 0 else -1 then
;

\ setup-cam ( -- ) place the camera for the current frame counter (32974).
\ Orbit radius R=448, centre (0,0,-448); yaw = frame*15deg; vertical bob.
: setup-cam
  32974 @ 24 mod          ( idx )
  dup sin@ 32752 !        \ sin(theta)
  6 + 24 mod sin@ 32750 ! \ cos(theta) = sin(theta + 90deg)
  448 32752 @ f*  32756 !          \ ox = R*sin
  -448 448 32750 @ f* +  32760 !   \ oz = -448 + R*cos
  96 32752 @ f*  32758 !           \ oy = bob amplitude * sin
;

\ =================== Sine table init (256 * sin) ===================
0 32976 !    66 32978 !   128 32980 !   181 32982 !   222 32984 !   247 32986 !
256 32988 !  247 32990 !  222 32992 !   181 32994 !   128 32996 !    66 32998 !
0 33000 !   -66 33002 !  -128 33004 !  -181 33006 !  -222 33008 !  -247 33010 !
-256 33012 ! -247 33014 ! -222 33016 ! -181 33018 ! -128 33020 !   -66 33022 !

0 32974 !   \ frame counter

\ =================== Main animation loop ===================
begin
  setup-cam

  32 0 do 64 0 do
    \ Camera-space ray (rx, ry, -1) rotated by yaw into world space.
    i 32 - 3 lshift                  ( rx )
    16 j - 3 lshift 32770 !           \ dy = ry (yaw leaves y untouched)
    dup 32750 @ f* 32752 @ -  32768 !   \ dx = rx*cos - sin
    32752 @ f* negate 32750 @ -  32764 ! \ dz = -(rx*sin) - cos

    j 2 rshift 32774 !  j 3 rshift 32776 !  31 j - 31 min 32778 !
    9999 32772 !

    \ Sphere 1: warm (80,0,-512) r=128
    80 0 -512 128 sphere-hit
    dup 0 > if dup 32772 @ < if
      32772 !  0 31 10 4 shade
    else drop then else drop then

    \ Sphere 2: cool (-80,-32,-384) r=96
    -80 -32 -384 96 sphere-hit
    dup 0 > if dup 32772 @ < if
      32772 !  0 6 18 31 shade
    else drop then else drop then

    \ Ground plane y=-128
    32770 @ dup 0< if
      negate ground-t
      dup 0 > if dup 32772 @ < if
        dup 32772 !
        dup 32768 @ f* 32756 @ + 32814 !   \ hit.x = ox + t*dx
        dup 32764 @ f* 32760 @ + 32816 !   \ hit.z = oz + t*dz
        drop
        32814 @ 256 and  32816 @ 256 and
        xor if  26 32774 ! 25 32776 ! 22 32778 !
        else    10 32774 ! 10 32776 ! 8 32778 !
        then
        \ Shadow from sphere 1
        0 80 0 -512 128 shadow?
        if  32774 @ 1 rshift 32774 !
            32776 @ 1 rshift 32776 !
            32778 @ 1 rshift 32778 !
        then
        \ Shadow from sphere 2
        0 -80 -32 -384 96 shadow?
        if  32774 @ 1 rshift 32774 !
            32776 @ 1 rshift 32776 !
            32778 @ 1 rshift 32778 !
        then
      else drop then else drop then
    else drop then

    32774 @ 10 lshift 32776 @ 5 lshift or 32778 @ or
    j i pixel16
  loop loop

  32974 @ 1 + 32974 !   \ frame complete; advance counter
again
