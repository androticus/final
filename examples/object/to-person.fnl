forget: reset-to-person
`
Example of creating a Final TypedObject from a schema object.

Try:
to-person
` const: to-person-blurb

// Step 1. get a prototype DO to fill, park on ys, name it

TypedObject.newTypeDef^ >ys

"Person" y:.name!

// Step 2. define the fields, by setting with the type

y:.fields@ >xs
String dup x:.firstName!
           x:.lastName!
Int        x:.age!
x:drop

// Step 2b. class constants (value .name!)

y:.constants@ >xs
111 x:.defaultAge!
x:drop

// Step 3. seal the type

ys> TypedObject.makeType^

// instance method
// o( inst:to )
: Person#print
  ( o:.firstName@ o:.lastName@ o:.age@ ) "%s %s (%d)\n" fmt $
;

// class method — o: is the Type
: Person##newDefault
  o:.new^ >ts
  "anon" t:.firstName!
  "" t:.lastName!
  Person.defaultAge@ t:.age!
  ts>
;

// Step 4. try it out

: to-person
  "Default age: " $ Person.defaultAge@ $ nl

  Person.newDefault^
  >ys
  y:.print^

  "Bilbo" y:.firstName!
  "Baggins" y:.lastName!
  0 y:@() $ sp
  1 y:@() $ nl
  y:.print^
  y:drop
;

to-person-blurb $ nl
