package core

// Contains the OpCodes and Dict entry of every Final axiom
// Values MUST be an exact range from 0..N-1, where N is the number of axioms
// Dictionary entries MUST be made in OpCode ordinal order
// Dictionary[OpCode] = DictEntry
// OpCode is as a value in the 	Execution Table
// A switch in Run() uses the OpCode to execute the appropriate axiom
// OpCode is a type alias for int (CT cells, switch dispatch). Debug/lookup names come from
// registerOpCode in regAxiom during registerAxioms — not stringer (aliases cannot have methods).
type OpCode = int

const (
	OpINVALID OpCode = iota // protects from 0 execution; panics in the VM
	OpStop                  // stops the execution of the current task, mostly used for testing

	// stack
	OpDup     // c( a -- a a ) duplicate the top of the DS
	OpDDup     // ( a -- a a ) duplicate the top of the DS
	OpQDup    // c( a -- a a | a) conditional duplicate the top of the DS if truthy
	OpDup2    // c( a b -- a b a b ) duplicate the top two elements of the DS
	OpDDup2    // ( a b -- a b a b ) duplicate the top two elements of the DS
	OpDrop    // c( a -- ) discard the top of the DS
	OpDDrop    // ( a -- ) discard the top of the DS
	OpDrop2   // c( a b -- ) discard the top two elements of the DS
	OpDDrop2   // ( a b -- ) discard the top two elements of the DS
	OpSwap    // c( a b -- b a ) swap the top two elements of the DS
	OpDSwap    // ( a b -- b a ) swap the top two elements of the DS
	OpOver    // ( a b -- a b a ) copy the second element of the DS to the top
	OpNip     // c( a b -- b ) discard the second element of the DS
	OpTuck    // c( a b -- b a b ) copy the top of the DS to third-from-top
	OpPick    // c( z..a +n:iu -- z..a [n] ) copy the nth element of the DS to the top
	OpRot     // ( a b c -- b c a ) rotate the top three elements of the DS
	OpDStoTS  // ( a -- ) t( -- a) move the top of the DS to the TS
	OpTStoDS  // ( -- a) t( a -- ) move the top of the TS to the DS
	OpTDrop   // t( -- ) discard temp stack TOS
	OpRPeek   // ( -- cfm:ColonFrameMark) r( cfm:ColonFrameMark ) peek the top of the RS
	OpXStoDS  // ( -- a ) xs( a -- ) move the top of the XS to the DS
	OpDStoXS  // ( a -- ) xs( -- a) move the top of the DS to the XS
	OpXDrop   // ( -- ) xs( any -- ) discard the top of the XS
	OpYStoDS  // ( -- a ) ys( a -- ) move the top of the YS to the DS
	OpDStoYS  // ( a -- ) ys( -- a) move the top of the DS to the YS
	OpYDrop   // ( -- ) ys( any -- ) discard the top of the YS
	OpZStoDS  // ( -- a ) zs( a -- ) move the top of the ZS to the DS
	OpDStoZS  // ( a -- ) zs( -- a) move the top of the DS to the ZS
	OpZDrop   // ( -- ) zs( any -- ) discard the top of the ZS
	OpDepth   // ( -- n:Int ) depth of the indicated stack
	OpIsTyped // ( value:Any -- value b:b ) check if value is typed
	OpDSto    // cs( x:any -- ) ( -- x ) move top of CS to DS
	OpToDS    // ( x:any -- ) cs( -- x ) move top of DS to CS
	OpPeek    // cs( x:any -- x ) ( -- x ) peek top of CS
	OpODrop   // ( -- ) o( any -- ) discard the top of the OS

	// bitwise
	OpBitAnd    // ( a:iu b:iu -- c:iu ) bitwise and
	OpBitOr     // ( a:iu b:iu -- c:iu ) bitwise or
	OpBitXor    // ( a:iu b:iu -- c:iu ) bitwise xor
	OpBitNot    // ( a:iu -- c:iu ) bitwise not
	OpBitSet    // ( flags:iu mask:iu -- c:iu ) bitwise set
	OpBitClear  // ( flags:iu mask:iu -- c:iu ) bitwise clear
	OpBitToggle // ( flags:iu mask:iu -- c:iu ) bitwise toggle

	// boolean (poly)
	OpBoolAnd // ( a:t b:t -- c:b ) boolean and
	OpBoolOr  // ( a:t b:t -- c:b ) boolean or
	OpBoolNot // ( a:t -- c:b ) boolean not

	// equatable (poly)
	OpEqual        // ( a:e b:e -- c:b ) equal
	OpZeroEqual    // ( a:e -- c:b ) equal to zero
	OpNotEqual     // ( a:e b:e -- c:b ) not equal
	OpZeroNotEqual // ( a:e -- c:b ) not equal to zero

	// comparable (poly)
	OpLessThan               // ( a:c b:c -- c:b ) less than
	OpLessThanOrEqual        // ( a:c b:c -- c:b ) less than or equal
	OpGreaterThan            // ( a:c b:c -- c:b ) greater than
	OpGreaterThanOrEqual     // ( a:c b:c -- c:b ) greater than or equal
	OpCompare                // ( a:c b:c -- c:i ) compare → −1 / 0 / +1
	OpZeroLessThan           // ( a:e -- c:b ) less than zero
	OpZeroLessThanOrEqual    // ( a:e -- c:b ) less than or equal to zero
	OpZeroGreaterThan        // ( a:e -- c:b ) greater than zero
	OpZeroGreaterThanOrEqual // ( a:e -- c:b ) greater than or equal to zero

	// arith (poly)
	OpAdd // ( a:n b:n -- c:* ) poly: add two numbers
	OpSub // ( a:n b:n -- c:* ) poly: subtract two numbers
	OpMul // ( a:n b:n -- c:* ) poly: multiply two numbers
	OpDiv // ( a:n b:n -- c:* ) poly: divide two numbers
	OpAbs // ( a:n -- b:n ) absolute value
	OpNeg // ( a:n -- b:n ) negate (Uint panics)

	// integral
	OpShiftLeft   // ( value:iu by:iu -- res:iu ) logical shift left by count
	OpShiftRight  // ( value:iu by:iu -- res:iu ) shift right; arithmetic if Int, logical if Uint
	OpRotateLeft  // ( value:iu by:iu -- res:iu ) rotate left; result type matches value
	OpRotateRight // ( value:iu by:iu -- res:iu ) rotate right; result type matches value
	OpRem         // ( dividend:iu divisor:iu -- rem:iu ) truncated remainder
	OpModRem      // ( dividend:iu divisor:iu -- quot:iu rem:iu ) integer quotient (/) and truncated rem (%)

	// control (colon words; global CT only in v1)
	// literals
	OpCompileLit // (v:Any -- ) c( -- op v) compile literal to CT; op is the OpCode to execute
	OpDoLit0     // c(_ -- ) ( -- 0 ) load 0
	OpDoLitB     // c(_ v:b-- ) ( -- v ) load Bool literal from CT (immediate)
	OpDoLitI     // c(_ v:i-- ) ( -- v ) load Int literal from CT (immediate)
	OpDoLitU     // c(_ v:u-- ) ( -- v ) load Uint literal from CT (immediate)
	OpDoLitF     // c(_ v:f-- ) ( -- v ) load Float literal from CT (immediate)
	OpDoLitNull  // c(_ -- ) ( -- nil ) load untyped null
	OpDoLit      // c(_ j:dtIdx -- ) ( -- v ) load literal from DT

	OpNest             // c(_ j:entry -- ) r( -- ths ret:VA ) i( -- entry ) subroutine enter
	OpUnnest           // c(_ -- ) r(ths ret:VA -- ths ) i( -- ret ) subroutine return
	OpInterrupt        // (_ moduleID:i entry:i -- ) nest from Go (Final method / interrupt); OS cleanup via compiled odrop
	OpDefer            // s( dtIdx:i -- xt ) DT[dtIdx] → XT on DS; reenter OpExec
	OpBootstrapDefer   // s( dtIdx:i -- xt ) core DT[slot] → XT on DS; reenter OpExec
	OpExec             // c(_ -- ) ( xt:XT -- ) dynamic XT dispatch (RS/ip same as Nest)
	OpExecWord         // ( w:Word -- ) run word via standard exec path, with arming
	OpBranch           // c(_ j:i -- ) ( -- ) relative branch to j
	OpBranchQ          // c(_ j:i -- ) ( a:t-- ) relative branch to j if truthy
	OpBrSub            // c(_ j:i -- ) push return IP on frame subRets, relative branch (subword call)
	OpRetSub           // c(_ -- ) pop return IP from frame subRets (subword return)
	OpTryGetIter       // ( provider:Iterable? -- ? ) ( -- iter true) || ( -- providerfalse ) foreach prologue
	OpForeachLoopTop   // r( cm:ColonFrameMark -- cm ) Slot.Foreach TOS getNext → ds true|false; epilogue drops iter
	OpI                // r( cm:ColonFrameMark -- cm ) ( -- curr ); IteratorAt(0).Current
	OpJ                // r( cm:ColonFrameMark -- cm ) ( -- curr ); IteratorAt(1).Current
	OpK                // r( cm:ColonFrameMark -- cm ) ( -- curr ); IteratorAt(2).Current
	OpPushIterator     // ( iter:Iterator -- ) r( cm:ColonFrameMark -- cm ); PushIterator on RS TOS ColonFrameMark
	OpDropIterator     // r( cm:ColonFrameMark -- cm ); DropIterator on RS TOS ColonFrameMark
	OpLocals           // r( cm:ColonFrameMark -- cm ) ( -- locals:DynamicObject ); GetLocals
	OpPushSwitch       // ( target:a -- ) r( cm -- cm ); PushSwitchTarget
	OpDropSwitch       // r( cm -- cm ); DropSwitchTarget
	OpSwitchA          // r( cm -- cm ) ( -- target ); SwitchTargetAt(0)
	OpSwitchB          // r( cm -- cm ) ( -- target ); SwitchTargetAt(1)
	OpSwitchC          // r( cm -- cm ) ( -- target ); SwitchTargetAt(2)
	OpAssertSwitchNest // t( -- ); panic if opening switch not outside or inside case/default

	// compile-time CT helpers ([imm … imm] interpret islands; global CT only in v1)
	OpCTHere             // c( -- here:i ) push next CT append index (current Len)
	OpCTComma            // c( i:Int -- ) append raw int cell to CT ([]int; branch offset placeholder, etc.)
	OpCTStore            // c( val:i index:i -- ) Write CT[index] = val
	OpCTPatchForward     // c( orig:i -- ) j←here-(orig+1); CT[orig]←j; pop orig (forward >resolve)
	OpCTPatchForwardNext // c( orig:i -- ) j←(here+1)-(orig+1); CT[orig]←j; loop break exit (>resolve after branch emit)

	OpSetWordFlags // ( flags:iu -- ) set WordFlags on most recent colon dict entry

	OpTransition // WKNative runtime dispatch; uses runSubCode as dispatch id

	OpFetch // @: cs( ValueRef -- value ); consumes ValueRef in prologue (R-M-W: dup @ … !)
	OpStore // !: cs( value ValueRef -- ); value under ref; ref on CS TOS (Forth addr!); consumes both in prologue

	OpDoVar          // ( dtIdx:i -- ref:ValueRef ) pop slot index from DS; push DTSlotRef on CS
	OpBootstrapDoVar // ( dtIdx:i -- ref:ValueRef ) core DT[slot] → DTSlotRef on CS

	// compiling and defining words
	OpCtImmDs // c(_ idx:i -- ) ( -- idx:i ) read next CT cell, push dtIdx on DS; prefix for dtIdx workers

	OpS     // ( a:any -- ) prints a as a naturalString to stdout
	OpSS    // ( a:any -- ) prints a as a naturalString to stderr
	OpRtoS  // ( rune:i -- s ) Int code point → one-character String
	OpDsS   // bootstrap debug: ( -- ) print DS deepest→TOS to stdout (-> .); not official language
	OpStarS // debug: ( -- ) print RS/DS/TS snapshot to stdout (fault-report layout)

	OpHandleCompound // s( path:s -- … ) compound demux; compile: OpDoLit + OpHandleCompound

	OpTypeMemberDispatch // s( index:i -- … ) uses runSubCode as dispatch kind; CS TOS is ths

	OpNativeInvoke // cs( ths -- ths ) RuntimeDispatch.Invoke on CS TOS; kind from runSubCode

	OpAtIndex    // cs( receiver index -- receiver value ) operand index from DS; value on CS (@())
	OpBangIndex  // cs( receiver value index -- receiver ) operands from DS (!())
	OpAtBrace    // cs( receiver key -- receiver value ) operand key from DS; value on CS (@{})
	OpBangBrace  // cs( receiver value key -- receiver ) operands from DS (!{})
	OpQueryBrace // cs( receiver key -- receiver has:b ) operand key from DS; has:b on DS (?{})
	// OpAtMulti / OpBangMulti registered before OpNop (end of const block).

	OpMethodInvoke // cs( MethodRef -- … ) consumes MethodRef in prologue, then dispatch (^)

	OpGetDictionaryContext // s( -- dc:DictionaryContext ) push task compile-target DC (task.dc), not execDC
	OpGetSession    // s( -- ses:Session ) push current Session onto DS
	OpGetLogger     // s( -- logger:Logger ) push current Logger onto DS
	OpLogAt         // s(msg:s data:a level:i  -- ) log a message at the given level
	OpGetSys        // s( -- sys:Sys ) push global Sys onto DS
	
	// context stack selectors
	OpCompileCSSelect // ( idx:i -- ct( op idx -- ) compile CS selector prefix;
	OpDoCSSelect      // ct( _ selector -- ) set CS to selector stack; prefix for context-stack-aware ops
	OpCSSelectFromDS  // s( idx:i -- ) set CS from DS operand (interpret / interactive helper)

	OpCompileSubCode // ( code:i -- ct( op code -- ) compile subCode prefix
	OpDoSubCode      // ct( _ code -- ) set runSubCode from CT operand
	OpSubCodeFromDS  // s( code:i -- ) set runSubCode from DS operand
	OpPushType       // s( -- typ:*Type ) push sealed type from runSubCode via DC TypeRegistrar
	OpError          // s( error:String -- ) panics with error

	OpAllotDtSlotToWord // s( def:any -- ) allot next DT slot, DT[idx]=def, PeekBuilding().dtIdx=idx
	OpDTRef             // s( dtIdx:i -- ref ) push mutable DT value ref (NewDTRef) for DT[dtIdx]
	OpDTConst           // s( dtIdx:i -- value ) push DT[dtIdx]

	// compile block state
	OpPushBlock             // c( kind:i -- ) t( -- bt ) push block token onto ts
	OpAcceptDropBlock       // c( kind:i -- ) t( bt -- ? ) accept block, and drop if found
	OpPanicDropBlock        // c( msg:s kind:i -- ) t( bt -- ) panic if block was not found, else drop it
	OpPanicRequireBlock     // c( msg:s kind:i -- ) t( bt -- ) panic if block was not found, but don't drop it
	OpPanicRequireNKind     // c( n:i msg:s kind:i -- ) t( -- ) panic unless n blocks of kind under colon; no-op inside :s
	OpPanicRequireOneOfKind // c( kinds:List msg:s -- ) t( -- ) panic unless innermost allow-listed kind under colon
	OpSetBlockMark          // c( mark:i -- ) t( bt -- bt ) set continue mark on TS TOS BlockToken
	OpContinueMark          // c( -- dest:i ) t( -- ) push continue CT mark from innermost foreach/forever

	OpRange                // ( start:i end:i -- r:Range ) make/push Range object onto ds
	OpRangeBy              // ( beg:n end:n step:n -- r:RangeBy ) make/push RangeBy onto ds
	OpGetCompileStateFlags // ( -- state:CompileStateFlags ) get compile state flags onto ds
	OpSetCompileStateFlags // ( state:CompileStateFlags -- ) set compile state flags from ds

	OpRecurse // c( -- ) emit OpNest to PeekBuilding().colonCode (compile-time self-call)

	OpDoEmit // c(_ xt:VA -- ) compile word at XT into open colon (Word.CompileWord)

	OpMarkContext    // r( cm:ColonFrameMark -- cm ) push Slot.ContextMark stack (ensure floor if needed)
	OpParen          // alias OpMarkContext
	OpBrace          // alias OpMarkContext
	OpBracket        // alias OpMarkContext
	OpGetContextMark // r( … cm:ColonFrameMark … -- … ) cs( -- mark ); walk ColonFrames, pop ContextMark TOS

	OpMathMonadic // ( a:n -- b:f ) apply MathMonadicFn to TOS

	OpDtDyadic // ( op1 op2 -- result ) DateTime/Interval dyadic via subcode (dt::+ …)

	OpStrOp // ( … -- … ) String stack ops via subcode (str::len …)

	OpPipeOp // ( … -- … ) Pipe stack ops via subcode (pipe::new, >pipe, …)

	OpCast // s( v:a -- b ) type probe; s( v:a -- v':t true | false ) conversion — subcode selects (Int?, >Int?, >>Int?, …)

	OpDelayMs // ( delayMs:i -- ) delays for the specified number of milliseconds
	OpDelayUs // ( delayUs:i -- ) delays for the specified number of microseconds

	OpSpawnPip // ( pp:AtomicDynamicObject xt:XT -- ) launch child pip; VirtXT is TOS

	OpCompileForget // c( -- ) stamp dict/CT/DT/type keeps as lits + emit doForget (mark create)
	OpDoForget      // ( dictKeep ctKeep dtKeep typeKeep -- ) inclusive forget to keeps

	OpAtMulti   // cs( receiver idx… -- receiver value ) N indexes from DS; value on CS (@[])
	OpBangMulti // cs( receiver value idx… -- receiver ) value + N indexes from DS (![])

	// LA shape families — subcode selects the op within each table (mat:: / vec::).
	OpVecUnary   // ( v -- … ) VecNf unary
	OpVecVec     // ( u v -- … ) VecNf × VecNf
	OpVecScalar  // ( v s:f -- … ) VecNf × Float
	OpMatUnary   // ( A -- … ) Mat2f unary
	OpMatMat     // ( A B -- … ) Mat2f × Mat2f
	OpMatVec     // ( A v -- … ) Mat2f × VecNf
	OpMatScalar  // ( A s:f -- … ) Mat2f × Float
	OpMatIndex   // ( A i:i -- … ) Mat2f × Int (row/col)
	OpMatFactory // ( n:i -- I ) / factory shapes as added

	OpPushCatchMark // c(_ catchJ:i -- ) r( cm:ColonFrameMark -- cm ); PushCatchMark on Slot (nested try = stack)
	OpDropCatchMark // r( cm:ColonFrameMark -- cm ); soft PopCatchMark if present (normal /try; no-op after throw)
	OpThrow         // s( msg:String -- ) unwind to CatchMark or panic; after restore msg on DS

	OpPls // ( -- pls:DynamicObject ) push pls:DynamicObject onto DS

	OpSetOp // ( a:Set b:Set -- c:Set ) dyadic set ops via subcode (set::union …); same sealed type both sides

	OpUsing        // ( prefixes:StringList -- ) put listed module prefixes at front of currentSearchOrder
	OpUsingDefault // ( -- ) reset currentSearchOrder to Session.modules

	OpInstall     // ( name:s opts:DynamicObject -- ) leaf install via _REPLACES_/path; Seal + Module table
	OpImport      // ( syn:s name:s -- ) attach already-installed module to Session (first-attach wins)
	OpLoadProject // ( projectPath:s -- ) load .fnlp: requires install+import, compile files into Local

	OpNop // ( -- ) do nothing; traditionally always last: "best for last"

	opAxiomLAST = OpNop
	opLAST      = OpNop
)

