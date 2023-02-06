/*
 * Corosync CPG Go bindings - Copyright (C) 2023-present David Coles
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
 */

package cpg

/*
#cgo LDFLAGS: -lcpg
#include <corosync/cpg.h>
typedef struct cpg_address cpg_address_s;
int _cpg_initialize(unsigned long int *, int, int);
int _cpg_join(unsigned long int, char *, int);
int _cpg_leave(unsigned long int, char *, int);
int _cpg_dispatch_one(unsigned long int);
int _cpg_mcast_joined(unsigned long int, char *, int);
int _cpg_local_get(unsigned long int, unsigned int *);

*/
import "C"

import "os"
import "unsafe"

const (
	CS_DISPATCH_ONE             = C.CS_DISPATCH_ONE
	CS_DISPATCH_ALL             = C.CS_DISPATCH_ALL
	CS_DISPATCH_BLOCKING        = C.CS_DISPATCH_BLOCKING
	CS_DISPATCH_ONE_NONBLOCKING = C.CS_DISPATCH_ONE_NONBLOCKING

	CS_OK                      = C.CS_OK
	CS_ERR_LIBRARY             = C.CS_ERR_LIBRARY
	CS_ERR_VERSION             = C.CS_ERR_VERSION
	CS_ERR_INIT                = C.CS_ERR_INIT
	CS_ERR_TIMEOUT             = C.CS_ERR_TIMEOUT
	CS_ERR_TRY_AGAIN           = C.CS_ERR_TRY_AGAIN
	CS_ERR_INVALID_PARAM       = C.CS_ERR_INVALID_PARAM
	CS_ERR_NO_MEMORY           = C.CS_ERR_NO_MEMORY
	CS_ERR_BAD_HANDLE          = C.CS_ERR_BAD_HANDLE
	CS_ERR_BUSY                = C.CS_ERR_BUSY
	CS_ERR_ACCESS              = C.CS_ERR_ACCESS
	CS_ERR_NOT_EXIST           = C.CS_ERR_NOT_EXIST
	CS_ERR_NAME_TOO_LONG       = C.CS_ERR_NAME_TOO_LONG
	CS_ERR_EXIST               = C.CS_ERR_EXIST
	CS_ERR_NO_SPACE            = C.CS_ERR_NO_SPACE
	CS_ERR_INTERRUPT           = C.CS_ERR_INTERRUPT
	CS_ERR_NAME_NOT_FOUND      = C.CS_ERR_NAME_NOT_FOUND
	CS_ERR_NO_RESOURCES        = C.CS_ERR_NO_RESOURCES
	CS_ERR_NOT_SUPPORTED       = C.CS_ERR_NOT_SUPPORTED
	CS_ERR_BAD_OPERATION       = C.CS_ERR_BAD_OPERATION
	CS_ERR_FAILED_OPERATION    = C.CS_ERR_FAILED_OPERATION
	CS_ERR_MESSAGE_ERROR       = C.CS_ERR_MESSAGE_ERROR
	CS_ERR_QUEUE_FULL          = C.CS_ERR_QUEUE_FULL
	CS_ERR_QUEUE_NOT_AVAILABLE = C.CS_ERR_QUEUE_NOT_AVAILABLE
	CS_ERR_BAD_FLAGS           = C.CS_ERR_BAD_FLAGS
)

type deliver_fn func(uint64, []byte, uint32, uint32, []byte)
type confchg_fn func(uint64, []byte, []Address, []Address, []Address)

type callback_t struct {
	d deliver_fn
	c confchg_fn
}

type Address struct {
	Nodeid uint32
	Pid    uint32
	reason uint32
}

var callbacks map[uint64]callback_t

func mapa2A(m []C.struct_cpg_address) []Address {
	mem := make([]Address, len(m))

	for i, _ := range m {
		mem[i].Nodeid = uint32(m[i].nodeid)
		mem[i].Pid = uint32(m[i].pid)
		mem[i].reason = uint32(m[i].reason)
	}
	return mem
}

//export deliver
func deliver(handle uint64, value *C.char, length C.int, nodeid C.uint, pid C.uint, msg *C.char, msglen C.int) {
	name := C.GoBytes(unsafe.Pointer(value), length)
	message := C.GoBytes(unsafe.Pointer(msg), msglen)
	callbacks[handle].d(handle, name, uint32(nodeid), uint32(pid), message)
}

//export confchg
func confchg(handle uint64, value *C.char, length C.int,
	member_list *C.cpg_address_s, member_list_entries C.int,
	left_list *C.cpg_address_s, left_list_entries C.int,
	joined_list *C.cpg_address_s, joined_list_entries C.int) {

	name := C.GoBytes(unsafe.Pointer(value), length)

	m := (*[1 << 30]C.cpg_address_s)(unsafe.Pointer(member_list))[0:member_list_entries]
	j := (*[1 << 30]C.cpg_address_s)(unsafe.Pointer(left_list))[0:left_list_entries]
	l := (*[1 << 30]C.cpg_address_s)(unsafe.Pointer(joined_list))[0:joined_list_entries]

	callbacks[handle].c(handle, name, mapa2A(m), mapa2A(l), mapa2A(j))
}

func Initialize(d deliver_fn, c confchg_fn) (uint64, int) {
	if callbacks == nil {
		callbacks = make(map[uint64]callback_t)
	}

	var cb callback_t
	cb.d = d
	cb.c = c

	del := 0
	if d != nil {
		del = 1
	}
	con := 0
	if c != nil {
		con = 1
	}

	var h uint64
	err := int(C._cpg_initialize((*C.ulong)(&h), C.int(del), C.int(con)))

	callbacks[h] = cb

	return h, err
}

func Finalize(handle uint64) int {
	return int(C.cpg_finalize(C.ulong(handle)))
}

func Join(handle uint64, name []byte) int {
	return int(C._cpg_join(C.ulong(handle), (*C.char)(unsafe.Pointer(&name[0])), C.int(len(name))))
}

func Leave(handle uint64, name []byte) int {
	return int(C._cpg_leave(C.ulong(handle), (*C.char)(unsafe.Pointer(&name[0])), C.int(len(name))))
}

func Dispatch(handle uint64, flag int) int {
	return int(C.cpg_dispatch(C.ulong(handle), C.cs_dispatch_flags_t(flag)))
}

func McastJoined(handle uint64, message []byte) int {
	return int(C._cpg_mcast_joined(C.ulong(handle), (*C.char)(unsafe.Pointer(&message[0])), C.int(len(message))))
}

func LocalGet(handle uint64) (Address, int) {
	var a Address
	a.Pid = uint32(os.Getpid())
	err := C._cpg_local_get(C.ulong(handle), (*C.uint)(&a.Nodeid))
	return a, int(err)
}
