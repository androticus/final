# Compiler

The Final compiler creates new compiled words from existing words. Final code gets compiled into a Code Block (CT) table that contains the OpCodes of Final axioms, as well as some prefix modifiers, and immediate values to an OpCode, branches, and so on. The code is then exececuted in a tight run loop with an IP pointing the current instruction. OpCodes get dispatched via an efficient Go-compiled jump table. The resulting execution is highly efficient and fast, and for some operations can rival the performance of natively compiled machine code, due to CPU code and data caching. For details on how Final works, see [How Final Works](../how-final-works.md).

## Final Compiler

The Final Compiler is used to compile a new colon word. It follows rules somewhat parallel to that of the Interpreter. Here is how it handles tokens;
* Comments are dropped and ignored.
* A literal value is compiled into the code as a literal; at runtime, the value gets pushed on the DS
* a compound path accessor is broken into its parts, and compiled into the CT
* an identifier is looked up as a word  
  if found, it is compiled into the word as a Nest call to that word's DC/CT index  
  if not found, an error is raised, and compilation stops

Some words, such as control flow words, have a more complex compile-time behavior; these words combine immediate interpretive behavior, along with emission of compiled code. For example, control flow constructs need to keep track of the IP address during compilation, place dummy values into the CT that will hold branch addresses, then resolve those branches when closing words are reached. But the outcome is the same: directly executed compiled code is what executes at runtime. 

Speaking of control flow words.. enough theory! Let's move on to a quick overview of Basic Operators, then from there to control flow words, along with examples.