package core

// Manually compiled core words (if/then/loop) live here until bootstrapInitScript
// defines them in Final source. See docs/003-design-notes/_archived/106-lbforth-bootstrap-pattern.md.

/*
From Grok (Forth names — Final init script will use lowercase):

: >mark    ( -- orig )   here 0 , ;
: >resolve ( orig -- )   here swap ! ;

: if       [[ branch? ]] >mark ; immediate
: then     >resolve ; immediate
...
*/
