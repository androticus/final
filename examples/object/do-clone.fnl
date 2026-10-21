/*
Example of DynamicObject.clone
x: prototype
y: clone1
z: clone2

Demonstrates we can set y: and z: props independently, with no change to x:
Shows that .clone does work, and object has the fields of prototype.
*/

: testDoClone
  newdo >xs
  // prototype, just default values
  "" x:.string!
  0  x:.int!
    
  // print the prototype
  "prototype:\n" $
  x:peek >jsonp $ nl

  // make clone y
  x:.clone^ >ys
    
  // print cloned y, to show it is initialized
  "y: (just after clone)\n" $
  y:peek >jsonp $ nl
    
  // mutate y
  "y mutate" y:.string!
  1 y:.int!
    
  // make clone z
  x:.clone^ >zs
  
  // mutate z
  "z mutate" z:.string!
  2 z:.int!
    
  // json print all 3, to show there was no leaking/overwriting
  "prototype:\n" $
  xs> >jsonp $ nl
    
  "y:\n" $
  ys> >jsonp $ nl
    
  "z:\n" $
  zs> >jsonp $ nl
;
