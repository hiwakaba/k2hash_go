//
// k2hash_go
//
// Copyright 2018 Yahoo Japan Corporation.
//
// Go driver for k2hash that is a NoSQL Key Value Store(KVS) library.
// For k2hash, see https://github.com/yahoojapan/k2hash for the details.
//
// For the full copyright and license information, please view
// the license file that was distributed with this source code.
//
// AUTHOR:   Hirotaka Wakabayashi
// CREATE:   Fri, 14 Sep 2018
// REVISION:
//

package k2hash

import (
	// #cgo CFLAGS: -g -O2 -Wall -Wextra -Wno-unused-variable -Wno-unused-parameter -I. -I/usr/include/k2hash
	// #cgo LDFLAGS: -L/usr/lib -lk2hash
	// #include <stdlib.h>
	// #include "k2hash.h"
	"C"
)

import (
	"fmt"
	"unsafe"
)

// KeyQueue keeps a queue configurations.
type KeyQueue struct {
	// K2HASH file handle
	handle C.k2h_h
	// K2HASH queue handle
	keyqhandle C.k2h_keyq_h
	// fifo
	fifo bool
	// prefix
	prefix string
}

// String returns a text representation of the object.
func (q *KeyQueue) String() string {
	return fmt.Sprintf("[%v, %v, %v, %v]", q.handle, q.keyqhandle, q.fifo, q.prefix)
}

// NewKeyQueue returns a new k2hash queue instance.
func NewKeyQueue(h *K2hash, options ...func(*KeyQueue)) (*KeyQueue, error) {
	// 1. set defaults
	q := KeyQueue{
		handle:     h.GetHandle(),
		keyqhandle: C.K2H_INVALID_HANDLE,
		fifo:       true,
		prefix:     "",
	}
	// 2. set options
	for _, option := range options {
		option(&q)
	}
	// 3. open
	var qh C.k2h_keyq_h
	if q.prefix == "" {
		qh = C.k2h_keyq_handle(q.handle, C._Bool(q.fifo))
	} else {
		cPrefix := C.CBytes([]byte(q.prefix))
		defer C.free(unsafe.Pointer(cPrefix))
		qh = C.k2h_keyq_handle_prefix(q.handle, C._Bool(q.fifo), (*C.uchar)(cPrefix), C.size_t(len([]byte(q.prefix))))
	}
	// 4. check qeueu handle
	if qh == C.K2H_INVALID_HANDLE {
		return nil, fmt.Errorf("C.k2h_keyq_handle return false")
	}
	// 5. reset keyqhandle
	q.keyqhandle = qh
	return &q, nil
}

/* -- QueueQueue methods -- */

// Push adds a value to the queue.
func (q *KeyQueue) Push(v interface{}, options ...func(*Params)) (bool, error) {
	// 1. binary or text
	var val string
	switch v.(type) {
	default:
		return false, fmt.Errorf("unsupported key data format %T", v)
	case string:
		val = v.(string)
	}
	// 2. set params
	params := Params{
		password:           "",
		expirationDuration: 0,
	}
	for _, option := range options {
		option(&params)
	}
	cPass := C.CString(params.password)
	defer C.free(unsafe.Pointer(cPass))
	var expire *C.time_t
	// WARNING: You can't set zero expire.
	if params.expirationDuration != 0 {
		expire = (*C.time_t)(&params.expirationDuration)
	}
	cVal := C.CString(val)
	defer C.free(unsafe.Pointer(cVal))
	if ok := C.k2h_keyq_str_push_wa(q.keyqhandle, (*C.char)(cVal), cPass, expire); !ok {
		return false, fmt.Errorf("C.k2h_keyq_str_push return false")
	}
	return true, nil
}

// Pop retrieves a value from the queue.
func (q *KeyQueue) Pop(options ...func(*Params)) (string, error) {
	var cRetVal (*C.char)
	defer C.free(unsafe.Pointer(cRetVal))
	ok := C.k2h_keyq_str_pop(q.keyqhandle, &cRetVal)
	defer C.free(unsafe.Pointer(cRetVal))
	if !ok {
		return "", fmt.Errorf("C.k2h_keyq_str_pop return false")
	}
	val := C.GoString(cRetVal)
	return val, nil
}

// Free destroys a k2hash queue handle.
func (q *KeyQueue) Free() (bool, error) {
	if ok := C.k2h_keyq_free(q.keyqhandle); !ok {
		return false, fmt.Errorf("k2h_keyq_free() returns false")
	}
	q.keyqhandle = C.K2H_INVALID_HANDLE
	return true, nil
}

// Count returns the number of k2hash queue elements.
func (q *KeyQueue) Count() (int, error) {
	count := C.k2h_keyq_count(q.keyqhandle)
	return int(count), nil
}

// Local Variables:
// c-basic-offset: 4
// tab-width: 4
// indent-tabs-mode: t
// End:
// vim600: noexpandtab sw=4 ts=4 fdm=marker
// vim<600: noexpandtab sw=4 ts=4
