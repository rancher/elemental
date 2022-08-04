# Configuration Reference

All custom configuration applied on top of a fresh deployment should come
from a minimal `cloud-config` data. The `cloud-config` data can eventually
be included within the OS image as a file in `/system/oem` or,
alternatively, it can also be distributed from the Kubernetes management
cluster as part of the machine registration data.

Below is a reference of supported configuration. 

```yaml
#cloud-config

# Add additional users or set the password/ssh keys for root
users:
- name: "bar"
  passwd: "foo"
  groups: "users"
  homedir: "/home/foo"
  shell: "/bin/bash"
  ssh_authorized_keys:
  - faaapploo

# Assigns these keys to the first user in users or root if there
# is none
ssh_authorized_keys:
  - asdd

# Run these commands once the system has fully booted
runcmd:
- foo
 
# Hostname to assign
hostname: "bar"

# Write arbitrary files
write_files:
- encoding: b64
  content: CiMgVGhpcyBmaWxlIGNvbnRyb2xzIHRoZSBzdGF0ZSBvZiBTRUxpbnV4
  path: /foo/bar
  permissions: "0644"
  owner: "bar"
```
