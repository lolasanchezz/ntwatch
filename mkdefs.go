//go:build exclude

package main

/*
#include "c_files/customTypes.h"
*/
import "C"

type socketsDef C.struct_socketInfo
