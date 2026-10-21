package finalpi

import "github.com/androticus/final/finalembed"

// registerPiSupport is the Final pseudoclass for board profile constants
// (PiSupport.None / .PiZ2 / .Pi5). Values match Go PiSupportType / _piSupport.
func registerPiSupport(ty *finalembed.Type) {
	ty.RegisterClassProperty("None", "Int",
		func(_ any) any { return finalembed.Int(PiSupportNone) }, nil)
	ty.RegisterClassProperty("PiZ2", "Int",
		func(_ any) any { return finalembed.Int(PiSupportPiZ2) }, nil)
	ty.RegisterClassProperty("Pi5", "Int",
		func(_ any) any { return finalembed.Int(PiSupportPi5) }, nil)
}
