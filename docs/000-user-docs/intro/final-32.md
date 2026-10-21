# 32-bit Final

The 32-bit port of Final was done primarily to support Pi Zero 2W running Pi OS 32. It is available in the TUI version only. It has had some of Final-64's libraries removed, although remains substantially the same as Final-64. The main differences:

* Int, Uint, and Float are 32 bit values
* there are no 64 bit basic types exposed
* IntList, UintList, and FloatList hold 32 bit values
* Mat2i, Mat2u, and Mat2f hold 32 bit values
* the *32 bit versions of Lists and Mats are basically
  synonyms for the i/u/f versions (for compatibility)
* there are no strongly typed 64-bit lists or mats

* some Cryptography classes are not available
* FileInfo.size is an OS Int64, so reading this as an Int32 can truncate or go negative; compatibility properties sizeLow/sizeHigh:Uint32 have been provided

Everything else should work, including Threading and web, so you should be able to use it for a 32 bit IoT/Maker/Embedded project.