func registerAxioms() []*Word {
	resetOpCodeNames()
	d := make([]*Word, 0, opAxiomLAST+1)
	// NOTE: compiler never compiles OpINALID; impossible to compile or execute
	regAxiom(&d, "INVALID", "_INVALID", OpINVALID, WordFlagCompileOnly|WordFlagCompilerOnly)
	regAxiom(&d, "Stop", "stop", OpStop, 0)

	// stack
	const csManip = WordFlagRequiresStackContext | WordFlagOSContextRestrict
	regAxiom(&d, "Dup", "dup", OpDup, csManip)
	regAxiom(&d, "DDup", "ddup", OpDDup, 0)
	regAxiom(&d, "QDup", "?dup", OpQDup, csManip)
	regAxiom(&d, "Dup2", "dup2", OpDup2, csManip)
	regAxiom(&d, "DDup2", "ddup2", OpDDup2, 0)
	regAxiom(&d, "Drop", "drop", OpDrop, csManip)
	regAxiom(&d, "DDrop", "ddrop", OpDDrop, 0)
	regAxiom(&d, "Drop2", "drop2", OpDrop2, csManip)
	regAxiom(&d, "DDrop2", "ddrop2", OpDDrop2, 0)
	regAxiom(&d, "Swap", "swap", OpSwap, csManip)
	regAxiom(&d, "DSwap", "dswap", OpDSwap, 0)
	regAxiom(&d, "Over", "over", OpOver, csManip)
	regAxiom(&d, "Nip", "nip", OpNip, csManip)
	regAxiom(&d, "Tuck", "tuck", OpTuck, csManip)
	regAxiom(&d, "Pick", "pick", OpPick, csManip)
	regAxiom(&d, "Rot", "rot", OpRot, csManip)
	regAxiom(&d, "DStoTS", ">ts", OpDStoTS, 0)
	regAxiom(&d, "TStoDS", "ts>", OpTStoDS, 0)
	regAxiom(&d, "TDrop", "tdrop", OpTDrop, 0)
	regAxiom(&d, "RPeek", "rpeek", OpRPeek, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly|WordFlagSystem)
	regAxiom(&d, "XStoDS", "xs>", OpXStoDS, 0)
	regAxiom(&d, "DStoXS", ">xs", OpDStoXS, 0)
	regAxiom(&d, "XDrop", "xdrop", OpXDrop, 0)
	regAxiom(&d, "YStoDS", "ys>", OpYStoDS, 0)
	regAxiom(&d, "DStoYS", ">ys", OpDStoYS, 0)
	regAxiom(&d, "YDrop", "ydrop", OpYDrop, 0)
	regAxiom(&d, "ZStoDS", "zs>", OpZStoDS, 0)
	regAxiom(&d, "DStoZS", ">zs", OpDStoZS, 0)
	regAxiom(&d, "ZDrop", "zdrop", OpZDrop, 0)
	regAxiom(&d, "Depth", "depth", OpDepth, WordFlagRequiresStackContext)

	regAxiom(&d, "IsTyped", "isTyped", OpIsTyped, 0)
	regAxiom(&d, "DSto", "ds>", OpDSto, csManip)
	regAxiom(&d, "ToDS", ">ds", OpToDS, csManip)
	regAxiom(&d, "Peek", "peek", OpPeek, WordFlagRequiresStackContext)
	regAxiom(&d, "ODrop", "odrop", OpODrop, WordFlagHidden)

	// bitwise
	regAxiom(&d, "BitAnd", "&", OpBitAnd, 0)
	regAxiom(&d, "BitOr", "|", OpBitOr, 0)
	regAxiom(&d, "BitXor", "xor", OpBitXor, 0)
	regAxiom(&d, "BitNot", "bitnot", OpBitNot, 0)
	regAxiom(&d, "BitSet", "bitset", OpBitSet, 0)
	regAxiom(&d, "BitClear", "bitclear", OpBitClear, 0)
	regAxiom(&d, "BitToggle", "bittoggle", OpBitToggle, 0)

	// boolean (poly)
	regAxiom(&d, "BoolAnd", "&&", OpBoolAnd, 0)
	regAxiom(&d, "BoolOr", "||", OpBoolOr, 0)
	regAxiom(&d, "BoolNot", "not", OpBoolNot, 0)

	// equatable (poly)
	regAxiom(&d, "Equal", "==", OpEqual, 0)
	regAxiom(&d, "ZeroEqual", "==0", OpZeroEqual, 0)
	regAxiom(&d, "NotEqual", "<>", OpNotEqual, 0)
	regAxiom(&d, "ZeroNotEqual", "<>0", OpZeroNotEqual, 0)

	// comparable (poly)
	regAxiom(&d, "LessThan", "<", OpLessThan, 0)
	regAxiom(&d, "LessThanOrEqual", "<=", OpLessThanOrEqual, 0)
	regAxiom(&d, "GreaterThan", ">", OpGreaterThan, 0)
	regAxiom(&d, "GreaterThanOrEqual", ">=", OpGreaterThanOrEqual, 0)
	regAxiom(&d, "Compare", "compare", OpCompare, 0)
	regAxiom(&d, "ZeroLessThan", "<0", OpZeroLessThan, 0)
	regAxiom(&d, "ZeroLessThanOrEqual", "<=0", OpZeroLessThanOrEqual, 0)
	regAxiom(&d, "ZeroGreaterThan", ">0", OpZeroGreaterThan, 0)
	regAxiom(&d, "ZeroGreaterThanOrEqual", ">=0", OpZeroGreaterThanOrEqual, 0)

	// arith (poly)
	regAxiom(&d, "Add", "+", OpAdd, 0)
	regAxiom(&d, "Sub", "-", OpSub, 0)
	regAxiom(&d, "Mul", "*", OpMul, 0)
	regAxiom(&d, "Div", "/", OpDiv, 0)
	regAxiom(&d, "Abs", "abs", OpAbs, 0)
	regAxiom(&d, "Neg", "neg", OpNeg, 0)

	// integral
	regAxiom(&d, "ShiftLeft", "<<", OpShiftLeft, 0)
	regAxiom(&d, "ShiftRight", ">>", OpShiftRight, 0)
	regAxiom(&d, "RotateLeft", "rotl", OpRotateLeft, 0)
	regAxiom(&d, "RotateRight", "rotr", OpRotateRight, 0)
	regAxiom(&d, "Rem", "%", OpRem, 0)
	regAxiom(&d, "ModRem", "modrem", OpModRem, 0)

	// control
	regAxiom(&d, "CompileLit", "literal,", OpCompileLit, WordFlagCompileOnly|WordFlagImmediate)
	regAxiom(&d, "DoLit0", "_do-lit0", OpDoLit0, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "DoLitB", "_do-litb", OpDoLitB, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "DoLitI", "_do-liti", OpDoLitI, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "DoLitU", "_do-litu", OpDoLitU, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "DoLitF", "_do-litf", OpDoLitF, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "DoLitNull", "_do-litnull", OpDoLitNull, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "DoLit", "_do-lit", OpDoLit, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "Nest", "_nest", OpNest, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "Unnest", "_unnest", OpUnnest, WordFlagHidden|WordFlagCompilerOnly|WordFlagSystem)
	regAxiom(&d, "Interrupt", "interrupt", OpInterrupt, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "Defer", "defer", OpDefer, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "BootstrapDefer", "_bootstrapDefer", OpBootstrapDefer, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "Exec", "exec", OpExec, 0)
	regAxiom(&d, "ExecWord", "execWord", OpExecWord, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "Branch", "branch", OpBranch, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "BranchQ", "branch?", OpBranchQ, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "BrSub", "brsub", OpBrSub, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "RetSub", "retsub", OpRetSub, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "TryGetIter", "(tryGetIter)", OpTryGetIter, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "ForeachLoopTop", "(foreachTop)", OpForeachLoopTop, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "I", "(i)", OpI, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "J", "(j)", OpJ, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "K", "(k)", OpK, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "PushIterator", "(pushIter)", OpPushIterator, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "DropIterator", "(dropIter)", OpDropIterator, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "Locals", "l", OpLocals, WordFlagCompileOnly)
	regAxiom(&d, "PushSwitch", "(pushSwitch)", OpPushSwitch, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "DropSwitch", "(dropSwitch)", OpDropSwitch, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "SwitchA", "(a)", OpSwitchA, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "SwitchB", "(b)", OpSwitchB, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "SwitchC", "(c)", OpSwitchC, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "AssertSwitchNest", "(assertSwitchNest)", OpAssertSwitchNest, WordFlagHidden|WordFlagCompileOnly)

	regAxiom(&d, "CTHere", "here", OpCTHere, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly)
	regAxiom(&d, "CTComma", ",", OpCTComma, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly)
	regAxiom(&d, "CTStore", "ct!", OpCTStore, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly)
	regAxiom(&d, "CTPatchForward", ">resolve", OpCTPatchForward, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly)
	regAxiom(&d, "CTPatchForwardNext", "breakResolve", OpCTPatchForwardNext, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly)
	regAxiom(&d, "SetWordFlags", "setWordFlags", OpSetWordFlags, WordFlagSystem)

	regAxiom(&d, "Transition", "_transition", OpTransition, WordFlagHidden|WordFlagCompileOnly)

	regAxiom(&d, "Fetch", "@", OpFetch, WordFlagRequiresStackContext)
	regAxiom(&d, "Store", "!", OpStore, WordFlagRequiresStackContext)

	regAxiom(&d, "DoVar", "(var)", OpDoVar, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "BootstrapDoVar", "_bootstrapDoVar", OpBootstrapDoVar, WordFlagHidden|WordFlagCompileOnly)

	regAxiom(&d, "CtImmDs", "(dtIdx)", OpCtImmDs, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly)

	regAxiom(&d, "S", "$", OpS, 0)
	regAxiom(&d, "SS", "$$", OpSS, 0)
	regAxiom(&d, "RtoS", "r>s", OpRtoS, 0)
	regAxiom(&d, "DsS", "ds$", OpDsS, WordFlagHidden)    // bootstrap debug; not official language
	regAxiom(&d, "StarS", "*$", OpStarS, WordFlagHidden) // debug: all stacks (fault-report layout)

	regAxiom(&d, "HandleCompound", "handleCompound", OpHandleCompound, WordFlagHidden)

	regAxiom(&d, "TypeMemberDispatch", "typeMemberDispatch", OpTypeMemberDispatch, WordFlagHidden|WordFlagRequiresStackContext)
	regAxiom(&d, "NativeInvoke", "nativeInvoke", OpNativeInvoke, WordFlagHidden|WordFlagRequiresStackContext)

	regAxiom(&d, "AtIndex", "@()", OpAtIndex, WordFlagRequiresStackContext)
	regAxiom(&d, "BangIndex", "!()", OpBangIndex, WordFlagRequiresStackContext)
	regAxiom(&d, "AtBrace", "@{}", OpAtBrace, WordFlagRequiresStackContext)
	regAxiom(&d, "BangBrace", "!{}", OpBangBrace, WordFlagRequiresStackContext)
	regAxiom(&d, "QueryBrace", "?{}", OpQueryBrace, WordFlagRequiresStackContext)
	// @[] / ![] registered before Nop (ordinals appended there).

	regAxiom(&d, "MethodInvoke", "^", OpMethodInvoke, WordFlagRequiresStackContext)

	regAxiom(&d, "GetDictionaryContext", "dc", OpGetDictionaryContext, 0)
	regAxiom(&d, "GetSession", "sess", OpGetSession, 0)
	regAxiom(&d, "GetLogger", "logger", OpGetLogger, 0)
	regAxiom(&d, "LogAt", "logAt", OpLogAt, 0)
	regAxiomSec(&d, "GetSys", "sys", OpGetSys, 0, SR_System)

	regAxiom(&d, "CompileCSSelect", "cs,", OpCompileCSSelect, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly|WordFlagImmediate|WordFlagSystem)
	regAxiom(&d, "DoCSSelect", "(cs)", OpDoCSSelect, WordFlagHidden|WordFlagCompileOnly|WordFlagSystem)
	regAxiom(&d, "CSSelectFromDS", "armCs", OpCSSelectFromDS, WordFlagHidden|WordFlagSystem)

	regAxiom(&d, "CompileSubCode", "subCode,", OpCompileSubCode, WordFlagHidden|WordFlagCompileOnly|WordFlagCompilerOnly|WordFlagImmediate)
	regAxiom(&d, "DoSubCode", "(subCode)", OpDoSubCode, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "SubCodeFromDS", "armSubCode", OpSubCodeFromDS, WordFlagHidden)
	regAxiom(&d, "PushType", "_push-type", OpPushType, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "Error", "error", OpError, 0)

	regAxiom(&d, "AllotDtSlotToWord", "allotDtSlotToWord", OpAllotDtSlotToWord, WordFlagCompilerOnly)
	regAxiom(&d, "DTRef", "_DTRef", OpDTRef, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "DTConst", "_DTConst", OpDTConst, WordFlagHidden|WordFlagCompilerOnly)

	regAxiom(&d, "PushBlock", "pushBlock", OpPushBlock, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "AcceptDropBlock", "acceptDropBlock", OpAcceptDropBlock, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "PanicDropBlock", "panicDropBlock", OpPanicDropBlock, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "PanicRequireBlock", "panicRequireBlock", OpPanicRequireBlock, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "PanicRequireNKind", "panicRequireNKind", OpPanicRequireNKind, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "PanicRequireOneOfKind", "panicRequireOneOfKind", OpPanicRequireOneOfKind, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "SetBlockMark", "setBlockMark", OpSetBlockMark, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "ContinueMark", "continueMark", OpContinueMark, WordFlagHidden|WordFlagCompileOnly)

	regAxiom(&d, "Range", "range", OpRange, 0)
	regAxiom(&d, "RangeBy", "rangeby", OpRangeBy, 0)

	regAxiom(&d, "GetCompileStateFlags", "getCompileStateFlags", OpGetCompileStateFlags, WordFlagHidden|WordFlagCompileOnly)
	regAxiom(&d, "SetCompileStateFlags", "setCompileStateFlags", OpSetCompileStateFlags, WordFlagHidden|WordFlagCompileOnly)

	regAxiom(&d, "Recurse", "recurse", OpRecurse, WordFlagCompileOnly|WordFlagImmediate)

	regAxiom(&d, "DoEmit", "doEmit", OpDoEmit, WordFlagHidden|WordFlagCompileOnly)

	regAxiom(&d, "MarkContext", "markContext", OpMarkContext, 0)
	regAxiom(&d, "Paren", "(", OpParen, 0)
	regAxiom(&d, "Brace", "{", OpBrace, 0)
	regAxiom(&d, "Bracket", "[", OpBracket, 0)
	regAxiom(&d, "GetContextMark", "getContextMark", OpGetContextMark, csManip)

	regAxiom(&d, "MathMonadic", "(mathMonadic)", OpMathMonadic, WordFlagHidden)

	regAxiom(&d, "DtDyadic", "(dtDyadic)", OpDtDyadic, WordFlagHidden)

	regAxiom(&d, "StrOp", "(strOp)", OpStrOp, WordFlagHidden)

	regAxiom(&d, "PipeOp", "(pipeOp)", OpPipeOp, WordFlagHidden)

	regAxiom(&d, "Cast", "(cast)", OpCast, WordFlagHidden)
	regAxiom(&d, "DelayMs", "delayMs", OpDelayMs, 0)
	regAxiom(&d, "DelayUs", "delayUs", OpDelayUs, 0)
	regAxiomSec(&d, "SpawnPip", "(spawnPip)", OpSpawnPip, 0, SR_Threads) // ( pp:AtomicDynamicObject xt:XT -- )

	regAxiom(&d, "CompileForget", "compileForget", OpCompileForget, WordFlagCompileOnly|WordFlagImmediate)
	regAxiom(&d, "DoForget", "doForget", OpDoForget, WordFlagHidden)

	regAxiom(&d, "AtMulti", "@[]", OpAtMulti, WordFlagRequiresStackContext)
	regAxiom(&d, "BangMulti", "![]", OpBangMulti, WordFlagRequiresStackContext)

	regAxiom(&d, "VecUnary", "(vecUnary)", OpVecUnary, WordFlagHidden)
	regAxiom(&d, "VecVec", "(vecVec)", OpVecVec, WordFlagHidden)
	regAxiom(&d, "VecScalar", "(vecScalar)", OpVecScalar, WordFlagHidden)
	regAxiom(&d, "MatUnary", "(matUnary)", OpMatUnary, WordFlagHidden)
	regAxiom(&d, "MatMat", "(matMat)", OpMatMat, WordFlagHidden)
	regAxiom(&d, "MatVec", "(matVec)", OpMatVec, WordFlagHidden)
	regAxiom(&d, "MatScalar", "(matScalar)", OpMatScalar, WordFlagHidden)
	regAxiom(&d, "MatIndex", "(matIndex)", OpMatIndex, WordFlagHidden)
	regAxiom(&d, "MatFactory", "(matFactory)", OpMatFactory, WordFlagHidden)

	regAxiom(&d, "PushCatchMark", "pushCatchMark", OpPushCatchMark, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "DropCatchMark", "dropCatchMark", OpDropCatchMark, WordFlagHidden|WordFlagCompilerOnly)
	regAxiom(&d, "Throw", "throw", OpThrow, 0)
	regAxiom(&d, "Pls", "pls", OpPls, 0)
	regAxiom(&d, "SetOp", "(setOp)", OpSetOp, WordFlagHidden)
	regAxiom(&d, "Using", "using", OpUsing, 0)
	regAxiom(&d, "UsingDefault", "using-default", OpUsingDefault, 0)
	regAxiomSec(&d, "Install", "install", OpInstall, 0, SR_ModuleInstall)
	regAxiomSec(&d, "Import", "import", OpImport, 0, SR_ModuleImport)
	regAxiom(&d, "LoadProject", "loadProject", OpLoadProject, 0)
	regAxiom(&d, "Nop", "nop", OpNop, 0)

	registerMathMonadics(&d)
	registerDtDyadics(&d)
	registerStrOps(&d)
	registerPipeOps(&d)
	registerCastOps(&d)
	registerVecOps(&d)
	registerMatOps(&d)
	registerSetOps(&d)
	return d
}
