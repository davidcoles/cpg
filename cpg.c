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

#include "_cgo_export.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/uio.h>
#include <unistd.h>

#include <corosync/cpg.h>

cpg_handle_t handle;

void _cpg_deliver_fn(cpg_handle_t handle,
		    const struct cpg_name *group_name,
		    uint32_t nodeid,
		    uint32_t pid,
		    void *msg,
		    size_t msg_len) {
  deliver(handle, (char *) group_name->value, group_name->length,
	  nodeid, pid, msg, msg_len);
}

void _cpg_confchg_fn(cpg_handle_t handle,
		    const struct cpg_name *group_name,
		    const struct cpg_address *member_list, size_t member_list_entries,
		    const struct cpg_address *left_list, size_t left_list_entries,
		    const struct cpg_address *joined_list, size_t joined_list_entries) {
  confchg(handle, (char *) group_name->value, group_name->length,
	  (struct cpg_address *) member_list, member_list_entries,
	  (struct cpg_address *) left_list, left_list_entries,
	  (struct cpg_address *) joined_list, joined_list_entries);
}

int _cpg_initialize(unsigned long int *handle, int deliver, int confchg) {
  cpg_callbacks_t callbacks;
  callbacks.cpg_deliver_fn = deliver ? _cpg_deliver_fn : NULL;
  callbacks.cpg_confchg_fn = confchg ? _cpg_confchg_fn : NULL;
  return cpg_initialize(handle, &callbacks);
}

int _cpg_join(unsigned long int handle, char *name, int length) {
  struct cpg_name n;
  n.length = length;
  memcpy(n.value, name, length);
  return cpg_join(handle, &n);
}

int _cpg_leave(unsigned long int handle, char *name, int length) {
  struct cpg_name n;
  n.length = length;
  memcpy(n.value, name, length);
  return cpg_leave(handle, &n);
}

int _cpg_mcast_joined(unsigned long int handle, char *m, int l) {
  struct iovec iov;
  cs_error_t err;

  iov.iov_len = l;
  iov.iov_base = (void *) m;

  return cpg_mcast_joined(handle, CPG_TYPE_FIFO,  &iov, 1);
}

int _cpg_local_get(unsigned long int handle, unsigned int *nodeid) {
  return cpg_local_get(handle, nodeid);
}
