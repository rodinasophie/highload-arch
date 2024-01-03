package openssl

// #include "shim.h"
import "C"

import (
	"unsafe"

	"github.com/mattn/go-pointer"
)

//export go_ssl_crypto_ex_free
func go_ssl_crypto_ex_free(
	parent unsafe.Pointer, ptr unsafe.Pointer,
	cryptoData *C.CRYPTO_EX_DATA, idx C.int,
	argl C.long, argp unsafe.Pointer,
) {
	pointer.Unref(ptr)
}
